package main

import (
	"fyne.io/fyne/v2/layout"
	"context"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
	"runtime"
	"ubnt-disgovery/gui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/net/ipv4"

	"ubnt-disgovery/core"
)

//go:embed icon.png
var iconBytes []byte

type UBNTDevice struct {
	IP       string
	MAC      string
	Hostname string
	Model    string
	Firmware string
	Hints    []string
	Iface    string
}

type doubleTappableCell struct {
	widget.BaseWidget
	text    *widget.Label
	onTap   func()
	onDbTap func()
}

func newDoubleTappableCell(txt *widget.Label) *doubleTappableCell {
	c := &doubleTappableCell{text: txt}
	c.ExtendBaseWidget(c)
	return c
}

func (c *doubleTappableCell) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.text)
}

func (c *doubleTappableCell) Tapped(e *fyne.PointEvent) {
	if c.onTap != nil {
		c.onTap()
	}
}

func (c *doubleTappableCell) DoubleTapped(e *fyne.PointEvent) {
	if c.onDbTap != nil {
		c.onDbTap()
	}
}

type TempIP struct {
	IP    string
	Iface string
	Type  string // "nmcli" or "ip"
}

var (
	tempIPs   []TempIP
	tempIPsMu sync.Mutex
)

func AddTempIP(ip, iface, ipType string) {
	tempIPsMu.Lock()
	tempIPs = append(tempIPs, TempIP{IP: ip, Iface: iface, Type: ipType})
	tempIPsMu.Unlock()
}

func CleanupTempIPs() {
	tempIPsMu.Lock()
	defer tempIPsMu.Unlock()
	for _, tip := range tempIPs {
		if tip.Type == "netsh" {
			cmd := exec.Command("netsh", "interface", "ipv4", "delete", "address", tip.Iface, strings.Split(tip.IP, "/")[0])
			_ = cmd.Run()
		} else if tip.Type == "nmcli" {
			cmd := exec.Command("nmcli", "device", "modify", tip.Iface, "-ipv4.addresses", tip.IP)
			_ = cmd.Run()
		} else if tip.Type == "nmcli-profile" {
			cmd := exec.Command("nmcli", "connection", "delete", "USSDisGOvery-"+tip.Iface)
			_ = cmd.Run()
		} else {
			cmd := exec.Command("ip", "addr", "del", tip.IP, "dev", tip.Iface)
			if err := cmd.Run(); err != nil {
				cmd = exec.Command("sudo", "ip", "addr", "del", tip.IP, "dev", tip.Iface)
				_ = cmd.Run()
			}
		}
	}
	tempIPs = nil
}

func injectDefaultSubnet() {
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, i := range ifaces {
		if i.Flags&net.FlagUp != 0 && i.Flags&net.FlagRunning != 0 && i.Flags&net.FlagLoopback == 0 {
			name := strings.ToLower(i.Name)
			if !strings.HasPrefix(name, "wl") && !strings.HasPrefix(name, "wi") && !strings.Contains(name, "vir") && !strings.Contains(name, "veth") && !strings.Contains(name, "br") {
				if !hasSubnetIP(i.Name, "192.168.1.20") {
					ipType, err := assignTempIP(i.Name, "192.168.1.253/24", false)
					if err == nil {
						AddTempIP("192.168.1.253/24", i.Name, ipType)
					}
				}
				break
			}
		}
	}
}

func detectTargetInterface(deviceIPStr string) string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	
	// 1. Check if any interface already has an IP in the same /24 subnet
	devIP := net.ParseIP(deviceIPStr)
	if devIP != nil {
		devIP4 := devIP.To4()
		if devIP4 != nil {
			_, subnet, _ := net.ParseCIDR(fmt.Sprintf("%d.%d.%d.0/24", devIP4[0], devIP4[1], devIP4[2]))
			for _, ifi := range ifaces {
				if ifi.Flags&net.FlagUp != 0 && ifi.Flags&net.FlagLoopback == 0 {
					addrs, _ := ifi.Addrs()
					for _, addr := range addrs {
						if ipnet, ok := addr.(*net.IPNet); ok {
							if ipnet.IP.To4() != nil && subnet.Contains(ipnet.IP) {
								return ifi.Name
							}
						}
					}
				}
			}
		}
	}

	// 2. Look for active physical interfaces with a connected cable (carrier == 1)
	var candidates []string
	for _, ifi := range ifaces {
		if ifi.Flags&net.FlagUp != 0 && ifi.Flags&net.FlagLoopback == 0 {
			carrierPath := fmt.Sprintf("/sys/class/net/%s/carrier", ifi.Name)
			if data, err := os.ReadFile(carrierPath); err == nil {
				if strings.TrimSpace(string(data)) == "1" {
					candidates = append(candidates, ifi.Name)
				}
			} else {
				candidates = append(candidates, ifi.Name)
			}
		}
	}
	
	if len(candidates) == 1 {
		return candidates[0]
	}
	
	// Prefer Ethernet interfaces
	for _, c := range candidates {
		if strings.HasPrefix(c, "e") {
			return c
		}
	}
	
	if len(candidates) > 0 {
		return candidates[0]
	}
	
	return ""
}

