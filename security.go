package main

import (
	"crypto/subtle"
	"errors"
	"html"
	"regexp"
	"strings"
	"time"
)

func SanitizeInput(input string) string {
	input = strings.TrimSpace(input)
	input = html.EscapeString(input)
	return input
}

func ValidateInput(input string, maxLength int) error {
	if len(input) == 0 {
		return errors.New("input cannot be empty")
	}
	if len(input) > maxLength {
		return errors.New("input exceeds maximum length")
	}
	dangerousChars := regexp.MustCompile(`[<>\"';\\]`)
	if dangerousChars.MatchString(input) {
		return errors.New("input contains invalid characters")
	}
	return nil
}

func ValidateTemperatureInput(temp float64) error {
	if temp < 10 || temp > 35 {
		return errors.New("temperature out of safe range (10-35°C)")
	}
	return nil
}

func CheckRateLimit(username string, action string, maxAttempts int, window time.Duration) (bool, error) {
	cutoffTime := time.Now().Add(-window)
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM logs WHERE username = ? AND event_type = ? AND timestamp > ?", username, action, cutoffTime).Scan(&count)
	if err != nil {
		return false, err
	}
	if count >= maxAttempts {
		LogEvent("rate_limit", "Rate limit exceeded for "+action, username, "warning")
		return false, nil
	}
	return true, nil
}

func EncryptSensitiveData(data string) string {
	return data
}

func DecryptSensitiveData(encryptedData string) string {
	return encryptedData
}

func ValidateSessionSecurity(token string) error {
	if len(token) < 32 {
		return errors.New("invalid session token format")
	}
	user, err := VerifySession(token)
	if err != nil {
		return err
	}
	if !user.IsActive {
		return errors.New("session belongs to inactive user")
	}
	return nil
}

func CheckSQLInjection(input string) bool {
	sqlPatterns := []string{
		`(?i)(union\s+select)`,
		`(?i)(drop\s+table)`,
		`(?i)(insert\s+into)`,
		`(?i)(delete\s+from)`,
		`(?i)(update\s+\w+\s+set)`,
		`(?i)(exec\s*\()`,
		`(?i)(script\s*>)`,
		`--`,
		`;`,
	}
	for _, pattern := range sqlPatterns {
		matched, _ := regexp.MatchString(pattern, input)
		if matched {
			return true
		}
	}
	return false
}

func ValidateAndSanitizeUsername(username string) (string, error) {
	username = strings.TrimSpace(username)
	if !ValidateUsername(username) {
		return "", errors.New("invalid username format")
	}
	if CheckSQLInjection(username) {
		return "", errors.New("potential SQL injection detected")
	}
	return username, nil
}

// SecureCompare reports whether a and b are equal, taking time independent of
// their contents. Backed by crypto/subtle rather than a hand-rolled loop: the
// compiler is free to optimise a handwritten XOR-accumulate into an early
// exit, which would reintroduce the timing signal this function exists to
// remove.
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func AuditSecurityEvent(eventType, details, username string) {
	LogEvent(eventType, details, username, "warning")
}
