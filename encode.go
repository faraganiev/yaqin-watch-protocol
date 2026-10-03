package watchprotocol

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"unicode/utf16"
)

// ErrUnsupportedCommand is returned for command types the watch cannot take.
var ErrUnsupportedCommand = errors.New("watch: unsupported command")

// EncodeCommand turns a platform command into the content the watch expects.
// The caller wraps it with Build using the vendor prefix of the live session.
func EncodeCommand(c Command) ([]byte, error) {
	arg := func(i int) string {
		if i < len(c.Args) {
			return strings.TrimSpace(c.Args[i])
		}
		return ""
	}
	switch c.Type {
	case CmdPositionSingle:
		return []byte("CR"), nil
	case CmdPositionPeriodic:
		if arg(0) == "" {
			return nil, fmt.Errorf("%w: %s needs seconds", ErrUnsupportedCommand, c.Type)
		}
		return []byte("UPLOAD," + arg(0)), nil
	case CmdSOSNumber:
		idx, phone := arg(0), arg(1)
		if idx < "1" || idx > "3" || len(idx) != 1 || phone == "" {
			return nil, fmt.Errorf("%w: %s needs index 1-3 and phone", ErrUnsupportedCommand, c.Type)
		}
		return []byte("SOS" + idx + "," + phone), nil
	case CmdPhonebook:
		if len(c.Args) == 0 || len(c.Args)%2 != 0 {
			return nil, fmt.Errorf("%w: %s needs phone,name pairs", ErrUnsupportedCommand, c.Type)
		}
		parts := make([]string, 0, len(c.Args))
		for i := 0; i < len(c.Args); i += 2 {
			parts = append(parts, arg(i), utf16Hex(arg(i+1)))
		}
		return []byte("PHB," + strings.Join(parts, ",")), nil
	case CmdTimezone:
		if arg(0) == "" || arg(1) == "" {
			return nil, fmt.Errorf("%w: %s needs language and tz", ErrUnsupportedCommand, c.Type)
		}
		return []byte("LZ," + arg(0) + "," + arg(1)), nil
	case CmdFind:
		return []byte("FIND"), nil
	case CmdReboot:
		return []byte("RESET"), nil
	case CmdPowerOff:
		return []byte("POWEROFF"), nil
	case CmdCall:
		if arg(0) == "" {
			return nil, fmt.Errorf("%w: %s needs phone", ErrUnsupportedCommand, c.Type)
		}
		return []byte("CALL," + arg(0)), nil
	case CmdMessage:
		return []byte("MESSAGE," + utf16Hex(arg(0))), nil
	case CmdRemoveAlarm:
		return []byte("REMOVE," + onOff(arg(0))), nil
	case CmdLowBatteryAlarm:
		return []byte("LOWBAT," + onOff(arg(0))), nil
	case CmdSOSSMS:
		return []byte("SOSSMS," + onOff(arg(0))), nil
	case CmdSilenceTime:
		return []byte("SILENCETIME," + strings.Join(c.Args, ",")), nil
	}
	return nil, fmt.Errorf("%w: %q", ErrUnsupportedCommand, c.Type)
}

// utf16Hex encodes text the way the watches expect free text: upper-case hex
// of the UTF-16 big-endian bytes.
func utf16Hex(s string) string {
	units := utf16.Encode([]rune(s))
	b := make([]byte, 0, len(units)*2)
	for _, u := range units {
		b = append(b, byte(u>>8), byte(u))
	}
	return strings.ToUpper(hex.EncodeToString(b))
}

func onOff(v string) string {
	switch strings.ToLower(v) {
	case "", "1", "on", "true", "yes":
		return "1"
	}
	return "0"
}