func hasSubnetIP(ifaceName string, deviceIPStr string) bool {
	devIP := net.ParseIP(deviceIPStr)
	if devIP == nil {
		return false
	}
	devIP4 := devIP.To4()
	if devIP4 == nil {
		return false
	}
	
	_, subnet, err := net.ParseCIDR(fmt.Sprintf("%d.%d.%d.0/24", devIP4[0], devIP4[1], devIP4[2]))
	if err != nil {
		return false
	}
	
	ifi, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return false
	}
	addrs, err := ifi.Addrs()
	if err != nil {
		return false
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.To4() != nil && subnet.Contains(ipnet.IP) {
				return true
			}
		}
	}
	return false
}

func getTempIPForUBNT(deviceIPStr string, existingDevices []UBNTDevice) string {
	devIP := net.ParseIP(deviceIPStr)
	if devIP == nil {
		return ""
	}
	devIP4 := devIP.To4()
	if devIP4 == nil {
		return ""
	}
	
	candidates := []int{253, 252, 99, 98, 254}
	for _, lastOctet := range candidates {
		candidateIP := fmt.Sprintf("%d.%d.%d.%d", devIP4[0], devIP4[1], devIP4[2], lastOctet)
		if candidateIP == deviceIPStr {
			continue
		}
		conflict := false
		for _, d := range existingDevices {
			if d.IP == candidateIP {
				conflict = true
				break
			}
		}
		if !conflict {
			return candidateIP + "/24"
		}
	}
	return fmt.Sprintf("%d.%d.%d.253/24", devIP4[0], devIP4[1], devIP4[2])
}

func assignTempIP(iface string, tempIP string, allowSudo bool) (string, error) {
	if iface == "" {
		return "", fmt.Errorf("no interface specified for temp IP")
	}

	if runtime.GOOS == "windows" {
		cmd := exec.Command("netsh", "interface", "ipv4", "add", "address", iface, strings.Split(tempIP, "/")[0], "255.255.255.0")
		if err := cmd.Run(); err == nil {
			return "netsh", nil
		}
		return "", fmt.Errorf("Windows assignment failed")
	}

	// Try NetworkManager first
	if debugMode {
	fmt.Printf("[DEBUG] Trying nmcli device modify %s +ipv4.addresses %s\n", iface, tempIP)
	}
	cmd := exec.Command("nmcli", "device", "modify", iface, "+ipv4.addresses", tempIP)
	out, err := cmd.CombinedOutput()
	if err == nil {
		if debugMode {
		fmt.Printf("[DEBUG] nmcli device modify succeeded.\n")
		}
		return "nmcli", nil
	}
	if debugMode {
	fmt.Printf("[DEBUG] nmcli device modify failed: %v, Output: %s\n", err, string(out))
	}

	// Fallback 1: Try creating/modifying a dedicated temporary profile (Approach A)
	profileName := "USSDisGOvery-" + iface
	if debugMode {
	fmt.Printf("[DEBUG] Trying nmcli connection modify %s\n", profileName)
	}
	cmd = exec.Command("nmcli", "connection", "modify", profileName, "+ipv4.addresses", tempIP)
	out, err = cmd.CombinedOutput()
	if err == nil {
		exec.Command("nmcli", "connection", "up", profileName).Run()
		return "nmcli-profile", nil
	}
	if debugMode {
	fmt.Printf("[DEBUG] nmcli connection modify failed: %v, Output: %s\n", err, string(out))
	}
	
	if debugMode {
	fmt.Printf("[DEBUG] Trying nmcli connection add %s\n", profileName)
	}
	cmd = exec.Command("nmcli", "connection", "add", "type", "ethernet", "ifname", iface, "con-name", profileName, "ipv4.method", "manual", "ipv4.addresses", tempIP)
	out, err = cmd.CombinedOutput()
	if err == nil {
		exec.Command("nmcli", "connection", "up", profileName).Run()
		return "nmcli-profile", nil
	}
	if debugMode {
	fmt.Printf("[DEBUG] nmcli connection add failed: %v, Output: %s\n", err, string(out))
	}

	// Fallback 2: ip directly
	if debugMode {
	fmt.Printf("[DEBUG] Trying ip addr add %s dev %s\n", tempIP, iface)
	}
	cmd = exec.Command("ip", "addr", "add", tempIP, "dev", iface)
	out, err = cmd.CombinedOutput()
	if err == nil {
		return "ip", nil
	}
	if debugMode {
	fmt.Printf("[DEBUG] ip addr add failed: %v, Output: %s\n", err, string(out))
	}

	if !allowSudo {
		return "", fmt.Errorf("assignment failed")
	}

	// Fallback 3: pkexec (opens graphical password dialog)
	if debugMode {
	fmt.Printf("[DEBUG] Reaching pkexec fallback for %s on %s\n", tempIP, iface)
	}
	cmd = exec.Command("pkexec", "ip", "addr", "add", tempIP, "dev", iface)
	if err := cmd.Run(); err == nil {
		return "ip", nil
	}

	return "", fmt.Errorf("assignment failed")
}

