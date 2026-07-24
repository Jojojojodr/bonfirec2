package commands

import (
	"runtime"
	"sort"
	"strings"
)

var available = map[string]string{
	"whoami":   "whoami",
	"ping":     "ping",
	"hostname": "hostname",
	"ip":       "ip",
	"list":     "list",
	"cmd":      "cmd",
	"exit":     "exit",
	"getinfo":  "getinfo",
}

// ParseSlashCommand splits "/whoami" into "whoami" and an optional payload.
func ParseSlashCommand(input string) (string, string, bool) {
	trimmed := strings.TrimSpace(input)
	if !strings.HasPrefix(trimmed, "/") {
		return "", "", false
	}

	trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "/"))
	if trimmed == "" {
		return "", "", false
	}

	parts := strings.Fields(trimmed)
	name := parts[0]
	if _, ok := available[name]; !ok {
		return "", "", false
	}

	payload := strings.TrimSpace(strings.TrimPrefix(trimmed, name))
	return name, payload, true
}

func IsKnown(name string) bool {
	_, ok := available[name]
	return ok
}

func ListSlashCommands() []string {
	list := make([]string, 0, len(available))
	for name := range available {
		list = append(list, "/"+name)
	}
	sort.Strings(list)
	return list
}

func GetCommand(command string) string {
	var resolved string
	switch command {
	case "ping":
		if runtime.GOOS == "windows" {
			resolved = "ping -n 1 127.0.0.1"
		} else {
			resolved = "ping -c 1 127.0.0.1"
		}
	case "ip":
		if runtime.GOOS == "windows" {
			resolved = "ipconfig"
		} else {
			resolved = "ip addr"
		}
	case "list":
		if runtime.GOOS == "windows" {
			resolved = "dir"
		} else {
			resolved = "ls"
		}
	default:
		resolved = command
	}
	return resolved
}
