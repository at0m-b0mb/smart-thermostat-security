package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

// OWASP - Broken Access Control
// CWE-639: Authorization Bypass Through User-Controlled Key

type AccessLevel int

const (
	AccessNone AccessLevel = iota
	AccessGuest
	AccessHomeowner
	AccessTechnician
)

type User struct {
	ID          string
	Username    string
	Role        string
	AccessLevel AccessLevel
	SessionID   string
	CreatedAt   time.Time
	LastAccess  time.Time
}

func EnforceAccessControl(user User, requiredLevel AccessLevel) error {
	if user.AccessLevel < requiredLevel {
		LogSecurityEvent("ACCESS_DENIED", fmt.Sprintf("User %s attempted action requiring level %d", user.Username, requiredLevel))
		return errors.New("No permission")
	}
	return nil
}

// ensures users can only access their own resources
func ValidateResourceOwnership(userID, resourceOwnerID string) error {
	if userID != resourceOwnerID {
		return errors.New("Unauthorized access to resource")
	}
	return nil
}

// OWASP - Cryptographic Failures
// CWE-916: Use of Password Hash With Insufficient Computational Effort

type PasswordHash struct {
	Hash string
	Salt string
}

// HashPassword uses Argon2id with secure parameters
// Prevents CWE-916: Use of Password Hash With Insufficient Computational Effort
// Prevents rainbow table attacks
func HashPassword(password string) (PasswordHash, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return PasswordHash{}, err
	}

	hash := argon2.IDKey([]byte(password), salt, 2, 64*1024, 4, 32)

	return PasswordHash{
		Hash: base64.StdEncoding.EncodeToString(hash),
		Salt: base64.StdEncoding.EncodeToString(salt),
	}, nil
}

// Verify password
func VerifyPassword(password string, stored PasswordHash) bool {
	salt, err := base64.StdEncoding.DecodeString(stored.Salt)
	if err != nil {
		return false
	}

	hash := argon2.IDKey([]byte(password), salt, 2, 64*1024, 4, 32)
	storedHash, err := base64.StdEncoding.DecodeString(stored.Hash)
	if err != nil {
		return false
	}

	return subtle.ConstantTimeCompare(hash, storedHash) == 1
}

// OWASP - Injection
// CWE-78: OS Command Injection
func ValidateInput(input string, inputType string) error {
	if input == "" {
		return errors.New("input cannot be empty")
	}

	switch inputType {
	case "username":
		if !regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`).MatchString(input) {
			return errors.New("invalid username format")
		}
	case "email":
		if !regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(input) {
			return errors.New("invalid email format")
		}
	case "temperature":
		if !regexp.MustCompile(`^[0-9]{1,3}(\.[0-9]{1,2})?$`).MatchString(input) {
			return errors.New("invalid temperature format")
		}
	}

	return nil
}

func SanitizeCommand(cmd string) (string, error) {
	// Remove any shell metacharacters
	dangerous := []string{";", "&", "|", ">", "<", "`", "$", "(", ")", "{", "}", "[", "]", "\\", "'", "\"", "*", "?"}
	for _, char := range dangerous {
		if strings.Contains(cmd, char) {
			return "", errors.New("invalid characters in command")
		}
	}

	return cmd, nil
}
