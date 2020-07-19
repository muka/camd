package device

//DeviceChanged base enum type
type DeviceChanged uint8

const (
	//DeviceAdded notify of a device added
	DeviceAdded DeviceChanged = 1
	//DeviceRemoved notify of a device removed
	DeviceRemoved DeviceChanged = 2
)

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

//OnChangeEvent notify of an event for a device
type OnChangeEvent struct {
	Device Device
	Event  DeviceChanged
}

// OnChanged return a OnChangeEvent instance
func OnChanged(dev Device, ev DeviceChanged) OnChangeEvent {
	return OnChangeEvent{
		Device: dev,
		Event:  ev,
	}
}
