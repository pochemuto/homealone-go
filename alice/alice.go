package alice

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/azzzak/alice"
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

	updates.Loop(func(k alice.Kit) *alice.Response {
		req, resp := k.Init()
		if req.IsNewSession() {
			return resp.Text("привет")
		}
		return resp.Text(req.OriginalUtterance())
	})

	return nil
}
