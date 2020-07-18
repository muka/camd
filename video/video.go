package video

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/muka/camd/device"
)

var v4lPath = "/sys/class/video4linux/"

// ListDevices list available devices
func ListDevices() ([]device.Device, error) {

	var devices []device.Device

	err := filepath.Walk(v4lPath, func(path string, info os.FileInfo, err error) error {
		devName := path[len(v4lPath):]
		if len(devName) > 0 {
			device := device.Device{
				Path: "/dev/" + devName,
			}

			b, err := ioutil.ReadFile(path + "/name")
			if err == nil {
				device.Name = string(b)
			}

			log.Printf("Found device name=%s source=%s", device.Name, device.Path)
			devices = append(devices, device)
		}
		return nil
	})

	if err != nil {
		return devices, err
	}

	return devices, nil
}
