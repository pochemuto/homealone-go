//go:build wireinject
// +build wireinject

package main

import (
	"fmt"
	"os"

	"github.com/google/wire"
	"github.com/pochemuto/homealone-go/alice"
	"github.com/pochemuto/homealone-go/homealone"
	"github.com/pochemuto/homealone-go/internal/db"
)

type Application struct {
	Bot   homealone.Bot
	Alice alice.Alice
}

func NewApplication(
	bot homealone.Bot,
	alice alice.Alice,
) Application {
	return Application{
		Bot:   bot,
		Alice: alice,
	}
}

func InitializeApplication() (Application, error) {

	connectionString := db.ConnectionString(os.Getenv("PGCONNECTION"))
	if connectionString == "" {
		return Application{}, fmt.Errorf("provide a connection string PGCONNECTION")
	}

	app, err := initializeApp(connectionString)

	if err != nil {
		return Application{}, fmt.Errorf("app initialization error: %w", err)
	}

	return app, nil
}

func initializeApp(
	connection db.ConnectionString,
) (Application, error) {
	wire.Build(
		NewApplication,
		alice.NewAlice,
		homealone.NewBot,
		db.NewPgxPool,
		db.NewDB,
	)
	return Application{}, nil
}
