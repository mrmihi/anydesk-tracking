package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kardianos/service"
)

func main() {
	svcConfig := &service.Config{
		Name:        "AnyDeskTracker",
		DisplayName: "AnyDesk Unified Tracker",
		Description: "Monitors AnyDesk logs for login/logout events and notifies Slack.",
		Arguments:   []string{},
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		cmd := os.Args[1]
		switch cmd {
		case "version":
			fmt.Printf("%s v%s (Built: %s)\n", Description, Version, BuildDate)
			return
		case "install":
			err = s.Install()
			if err != nil {
				log.Fatalf("Failed to install: %v", err)
			}
			fmt.Println("Service installed successfully.")
			return
		case "uninstall":
			err = s.Uninstall()
			if err != nil {
				log.Fatalf("Failed to uninstall: %v", err)
			}
			fmt.Println("Service uninstalled successfully.")
			return
		case "start":
			err = s.Start()
			if err != nil {
				log.Fatalf("Failed to start: %v", err)
			}
			fmt.Println("Service started successfully.")
			return
		case "stop":
			err = s.Stop()
			if err != nil {
				log.Fatalf("Failed to stop: %v", err)
			}
			fmt.Println("Service stopped successfully.")
			return
		case "run":
		default:
			fmt.Printf("Usage: %s [install|uninstall|start|stop|run|version]\n", os.Args[0])
			return
		}
	}

	err = s.Run()
	if err != nil {
		log.Fatal(err)
	}
}
