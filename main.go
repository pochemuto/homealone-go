package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/pochemuto/homealone-go/homealone"
)

func main() {
	err := godotenv.Load()
	if err != nil && os.IsNotExist(err) {
		log.Fatal("Error loading .env file, %v", err)
	}

	log.Println("Started")
	homealone.Start()
}
