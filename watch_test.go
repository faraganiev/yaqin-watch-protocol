package watchprotocol

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

func TestParseSingleLK(t *testing.T) {
	buf := []byte("[SG*8800000015*0002*LK]")
	f, n, err := Parse(buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(buf) || f.Vendor != "SG" || f.ID != "8800000015" || string(f.Content) != "LK" {
		t.Fatalf("got %+v consumed %d", f, n)
	}
	reply := Reply(Split(f.Content))
	if string(reply) != "LK" {
		t.Fatalf("reply = %q", reply)
	}
	if got := Build(f.Vendor, f.ID, reply); string(got) != "[SG*8800000015*0002*LK]" {
		t.Fatalf("build = %q", got)
	}
}

func TestParseTwoFramesAndPartial(t *testing.T) {
	a := Build("3G", "4700186508", []byte("LK,10,2,85"))
	b := Build("3G", "4700186508", []byte("TKQ"))
	buf := append(append(append([]byte("junk"), a...), b...), []byte("[3G*4700186508*00")...)

	f1, n1, err := Parse(buf)
	if err != nil || string(f1.Content) != "LK,10,2,85" || n1 != 4+len(a) {
		t.Fatalf("first: %v %+v %d", err, f1, n1)
	}
	f2, n2, err := Parse(buf[n1:])
	if err != nil || string(f2.Content) != "TKQ" || n2 != len(b) {
		t.Fatalf("second: %v %+v %d", err, f2, n2)
	}
	_, n3, err := Parse(buf[n1+n2:])
	if !errors.Is(err, ErrIncomplete) || n3 != 0 {
		t.Fatalf("third: %v %d", err, n3)
	}
}

func TestParseWrongLengthFallsBackToBracket(t *testing.T) {
	// Declared length 0x0005 but the content is "LK,0,0,90" (9 bytes).
	buf := []byte("[SG*1234567890*0005*LK,0,0,90][SG*1234567890*0002*LK]")
	f, n, err := Parse(buf)
	if err != nil || string(f.Content) != "LK,0,0,90" {
		t.Fatalf("%v %+v", err, f)
	}
	f2, _, err := Parse(buf[n:])
	if err != nil || string(f2.Content) != "LK" {
		t.Fatalf("%v %+v", err, f2)
	}
}

func TestParseMalformedResyncs(t *testing.T) {
	buf := []byte("[garbage without stars that is long enough[SG*1*0002*LK]")
	_, n, err := Parse(buf)
	if !errors.Is(err, ErrMalformed) {
		t.Fatalf("want malformed, got %v", err)
	}
	f, _, err := Parse(buf[n:])
	if err != nil || string(f.Content) != "LK" || f.ID != "1" {
		t.Fatalf("%v %+v", err, f)
	}
}

func TestDecodeUDWithCells(t *testing.T) {
	raw := "UD,220414,134652,A,22.571707,N,113.8613968,E,0.1,0.0,100,7,60,90,1000,50,0000,4,1,460,0,9360,4082,131,9360,4092,148,9360,4091,143,9360,4153,141"
	buf := Build("3G", "4700186508", []byte(raw))
	f, _, err := Parse(buf)
	if err != nil {
		t.Fatal(err)
	}
	m := Split(f.Content)
	if !IsPosition(m) || Reply(m) != nil {
		t.Fatalf("UD must be a position without reply")
	}
	now := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
	p, err := DecodePosition(f, m, now)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2014, 4, 22, 13, 46, 52, 0, time.UTC)
	if !p.DeviceTime.Equal(want) {
		t.Fatalf("device time %v", p.DeviceTime)
	}
	if !p.Valid || p.Latitude != 22.571707 || p.Longitude != 113.8613968 {
		t.Fatalf("coords %+v", p)
	}
	if p.Satellites != 7 || p.GSMSignal != 60 || p.Battery != 90 || p.Steps != 1000 || p.Rolls != 50 {
		t.Fatalf("fields %+v", p)
	}
	if p.MCC != 460 || p.MNC != 0 || len(p.Cells) != 4 || p.Cells[3].CID != 4153 || p.Cells[3].RSSI != 141 {
		t.Fatalf("cells %+v", p.Cells)
	}
	if p.Alarm != "" || p.Cached {
		t.Fatalf("alarm %q cached %v", p.Alarm, p.Cached)
	}
}

