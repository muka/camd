package discovery

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

var errWrongDiscoveryResponse = errors.New("Response is not related to discovery request")

const (
	maxDatagramSize = 8192
)

//Device API wrapper
type Device struct {
	LastUpdate int64
	MediaURI   string
	UUID       string
	Address    string
	Name       string
	Types      []string
	Hardware   string
	Country    string
}

// NewDiscovery init a new discovery wrapper
func NewDiscovery() *Discovery {
	return &Discovery{}
}

// Discovery process wrapper
type Discovery struct {
	stop    chan bool
	ticker  *time.Ticker
	Matches chan Device
}

// Stop blocks the discovery process
func (ws *Discovery) Stop() {
	if ws.stop != nil {
		ws.stop <- true
	}
	if ws.ticker != nil {
		ws.ticker.Stop()
	}
}

func (ws *Discovery) getAddrs() ([]string, error) {

	addrs := []string{}

	interfaces, err := net.InterfaceAddrs()
	if err != nil {
		return addrs, err
	}

	for _, iface := range interfaces {
		addr, ok := iface.(*net.IPNet)
		if ok && !addr.IP.IsLoopback() && addr.IP.To4() != nil {
			addrs = append(addrs, addr.IP.String())
		}
	}

	log.Printf("Got addrs=%s", addrs)
	return addrs, err
}

// Start send a WS-Discovery message and wait for all matching device to respond
func (ws *Discovery) Start() error {

	ws.stop = make(chan bool)
	ws.Matches = make(chan Device)
	ws.ticker = time.NewTicker(30 * time.Second)

	addrs, err := ws.getAddrs()
	if err != nil {
		return fmt.Errorf("Failed to get addrs: %s\n", err)
	}

	discover := func(addr string) {
		log.Printf("Discovering on addr=%s\n", addr)
		err := runWSDiscovery(ws.Matches, addr)
		if err != nil {
			log.Printf("discover error on addr=%s: %s", addr, err)
		}
	}

	for _, addr := range addrs {
		go func(addr string) {
			discover(addr)
			for {
				select {
				case <-ws.stop:
					log.Printf("Stop discovery on addr=%s\n", addr)
					return
				case <-ws.ticker.C:
					discover(addr)
					break
				}

			}
		}(addr)
	}

	return nil
}

func runWSDiscovery(matches chan Device, addr string) error {

	requestUUID, err := uuid.NewV4()
	if err != nil {
		return err
	}

	requestID := requestUUID.String()
	request := CreateProbeMessage(requestID)

	// Create UDP address for local and multicast address
	localAddress, err := net.ResolveUDPAddr("udp4", addr+":0")
	if err != nil {
		return err
	}

	multicastAddress, err := net.ResolveUDPAddr("udp4", "239.255.255.250:3702")
	if err != nil {
		return err
	}

	// Create UDP connection to listen for respond from matching device
	conn, err := net.ListenUDP("udp", localAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Set connection's timeout
	err = conn.SetDeadline(time.Now().Add(time.Second * 30))
	if err != nil {
		return err
	}

	// Send WS-Discovery request to multicast address
	_, err = conn.WriteToUDP([]byte(request), multicastAddress)
	if err != nil {
		return err
	}

	for {

		buffer := make([]byte, 10*1024)
		_, _, err = conn.ReadFromUDP(buffer)

		if err != nil {
			if udpErr, ok := err.(net.Error); ok && udpErr.Timeout() {
				break
			} else {
				return err
			}
		}

		device, err := parseResponse(requestID, buffer)
		if err != nil && err != errWrongDiscoveryResponse {
			return err
		}

		matches <- device
	}

	return nil
}

func parseResponse(messageID string, buffer []byte) (Device, error) {

	response := ProbeMatchEnvelope{}

	err := xml.Unmarshal(buffer, &response)
	if err != nil {
		return Device{}, err
	}

	relatesTo := strings.ReplaceAll(strings.Trim(response.Header.RelatesTo, "\n \t"), "uuid:", "")
	if relatesTo != messageID {
		log.Printf("Skip unrelated response [%s<>%s]\n", relatesTo, messageID)
		return Device{}, nil
	}

	dev := Device{}
	for _, probeMatch := range response.Body.ProbeMatches.ProbeMatch {

		addrs := strings.Split(probeMatch.XAddrs, " ")
		if len(addrs) == 0 {
			continue
		}

		dev.UUID = probeMatch.EndpointReference.Address
		dev.Address = addrs[0]
		dev.Types = []string{}

		scopes := strings.Split(probeMatch.Scopes, " ")
		for _, scope := range scopes {

			scope = strings.Replace(scope, "onvif://www.onvif.org/", "", 1)
			if len(scope) == 0 {
				continue
			}

			pts := strings.Split(scope, "/")
			switch pts[0] {
			case "name":
				{
					dev.Name = pts[1]
				}
			case "hardware":
				{
					dev.Hardware = pts[1]
				}
			case "type":
				{
					dev.Types = append(dev.Types, pts[1])
				}
			case "location":
				{
					if len(pts) > 2 {
						if pts[1] == "country" {
							dev.Country = pts[2]
						}
					}
				}
			}

		}

	}

	return dev, nil
}
