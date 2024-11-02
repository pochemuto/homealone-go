package homealone

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang/glog"
	"github.com/pochemuto/homealone-go/plex"
)

var buildDate string

type Bot struct {
	ctx context.Context
	api *tgbotapi.BotAPI
}

func (bot Bot) updateMessage(original tgbotapi.Message, text string) {
	bot.api.Send(tgbotapi.NewEditMessageText(original.Chat.ID, original.MessageID,
		text))
}

type longProcess struct {
	longProcess func(context.Context) (<-chan struct{}, error)
	tick        func(seconds_elapsed int64)
	done        func()
}

func doLongProcess(ctx context.Context, cfg longProcess) error {
	done, err := cfg.longProcess(ctx)
	if err != nil {
		return err
	}
	ticker := time.NewTicker(time.Second)
	started := time.Now()
	defer ticker.Stop()
	for {
		select {
		case now := <-ticker.C:
			seconds := now.Unix() - started.Unix()
			cfg.tick(seconds)
		case <-done:
			cfg.done()
			return nil
		}
	}
}

func (bot Bot) handleShutdown(update tgbotapi.Update) error {
	status_message, err := bot.api.Send(Message(update, "Выключение..."))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(bot.ctx, 10*time.Minute)
	defer cancel()
	return doLongProcess(ctx, longProcess{
		longProcess: plex.ShutdownAndWait,
		tick: func(seconds_elapsed int64) {
			bot.updateMessage(status_message, fmt.Sprintf("Выключение (%d)...", seconds_elapsed))
		},
		done: func() {
			bot.updateMessage(status_message, "Выключен")
		},
	})
}

func (bot Bot) handleWakeup(update tgbotapi.Update) error {
	status_message, err := bot.api.Send(Message(update, "Включение..."))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(bot.ctx, 10*time.Minute)
	defer cancel()
	return doLongProcess(ctx, longProcess{
		longProcess: plex.WakeupAndWait,
		tick: func(seconds_elapsed int64) {
			bot.updateMessage(status_message, fmt.Sprintf("Включение (%d)...", seconds_elapsed))
		},
		done: func() {
			bot.updateMessage(status_message, "Включен")
		},
	})
}

func (bot Bot) replyText(update tgbotapi.Update, text string) {
	bot.api.Send(Message(update, text))
}

func (bot Bot) handleUpdate(update tgbotapi.Update) (err error) {
	switch update.Message.Command() {
	case "wakeup":
		err = bot.handleWakeup(update)
	case "shutdown":
		err = bot.handleShutdown(update)
	case "check":
		if plex.IsAlive() {
			bot.replyText(update, "Работает")
		} else {
			bot.replyText(update, "Не работает")
		}
	case "version":
		version, err := getVersion()
		if err != nil {
			bot.replyText(update, err.Error())
		} else {
			bot.replyText(update, version)
		}
	default:
		bot.replyText(update, update.Message.Text)
	}
	return
}

func getVersion() (string, error) {
	fmt.Printf("Build date: %s\n", buildDate)
	if buildDate == "" {
		fmt.Println("Build date is not set.")
		return "", fmt.Errorf("build date is not set")
	}

	buildTime, err := time.Parse("2006-01-02T15:04:05", buildDate)
	if err != nil {
		return "", err
	}

	duration := time.Since(buildTime)
	if duration < time.Minute {
		return fmt.Sprintf("%s, %d seconds ago", buildDate, int(duration.Seconds())), nil
	} else if duration < time.Hour {
		return fmt.Sprintf("%s, %d minutes ago", buildDate, int(duration.Minutes())), nil
	} else if duration < time.Hour*24 {
		return fmt.Sprintf("%s, %d hours ago", buildDate, int(duration.Hours())), nil
	} else {
		return fmt.Sprintf("%s, %d days ago", buildDate, int(duration.Hours()/24)), nil
	}
}

func (bot Bot) Start(ctx context.Context) (err error) {
	bot.ctx = ctx
	bot.api, err = tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		return fmt.Errorf("can't start bot: %v", err)
	}

	glog.Infof("Authorized on account %s", bot.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	_, err = bot.api.MakeRequest("deleteWebhook", tgbotapi.Params{})
	if err != nil {
		return fmt.Errorf("error deleting WebHook request %v", err)
	}
	updates := bot.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			glog.Infof("[%s] %s", update.Message.From.UserName, update.Message.Text)
			if !IsAuthorized(*update.SentFrom()) {
				bot.replyText(update, notAuthorizedUser(update))
				continue
			}

			go func() {
				if err := bot.handleUpdate(update); err != nil {
					bot.errorMessage(update, err)
				}
			}()
		}
	}

	return nil
}

func (bot Bot) errorMessage(incomming tgbotapi.Update, err error) {
	bot.replyText(incomming, "Произошла ошибка: "+err.Error())
}

func Message(incomming tgbotapi.Update, text string) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(incomming.Message.Chat.ID, text)
	msg.ReplyToMessageID = incomming.Message.MessageID
	return msg
}

func notAuthorizedUser(incoming tgbotapi.Update) string {
	return "Пользователь " + incoming.SentFrom().UserName + " не авторизован"
}

func IsAuthorized(user tgbotapi.User) bool {
	id, err := strconv.ParseInt(os.Getenv("TELEGRAM_AUTHORIZED_USER_ID"), 10, 64)
	if err != nil {
		log.Panic(err)
	}
	return user.ID == id
}
