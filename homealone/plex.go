package homealone

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"

	"github.com/mdlayher/wol"
)

func Shutdown() (bool, error) {
	trigger := os.Getenv("PLEX_SHUTDOWN_TRIGGER")
	body := []byte(os.Getenv("ZINA_SECRET"))
	resp, err := http.Post(trigger, "Content-Type: text/plain", bytes.NewReader(body))
	if errors.Is(err, syscall.ECONNREFUSED) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
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

	ip := os.Getenv("WAKEUP_BROADCAST_ADDRESS")
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
