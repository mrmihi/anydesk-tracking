package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/kardianos/service"
)

type program struct {
	exit chan struct{}
}

func (p *program) Start(s service.Service) error {
	p.exit = make(chan struct{})
	go p.run()
	return nil
}

func (p *program) run() {
	LoadConfig("config.yaml")

	InitUserTracker()

	if !filepath.IsAbs(Config.AppLogFile) {
		ex, err := os.Executable()
		if err == nil {
			Config.AppLogFile = filepath.Join(filepath.Dir(ex), Config.AppLogFile)
		}
	}

	logDir := filepath.Dir(Config.AppLogFile)
	_ = os.MkdirAll(logDir, 0755)

	f, err := os.OpenFile(Config.AppLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
	} else {
		log.SetOutput(f)
	}

	log.Println("Starting AnyDesk Tracker Service...")
	log.Printf("Version: %s (Built: %s)", Version, BuildDate)
	log.Printf("Loaded Configuration: VMName=%s, Webhook configured: %v", Config.VMName, Config.WebhookURL != "")

	StartMonitoring(Config.UserTraceFile, "Login Track", HandleLogin)

	StartMonitoring(Config.ServiceTraceFile, "Logout Track", HandleLogout)

	if Config.ExternalFile != "" {
		log.Printf("Starting file watcher for: %s", Config.ExternalFile)
		if err := StartFileWatcher(Config.ExternalFile); err != nil {
			log.Printf("Warning: Could not start file watcher: %v", err)
		}
	} else {
		log.Println("External file monitoring disabled (no path configured)")
	}

	<-p.exit
}

func (p *program) Stop(s service.Service) error {
	log.Println("Stopping AnyDesk Tracker Service...")
	close(p.exit)
	return nil
}