func findOpenWebPort(ip string, ports []string) string {
	if len(ports) == 0 {
		return ""
	}
	if len(ports) == 1 {
		return ports[0]
	}
	
	type result struct {
		port string
		open bool
	}
	ch := make(chan result, len(ports))
	
	for _, p := range ports {
		go func(port string) {
			conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, port), 200*time.Millisecond)
			if err == nil {
				conn.Close()
				ch <- result{port: port, open: true}
			} else {
				ch <- result{port: port, open: false}
			}
		}(p)
	}
	
	openPorts := make(map[string]bool)
	for i := 0; i < len(ports); i++ {
		res := <-ch
		if res.open {
			openPorts[res.port] = true
		}
	}
	
	for _, p := range ports {
		if openPorts[p] {
			return p
		}
	}
	
	return ""
}

func openWebGUI(d UBNTDevice, selectedIface string, customPorts []string) {
	ip := d.IP
	if ip == "" {
		return
	}
	
	iface := selectedIface
	if iface == "All Interfaces" || iface == "Alle Schnittstellen" || iface == gui.T("All") || iface == "Alle" {
		iface = ""
	}
	if iface == "" {
		iface = d.Iface
	}
	if iface == "" {
		iface = detectTargetInterface(ip)
	}
	
	if iface != "" {
		if !hasSubnetIP(iface, ip) {
			mu.Lock()
			tempIP := getTempIPForUBNT(ip, data)
			mu.Unlock()
			if tempIP != "" {
				ipType, err := assignTempIP(iface, tempIP, true)
				if err == nil {
					AddTempIP(tempIP, iface, ipType)
					time.Sleep(200 * time.Millisecond)
				}
			}
		}
	}
	
	// Collect candidates
	var candidatePorts []string
	portSchemes := make(map[string]string)
	
	if len(customPorts) > 0 {
		for _, cp := range customPorts {
			p := cp
			if strings.HasPrefix(p, "http:") {
				p = strings.TrimPrefix(p, "http:")
				portSchemes[p] = "http"
			} else if strings.HasPrefix(p, "https:") {
				p = strings.TrimPrefix(p, "https:")
				portSchemes[p] = "https"
			}
			candidatePorts = append(candidatePorts, p)
		}
	}
	
	for _, h := range d.Hints {
		if strings.HasPrefix(h, "http:") || strings.HasPrefix(h, "https:") {
			parts := strings.Split(h, " ")
			prefix := "http:"
			scheme := "http"
			if strings.HasPrefix(h, "https:") {
				prefix = "https:"
				scheme = "https"
			}
			p := strings.TrimPrefix(parts[0], prefix)
			if p != "" {
				dup := false
				for _, cp := range candidatePorts {
					if cp == p {
						dup = true
						break
					}
				}
				if !dup {
					candidatePorts = append(candidatePorts, p)
					if _, exists := portSchemes[p]; !exists {
						portSchemes[p] = scheme
					}
				}
			}
		}
	}
	
	has80 := false
	has443 := false
	for _, cp := range candidatePorts {
		if cp == "80" {
			has80 = true
		}
		if cp == "443" {
			has443 = true
		}
	}
	if !has443 {
		candidatePorts = append(candidatePorts, "443")
	}
	if !has80 {
		candidatePorts = append(candidatePorts, "80")
	}

	webPort := findOpenWebPort(ip, candidatePorts)
	if webPort == "" {
		webPort = candidatePorts[0]
	}

	scheme := "https"
	if s, ok := portSchemes[webPort]; ok {
		scheme = s
	} else if webPort == "80" {
		scheme = "http"
	} else if webPort == "443" {
		scheme = "https"
	} else if strings.Contains(webPort, "80") {
		scheme = "http"
	}

	portStr := ":" + webPort
	if (scheme == "http" && webPort == "80") || (scheme == "https" && webPort == "443") {
		portStr = ""
	}
	
	u, _ := url.Parse(scheme + "://" + ip + portStr)
	fyne.CurrentApp().OpenURL(u)
}



