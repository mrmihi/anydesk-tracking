package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	lastAnydeskUser     string
	lastAnydeskUserTime time.Time
	userMutex           sync.RWMutex
	userStateFile       string
)

type UserState struct {
	UserName string    `json:"user_name"`
	LastSeen time.Time `json:"last_seen"`
}

func InitUserTracker() {
	ex, err := os.Executable()
	if err == nil {
		userStateFile = filepath.Join(filepath.Dir(ex), ".anydesk_user_state.json")
	} else {
		userStateFile = ".anydesk_user_state.json"
	}
	loadUserState()
}

func loadUserState() {
	data, err := os.ReadFile(userStateFile)
	if err != nil {
		log.Printf("No previous user state found (this is normal on first run)")
		return
	}

	var state UserState
	if err := json.Unmarshal(data, &state); err != nil {
		log.Printf("Error loading user state: %v", err)
		return
	}

	userMutex.Lock()
	lastAnydeskUser = state.UserName
	lastAnydeskUserTime = state.LastSeen
	userMutex.Unlock()

	log.Printf("Restored last AnyDesk user: %s (last seen: %s)", state.UserName, state.LastSeen.Format(time.RFC3339))
}

func saveUserState() {
	userMutex.RLock()
	state := UserState{
		UserName: lastAnydeskUser,
		LastSeen: lastAnydeskUserTime,
	}
	userMutex.RUnlock()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		log.Printf("Error marshaling user state: %v", err)
		return
	}

	if err := os.WriteFile(userStateFile, data, 0600); err != nil {
		log.Printf("Error saving user state: %v", err)
	}
}

// SetLastAnydeskUser records the most recent AnyDesk user
func SetLastAnydeskUser(userName string) {
	userMutex.Lock()
	lastAnydeskUser = userName
	lastAnydeskUserTime = time.Now()
	userMutex.Unlock()
	saveUserState()
}

// GetLastAnydeskUser returns the most recent AnyDesk user
func GetLastAnydeskUser() string {
	userMutex.RLock()
	defer userMutex.RUnlock()

	if lastAnydeskUser == "" {
		return "Unknown (no recent login)"
	}

	timeSince := time.Since(lastAnydeskUserTime)
	if timeSince > 24*time.Hour {
		return lastAnydeskUser + " (last seen " + formatDuration(timeSince) + " ago)"
	}

	return lastAnydeskUser
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	}
	return fmt.Sprintf("%d days", int(d.Hours()/24))
}
