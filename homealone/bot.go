package homealone

import (
	"log"
	"time"
)

func Start() {
	timer := time.NewTicker(1 * time.Second)
	var i = 0
	for range timer.C {
		i++
		log.Println("Tick", i)
	}
}
