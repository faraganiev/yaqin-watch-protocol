package watchprotocol

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
	"time"
)

// Message is the content of a frame split into its command name and arguments.
type Message struct {
	Type string
	Args []string
	Raw  []byte
}

// Split separates the command name from its arguments. Binary commands keep
// their payload as a single untouched argument.
func Split(content []byte) Message {
	comma := bytes.IndexByte(content, ',')
	if comma < 0 {
		return Message{Type: string(content), Raw: content}
	}
	typ := string(content[:comma])
	rest := content[comma+1:]
	if isBinaryContent(content) {
		return Message{Type: typ, Args: []string{string(rest)}, Raw: content}
	}
	return Message{Type: typ, Args: strings.Split(string(rest), ","), Raw: content}
}

// Reply returns the content the server must send back for a message, or nil
// when the protocol expects no answer.
func Reply(m Message) []byte {
	switch m.Type {
	case "LK", "AL", "TKQ", "TKQ2":
		return []byte(m.Type)
	case "TK":
		return []byte("TK,1")
	}
	return nil
}

// IsPosition reports whether the message carries a location report.
func IsPosition(m Message) bool {
	switch m.Type {
	case "UD", "UD2", "UD_LTE", "AL":
		return true
	}
	return false
}

// IsHeartbeat reports whether the message is a keep-alive.
func IsHeartbeat(m Message) bool {
	switch m.Type {
	case "LK", "TKQ", "TKQ2":
		return true
	}
	return false
}

var errShortPosition = errors.New("watch: position report too short")

// DecodePosition parses UD / UD2 / AL reports.
//
// Field order (all optional past the coordinates):
//
//	date(DDMMYY), time(HHMMSS), A|V, lat, N|S, lon, E|W, speed, course,
//	altitude, satellites, gsm signal, battery, steps, rolls, status(hex),
//	cell count, timing advance, mcc, mnc, [lac, cid, rssi]*n,
//	wifi count, [ssid, mac, rssi]*n, accuracy
func DecodePosition(f Frame, m Message, now time.Time) (Position, error) {
	a := m.Args
	if len(a) < 7 {
		return Position{}, errShortPosition
	}
	p := Position{
		DeviceID:   f.ID,
		Vendor:     f.Vendor,
		Protocol:   "watch",
		ServerTime: now,
		Cached:     m.Type == "UD2",
		Raw:        string(m.Raw),
	}

	if t, err := time.Parse("020106150405", a[0]+a[1]); err == nil {
		p.DeviceTime = t.UTC()
	} else {
		p.DeviceTime = now
	}
	p.Valid = a[2] == "A"

	lat, err1 := strconv.ParseFloat(a[3], 64)
	lon, err2 := strconv.ParseFloat(a[5], 64)
	if err1 != nil || err2 != nil {
		return Position{}, errShortPosition
	}
	if a[4] == "S" {
		lat = -lat
	}
	if a[6] == "W" {
		lon = -lon
	}
	p.Latitude, p.Longitude = lat, lon

	p.SpeedKmh = floatAt(a, 7)
	p.Course = floatAt(a, 8)
	p.Altitude = floatAt(a, 9)
	p.Satellites = intAt(a, 10)
	p.GSMSignal = intAt(a, 11)
	p.Battery = intAt(a, 12)
	p.Steps = intAt(a, 13)
	p.Rolls = intAt(a, 14)
	if len(a) > 15 {
		if v, err := strconv.ParseUint(strings.TrimSpace(a[15]), 16, 32); err == nil {
			p.Status = uint32(v)
		}
	}
	p.Alarm = decodeAlarm(p.Status)
	if p.Alarm == "" && m.Type == "AL" {
		p.Alarm = "general"
	}

	i := 16
	if len(a) > i {
		cellCount := intAt(a, i)
		i++ // cell count
		i++ // timing advance
		p.MCC = intAt(a, i)
		i++
		p.MNC = intAt(a, i)
		i++
		for c := 0; c < cellCount && i+2 < len(a); c++ {
			p.Cells = append(p.Cells, Cell{LAC: intAt(a, i), CID: intAt(a, i+1), RSSI: intAt(a, i+2)})
			i += 3
		}
		if len(a) > i {
			wifiCount := intAt(a, i)
			i++
			for w := 0; w < wifiCount && i+2 < len(a); w++ {
				p.WiFi = append(p.WiFi, WiFi{SSID: a[i], MAC: a[i+1], RSSI: intAt(a, i+2)})
				i += 3
			}
		}
		if len(a) > i {
			p.Accuracy = floatAt(a, i)
		}
	}
	return p, nil
}

// DecodeHeartbeat parses LK / TKQ keep-alives. LK optionally carries
// steps, rolls and battery.
func DecodeHeartbeat(f Frame, m Message, now time.Time, remote string) Heartbeat {
	h := Heartbeat{DeviceID: f.ID, Vendor: f.Vendor, Kind: m.Type, ServerTime: now, Remote: remote}
	if m.Type == "LK" && len(m.Args) >= 3 {
		h.Steps = intAt(m.Args, 0)
		h.Rolls = intAt(m.Args, 1)
		h.Battery = intAt(m.Args, 2)
	}
	return h
}

// decodeAlarm maps the status bit field to an alarm name. Bit layout follows
// the vendor documentation and matches what SeTracker shows.
func decodeAlarm(status uint32) string {
	switch {
	case status&(1<<16) != 0:
		return AlarmSOS
	case status&(1<<17) != 0, status&1 != 0:
		return AlarmLowBattery
	case status&(1<<18) != 0, status&(1<<1) != 0:
		return AlarmGeofenceExit
	case status&(1<<19) != 0, status&(1<<2) != 0:
		return AlarmGeofenceEnter
	case status&(1<<20) != 0:
		return AlarmRemoved
	case status&(1<<21) != 0:
		return AlarmFall
	case status&(1<<3) != 0:
		return AlarmOverspeed
	}
	return ""
}

func floatAt(a []string, i int) float64 {
	if i >= len(a) {
		return 0
	}
	v, _ := strconv.ParseFloat(strings.TrimSpace(a[i]), 64)
	return v
}

func intAt(a []string, i int) int {
	if i >= len(a) {
		return 0
	}
	v, _ := strconv.Atoi(strings.TrimSpace(a[i]))
	return v
}
