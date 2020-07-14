package onvif

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"

	"github.com/muka/camd/onvif/discovery"
	"github.com/use-go/onvif"
	"github.com/use-go/onvif/device"
)

// Discover devices on the network
func Discover() error {

	wsDiscovery := discovery.NewDiscovery()

	err := wsDiscovery.Start()
	if err != nil {
		return fmt.Errorf("listen failed: %s", err)
	}

	devices := map[string]*discovery.Device{}

	for {
		select {
		case dev := <-wsDiscovery.Matches:
			if _, ok := devices[dev.UUID]; ok {
				continue
			}
			log.Printf("dev %s\n", dev)
			uri, err := url.Parse(dev.Address)
			if err != nil {
				log.Printf("cannot parse %s: %s", uri, err)
			}

			client(uri.Host)
			break
		}
	}

}

func client(ipaddr string) {
	dev, err := onvif.NewDevice(ipaddr)
	if err != nil {
		panic(err)
	}
	dev.Authenticate("admin", "zsyy12345")

	log.Printf("output %+v", dev.GetServices())

	res, err := dev.CallMethod(device.GetUsers{})
	bs, _ := ioutil.ReadAll(res.Body)
	log.Printf("output %+v %s", res.StatusCode, bs)
}
