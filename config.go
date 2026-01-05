package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Version information
const (
	Version     = "2.1.0"
	BuildDate   = "2025-12-02"
	Description = "AnyDesk Tracker with External File Monitoring"
)

// Default Configuration Values
const (
	DefaultWebhookURL            = ""
	DefaultVMName                = ""
	DefaultUserTraceFile         = `C:\Users\This\AppData\Roaming\AnyDesk\ad.trace`
	DefaultServiceTraceFile      = `C:\ProgramData\AnyDesk\ad_svc.trace`
	DefaultAppLogFile            = `AnyDeskGoTracker.log`
	DefaultExternalFile          = `C:\User\user.yml`
	DefaultAllowedRecentDuration = 5 * time.Second
	DefaultLogTimeLayout         = "2006-01-02 15:04:05.000"
)

type AppConfig struct {
	WebhookURL            string        `yaml:"webhook_url"`
	VMName                string        `yaml:"vm_name"`
	UserTraceFile         string        `yaml:"user_trace_file"`
	ServiceTraceFile      string        `yaml:"service_trace_file"`
	AppLogFile            string        `yaml:"app_log_file"`
	ExternalFile          string        `yaml:"external_file"`
	AllowedRecentDuration time.Duration `yaml:"allowed_recent_duration"`
	LogTimeLayout         string        `yaml:"log_time_layout"`
}

var Config = AppConfig{
	WebhookURL:            DefaultWebhookURL,
	VMName:                DefaultVMName,
	UserTraceFile:         DefaultUserTraceFile,
	ServiceTraceFile:      DefaultServiceTraceFile,
	AppLogFile:            DefaultAppLogFile,
	ExternalFile:          DefaultExternalFile,
	AllowedRecentDuration: DefaultAllowedRecentDuration,
	LogTimeLayout:         DefaultLogTimeLayout,
}

func LoadConfig(filename string) {
	ex, err := os.Executable()
	if err != nil {
		log.Printf("Warning: Could not get executable path: %v", err)
		return
	}
	exPath := filepath.Dir(ex)
	configPath := filepath.Join(exPath, filename)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Config file not found at %s, using defaults", configPath)
		return
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Error reading config file: %v", err)
		return
	}

	err = yaml.Unmarshal(data, &Config)
	if err != nil {
		log.Printf("Error parsing config file: %v", err)
		return
	}

	log.Printf("Configuration loaded from %s", configPath)
}
