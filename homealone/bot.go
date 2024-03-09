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
	api *tgbotapi.BotAPI
}

func (bot Bot) updateMessage(original tgbotapi.Message, text string) {
	bot.api.Send(tgbotapi.NewEditMessageText(original.Chat.ID, original.MessageID,
		text))
}

func (bot Bot) handleShutdown(update tgbotapi.Update) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Minute)
	done, err := plex.ShutdownAndWait(ctx)
	if err != nil {
		cancel()
		return err
	}
	ticker := time.NewTicker(time.Second)
	status_message, err := bot.api.Send(Message(update, "Выключаем..."))
	started := time.Now()
	if err != nil {
		bot.errorMessage(update, err)
		cancel()
		return nil
	}
	go func(update tgbotapi.Update) {
		defer ticker.Stop()
		defer cancel()
		for {
			select {
			case now := <-ticker.C:
				seconds := now.Unix() - started.Unix()
				bot.updateMessage(status_message, fmt.Sprintf("Выключаем (%d)...", seconds))
			case turned_off := <-done:
				if turned_off {
					bot.updateMessage(status_message, "Выключен")
				} else {
					bot.updateMessage(status_message, "Не удалось выключить")
				}
				return
			}
		}
	}(update)
	return nil
}

func (bot Bot) handleWakeup(update tgbotapi.Update) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Minute)
	done, err := plex.WakeupAndWait(ctx)
	if err != nil {
		cancel()
		return err
	}
	ticker := time.NewTicker(time.Second)
	status_message, err := bot.api.Send(Message(update, "Включение..."))
	started := time.Now()
	if err != nil {
		bot.errorMessage(update, err)
		cancel()
		return nil
	}
	go func(update tgbotapi.Update) {
		defer ticker.Stop()
		defer cancel()
		for {
			select {
			case now := <-ticker.C:
				seconds := now.Unix() - started.Unix()
				bot.updateMessage(status_message, fmt.Sprintf("Включение (%d)...", seconds))
			case turned_off := <-done:
				if turned_off {
					bot.updateMessage(status_message, "Включен")
				} else {
					bot.updateMessage(status_message, "Не удалось включить")
				}
				return
			}
		}
	}(update)
	return nil
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
		bot.echo(update)
	}
	return
}

func (bot Bot) Start() (err error) {
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

func (bot Bot) echo(incomming tgbotapi.Update) {
	bot.replyText(incomming, incomming.Message.Text)
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
