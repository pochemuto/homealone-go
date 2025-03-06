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
	// Sends a shutdown command to the plex server.
	trigger := os.Getenv("PLEX_SHUTDOWN_TRIGGER")
	body := []byte(os.Getenv("PLEX_SHUTDOWN_SECRET"))
	resp, err := http.Post(trigger, "Content-Type: text/plain", bytes.NewReader(body))
	if err != nil {
		return err
	}
	// Check the response status code.
	if resp.StatusCode != 200 {
		if resp.Body != nil {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			return fmt.Errorf("failed response (%d): %s", resp.StatusCode, body)
		}
		return nil
	}
	return nil
}

func ShutdownAndWait(ctx context.Context) (<-chan struct{}, error) {
	// ShutdownAndWait sends a shutdown command to the plex server and waits until it is shut down.
	response := make(chan struct{})
	if !IsAlive() {
		close(response)
		return response, nil
	}
	// Send shutdown command
	err := shutdown()
	if err != nil {
		if isConnectionError(err) {
			close(response)
			return response, nil
		}
		return nil, err
	}
	// Waiting for shutdown.
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		defer ticker.Stop()
		defer close(response)
		for {
			select {
			case <-ticker.C:
				if !IsAlive() {
					return
				}
			case <-ctx.Done():
			}
		}
	}()
	return response, nil
}

func WakeupAndWait(ctx context.Context) (<-chan struct{}, error) {
	// WakeupAndWait sends a wakeup command to the plex server and waits until it is up.
	response := make(chan struct{})
	if IsAlive() {
		close(response)
		glog.Info("Is already alive")
		return response, nil
	}
	// Wake up command.
	err := wakeup()
	if err != nil {
		glog.Warningf("Wake up command error: %v", err)
		if isConnectionError(err) {
			close(response)
			return response, nil
		}
		return nil, err
	}
	// Waiting for wakeup
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		defer ticker.Stop()
		defer close(response)
		for {
			select {
			case <-ticker.C:
				if IsAlive() {
					glog.Info("Alive!")
					return
				}
				glog.Info("Still isn't alive")
			case <-ctx.Done():
				glog.Info("Context Done()")
				return
			}
		}
	}()
	return response, nil
}

func wakeup() error {
	// Sends a magic packet to wake up the plex server.
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
	// Checks if an error is a connection error.
	if errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.EHOSTDOWN) ||
		errors.Is(err, syscall.EHOSTUNREACH) {
		return true
	}
	return false
}

func IsAlive() bool {
	// Checks if the plex server is alive.
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
	plexProtocol := resp.Header.Get("X-Plex-Protocol")
	glog.Infof("Plex protocol: %s", plexProtocol)
	return plexProtocol != ""
}
