package homealone

import (
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	id, err := strconv.ParseInt(os.Getenv("AUTHORIZED_USER_ID"), 10, 64)
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

			switch update.Message.Command() {
			case "wakeup":
				if !IsAuthorized(*update.SentFrom()) {
					bot.Send(NotAuthorizedUser(update))
					continue
				}
				err = Wakeup()
				if err != nil {
					bot.Send(ErrorMessage(update, err))
					continue
				}

				bot.Send(Message(update, "Пробуждаем"))
			case "shutdown":
				if !IsAuthorized(*update.SentFrom()) {
					bot.Send(NotAuthorizedUser(update))
					continue
				}
				err = Shutdown()
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