// debugMode enables verbose diagnostic output to stdout.
// Set to true during development to trace network operations.
var debugMode = false

var (
	devices            = make(map[string]UBNTDevice)
	mu                 sync.Mutex
	table              *widget.Table
	data               []UBNTDevice
	statusLbl          *widget.Label
	ifaceMap           map[string]string
	currentScanVersion string
)

func clearDevices() {
	mu.Lock()
	devices = make(map[string]UBNTDevice)
	data = nil
	mu.Unlock()
	table.Refresh()
}

func getMinitoolHeaders() []string {
	return []string{gui.T("IPAddress"), gui.T("MACAddress"), gui.T("Hostname"), gui.T("Model"), gui.T("Version")}
}

func main() {
	gui.LoadConfig()
	
	var ifaceSelect *widget.Select
	var ifaceMap map[string]string
	var httpsPortsEntry, httpPortsEntry *widget.Entry
	var checkV1, checkV2 *widget.Check
	var startBtn *gui.ColorButton
	var clearBtn *widget.Button
	var ifaceLabel *widget.Label

	a := app.New()
	a.Settings().SetTheme(&gui.USSTheme{})
	if len(iconBytes) > 0 {
		a.SetIcon(fyne.NewStaticResource("icon.png", iconBytes))
	}
	w := a.NewWindow("UBNT DisGOvery")
	w.Resize(fyne.NewSize(1000, 600))

	statusLbl = widget.NewLabel(gui.T("Ready"))
	
	versionLbl := widget.NewLabel("v1.0.0")
	aboutBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		lbl := widget.NewLabel(gui.T("AboutTextMini"))
		lbl.Alignment = fyne.TextAlignCenter
		dialog.ShowCustom(gui.T("About"), "OK", lbl, w)
	})
	aboutBtn.Importance = widget.LowImportance
	

	
	themeSwitcherBtn := widget.NewButtonWithIcon("", theme.ColorPaletteIcon(), func() {
		gui.ForceLightMode = !gui.ForceLightMode
		a.Settings().SetTheme(&gui.USSTheme{})
		gui.SaveConfig()
	})
	themeSwitcherBtn.Importance = widget.LowImportance

	langSwitcherBtn := widget.NewButton("EN", func() {})
	langSwitcherBtn.OnTapped = func() {
		if gui.Lang == "en" {
			gui.Lang = "de"
			langSwitcherBtn.SetText("DE")
		} else {
			gui.Lang = "en"
			langSwitcherBtn.SetText("EN")
		}
		gui.SaveConfig()
		
		statusLbl.SetText(gui.T("Ready"))
		ifaceLabel.SetText(gui.T("Interface"))
		startBtn.SetButtonText(gui.T("Start"))
		httpsPortsEntry.SetPlaceHolder(gui.T("PH_HTTPS"))
		httpPortsEntry.SetPlaceHolder(gui.T("PH_HTTP"))
		checkV1.SetText(gui.T("CheckV1"))
		checkV2.SetText(gui.T("CheckV2"))
		ifaceSelect.Options[0] = gui.T("All")
		if ifaceSelect.Selected == "All Interfaces" || ifaceSelect.Selected == "Alle Schnittstellen" {
			ifaceSelect.SetSelected(gui.T("All"))
		}
		table.Refresh()
	}
	if gui.Lang == "de" {
		langSwitcherBtn.SetText("DE")
	} else {
		langSwitcherBtn.SetText("EN")
	}
	langSwitcherBtn.Importance = widget.LowImportance

	rightPart := container.NewHBox(versionLbl, aboutBtn, langSwitcherBtn, themeSwitcherBtn)
	statusBar := container.NewBorder(nil, nil, statusLbl, rightPart)



	ifaces, _ := net.Interfaces()
	ifaceNames := []string{gui.T("All")}
	ifaceMap = make(map[string]string)
	
	var defaultIface string
	var fallbackIface string
	
	for _, i := range ifaces {
		if i.Flags&net.FlagUp != 0 && i.Flags&net.FlagRunning != 0 && i.Flags&net.FlagLoopback == 0 {
			display := gui.GetIfaceDisplayName(i)
			ifaceNames = append(ifaceNames, display)
			ifaceMap[display] = i.Name
			
			name := strings.ToLower(i.Name)
			if fallbackIface == "" {
				fallbackIface = display
			}
			if defaultIface == "" && !strings.HasPrefix(name, "wl") && !strings.HasPrefix(name, "wi") && !strings.Contains(name, "vir") && !strings.Contains(name, "veth") && !strings.Contains(name, "br") {
				defaultIface = display
			}
		}
	}
	ifaceSelect = widget.NewSelect(ifaceNames, nil)
	
	targetDefault := gui.T("All")
	if defaultIface != "" {
		targetDefault = defaultIface
	} else if fallbackIface != "" {
		targetDefault = fallbackIface
	}
	ifaceSelect.SetSelected(targetDefault)

	portValidator := func(s string) error {
		if s == "" {
			return nil
		}
		parts := strings.Split(s, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			p = strings.TrimPrefix(p, "http:")
			p = strings.TrimPrefix(p, "https:")
			port, err := strconv.Atoi(p)
			if err != nil {
				return errors.New("Ungültiges Format (nur Zahlen)")
			}
			if port < 1 || port > 65535 {
				return errors.New("Port muss zwischen 1 und 65535 liegen")
			}
		}
		return nil
	}

	httpsPortsEntry = widget.NewEntry()
	httpsPortsEntry.SetPlaceHolder(gui.T("PH_HTTPS"))
	httpsPortsEntry.Validator = portValidator
	httpsPortsEntry.SetText(gui.CurrentConfig.HttpsPortsStr)
	httpsPortsEntry.OnChanged = func(s string) {
		gui.CurrentConfig.HttpsPortsStr = s
		gui.SaveConfig()
	}

	httpPortsEntry = widget.NewEntry()
	httpPortsEntry.SetPlaceHolder(gui.T("PH_HTTP"))
	httpPortsEntry.Validator = portValidator
	httpPortsEntry.SetText(gui.CurrentConfig.HttpPortsStr)
	httpPortsEntry.OnChanged = func(s string) {
		gui.CurrentConfig.HttpPortsStr = s
		gui.SaveConfig()
	}

	

	table = widget.NewTable(
		func() (int, int) {
			mu.Lock()
			defer mu.Unlock()
			return len(data), len(getMinitoolHeaders())
		},
		func() fyne.CanvasObject {
			return newDoubleTappableCell(widget.NewLabel("Vorlage"))
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			cell := o.(*doubleTappableCell)
			l := cell.text
			mu.Lock()
			defer mu.Unlock()
			if i.Row >= len(data) {
				l.SetText("")
				return
			}
			d := data[i.Row]
			switch i.Col {
			case 0:
				l.SetText(d.IP)
			case 1:
				l.SetText(d.MAC)
			case 2:
				l.SetText(d.Hostname)
			case 3:
				l.SetText(d.Model)
			case 4:
				l.SetText(d.Firmware)
			}
			cell.onTap = func() {
				table.Select(i)
			}
			cell.onDbTap = func() {
				if err := httpsPortsEntry.Validate(); err != nil {
					dialog.ShowError(fmt.Errorf("Fehler bei HTTPS-Ports:\n%v", err), w)
					return
				}
				if err := httpPortsEntry.Validate(); err != nil {
					dialog.ShowError(fmt.Errorf("Fehler bei HTTP-Ports:\n%v", err), w)
					return
				}
				
				var customPorts []string
				httpsText := httpsPortsEntry.Text
				if httpsText != "" {
					parts := strings.Split(httpsText, ",")
					for _, p := range parts {
						p = strings.TrimSpace(p)
						if p != "" {
							if !strings.HasPrefix(p, "https:") && !strings.HasPrefix(p, "http:") {
								p = "https:" + p
							}
							customPorts = append(customPorts, p)
						}
					}
				}
				httpText := httpPortsEntry.Text
				if httpText != "" {
					parts := strings.Split(httpText, ",")
					for _, p := range parts {
						p = strings.TrimSpace(p)
						if p != "" {
							if !strings.HasPrefix(p, "http:") && !strings.HasPrefix(p, "https:") {
								p = "http:" + p
							}
							customPorts = append(customPorts, p)
						}
					}
				}
				
				sel := ifaceSelect.Selected
				if sel == gui.T("All") || sel == "All Interfaces" || sel == "Alle Schnittstellen" || sel == "All" || sel == "Alle" {
					sel = ""
				} else if mapped, ok := ifaceMap[sel]; ok {
					sel = mapped
				}

				openWebGUI(d, sel, customPorts)
			}
		},
	)

	table.ShowHeaderRow = true
	table.CreateHeader = func() fyne.CanvasObject {
		return widget.NewLabel("")
	}
	table.UpdateHeader = func(id widget.TableCellID, obj fyne.CanvasObject) {
		lbl := obj.(*widget.Label)
		if id.Row == -1 && id.Col >= 0 && id.Col < len(getMinitoolHeaders()) {
			lbl.SetText(getMinitoolHeaders()[id.Col])
		}
	}

	table.SetColumnWidth(0, 130) // IP
	table.SetColumnWidth(1, 150) // MAC
	table.SetColumnWidth(2, 220) // Hostname
	table.SetColumnWidth(3, 200) // Modell
	table.SetColumnWidth(4, 180) // Firmware

	checkV1 = widget.NewCheck(gui.T("CheckV1"), nil)
	checkV1.SetChecked(true)
	checkV2 = widget.NewCheck(gui.T("CheckV2"), nil)
	
	infoV1V2Btn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		lbl := widget.NewLabel(gui.T("InfoV1V2Text"))
		dialog.ShowCustom(gui.T("InfoV1V2Title"), "OK", lbl, w)
	})
	infoV1V2Btn.Importance = widget.LowImportance

	infoPortsBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		lbl := widget.NewLabel(gui.T("InfoPortsText"))
		dialog.ShowCustom(gui.T("InfoPortsTitle"), "OK", lbl, w)
	})
	infoPortsBtn.Importance = widget.LowImportance
	
	var cancel context.CancelFunc
	var isRunning bool

	
	clearBtn = widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		clearDevices()
		statusLbl.SetText(gui.T("Cleared"))
	})

	startBtn = gui.NewColorButton(gui.T("Start"), func() {
		if isRunning {
			if cancel != nil {
				cancel()
			}
			startBtn.SetButtonText(gui.T("Start"))
		startBtn.SetIsStop(false)
			clearBtn.Enable()
			isRunning = false
			if len(data) == 1 {
				statusLbl.SetText(gui.T("Stopped1"))
			} else {
				statusLbl.SetText(fmt.Sprintf(gui.T("StoppedN"), len(data)))
			}
			return
		}

		// Clear first
		clearDevices()

		isRunning = true
		startBtn.SetButtonText(gui.T("Stop"))
		startBtn.SetIsStop(true)
		clearBtn.Disable()
		statusLbl.SetText(fmt.Sprintf(gui.T("ScanningN"), currentScanVersion, 0))
		
		var ctx context.Context
		ctx, cancel = context.WithCancel(context.Background())

		sel := ifaceSelect.Selected
		if sel == gui.T("All") {
			sel = ""
		} else {
			sel = ifaceMap[sel]
		}
		
		go func() {
			if checkV1.Checked {
				runDiscovery(ctx, "1", sel)
			}
			if checkV2.Checked {
				runDiscovery(ctx, "2", sel)
			}
			
			fyne.Do(func() {
				if isRunning {
					cancel()
					startBtn.SetButtonText(gui.T("Start"))
		startBtn.SetIsStop(false)
					clearBtn.Enable()
					isRunning = false
					if len(data) == 1 {
						statusLbl.SetText(gui.T("Stopped1"))
					} else {
						statusLbl.SetText(fmt.Sprintf(gui.T("StoppedN"), len(data)))
					}
				}
			})
		}()
	})
	startBtn.SetIsStop(false)

	startBtnContainer := container.NewGridWrap(fyne.NewSize(200, 40), startBtn)
	clearBtnContainer := container.NewGridWrap(fyne.NewSize(40, 40), clearBtn)
	buttonsGroup := container.NewHBox(startBtnContainer, clearBtnContainer)

	ifaceLabel = widget.NewLabel(gui.T("Interface"))
	
	ifaceSelectContainer := container.NewGridWrap(fyne.NewSize(350, 40), ifaceSelect)

	row1 := container.NewHBox(
		buttonsGroup,
		layout.NewSpacer(),
		container.NewPadded(ifaceLabel), ifaceSelectContainer,
		layout.NewSpacer(),
		checkV1, checkV2, infoV1V2Btn,
	)
	row2 := container.NewBorder(nil, nil, nil, infoPortsBtn, container.NewGridWithColumns(2, httpsPortsEntry, httpPortsEntry))

	topControls := container.NewVBox(row1, row2)

	w.SetContent(container.NewBorder(
		topControls,
		statusBar, nil, nil, table,
	))

	w.SetCloseIntercept(func() {
		CleanupTempIPs()
		w.Close()
	})

	// Inject 192.168.1.x subnet automatically if missing on main ETH
	go injectDefaultSubnet()

	// Force translation refresh on start
	statusLbl.SetText(gui.T("Ready"))
	ifaceLabel.SetText(gui.T("Interface"))
	startBtn.SetButtonText(gui.T("Start"))
	if ifaceSelect.Selected == "All Interfaces" || ifaceSelect.Selected == "Alle Schnittstellen" || ifaceSelect.Selected == "All" || ifaceSelect.Selected == "Alle" {
		ifaceSelect.SetSelected(gui.T("All"))
	}
	table.Refresh()

	w.ShowAndRun()
}

