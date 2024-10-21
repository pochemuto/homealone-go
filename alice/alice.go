package alice

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
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
	// Создаем новый HTTP сервер
	a.server = &http.Server{
		Addr:    ":" + port,
		Handler: a.routes(),
	}

	go func() {
		log.Printf("HTTP server started on :%s\n", port)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start server: %v", err)
		}
	}()

	// Ожидаем завершения через контекст
	<-ctx.Done()

	// Завершаем работу сервера при получении сигнала завершения
	log.Println("Shutting down server...")

	if err := a.server.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	log.Println("Server exited properly")

	return nil
}

func (a *Alice) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.handleHome())
	return mux
}

func (a *Alice) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from Alice!")
	}
}
