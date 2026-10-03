// Package watch implements the text protocol spoken by the common Chinese kids
// GPS watches (SeTracker / Wonlex / 3g-elec family, Q50..Q90, A36E, Y95, KT*).
//
// Every packet is an envelope:
//
//	[VENDOR*DEVICE_ID*LLLL*CONTENT]
//
// VENDOR is a short prefix such as SG, 3G, CS or ZJ; DEVICE_ID is the number
// printed on the watch (10 digits, sometimes the 15-digit IMEI); LLLL is the
// content length as four upper-case hex digits; CONTENT is a command name
// followed by comma separated arguments. The server answers most commands by
// echoing the command name in the same envelope.
package watchprotocol

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

// Frame is one decoded envelope.
type Frame struct {
	Vendor  string
	ID      string
	Content []byte
}

var (
	// ErrIncomplete means more bytes are needed before a frame can be parsed.
	ErrIncomplete = errors.New("watch: incomplete frame")
	// ErrMalformed means the bytes at the start of the buffer cannot be a frame.
	ErrMalformed = errors.New("watch: malformed frame")
)

const (
	maxVendorLen  = 4
	maxIDLen      = 20
	maxContentLen = 64 * 1024
)

// Parse extracts one frame from the start of buf.
//
// On success it returns the frame and the number of bytes consumed (garbage
// before the opening bracket included). On ErrIncomplete the returned count
// is the number of leading bytes that can already be discarded. On
// ErrMalformed the count skips the offending opening bracket so the caller can
// resynchronise on the next one.
func Parse(buf []byte) (Frame, int, error) {
	start := bytes.IndexByte(buf, '[')
	if start < 0 {
		return Frame{}, len(buf), ErrIncomplete
	}
	b := buf[start:]

	p1 := bytes.IndexByte(b, '*')
	if p1 < 0 {
		if len(b) > maxVendorLen+1 {
			return Frame{}, start + 1, ErrMalformed
		}
		return Frame{}, start, ErrIncomplete
	}
	p2 := indexFrom(b, '*', p1+1)
	if p2 < 0 {
		if len(b) > p1+1+maxIDLen {
			return Frame{}, start + 1, ErrMalformed
		}
		return Frame{}, start, ErrIncomplete
	}
	p3 := indexFrom(b, '*', p2+1)
	if p3 < 0 {
		if len(b) > p2+1+4 {
			return Frame{}, start + 1, ErrMalformed
		}
		return Frame{}, start, ErrIncomplete
	}

	vendor := b[1:p1]
	id := b[p1+1 : p2]
	lenHex := b[p2+1 : p3]
	if len(vendor) == 0 || len(vendor) > maxVendorLen || len(id) == 0 || len(id) > maxIDLen || len(lenHex) != 4 {
		return Frame{}, start + 1, ErrMalformed
	}
	n, err := strconv.ParseUint(string(lenHex), 16, 32)
	if err != nil || n > maxContentLen {
		return Frame{}, start + 1, ErrMalformed
	}

	contentStart := p3 + 1
	end := contentStart + int(n)

	// Text commands never contain a closing bracket in their payload, so if one
	// shows up before the declared end the firmware miscounted the length
	// (this happens with multibyte characters). Trust the bracket in that case.
	if !isBinaryContent(b[contentStart:min(len(b), end)]) {
		if alt := indexFrom(b, ']', contentStart); alt >= 0 && alt < end {
			return Frame{Vendor: string(vendor), ID: string(id), Content: b[contentStart:alt]}, start + alt + 1, nil
		}
	}

	if len(b) < end+1 {
		return Frame{}, start, ErrIncomplete
	}
	if b[end] != ']' {
		// Declared length too short. Fall back to the next closing bracket.
		alt := indexFrom(b, ']', end)
		if alt < 0 {
			if len(b) > contentStart+maxContentLen {
				return Frame{}, start + 1, ErrMalformed
			}
			return Frame{}, start, ErrIncomplete
		}
		end = alt
	}
	return Frame{Vendor: string(vendor), ID: string(id), Content: b[contentStart:end]}, start + end + 1, nil
}

// Build wraps content into an envelope addressed to the device.
func Build(vendor, id string, content []byte) []byte {
	out := make([]byte, 0, len(vendor)+len(id)+len(content)+12)
	out = append(out, '[')
	out = append(out, vendor...)
	out = append(out, '*')
	out = append(out, id...)
	out = append(out, '*')
	out = append(out, fmt.Sprintf("%04X", len(content))...)
	out = append(out, '*')
	out = append(out, content...)
	out = append(out, ']')
	return out
}

// isBinaryContent reports whether the payload may legitimately contain a
// closing bracket. Voice notes (TK) and pictures (img) carry raw bytes.
func isBinaryContent(content []byte) bool {
	comma := bytes.IndexByte(content, ',')
	name := content
	if comma >= 0 {
		name = content[:comma]
	}
	switch string(name) {
	case "TK", "img", "IMG":
		return true
	}
	return false
}

func indexFrom(b []byte, c byte, from int) int {
	if from >= len(b) {
		return -1
	}
	i := bytes.IndexByte(b[from:], c)
	if i < 0 {
		return -1
	}
	return from + i
}
