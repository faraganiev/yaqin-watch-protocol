# Protocol notes

Many inexpensive children's GPS watches use a compact TCP text protocol commonly associated with SeTracker/Wonlex-compatible devices.

A typical frame looks like:

```text
[VENDOR*DEVICE_ID*LLLL*CONTENT]
```

`VENDOR` is a short device-family prefix, `DEVICE_ID` is the watch identifier, `LLLL` is the content length in uppercase hexadecimal, and `CONTENT` is a command followed by comma-separated fields.

Real devices are inconsistent: some firmware reports incorrect text lengths, some messages contain optional fields, and binary payloads need different framing rules. This package focuses on defensive parsing and normalization.

Supported parsing includes frame extraction, common heartbeats, `UD`/`UD2`/`AL` position reports, cell-tower and Wi-Fi observations, alarm flags, and a limited set of ordinary device commands.

## Privacy

Device IDs, locations, phone numbers, Wi-Fi observations, and cell-tower data can be sensitive. Do not commit production packets or credentials. Use synthetic fixtures in public tests and examples.
