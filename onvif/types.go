package onvif

import "encoding/xml"

// GetStremUriResponse soap message response
type GetStremUriResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		GetStreamUriResponse struct {
			MediaUri struct {
				URI                 string `xml:"Uri"`
				InvalidAfterConnect string `xml:"InvalidAfterConnect"`
				InvalidAfterReboot  string `xml:"InvalidAfterReboot"`
				Timeout             string `xml:"Timeout"`
			} `xml:"MediaUri"`
		} `xml:"GetStreamUriResponse"`
	} `xml:"Body"`
}

// GetURI return the stream URI
func (r *GetStremUriResponse) GetURI() string {
	return r.Body.GetStreamUriResponse.MediaUri.URI
}

// GetTimeout return the timeout
func (r *GetStremUriResponse) GetTimeout() string {
	return r.Body.GetStreamUriResponse.MediaUri.Timeout
}
