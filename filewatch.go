package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/goccy/go-yaml"
)

var (
	lastFileHash    string
	lastFileContent string
)

func StartFileWatcher(filePath string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %v", err)
	}

	lastFileHash, err = getFileHash(filePath)
	if err != nil {
		log.Printf("Warning: Could not read initial file hash for %s: %v", filePath, err)
		lastFileHash = ""
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Warning: Could not read initial file content for %s: %v", filePath, err)
		lastFileContent = ""
	} else {
		lastFileContent = string(content)
		log.Printf("Initial file content captured (%d bytes)", len(lastFileContent))
	}

	err = watcher.Add(filePath)
	if err != nil {
		log.Printf("Could not watch file directly: %v", err)
		return err
	}

	log.Printf("Started monitoring file: %s", filePath)

	go func() {
		defer watcher.Close()
		debounceTimer := time.NewTimer(0)
		debounceTimer.Stop()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					log.Printf("File change detected: %s (%s)", event.Name, event.Op)

					debounceTimer.Reset(500 * time.Millisecond)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("File watcher error: %v", err)

			case <-debounceTimer.C:
				handleFileChange(filePath)
			}
		}
	}()

	return nil
}

func handleFileChange(filePath string) {
	newHash, err := getFileHash(filePath)
	if err != nil {
		log.Printf("Error reading file %s: %v", filePath, err)
		return
	}

	if newHash == lastFileHash {
		log.Printf("File hash unchanged, ignoring event")
		return
	}

	log.Printf("File content changed, generating diff...")

	newContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading new file content: %v", err)
		return
	}

	diff := generateDiff(lastFileContent, string(newContent))

	user := GetLastAnydeskUser()
	if user == "" {
		user = "Unknown (no recent login)"
	}

	msg := fmt.Sprintf(":warning: *Configuration File Changed on %s*\nFile: `%s`\nModified by: %s\nTime: %s (GMT)\n\n*Changes:*\n```\n%s\n```",
		Config.VMName, filePath, user, time.Now().UTC().Format(time.RFC3339), diff)

	log.Printf("Sending Slack notification for file change")
	if err := SendSlack(msg); err != nil {
		log.Printf("Error sending Slack notification: %v", err)
	}

	lastFileHash = newHash
	lastFileContent = string(newContent)
}

func getFileHash(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func generateDiff(oldContent, newContent string) string {
	var oldData, newData map[string]interface{}
	if err := yaml.Unmarshal([]byte(oldContent), &oldData); err != nil {
		return fmt.Sprintf("error unmarshalling old YAML: %v", err)
	}
	if err := yaml.Unmarshal([]byte(newContent), &newData); err != nil {
		return fmt.Sprintf("error unmarshalling new YAML: %v", err)
	}

	changes := make(map[string][]string)
	findDifferences(changes, "", oldData, newData)

	var parents []string
	for parent := range changes {
		parents = append(parents, parent)
	}
	sort.Strings(parents)

	var builder strings.Builder
	for _, parent := range parents {
		if parent != "" {
			builder.WriteString(fmt.Sprintf("%s:\n", parent))
		}
		for _, diff := range changes[parent] {
			builder.WriteString(diff)
		}
	}
	return builder.String()
}

func findDifferences(changes map[string][]string, parentKey string, oldData, newData map[string]interface{}) {
	allKeys := make(map[string]bool)
	for key := range oldData {
		allKeys[key] = true
	}
	for key := range newData {
		allKeys[key] = true
	}

	for key := range allKeys {
		oldVal, oldOk := oldData[key]
		newVal, newOk := newData[key]

		if oldOk && newOk {
			if oldMap, ok := oldVal.(map[string]interface{}); ok {
				if newMap, ok := newVal.(map[string]interface{}); ok {
					newParent := key
					if parentKey != "" {
						newParent = parentKey + "." + key
					}
					findDifferences(changes, newParent, oldMap, newMap)
					continue
				}
			}

			if fmt.Sprintf("%v", oldVal) != fmt.Sprintf("%v", newVal) {
				changes[parentKey] = append(changes[parentKey], fmt.Sprintf("- %s: %v\n", key, oldVal))
				changes[parentKey] = append(changes[parentKey], fmt.Sprintf("+ %s: %v\n", key, newVal))
			}
		} else if oldOk {
			changes[parentKey] = append(changes[parentKey], fmt.Sprintf("- %s: %v\n", key, oldVal))
		} else if newOk {
			changes[parentKey] = append(changes[parentKey], fmt.Sprintf("+ %s: %v\n", key, newVal))
		}
	}
}
