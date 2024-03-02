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

func Start() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	_, err = bot.MakeRequest("deleteWebhook", tgbotapi.Params{})
	if err != nil {
		log.Panic(err)
	}
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			if !IsAuthorized(*update.SentFrom()) {
				bot.Send(NotAuthorizedUser(update))
				continue
			}

			switch update.Message.Command() {
			case "wakeup":
				err = plex.Wakeup()
				if err != nil {
					bot.Send(ErrorMessage(update, err))
					continue
				}

				bot.Send(Message(update, "Пробуждаем"))
			case "check":
				if plex.IsAlive() {
					bot.Send(Message(update, "Работает"))
				} else {
					bot.Send(Message(update, "Не работает"))
				}
			case "shutandwait":
				ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
				done, err := plex.ShutdownAndWait(ctx)
				if err != nil {
					bot.Send(ErrorMessage(update, err))
					cancel()
					continue
				}
				ticker := time.NewTicker(1 * time.Second)
				status_message, err := bot.Send(Message(update, "Выключаем..."))
				if err != nil {
					bot.Send(ErrorMessage(update, err))
					cancel()
					continue
				}
				go func(update tgbotapi.Update) {
					n := 0
					defer ticker.Stop()
					defer cancel()
					for {
						select {
						case <-ticker.C:
							n++
							bot.Send(tgbotapi.NewEditMessageText(status_message.Chat.ID, status_message.MessageID,
								fmt.Sprintf("Выключаем (%d)", n)))
						case turned_off := <-done:
							if turned_off {
								bot.Send(tgbotapi.NewEditMessageText(status_message.Chat.ID, status_message.MessageID,
									"Выключен"))
							} else {
								bot.Send(tgbotapi.NewEditMessageText(status_message.Chat.ID, status_message.MessageID,
									"Не удалось выключить"))
							}
							return
						}
					}
				}(update)
			case "shutdown":
				err := plex.Shutdown()
				if err != nil {
					bot.Send(ErrorMessage(update, err))
					continue
				}
				bot.Send(Message(update, "Выключаем"))
			default:
				bot.Send(Echo(update))
			}
		}
	}
}
