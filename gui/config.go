package gui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Lang           string `json:"lang"`
	ForceLightMode bool   `json:"force_light_mode"`
	HttpsPortsStr  string `json:"https_ports"`
	HttpPortsStr   string `json:"http_ports"`
}

var CurrentConfig = Config{
	Lang:           "en",
	ForceLightMode: false,
	HttpsPortsStr:  "",
	HttpPortsStr:   "",
}

func getConfigPath() string {
	exePath, err := os.Executable()
	if err != nil {
		return "ubnt-disgovery.cfg"
	}
	base := filepath.Base(exePath)
	ext := filepath.Ext(base)
	name := base[0 : len(base)-len(ext)]
	return filepath.Join(filepath.Dir(exePath), name+".cfg")
}

func LoadConfig() {
	path := getConfigPath()
	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &CurrentConfig)
		Lang = CurrentConfig.Lang
		ForceLightMode = CurrentConfig.ForceLightMode
	} else {
		Lang = CurrentConfig.Lang
		ForceLightMode = CurrentConfig.ForceLightMode
		SaveConfig()
	}
}

func SaveConfig() {
	CurrentConfig.Lang = Lang
	CurrentConfig.ForceLightMode = ForceLightMode

	data, err := json.MarshalIndent(CurrentConfig, "", "  ")
	if err == nil {
		_ = os.WriteFile(getConfigPath(), data, 0644)
	}
}
