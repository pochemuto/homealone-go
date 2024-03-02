package plex

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/mdlayher/wol"
	"golang.org/x/net/context"
)

func Shutdown() error {
	trigger := os.Getenv("PLEX_SHUTDOWN_TRIGGER")
	body := []byte(os.Getenv("PLEX_SHUTDOWN_SECRET"))
	resp, err := http.Post(trigger, "Content-Type: text/plain", bytes.NewReader(body))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return false, fmt.Errorf("failed response (%d): %s", resp.StatusCode, body)
	}
	return true, nil
}

func ShutdownAndWait(ctx context.Context) (<-chan bool, error) {
	response := make(chan bool, 1)
	shutdown_result, err := Shutdown()
	if err != nil {
		return nil, err
	}
	if !shutdown_result 
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		select {
		case <-ticker.C:
			if IsAlive() {
				response <- true
			}
		case <-ctx.Done():
			response <- false
		}
	}()
	return response, nil
}

func Wakeup() error {
	client, err := wol.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()

	ip := os.Getenv("PLEX_WAKEUP_BROADCAST_ADDRESS")
	mac, err := net.ParseMAC(os.Getenv("PLEX_MAC"))
	if err != nil {
		return err
	}
	log.Println("Sending wakup request to", ip, "addr", mac)
	err = client.Wake(ip, mac)
	if err != nil {
		return err
	}
	log.Println("Request sent to", ip)
	return nil
}

func isConnectionError(err error) bool {
	if errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.EHOSTDOWN) ||
		errors.Is(err, syscall.EHOSTUNREACH) {
		return true
	}
	return false
}

func IsAlive() bool {
	url := os.Getenv("PLEX_URL")
	log.Printf("Requesting %s", url)
	resp, err := http.Head(url)
	if err != nil {
		log.Printf("Error: %v", err.Error())
		return false
	}
	plex_protocol := resp.Header.Get("X-Plex-Protocol")
	log.Printf("Plex protocol: %s", plex_protocol)
	return plex_protocol != ""
}