func runDiscovery(ctx context.Context, version string, ifaceName string) {
	fyne.Do(func() {
		statusLbl.SetText(fmt.Sprintf(gui.T("ScanningN"), version, 0))
	})
	if debugMode {
	fmt.Printf("[DEBUG v%s] Starting discovery on interface: '%s'\n", version, ifaceName)
	}
	currentScanVersion = version
	port := 0
	if version == "2" {
		port = 10001
	}

	pc, err := net.ListenPacket("udp4", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("Listen error for v%s: %v\n", version, err)
		return
	}
	defer pc.Close()
	
	if udpConn, ok := pc.(*net.UDPConn); ok {
		_ = udpConn.SetReadBuffer(1048576)
	}

	pconn := ipv4.NewPacketConn(pc)
	pconn.SetMulticastLoopback(true)
	pconn.SetMulticastTTL(100)
	pconn.SetTTL(100)

	mcastIP := net.ParseIP("233.89.188.1")
	ifaces, _ := net.Interfaces()
	for _, ifi := range ifaces {
		if ifi.Flags&net.FlagUp != 0 && ifi.Flags&net.FlagLoopback == 0 {
			if ifaceName == "" || ifi.Name == ifaceName {
				if debugMode {
				fmt.Printf("[DEBUG v%s] Joining multicast group on %s\n", version, ifi.Name)
				}
				_ = pconn.JoinGroup(&ifi, &net.UDPAddr{IP: mcastIP})
			}
		}
	}
	
	msg := []byte{0x01, 0x00, 0x00, 0x00}
	var msg2 []byte
	if version == "2" {
		msg = []byte{0x02, 0x0a, 0x00, 0x00}
		msg2 = []byte{0x02, 0x06, 0x00, 0x00} // UniFi specific discovery
	}

	// Start receiver goroutine FIRST
	if debugMode {
	fmt.Printf("[DEBUG v%s] Starting receiver goroutine\n", version)
	}
	go func() {
		buf := make([]byte, 4096)
		for {
			_ = pc.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, addr, err := pc.ReadFrom(buf)
			if err != nil {
				// Check context
				select {
				case <-ctx.Done():
					if debugMode {
					fmt.Printf("[DEBUG v%s] Receiver goroutine exiting (context cancelled)\n", version)
					}
					return
				default:
					continue
				}
			}
			
			if n <= 4 {
				continue // Ignore echoed requests
			}
			
			if debugMode {
			fmt.Printf("[DEBUG v%s] Received %d bytes from %v\n", version, n, addr)
			}
			dev := core.ParseUBNTPacket(buf[:n])
			if dev != nil {
				if dev.IP == "" {
					if uaddr, ok := addr.(*net.UDPAddr); ok {
						dev.IP = uaddr.IP.String()
					}
				}
				addDevice(UBNTDevice{
					IP:       dev.IP,
					MAC:      dev.MAC,
					Hostname: dev.Hostname,
					Model:    dev.Model,
					Firmware: dev.Version,
					Hints:    dev.Hints,
				})
			}
		}
	}()

	// 1. Send to Multicast
	mcastAddr := &net.UDPAddr{IP: mcastIP, Port: 10001}
	for _, ifi := range ifaces {
		if ifaceName != "" && ifi.Name != ifaceName {
			continue
		}
		if ifi.Flags&net.FlagUp != 0 && ifi.Flags&net.FlagLoopback == 0 {
			if debugMode {
			fmt.Printf("[DEBUG v%s] Sending Multicast via %s\n", version, ifi.Name)
			}
			_ = pconn.SetMulticastInterface(&ifi)
			if _, err := pc.WriteTo(msg, mcastAddr); err != nil {
				fmt.Printf("Multicast WriteTo error (v%s) on %s: %v\n", version, ifi.Name, err)
			}
			if msg2 != nil {
				_, _ = pc.WriteTo(msg2, mcastAddr)
			}
		}
	}

	// 2. UDP Unicast Spray across all subnets
	if debugMode {
	fmt.Printf("[DEBUG v%s] Starting Unicast spray...\n", version)
	}
	for pass := 0; pass < 2; pass++ {
		if pass == 1 {
			time.Sleep(1 * time.Second) // Wait for ARP resolution from first pass
		}
		for _, ifi := range ifaces {
			if ifaceName != "" && ifi.Name != ifaceName {
				continue
			}
			addrs, _ := ifi.Addrs()
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if ipnet.IP.To4() != nil && !ipnet.IP.IsLoopback() {
						ones, _ := ipnet.Mask.Size()
						if ones <= 20 {
							if pass == 0 {
								if debugMode {
								fmt.Printf("[DEBUG v%s] Skipping large subnet %s on %s (prefix /%d)\n", version, ipnet.String(), ifi.Name, ones)
								}
							}
							continue
						}
						
						if pass == 0 {
							// Global Broadcast
							globalBcast := net.ParseIP("255.255.255.255")
							if debugMode {
							fmt.Printf("[DEBUG v%s] Sending Global Broadcast to 255.255.255.255 on %s\n", version, ifi.Name)
							}
							pc.WriteTo(msg, &net.UDPAddr{IP: globalBcast, Port: 10001})
							if msg2 != nil {
								pc.WriteTo(msg2, &net.UDPAddr{IP: globalBcast, Port: 10001})
							}
						}

						// Broadcast address
						bcast := getBroadcast(ipnet)
						if bcast != nil {
							if pass == 0 {
								if debugMode {
								fmt.Printf("[DEBUG v%s] Sending Subnet Broadcast to %s on %s\n", version, bcast.String(), ifi.Name)
								}
							}
							if _, err := pc.WriteTo(msg, &net.UDPAddr{IP: bcast, Port: 10001}); err != nil && pass == 0 {
								fmt.Printf("Broadcast WriteTo error (v%s): %v\n", version, err)
							}
						}

						// Unicast Spray
						if pass == 0 {
							if debugMode {
							fmt.Printf("[DEBUG v%s] Spraying %s on %s\n", version, ipnet.String(), ifi.Name)
							}
						}
						spraySubnet(ctx, pc, msg, ipnet)
						if msg2 != nil {
							spraySubnet(ctx, pc, msg2, ipnet)
						}
					}
				}
			}
		}
	}

	if debugMode {
	fmt.Printf("[DEBUG v%s] Finished sending all packets. Waiting 3s or until STOP...\n", version)
	}
	// Wait for responses or cancellation
	select {
	case <-ctx.Done():
		if debugMode {
		fmt.Printf("[DEBUG v%s] STOP requested. Exiting wait.\n", version)
		}
	case <-time.After(3 * time.Second):
		if debugMode {
		fmt.Printf("[DEBUG v%s] 3s wait finished.\n", version)
		}
	}
	if debugMode {
	fmt.Printf("[DEBUG v%s] Closing socket.\n", version)
	}
}

