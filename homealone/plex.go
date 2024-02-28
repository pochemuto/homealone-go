package homealone

import (
	"log"
	"net"
	"os"

	"github.com/mdlayher/wol"
)

func Wakeup() error {
	client, err := wol.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()

	ip := os.Getenv("PLEX_HOST")
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
