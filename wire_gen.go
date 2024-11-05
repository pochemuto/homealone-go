// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"fmt"
	"github.com/pochemuto/homealone-go/alice"
	"github.com/pochemuto/homealone-go/homealone"
	"github.com/pochemuto/homealone-go/internal/db"
	"os"
)

// Injectors from wire.go:

func initializeApp(connection db.ConnectionString) (Application, error) {
	pool, err := db.NewPgxPool(connection)
	if err != nil {
		return Application{}, err
	}
	dbDB, err := db.NewDB(pool)
	if err != nil {
		return Application{}, err
	}
	bot := homealone.NewBot(dbDB)
	aliceAlice := alice.NewAlice()
	application := NewApplication(bot, aliceAlice)
	return application, nil
}

// wire.go:

type Application struct {
	Bot   homealone.Bot
	Alice alice.Alice
}

func NewApplication(
	bot homealone.Bot, alice2 alice.Alice,

) Application {
	return Application{
		Bot:   bot,
		Alice: alice2,
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