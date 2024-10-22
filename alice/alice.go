package alice

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/azzzak/alice"
	"github.com/golang/glog"
	"github.com/pochemuto/homealone-go/plex"
)

type Alice struct {
	ctx    context.Context
	server *http.Server
}

func NewAlice() Alice {
	return Alice{}
}

func (a Alice) Start(ctx context.Context) error {
	a.ctx = ctx

	port := os.Getenv("ALICE_PORT")
	log.Printf("Listening port :%s", port)
	updates := alice.ListenForWebhook("/alice")
	go http.ListenAndServe(":"+port, nil)

	offPhrases := []string{
		"Выключаю",
		"Ага. Произвожу выключение",
		"Выдергиваю из розетки",
		"Угу. Тушу",
		"Выключаю устройство",
		"Секундочку. Отключаю",
		"Всё. Девайс гаснет",
		"Вырубаю агрегат",
		"Отключаю технику",
		"Готово. Устройство останавливается",
		"Выполнил отключение",
		"Процесс завершен. Отключаю",
		"Отключение выполнено",
		"Всё. Щас перестанет работать",
		"Снимаю питание",
	}

	onPhrases := []string{
		"Включаю",
		"Ага. Произвожу включение",
		"Втыкаю в розетку",
		"угу. Запускаю",
		"Включаю устройство",
		"Секундочку. Запускаю",
		"Всё. Девайс запускается",
		"Врубаю агрегат",
		"Включаю технику",
		"Готово. Устройство работает",
		"Выполнил включение",
		"Принято. Включаю",
		"Включение выполнено",
		"Всё. Щас заработает",
		"Подаю питание",
	}

	updates.Loop(func(k alice.Kit) *alice.Response {
		req, resp := k.Init()
		glog.Infof("Received message: %s", req.OriginalUtterance())
		if req.OriginalUtterance() == "выключиться" {
			glog.Info("Shutting down plex")
			ctx, cancel := context.WithCancel(a.ctx)
			ch, err := plex.ShutdownAndWait(ctx)
			cancel()
			if err != nil {
				return resp.Text(err.Error())
			}
			<-ch
			glog.Info("Request sent")
			return resp.RandomText(offPhrases...).EndSession()
		}
		if req.OriginalUtterance() == "включиться" {
			glog.Info("Turning up plex")
			ctx, cancel := context.WithCancel(a.ctx)
			ch, err := plex.WakeupAndWait(ctx)
			cancel()
			if err != nil {
				return resp.Text(err.Error())
			}
			<-ch
			glog.Info("Request sent")
			return resp.RandomText(onPhrases...).EndSession()
		}

		return resp.Text("не поняла. Я услышала: ").Pause(1).Text(req.OriginalUtterance())
	})

	return nil
}
