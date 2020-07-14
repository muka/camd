package discovery

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseResponse(t *testing.T) {

	xml, err := ioutil.ReadFile("./probe_match_example.xml")
	if err != nil {
		t.Fatalf("Cannot read xml: %s", err)
	}

	messageID := "uuid:0a6dc791-2be6-4991-9af1-454778a1917a"
	device, err := parseResponse(messageID, xml)

	assert.Equal(t, "http://prn-example/PRN42/b42-1668-a", device.Address)

}
