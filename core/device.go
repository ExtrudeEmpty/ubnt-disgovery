package core

import "time"

// EventKind identifies the type of discovery event.
type EventKind string

const (
	EventNewDevice     EventKind = "new_device"
	EventDeviceUpdated EventKind = "updated"
	EventLog           EventKind = "log"
	EventScanStart     EventKind = "scan_start"
	EventScanDone      EventKind = "scan_done"
)

// Device is the unified representation of any discovered network device.
type Device struct {
	Protocol  string    // UBNT, MikroTik, Grandstream, SSDP, WSD, Fingerprint
	IP        string
	IPs       []string // All discovered IPs for this device
	MAC       string
	Hostname  string
	Model     string
	Platform  string
	Version   string
	Board     string
	Identity  string
	Kind      string   // TCP fingerprint classification
	Hints     []string // TCP fingerprint hints
	Iface     string   // The interface where this device was discovered
	FirstSeen time.Time
	LastSeen  time.Time
}

// DeviceEvent is emitted by the Engine for every discovery event.
type DeviceEvent struct {
	Kind    EventKind
	Device  *Device // nil for log-only events
	Message string  // human-readable log message
}
