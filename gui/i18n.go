package gui

var Lang = "en"

var messages = map[string]map[string]string{
	"de": {
		"Interface": "Schnittstelle: ",
		"All": "Alle",
		"Start": "Start",
		"Stop": "STOPP",
		"Clear": "Leeren",
		"ScanningN": "Scanne V%s... %d Geräte gefunden",
		"Scanning1": "Scanne V%s... 1 Gerät gefunden",
		"Stopped1": "Angehalten — 1 Gerät gefunden",
		"StoppedN": "Angehalten — %d Geräte gefunden",
		"Ready": "Bereit — Start drücken",
		"Cleared": "Geleert",
		"PH_HTTPS": "HTTPS Ports (z.B. 10443, 2443)",
		"PH_HTTP": "HTTP Ports (z.B. 10080, 8080)",
		"IPAddress": "IP-Adresse",
		"MACAddress": "MAC-Adresse",
		"Hostname": "Hostname",
		"Model": "Modell",
		"Version": "Firmware / Version",
		"LAN": "LAN",
		"WLAN": "WLAN",
		"Virtual": "Virtuell",
		"Bridge": "Bridge",
		"CheckV1": "V1 (Alt)",
		"CheckV2": "V2 (Neu)",
		"InfoV1V2Title": "Ubiquiti Suchprotokolle (V1 & V2)",
		"InfoV1V2Text": "Ubiquiti-Geräte reagieren auf zwei verschiedene Suchanfragen (Probes).\nV1 ist das klassische Protokoll für ältere Hardware.\nV2 wird von modernen Geräten (wie aktuellen Switches oder Cloud Keys) genutzt und liefert detailliertere Infos.\n\nWenn ein Gerät nicht gefunden wird, aktiviere am besten beide Optionen.",
		"InfoPortsTitle": "Optionale Web-Ports",
		"InfoPortsText": "Wenn du in der Liste einen Doppelklick auf ein Gerät machst, versucht das Tool dessen Web-Oberfläche im Browser zu öffnen.\n\nTrage hier optionale Ports (kommagetrennt) ein, falls das Gerät vom Standard-Port abweicht.",
		"About": "Über / About",
		"AboutTextMini": "UBNT DisGOvery\nVersion 1.0.0\n\nEntwickelt von ExtrudeEmpty\nLizenziert unter der MIT-Lizenz\n\nBesonderer Dank geht an Carlos Guerrero (Entwickler des 'ubnt-discover' NodeJS-Pakets),\ndessen Arbeit als Inspiration für diese Software diente.",
	},
	"en": {
		"Interface": "Interface: ",
		"All": "All",
		"Start": "Start",
		"Stop": "STOP",
		"Clear": "Clear",
		"ScanningN": "Scanning V%s... %d devices found",
		"Scanning1": "Scanning V%s... 1 device found",
		"Stopped1": "Stopped — 1 device found",
		"StoppedN": "Stopped — %d devices found",
		"Ready": "Ready — press Start",
		"Cleared": "Cleared",
		"PH_HTTPS": "HTTPS Ports (e.g. 10443, 2443)",
		"PH_HTTP": "HTTP Ports (e.g. 10080, 8080)",
		"IPAddress": "IP Address",
		"MACAddress": "MAC Address",
		"Hostname": "Hostname",
		"Model": "Model",
		"Version": "Firmware / Version",
		"LAN": "LAN",
		"WLAN": "WiFi",
		"Virtual": "Virtual",
		"Bridge": "Bridge",
		"CheckV1": "V1 (Old)",
		"CheckV2": "V2 (New)",
		"InfoV1V2Title": "Ubiquiti Discovery Protocols (V1 & V2)",
		"InfoV1V2Text": "Ubiquiti devices respond to two different discovery probes.\nV1 is the classic protocol for older hardware.\nV2 is used by modern devices (like current switches or Cloud Keys) and provides more detailed info.\n\nIf a device is not found, it is best to enable both options.",
		"InfoPortsTitle": "Optional Web Ports",
		"InfoPortsText": "When you double-click a device in the list, the tool attempts to open its web interface in the browser.\n\nEnter optional ports here (comma-separated) in case the device deviates from the standard port.",
		"About": "About",
		"AboutTextMini": "UBNT DisGOvery\nVersion 1.0.0\n\nDeveloped by ExtrudeEmpty\nLicensed under MIT License\n\nSpecial thanks to Carlos Guerrero (developer of the 'ubnt-discover' NodeJS package),\nwhose work served as inspiration for this software.",
	},
}

func T(key string) string {
	if msgs, ok := messages[Lang]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}
	// Fallback to English
	if msgs, ok := messages["en"]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}
	return key
}
