package watchprotocol

import "time"

// Position is a normalized location report decoded from a watch packet.
type Position struct {
	DeviceID   string    `json:"device_id"`
	Vendor     string    `json:"vendor"`
	Protocol   string    `json:"protocol"`
	DeviceTime time.Time `json:"device_time"`
	ServerTime time.Time `json:"server_time"`
	Valid      bool      `json:"valid"`
	Latitude   float64   `json:"lat"`
	Longitude  float64   `json:"lon"`
	SpeedKmh   float64   `json:"speed_kmh"`
	Course     float64   `json:"course"`
	Altitude   float64   `json:"altitude"`
	Satellites int       `json:"satellites"`
	GSMSignal  int       `json:"gsm_signal"`
	Battery    int       `json:"battery"`
	Steps      int       `json:"steps"`
	Rolls      int       `json:"rolls"`
	Status     uint32    `json:"status"`
	Alarm      string    `json:"alarm,omitempty"`
	Cached     bool      `json:"cached"`
	MCC        int       `json:"mcc,omitempty"`
	MNC        int       `json:"mnc,omitempty"`
	Cells      []Cell    `json:"cells,omitempty"`
	WiFi       []WiFi    `json:"wifi,omitempty"`
	Accuracy   float64   `json:"accuracy,omitempty"`
	Raw        string    `json:"raw,omitempty"`
}

type Cell struct {
	LAC  int `json:"lac"`
	CID  int `json:"cid"`
	RSSI int `json:"rssi"`
}
type WiFi struct {
	SSID string `json:"ssid"`
	MAC  string `json:"mac"`
	RSSI int    `json:"rssi"`
}
type Heartbeat struct {
	DeviceID   string    `json:"device_id"`
	Vendor     string    `json:"vendor"`
	Kind       string    `json:"kind"`
	ServerTime time.Time `json:"server_time"`
	Battery    int       `json:"battery,omitempty"`
	Steps      int       `json:"steps,omitempty"`
	Rolls      int       `json:"rolls,omitempty"`
	Remote     string    `json:"remote,omitempty"`
}

const (
	AlarmSOS           = "sos"
	AlarmLowBattery    = "low_battery"
	AlarmGeofenceEnter = "geofence_enter"
	AlarmGeofenceExit  = "geofence_exit"
	AlarmOverspeed     = "overspeed"
	AlarmRemoved       = "removed"
	AlarmFall          = "fall"
)

type Command struct {
	Type string
	Args []string
}

const (
	CmdPositionSingle   = "position_single"
	CmdPositionPeriodic = "position_periodic"
	CmdSOSNumber        = "sos_number"
	CmdPhonebook        = "phonebook"
	CmdTimezone         = "timezone"
	CmdFind             = "find"
	CmdReboot           = "reboot"
	CmdPowerOff         = "power_off"
	CmdCall             = "call"
	CmdMessage          = "message"
	CmdRemoveAlarm      = "remove_alarm"
	CmdLowBatteryAlarm  = "low_battery_alarm"
	CmdSOSSMS           = "sos_sms"
	CmdSilenceTime      = "silence_time"
)
