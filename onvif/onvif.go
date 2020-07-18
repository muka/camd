package onvif

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"time"

	"github.com/muka/camd/device"
	"github.com/muka/camd/onvif/discovery"
	goonvif "github.com/use-go/onvif"
	"github.com/use-go/onvif/media"
)

// Discover devices on the network
func Discover(emitter chan device.Device) error {

	wsDiscovery := discovery.NewDiscovery()

	err := wsDiscovery.Start()
	if err != nil {
		return fmt.Errorf("listen failed: %s", err)
	}

	devices := map[string]*device.Device{}

	for {
		select {
		case dev := <-wsDiscovery.Matches:

			mediaURI, err := getMediaURI(dev.Address)
			if err != nil {
				log.Printf("getMediaURI error: %s\n", err)
				delete(devices, dev.UUID)
				continue
			}

			dev.MediaURI = mediaURI
			dev.LastUpdate = time.Now().UnixNano()
			devices[dev.UUID] = &dev

			log.Printf("Found ONVIF device name=%s source=%s\n", dev.Name, dev.MediaURI)

			emitter <- dev

			break
		}
	}

}

func getMediaURI(uri string) (string, error) {

	uriItems, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	dev, err := goonvif.NewDevice(uriItems.Host)
	if err != nil {
		return "", err
	}

	res, err := dev.CallMethod(media.GetStreamUri{})
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	getStremUriResponse := GetStremUriResponse{}
	if err = xml.Unmarshal(b, &getStremUriResponse); err != nil {
		return "", err
	}

	return getStremUriResponse.GetURI(), nil
}
