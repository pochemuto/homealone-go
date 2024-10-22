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
		"выключаю",
		"ага-, произвожу выключение",
		"выдергиваю из розетки",
		"угу-, тушу",
		"Выключаю устройство",
		"Секундочку-, отключаю",
		"Всё-, девайс гаснет",
		"Вырубаю агрегат",
		"Отключаю технику",
		"Готово-, устройство останавливается",
		"Выполнил отключение",
		"Процесс завершен, отключаю",
		"Отключение выполнено",
		"Всё-, щас перестанет работать",
		"снимаю питание",
	}

	onPhrases := []string{
		"включаю",
		"ага, произвожу включение",
		"втыкаю в розетку",
		"угу-, запускаю",
		"Включаю устройство",
		"Секундочку-, запускаю",
		"Всё, девайс запускается",
		"Врубаю агрегат",
		"Включаю технику",
		"Готово-, устройство работает",
		"Выполнил включение",
		"Процесс завершен-, включаю",
		"Включение выполнено",
		"Всё-, щас заработает",
		"подаю питание",
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
