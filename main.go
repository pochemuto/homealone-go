package main

import (
	"context"
	"flag"
	"log"
	"os"
	"sync"

	"github.com/golang/glog"
	"github.com/joho/godotenv"
	"github.com/pochemuto/homealone-go/alice"
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

	log.Println("Starting...")
	var wg sync.WaitGroup
	wg.Add(2)

	ctx := context.Background()
	go func() {
		defer wg.Done()
		var bot homealone.Bot
		err = bot.Start(ctx)

		if err != nil {
			glog.Fatalf("Error in bot: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		alice := alice.NewAlice()
		alice.Start(ctx)

		if err != nil {
			glog.Fatalf("Error in alice: %v", err)
		}
	}()
	wg.Wait()
	log.Println("Finished")
}
