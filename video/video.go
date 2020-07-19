package video

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/muka/camd/device"
)

var v4lPath = "/sys/class/video4linux/"

var cachedDevices = map[string]device.Device{}

// WatchDevices watch local video devices for changes
func WatchDevices(emitter chan device.OnChangeEvent) error {

	ticker := time.NewTicker(500 * time.Millisecond)

	go func() {
		for {
			select {
			case <-ticker.C:

				list, err := enumerateDevices()
				if err != nil {
					log.Printf("Failed to enumerate device: %s\n", err)
					continue
				}

				removed := map[string]bool{}
				for _, device := range cachedDevices {
					removed[device.UUID] = true
				}

				for _, dev := range list {

					if _, ok := cachedDevices[dev.UUID]; ok {
						removed[dev.UUID] = false
						continue
					}

					cachedDevices[dev.UUID] = dev
					log.Printf("Added device name=%s path=%s\n", dev.Name, dev.Path)
					emitter <- device.OnChanged(dev, device.DeviceAdded)
				}

				for uuid, isRemoved := range removed {
					if isRemoved {
						dev := cachedDevices[uuid]
						log.Printf("Removed device name=%s path=%s\n", dev.Name, dev.Path)
						emitter <- device.OnChanged(dev, device.DeviceRemoved)
						delete(cachedDevices, uuid)
					}
				}

			}
		}
	}()

	return nil
}

func enumerateDevices() ([]device.Device, error) {

	var devices []device.Device
	err := filepath.Walk(v4lPath, func(path string, info os.FileInfo, err error) error {
		devName := path[len(v4lPath):]
		if len(devName) > 0 {
			device := device.Device{
				Path: "/dev/" + devName,
			}

			b, err := ioutil.ReadFile(path + "/name")
			if err == nil {
				device.Name = strings.Trim(string(b), "\n\t ")
			}

			device.UUID = fmt.Sprintf("%x", md5.Sum([]byte(device.Path)))

			// log.Printf("Found device name=%s source=%s", device.Name, device.Path)
			devices = append(devices, device)
		}
		return nil
	})

	if err != nil {
		return devices, err
	}

	return devices, nil
}
