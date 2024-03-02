package homealone

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pochemuto/homealone-go/plex"
)

type Bot struct {
	api *tgbotapi.BotAPI
}

func (bot Bot) handleShutdownAndWait(update tgbotapi.Update) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	done, err := plex.ShutdownAndWait(ctx)
	if err != nil {
		cancel()
		return err
	}
	ticker := time.NewTicker(time.Second)
	status_message, err := bot.api.Send(Message(update, "Выключаем..."))
	started := time.Now()
	if err != nil {
		bot.api.Send(ErrorMessage(update, err))
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
				bot.api.Send(tgbotapi.NewEditMessageText(status_message.Chat.ID, status_message.MessageID,
					fmt.Sprintf("Выключаем (%d)...", seconds)))
			case turned_off := <-done:
				if turned_off {
					bot.api.Send(tgbotapi.NewEditMessageText(status_message.Chat.ID, status_message.MessageID,
						"Выключен"))
				} else {
					bot.api.Send(tgbotapi.NewEditMessageText(status_message.Chat.ID, status_message.MessageID,
						"Не удалось выключить"))
				}
				return
			}
		}
	}(update)
	return nil
}

func (bot Bot) handleUpdate(update tgbotapi.Update) (err error) {
	switch update.Message.Command() {
	case "wakeup":
		err = plex.Wakeup()
		if err != nil {
			return
		}
		bot.api.Send(Message(update, "Пробуждаем"))
	case "check":
		if plex.IsAlive() {
			bot.api.Send(Message(update, "Работает"))
		} else {
			bot.api.Send(Message(update, "Не работает"))
		}
	case "shutandwait":
		bot.handleShutdownAndWait(update)
	case "shutdown":
		err = plex.Shutdown()
		if err != nil {
			return
		}
		bot.api.Send(Message(update, "Выключаем"))
	default:
		bot.api.Send(Echo(update))
	}
	return
}

func (bot Bot) Start() (err error) {
	bot.api, err = tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		return fmt.Errorf("can't start bot: %v", err)
	}

	bot.api.Debug = true
	log.Printf("Authorized on account %s", bot.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	_, err = bot.api.MakeRequest("deleteWebhook", tgbotapi.Params{})
	if err != nil {
		return fmt.Errorf("error deleting WebHook request %v", err)
	}
	updates := bot.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			if !IsAuthorized(*update.SentFrom()) {
				bot.api.Send(NotAuthorizedUser(update))
				continue
			}

			if err := bot.handleUpdate(update); err != nil {
				bot.api.Send(ErrorMessage(update, err))
			}
		}
	}

	return nil
}

func ErrorMessage(incomming tgbotapi.Update, err error) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(incomming.Message.Chat.ID, "Произошла ошибка: "+err.Error())
	msg.ReplyToMessageID = incomming.Message.MessageID
	return msg
}

func Echo(incomming tgbotapi.Update) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(incomming.Message.Chat.ID, incomming.Message.Text)
	msg.ReplyToMessageID = incomming.Message.MessageID
	return msg
}

func Message(incomming tgbotapi.Update, text string) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(incomming.Message.Chat.ID, text)
	msg.ReplyToMessageID = incomming.Message.MessageID
	return msg
}

func NotAuthorizedUser(incoming tgbotapi.Update) tgbotapi.MessageConfig {
	return Message(incoming, "Пользователь "+incoming.SentFrom().UserName+" не авторизован")
}

func IsAuthorized(user tgbotapi.User) bool {
	id, err := strconv.ParseInt(os.Getenv("TELEGRAM_AUTHORIZED_USER_ID"), 10, 64)
	if err != nil {
		log.Panic(err)
	}
	return user.ID == id
}
