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

	"github.com/golang/glog"
	"github.com/mdlayher/wol"
	"golang.org/x/net/context"
)

func shutdown() error {
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
		return fmt.Errorf("failed response (%d): %s", resp.StatusCode, body)
	}
	return nil
}

func ShutdownAndWait(ctx context.Context) (<-chan bool, error) {
	response := make(chan bool, 1)
	if !IsAlive() {
		response <- true
		return response, nil
	}
	err := shutdown()
	if err != nil {
		if isConnectionError(err) {
			response <- true
			return response, nil
		}
		return nil, err
	}
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		defer ticker.Stop()
		defer close(response)
		for {
			select {
			case <-ticker.C:
				if !IsAlive() {
					response <- true
					return
				}
			case <-ctx.Done():
				response <- false
			}
		}
	}()
	return response, nil
}

func WakeupAndWait(ctx context.Context) (<-chan bool, error) {
	response := make(chan bool, 1)
	if IsAlive() {
		response <- true
		return response, nil
	}
	err := wakeup()
	if err != nil {
		if isConnectionError(err) {
			response <- true
			return response, nil
		}
		return nil, err
	}
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		defer ticker.Stop()
		defer close(response)
		for {
			select {
			case <-ticker.C:
				if IsAlive() {
					response <- true
					return
				}
			case <-ctx.Done():
				response <- false
			}
		}
	}()
	return response, nil
}

func wakeup() error {
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
	glog.Infof("Sending wakeup request to", ip, "addr", mac)
	err = client.Wake(ip, mac)
	if err != nil {
		return err
	}
	glog.Infof("Request sent to", ip)
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
	glog.Infof("Requesting %s", url)
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Head(url)
	if err != nil {
		log.Printf("Error: %v", err.Error())
		return false
	}
	plex_protocol := resp.Header.Get("X-Plex-Protocol")
	glog.Infof("Plex protocol: %s", plex_protocol)
	return plex_protocol != ""
}
