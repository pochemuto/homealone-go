//go:build wireinject
// +build wireinject

package main

import (
	"fmt"

	"github.com/google/wire"
	"github.com/pochemuto/homealone-go/alice"
	"github.com/pochemuto/homealone-go/homealone"
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
	app, err := initializeApp()
	if err != nil {
		return Application{}, fmt.Errorf("app initialization error: %w", err)
	}

	return app, nil
}

func initializeApp() (Application, error) {
	wire.Build(
		NewApplication,
		alice.NewAlice,
		homealone.NewBot,
	)
	return Application{}, nil
}
