# Yaqin Watch Protocol

A standalone Go library for parsing and encoding the TCP protocol used by many SeTracker/Wonlex-compatible children's GPS watches.

This package was extracted and sanitized from the device-integration layer used by **Yaqin Kids**, a child-safety platform being developed in Uzbekistan. The production platform remains private because it contains proprietary application code, infrastructure details, and security-sensitive logic around children's location data. This repository publishes only a reusable, non-sensitive protocol component.

**Yaqin Kids:** https://yaqinkids.uz/

## Why this exists

Low-cost GPS watches are widely available, but their server protocol is inconsistently documented and firmware behavior varies between models. This library provides a compact, tested implementation for common packets without requiring the rest of the Yaqin Kids platform.

## Features

- Parse framed TCP packets such as `[SG*DEVICE*000B*LK,120,2,87]`
- Recover from incomplete or malformed stream data
- Split text and binary commands safely
- Decode `UD`, `UD2`, and `AL` position reports
- Decode `LK`, `TKQ`, and `TKQ2` heartbeats
- Normalize GPS, battery, GSM, cell-tower, Wi-Fi, and alarm fields
- Build protocol frames and encode a limited set of common commands
- Unit tests covering framing edge cases and representative reports
- No external Go dependencies

## Before publishing

The module path is preconfigured for `github.com/faraganiev/yaqin-watch-protocol`.

## Quick example

```go
packet := watch.Build("SG", "1234567890", []byte("LK,120,2,87"))
frame, _, err := watch.Parse(packet)
if err != nil { panic(err) }
msg := watch.Split(frame.Content)
hb := watch.DecodeHeartbeat(frame, msg, time.Now().UTC(), "example:5093")
fmt.Printf("device=%s battery=%d%%\n", hb.DeviceID, hb.Battery)
```

## Repository scope

This is intentionally **not** the full Yaqin Kids codebase. It does not contain the mobile applications, admin/school products, customer data, databases, production deployment configuration, authentication, billing, or internal services.

The public repository shares a small piece of engineering useful to developers integrating compatible hardware while keeping the production child-safety system private.

## Development

```bash
gofmt -w .
go vet ./...
go test ./...
```

## Privacy

GPS packets can contain precise location, device identifiers, network observations, and phone-related configuration. Never commit real customer packets or credentials. All examples and test fixtures in this repository are synthetic.

## License

MIT. See [LICENSE](LICENSE).
