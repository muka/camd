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
func Discover(emitter chan device.OnChangeEvent) error {

	wsDiscovery := discovery.NewDiscovery()

	err := wsDiscovery.Start()
	if err != nil {
		return fmt.Errorf("listen failed: %s", err)
	}

	devices := map[string]*device.Device{}

	for {
		select {
		case ev := <-wsDiscovery.Matches:

			if ev.Device.MediaURI == "" && ev.Event == device.DeviceAdded {
				mediaURI, err := getMediaURI(ev.Device.Address)
				if err != nil {
					log.Printf("getMediaURI error: %s\n", err)
					delete(devices, ev.Device.UUID)
					continue
				}

				ev.Device.MediaURI = mediaURI
				ev.Device.LastUpdate = time.Now().UnixNano()
				devices[ev.Device.UUID] = &ev.Device
			}

			op := "Added"
			if ev.Event == device.DeviceRemoved {
				op = "Removed"
			}
			log.Printf("%s ONVIF device name=%s source=%s\n", op, ev.Device.Name, ev.Device.MediaURI)

			emitter <- ev

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
