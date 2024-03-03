package main

import (
	"flag"
	"log"
	"os"

	"github.com/golang/glog"
	"github.com/joho/godotenv"
	"github.com/pochemuto/homealone-go/homealone"
)

func init() {
	glog.CopyStandardLogTo("INFO")
}

func main() {
	flag.Parse()
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		glog.Fatalf("Error loading .env file, %v", err)
	}

	log.Println("Started")
	var bot homealone.Bot
	err = bot.Start()
	if err != nil {
		glog.Fatalf("Error: %v", err)
	}
}
