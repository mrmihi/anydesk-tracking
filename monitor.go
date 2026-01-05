package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/nxadm/tail"
)

var (
	// Regex for timestamp: 2025-01-01 12:00:00.123
	reTimestamp = regexp.MustCompile(`(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3})`)

	// Regex for Login: Incoming session request: User (ID)
	reLogin = regexp.MustCompile(`Incoming session request: (.+?) \(([^)]+)\)`)

	// Regex for Logout: Session closed by
	reLogout = regexp.MustCompile(`Session closed by`)

	notificationCache      = make(map[string]time.Time)
	notificationCacheMutex sync.RWMutex
	deduplicationWindow    = 30 * time.Second
)

type LineHandler func(line string, logTime time.Time, label string)

func StartMonitoring(filePath string, label string, handler LineHandler) {
	config := tail.Config{
		Follow:    true,
		ReOpen:    true, // Reopen if file is rotated/truncated
		MustExist: false,
		Poll:      true,                                 // Better for Windows
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, // Tail from end
		Logger:    tail.DiscardingLogger,
	}

	t, err := tail.TailFile(filePath, config)
	if err != nil {
		log.Printf("[%s] Error starting tail: %v", label, err)
		return
	}

	log.Printf("[%s] Started monitoring %s", label, filePath)

	go func() {
		for line := range t.Lines {
			if line.Err != nil {
				log.Printf("[%s] Error reading line: %v", label, line.Err)
				continue
			}
			processLine(line.Text, label, handler)
		}
	}()
}

func processLine(text string, label string, handler LineHandler) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return
	}

	// Normalize whitespace
	normalized := strings.Join(strings.Fields(trimmed), " ")

	// Extract timestamp
	matches := reTimestamp.FindStringSubmatch(normalized)
	if matches == nil {
		return
	}

	logTime, err := time.Parse(Config.LogTimeLayout, matches[1])
	if err != nil {
		return
	}

	handler(normalized, logTime, label)
}

func isRecent(t time.Time) bool {
	now := time.Now().UTC()
	diff := now.Sub(t)
	if diff < 0 {
		diff = -diff
	}
	return diff <= Config.AllowedRecentDuration
}

func HandleLogin(line string, logTime time.Time, label string) {
	matches := reLogin.FindStringSubmatch(line)
	if matches != nil {
		userName := matches[1]
		userID := matches[2]

		SetLastAnydeskUser(userName)

		if isRecent(logTime) {
			eventKey := fmt.Sprintf("%s|%s|%s", userName, userID, logTime.Format(time.RFC3339))
			eventHash := fmt.Sprintf("%x", sha256.Sum256([]byte(eventKey)))

			notificationCacheMutex.Lock()
			lastNotified, exists := notificationCache[eventHash]
			now := time.Now()

			for hash, notifTime := range notificationCache {
				if now.Sub(notifTime) > deduplicationWindow {
					delete(notificationCache, hash)
				}
			}

			if exists && now.Sub(lastNotified) < deduplicationWindow {
				notificationCacheMutex.Unlock()
				log.Printf("Skipped duplicate notification for %s (ID: %s) - already sent %v ago",
					userName, userID, now.Sub(lastNotified))
				return
			}

			notificationCache[eventHash] = now
			notificationCacheMutex.Unlock()

			msg := fmt.Sprintf(":rotating-light-red: *%s* AnyDesk session request detected\nUser: %s\nID: %s\nTime: %s (GMT)\nSource: %s",
				Config.VMName, userName, userID, logTime.Format(time.RFC3339), label)

			log.Printf("Sending Slack (Login): %s", msg)
			if err := SendSlack(msg); err != nil {
				log.Printf("Error sending Slack: %v", err)
			}
		} else {
			log.Printf("Skipped old session request (LogTime=%s)", logTime)
		}
	}
}

func HandleLogout(line string, logTime time.Time, label string) {
	if reLogout.MatchString(line) {
		if isRecent(logTime) {
			msg := fmt.Sprintf(":white_check_mark: *%s* AnyDesk session closed (logout)\nTime: %s (GMT)",
				Config.VMName, logTime.Format(time.RFC3339))

			log.Printf("Sending Slack (Logout): %s", msg)
			if err := SendSlack(msg); err != nil {
				log.Printf("Error sending Slack: %v", err)
			}
		} else {
			log.Printf("Skipped old logout event (LogTime=%s)", logTime)
		}
	}
}
