package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

// ────────────────────────────────────────────────────────────
// UBNT (Ubiquiti)
// ────────────────────────────────────────────────────────────

const (
	ubntTLV_MAC      = 0x01
	ubntTLV_IP       = 0x02
	ubntTLV_Version  = 0x03
	ubntTLV_Hostname = 0x0b
	ubntTLV_Model    = 0x0c
	ubntTLV_Platform = 0x0d
	ubntTLV_WEBUI    = 0x0f
)

func ParseUBNTPacket(data []byte) *Device {
	if len(data) < 4 {
		return nil
	}
	r := bytes.NewReader(data[4:])
	dev := &Device{Protocol: "UBNT"}
	for r.Len() > 3 {
		t, _ := r.ReadByte()
		var l uint16
		if binary.Read(r, binary.BigEndian, &l) != nil { break }
		if int(l) > r.Len() { break }
		val := make([]byte, l)
		r.Read(val)
		switch t {
		case ubntTLV_MAC:
			if len(val) == 6 {
				dev.MAC = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", val[0], val[1], val[2], val[3], val[4], val[5])
			}
		case ubntTLV_IP:
			var ipStr string
			if len(val) == 4 {
				ipStr = fmt.Sprintf("%d.%d.%d.%d", val[0], val[1], val[2], val[3])
			} else if len(val) == 10 { // format includes MAC
				ip := val[6:]
				ipStr = fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
			}
			if ipStr != "" {
				if dev.IP == "" {
					dev.IP = ipStr
				}
				dev.IPs = append(dev.IPs, ipStr)
			}
		case ubntTLV_Version:
			vStr := string(val)
			idx := strings.Index(vStr, ".v")
			if idx != -1 {
				vStr = vStr[idx+1:]
			} else {
				idx = strings.Index(vStr, "v")
				if idx != -1 {
					vStr = vStr[idx:]
				}
			}
			parts := strings.Split(vStr, ".")
			if len(parts) > 3 {
				vStr = strings.Join(parts[:3], ".")
			}
			dev.Version = vStr
		case ubntTLV_Hostname:
			dev.Hostname = string(val)
		case ubntTLV_Model:
			dev.Model = string(val)
		case ubntTLV_Platform:
			dev.Platform = string(val)
		case ubntTLV_WEBUI:
			if len(val) == 4 {
				num := binary.BigEndian.Uint32(val)
				port := num & 0xFFFF
				protocol := (num >> 16) & 0xFFFF
				if protocol > 0 {
					dev.Hints = append(dev.Hints, fmt.Sprintf("https:%d", port))
				} else {
					dev.Hints = append(dev.Hints, fmt.Sprintf("http:%d", port))
				}
			}
		}
	}
	if dev.MAC == "" && dev.IP == "" {
		return nil
	}
	return dev
}