func TestDecodeALSOSSouthWest(t *testing.T) {
	raw := "AL,130926,081500,V,41.311081,S,69.240562,W,0.0,0.0,450,0,80,45,12,0,00010000,1,1,434,7,101,2001,120,2,School,aa:bb:cc:dd:ee:ff,-50,Home,11:22:33:44:55:66,-70,120"
	f, _, err := Parse(Build("SG", "8800000015", []byte(raw)))
	if err != nil {
		t.Fatal(err)
	}
	m := Split(f.Content)
	if string(Reply(m)) != "AL" {
		t.Fatalf("AL must be acknowledged")
	}
	p, err := DecodePosition(f, m, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if p.Valid || p.Latitude != -41.311081 || p.Longitude != -69.240562 {
		t.Fatalf("coords %+v", p)
	}
	if p.Alarm != AlarmSOS {
		t.Fatalf("alarm %q", p.Alarm)
	}
	if p.MCC != 434 || p.MNC != 7 || len(p.Cells) != 1 || p.Cells[0].LAC != 101 {
		t.Fatalf("cells %+v", p.Cells)
	}
	if len(p.WiFi) != 2 || p.WiFi[1].SSID != "Home" || p.WiFi[1].RSSI != -70 || p.Accuracy != 120 {
		t.Fatalf("wifi %+v accuracy %v", p.WiFi, p.Accuracy)
	}
}

func TestDecodeShortUD2(t *testing.T) {
	f, _, _ := Parse(Build("CS", "123", []byte("UD2,130926,081500,A,41.3,N,69.2,E")))
	p, err := DecodePosition(f, Split(f.Content), time.Now())
	if err != nil || !p.Cached || p.Latitude != 41.3 {
		t.Fatalf("%v %+v", err, p)
	}
}

func TestHeartbeat(t *testing.T) {
	f, _, _ := Parse(Build("SG", "1", []byte("LK,1234,5,77")))
	h := DecodeHeartbeat(f, Split(f.Content), time.Now(), "1.2.3.4:5")
	if h.Steps != 1234 || h.Rolls != 5 || h.Battery != 77 || h.Kind != "LK" {
		t.Fatalf("%+v", h)
	}
}

func TestBinaryContentKeepsBrackets(t *testing.T) {
	payload := []byte("TK,#!AMR\n\x00]\x01[\x02")
	buf := Build("SG", "1", payload)
	f, n, err := Parse(buf)
	if err != nil || n != len(buf) || !bytes.Equal(f.Content, payload) {
		t.Fatalf("%v %d %q", err, n, f.Content)
	}
	if string(Reply(Split(f.Content))) != "TK,1" {
		t.Fatal("TK must be acknowledged with TK,1")
	}
}

func TestEncodeCommands(t *testing.T) {
	cases := []struct {
		cmd  Command
		want string
	}{
		{Command{Type: CmdPositionSingle}, "CR"},
		{Command{Type: CmdPositionPeriodic, Args: []string{"180"}}, "UPLOAD,180"},
		{Command{Type: CmdSOSNumber, Args: []string{"1", "+10000000000"}}, "SOS1,+10000000000"},
		{Command{Type: CmdTimezone, Args: []string{"9", "5"}}, "LZ,9,5"},
		{Command{Type: CmdFind}, "FIND"},
		{Command{Type: CmdCall, Args: []string{"+10000000000"}}, "CALL,+10000000000"},
		{Command{Type: CmdMessage, Args: []string{"Hi"}}, "MESSAGE,00480069"},
		{Command{Type: CmdPhonebook, Args: []string{"+10000000000", "Mom"}}, "PHB,+10000000000,004D006F006D"},
		{Command{Type: CmdRemoveAlarm, Args: []string{"off"}}, "REMOVE,0"},
	}
	for _, c := range cases {
		got, err := EncodeCommand(c.cmd)
		if err != nil || string(got) != c.want {
			t.Errorf("%s: got %q err %v, want %q", c.cmd.Type, got, err, c.want)
		}
	}
	if _, err := EncodeCommand(Command{Type: "dance"}); !errors.Is(err, ErrUnsupportedCommand) {
		t.Errorf("unknown command must fail, got %v", err)
	}
	if _, err := EncodeCommand(Command{Type: CmdSOSNumber, Args: []string{"4", "1"}}); err == nil {
		t.Errorf("sos index 4 must fail")
	}
}
