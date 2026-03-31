package conf

type LAN struct {
	Endpoint       string `json:"endpoint"`       // Host endpoint, e.g. http://192.168.1.10:6806
	BasePath       string `json:"basePath"`       // Host-side repository root
	AuthToken      string `json:"authToken"`      // Shared LAN sync token
	DeviceID       string `json:"deviceID"`       // Stable device identifier
	DeviceName     string `json:"deviceName"`     // Human-readable device name
	Serve          bool   `json:"serve"`          // Whether this device acts as a LAN sync host
	Timeout        int    `json:"timeout"`        // Timeout in seconds
	ConcurrentReqs int    `json:"concurrentReqs"` // Concurrent requests
}

const ProviderLAN = 5 // ProviderLAN syncs via another kernel instance on the local network

func NewLAN() *LAN {
	return &LAN{
		Timeout:        15,
		ConcurrentReqs: 4,
	}
}
