package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/pochemuto/homealone-go/homealone"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	log.Println("Started")
	homealone.Start()
}