func addDevice(d UBNTDevice) {
	mu.Lock()
	if _, exists := devices[d.MAC]; !exists {
		devices[d.MAC] = d
		data = append(data, d)
		count := len(data)
		mu.Unlock()
		fyne.Do(func() {
			table.Refresh()
			if statusLbl != nil {
				if count == 1 {
					statusLbl.SetText(fmt.Sprintf(gui.T("Scanning1"), currentScanVersion))
				} else {
					statusLbl.SetText(fmt.Sprintf(gui.T("ScanningN"), currentScanVersion, count))
				}
			}
		})
	} else {
		mu.Unlock()
	}
}

func spraySubnet(ctx context.Context, pc net.PacketConn, msg []byte, ipnet *net.IPNet) {
	ip := ipnet.IP.To4()
	if ip == nil {
		return
	}
	
	ipInt := binary.BigEndian.Uint32(ip)
	maskInt := binary.BigEndian.Uint32(ipnet.Mask)
	
	network := ipInt & maskInt
	broadcast := network | ^maskInt
	
	count := broadcast - network - 1
	
	b := make([]byte, 4)
	var i uint32
	for i = 1; i <= count; i++ {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return
		default:
		}

		binary.BigEndian.PutUint32(b, network+i)
		target := net.IP(b)
		_, _ = pc.WriteTo(msg, &net.UDPAddr{IP: target, Port: 10001})
		
		// Yield thread to prevent UI freeze and OS network buffer drops!
		if i%50 == 0 {
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func getBroadcast(ipnet *net.IPNet) net.IP {
	ip := ipnet.IP.To4()
	if ip == nil {
		return nil
	}
	ipInt := binary.BigEndian.Uint32(ip)
	maskInt := binary.BigEndian.Uint32(ipnet.Mask)
	broadcast := (ipInt & maskInt) | ^maskInt
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, broadcast)
	return net.IP(b)
}
