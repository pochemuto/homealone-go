package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/pochemuto/homealone-go/homealone"
)

func main() {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Error loading .env file, %v", err)
	}

	log.Println("Started")
	var bot homealone.Bot
	err = bot.Start()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
