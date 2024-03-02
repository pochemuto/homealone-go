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

	"github.com/mdlayher/wol"
)

func Shutdown() (bool, error) {
	trigger := os.Getenv("PLEX_SHUTDOWN_TRIGGER")
	body := []byte(os.Getenv("PLEX_SHUTDOWN_SECRET"))
	resp, err := http.Post(trigger, "Content-Type: text/plain", bytes.NewReader(body))
	if isConnectionError(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}
		return false, fmt.Errorf("failed response (%d): %s", resp.StatusCode, body)
	}
	return true, nil
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
