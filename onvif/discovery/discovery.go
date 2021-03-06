package discovery

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/muka/camd/device"
)

var errWrongDiscoveryResponse = errors.New("Response is not related to discovery request")

const (
	maxDatagramSize = 8192
)

var cachedDevices = map[string]map[string]device.Device{}

// NewDiscovery init a new discovery wrapper
func NewDiscovery() *Discovery {
	return &Discovery{}
}

// Discovery process wrapper
type Discovery struct {
	stop    chan bool
	ticker  *time.Ticker
	Matches chan device.OnChangeEvent
	mut     sync.Mutex
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

	// log.Printf("Got addrs=%s", addrs)
	return addrs, err
}

// Start send a WS-Discovery message and wait for all matching device to respond
func (ws *Discovery) Start() error {

	ws.stop = make(chan bool)
	ws.Matches = make(chan device.OnChangeEvent)
	ws.ticker = time.NewTicker(5 * time.Second)

	addrs, err := ws.getAddrs()
	if err != nil {
		return fmt.Errorf("Failed to get addrs: %s", err)
	}

	discover := func() {
		for _, addr := range addrs {
			go func(addr string) {
				// log.Printf("Discovering on addr=%s\n", addr)
				ws.mut.Lock()
				if _, ok := cachedDevices[addr]; !ok {
					cachedDevices[addr] = map[string]device.Device{}
				}
				ws.mut.Unlock()
				err := ws.runDiscovery(addr)
				if err != nil {
					log.Printf("discover error on addr=%s: %s", addr, err)
				}
			}(addr)
		}
	}

	discover()
	go func() {
		for {
			select {
			case <-ws.stop:
				log.Println("Stopped discovery")
				return
			case <-ws.ticker.C:
				discover()
				break
			}

		}
	}()

	return nil
}

func (ws *Discovery) runDiscovery(addr string) error {

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
	err = conn.SetDeadline(time.Now().Add(time.Second * 2))
	if err != nil {
		return err
	}

	// Send WS-Discovery request to multicast address
	_, err = conn.WriteToUDP([]byte(request), multicastAddress)
	if err != nil {
		return err
	}

	localCache := map[string]device.Device{}

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

		dev, err := parseResponse(requestID, buffer)
		if err != nil && err != errWrongDiscoveryResponse {
			return err
		}

		// matches <- dev
		localCache[dev.UUID] = dev
		// cachedDevices[addr][dev.UUID] = dev
	}

	removed := map[string]bool{}

	ws.mut.Lock()

	for uuid := range cachedDevices[addr] {
		removed[uuid] = true
	}

	for uuid, dev := range localCache {
		if _, ok := cachedDevices[addr][uuid]; ok {
			removed[uuid] = false
			continue
		}
		// added
		cachedDevices[addr][uuid] = dev
		ws.Matches <- device.OnChanged(cachedDevices[addr][uuid], device.DeviceAdded)
	}

	for uuid, isRemoved := range removed {
		if isRemoved {
			// removed
			ws.Matches <- device.OnChanged(cachedDevices[addr][uuid], device.DeviceRemoved)
			delete(cachedDevices[addr], uuid)

		}
	}

	ws.mut.Unlock()

	// log.Printf("Completed discovery on %s\n", addr)
	return nil
}

func parseResponse(messageID string, buffer []byte) (device.Device, error) {

	response := ProbeMatchEnvelope{}

	err := xml.Unmarshal(buffer, &response)
	if err != nil {
		return device.Device{}, err
	}

	relatesTo := strings.ReplaceAll(strings.Trim(response.Header.RelatesTo, "\n \t"), "uuid:", "")
	if relatesTo != messageID {
		log.Printf("Skip unrelated response [%s<>%s]\n", relatesTo, messageID)
		return device.Device{}, nil
	}

	dev := device.Device{}
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
