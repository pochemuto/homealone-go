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

type Bot struct {
	ctx context.Context
	api *tgbotapi.BotAPI
}

func (bot Bot) updateMessage(original tgbotapi.Message, text string) {
	bot.api.Send(tgbotapi.NewEditMessageText(original.Chat.ID, original.MessageID,
		text))
}

func (bot Bot) handleShutdown(update tgbotapi.Update) error {
	status_message, err := bot.api.Send(Message(update, "Выключение..."))
	if err != nil {
		return err
	}
	ctx, _ := context.WithTimeout(bot.ctx, 10*time.Minute)
	return bot.doLongProcess(ctx, longProcess{
		longProcessFunc: plex.ShutdownAndWait,
		tickFunc: func(seconds_elapsed int64) {
			bot.updateMessage(status_message, fmt.Sprintf("Выключение (%d)...", seconds_elapsed))
		},
		doneFunc: func() {
			bot.updateMessage(status_message, "Выключен")
		},
	})
}

type longProcess struct {
	longProcessFunc func(context.Context) (<-chan struct{}, error)
	tickFunc        func(seconds_elapsed int64)
	doneFunc        func()
}

func (bot Bot) doLongProcess(ctx context.Context, cfg longProcess) error {
	done, err := cfg.longProcessFunc(ctx)
	if err != nil {
		return err
	}
	ticker := time.NewTicker(time.Second)
	started := time.Now()
	go func() {
		defer ticker.Stop()
		for {
			select {
			case now := <-ticker.C:
				seconds := now.Unix() - started.Unix()
				cfg.tickFunc(seconds)
			case <-done:
				cfg.doneFunc()
				return
			}
		}
	}()
	return nil
}

func (bot Bot) handleWakeup(update tgbotapi.Update) error {
	// status_message, err := bot.api.Send(Message(update, "Включение..."))
	// if err != nil {
	// 	return err
	// }
	// ctx, cancel := context.WithTimeout(bot.ctx, 10*time.Minute)
	// defer cancel()
	// return bot.doLongProcess(ctx, longProcess{
	// 	longProcessFunc: plex.WakeupAndWait,
	// 	tickFunc: func(seconds_elapsed int64) {
	// 		bot.updateMessage(status_message, fmt.Sprintf("Включение (%d)...", seconds_elapsed))
	// 	},
	// 	doneFunc: func() {
	// 		bot.updateMessage(status_message, "Включен")
	// 	},
	// })
	ctx, cancel := context.WithTimeout(bot.ctx, 10*time.Minute)
	defer cancel()
	t := time.NewTicker(1 * time.Second)
	result := make(chan struct{})
	defer close(result)
	go func() {
		plex.ShutdownAndWait(ctx)
		close(result)
	}()

	for {
		select {
		case c := <-t.C:
			bot.updateMessage(status_message, fmt.Sprintf("Включение (%d)...", seconds_elapsed))
		case <-t:
			bot.updateMessage(status_message, "Включен")
		}
	}
}

func (bot Bot) replyText(update tgbotapi.Update, text string) {
	bot.api.Send(Message(update, text))
}

func (bot Bot) handleUpdate(update tgbotapi.Update) (err error) {
	switch update.Message.Command() {
	case "wakeup":
		bot.handleWakeup(update)
	case "check":
		if plex.IsAlive() {
			bot.replyText(update, "Работает")
		} else {
			bot.replyText(update, "Не работает")
		}
	case "shutdown":
		bot.handleShutdown(update)
	default:
		bot.replyText(update, update.Message.Text)
	}
	return
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

			if err := bot.handleUpdate(update); err != nil {
				bot.errorMessage(update, err)
			}
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
