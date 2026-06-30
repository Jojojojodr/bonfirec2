package commands

import (
	"sort"
	"strings"
)

var available = map[string]string{
	"whoami": "whoami",
	"ping":   "ping",
}

// ParseSlashCommand converts "/whoami" to "whoami" and validates it.
func ParseSlashCommand(input string) (string, bool) {
	trimmed := strings.TrimSpace(input)
	if !strings.HasPrefix(trimmed, "/") {
		return "", false
	}

	name := strings.TrimPrefix(trimmed, "/")
	name = strings.TrimSpace(name)
	if name == "" {
		return "", false
	}

	if _, ok := available[name]; !ok {
		return "", false
	}

	return name, true
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
