package hook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/muka/camd/device"
	"github.com/spf13/viper"
)

//CameraSource a json payload
type CameraSource struct {
	Live bool   `json:"live"`
	URI  string `json:"uri"`
	Type string `json:"type"`
}

// Request Perform an HTTP request based on the event
func Request(ev device.OnChangeEvent) error {

	url := viper.GetViper().GetString("request_url")
	if url == "" {
		url = "http://localhost:8778/api/config/source/%uuid"
	}
	contentType := "application/json"

	url = strings.ReplaceAll(url, "%uuid", ev.Device.UUID)

	var body io.Reader
	method := "DELETE"

	if ev.Event == device.DeviceAdded {
		method = "PUT"

		uri := ev.Device.MediaURI
		if uri == "" {
			uri = ev.Device.Path
		}

		source := CameraSource{
			Type: "video",
			URI:  uri,
			Live: true,
		}

		b, err := json.Marshal(source)
		if err != nil {
			return err
		}

		body = bytes.NewReader(b)
	}

	client := &http.Client{}

	// Create request
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	req.Header.Add("content-type", contentType)

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read Response Body
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Request failed: %d %s", resp.StatusCode, resp.Status)
	}

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	log.Printf("Response %s", responseBody)

	return nil
}
