package discovery

import "encoding/xml"

//CreateProbeMessage return a string with the XML payload
func CreateProbeMessage(uuid string) string {
	return `<?xml version="1.0" encoding="UTF-8"?><Envelope xmlns="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing"><Header><a:Action mustUnderstand="1">http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</a:Action><a:MessageID>uuid:` + uuid + `</a:MessageID><a:ReplyTo><a:Address>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address></a:ReplyTo><a:To mustUnderstand="1">urn:schemas-xmlsoap-org:ws:2005:04:discovery</a:To></Header><Body><Probe xmlns="http://schemas.xmlsoap.org/ws/2005/04/discovery"><d:Types xmlns:d="http://schemas.xmlsoap.org/ws/2005/04/discovery" xmlns:dp0="http://www.onvif.org/ver10/network/wsdl">dp0:NetworkVideoTransmitter</d:Types></Probe></Body></Envelope>`
}

//ProbeMatchEnvelope a struct to unmarshal a probe response
type ProbeMatchEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Text    string   `xml:",chardata"`
	S       string   `xml:"s,attr"`
	Enc     string   `xml:"enc,attr"`
	Xsi     string   `xml:"xsi,attr"`
	Xsd     string   `xml:"xsd,attr"`
	Wsa     string   `xml:"wsa,attr"`
	Wsa5    string   `xml:"wsa5,attr"`
	D       string   `xml:"d,attr"`
	Tdn     string   `xml:"tdn,attr"`
	Tt      string   `xml:"tt,attr"`
	Tds     string   `xml:"tds,attr"`
	Header  struct {
		Text      string `xml:",chardata"`
		MessageID string `xml:"MessageID"`
		RelatesTo string `xml:"RelatesTo"`
		To        struct {
			Text           string `xml:",chardata"`
			MustUnderstand string `xml:"mustUnderstand,attr"`
		} `xml:"To"`
		Action struct {
			Text           string `xml:",chardata"`
			MustUnderstand string `xml:"mustUnderstand,attr"`
		} `xml:"Action"`
	} `xml:"Header"`
	Body struct {
		Text         string `xml:",chardata"`
		ProbeMatches struct {
			Text       string `xml:",chardata"`
			ProbeMatch []struct {
				Text              string `xml:",chardata"`
				EndpointReference struct {
					Text    string `xml:",chardata"`
					Address string `xml:"Address"`
				} `xml:"EndpointReference"`
				Types           string `xml:"Types"`
				Scopes          string `xml:"Scopes"`
				XAddrs          string `xml:"XAddrs"`
				MetadataVersion string `xml:"MetadataVersion"`
			} `xml:"ProbeMatch"`
		} `xml:"ProbeMatches"`
	} `xml:"Body"`
}
