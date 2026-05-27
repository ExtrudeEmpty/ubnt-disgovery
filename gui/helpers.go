package gui

import (
	"fmt"
	"net"
	"strings"
)

func GetIfaceDisplayName(i net.Interface) string {
	typ := "LAN"
	name := strings.ToLower(i.Name)
	if strings.HasPrefix(name, "wl") || strings.HasPrefix(name, "wi") {
		typ = T("WLAN")
	} else if strings.Contains(name, "vir") || strings.Contains(name, "veth") || strings.Contains(name, "br") {
		typ = T("Virtual")
		if strings.Contains(name, "br") {
			typ = T("Bridge")
		}
	} else {
		typ = T("LAN")
	}

	ipStr := ""
	addrs, _ := i.Addrs()
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok {
			if ipnet.IP.To4() != nil {
				ipStr = ipnet.IP.String()
				break
			}
		}
	}

	if ipStr != "" {
		return fmt.Sprintf("%s (%s) - %s", i.Name, typ, ipStr)
	}
	return fmt.Sprintf("%s (%s)", i.Name, typ)
}
