package config

import (
	"log"
	"strings"
)

// MaskSecret masks a secret string for logging
func MaskSecret(secret string) string {
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "***" + secret[len(secret)-4:]
}

// LogConfigValue logs a configuration value with appropriate formatting
func LogConfigValue(key string, value interface{}, indent int) {
	prefix := strings.Repeat("  ", indent)
	log.Printf("%s%s: %v", prefix, key, value)
}

// LogConfigSection logs a configuration section header
func LogConfigSection(section string) {
	log.Printf("%s Configuration:", section)
}

// ParseCSV parses a comma-separated string into a slice
func ParseCSV(value string) []string {
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
} 