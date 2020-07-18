package device

//Device API wrapper
type Device struct {
	LastUpdate int64
	Path       string
	MediaURI   string
	UUID       string
	Address    string
	Name       string
	Types      []string
	Hardware   string
	Country    string
}
