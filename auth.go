package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"regexp"
	"time"
	"golang.org/x/crypto/bcrypt"
)

const (
	MinPasswordLen         = 8
	MinUsernameLen         = 3
	MaxFailedLoginAttempts = 5
	AccountLockDuration    = 15 * time.Minute
		MinPinLen              = 4
)

type User struct {
	ID           int
	Username     string
	PasswordHash string
	Role         string
	SessionToken string
	LastLogin    time.Time
	IsActive     bool
}

var validUsername = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("password hash failed")
	}
	return string(hash), nil
}

func CheckPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ValidateUsername(username string) bool {
	return validUsername.MatchString(username)
}

func ValidatePassword(password string) error {
	if len(password) < MinPasswordLen {
		return errors.New("password must be at least 8 characters")
	}
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasUpper || !hasLower || !hasDigit {
		return errors.New("password must contain uppercase, lowercase, and digit")
	}
	return nil
}

// ValidatePin validates a PIN for guest accounts (numeric, minimum 4 digits)
func ValidatePin(pin string) error {
	if len(pin) < MinPinLen {
		return errors.New("PIN must be at least 4 digits")
	}
	// PIN should only contain digits
	hasDigit := regexp.MustCompile(`^[0-9]+$`).MatchString(pin)
	if !hasDigit {
		return errors.New("PIN must contain only digits")
	}
	return nil
}

func RegisterUser(username, password, role string) error {
	if len(username) < MinUsernameLen {
		return errors.New("username too short")
	}
	if !ValidateUsername(username) {
		return errors.New("invalid username format")
	}
	if err := ValidatePassword(password); err != nil {
		return err
	}
	if role != "homeowner" && role != "technician" && role != "guest" {
		return errors.New("invalid role")
	}
	passHash, err := HashPassword(password)
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)", username, passHash, role)
	if err != nil {
		return errors.New("user already exists")
	}
	LogEvent("register", "User registered", username, "info")
	return nil
}

// RegisterGuestUser registers a guest with a PIN instead of a password
func RegisterGuestUser(username, pin string) error {
	if len(username) < MinUsernameLen {
		return errors.New("username too short")
	}
	if !ValidateUsername(username) {
		return errors.New("invalid username format")
	}
	// Validate PIN instead of password
	if err := ValidatePin(pin); err != nil {
		return err
	}
	// Hash the PIN using bcrypt (same as password)
	pinHash, err := HashPassword(pin)
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)", username, pinHash, "guest")
	if err != nil {
		return errors.New("user already exists")
	}
	LogEvent("register", "Guest registered with PIN", username, "info")
	return nil
}

func isAccountLocked(username string) (bool, error) {
	var lockedUntil sql.NullTime
	var failedAttempts int
	err := db.QueryRow("SELECT failed_login_attempts, locked_until FROM users WHERE username = ?", username).Scan(&failedAttempts, &lockedUntil)
	if err != nil {
		return false, err
	}
	if lockedUntil.Valid && time.Now().Before(lockedUntil.Time) {
		return true, nil
	}
	if lockedUntil.Valid && time.Now().After(lockedUntil.Time) {
		db.Exec("UPDATE users SET failed_login_attempts = 0, locked_until = NULL WHERE username = ?", username)
	}
	return false, nil
}

func IsTechnicianAccessAllowed(db *sql.DB, username string) bool {
    var expiresAt time.Time
    err := db.QueryRow(
        "SELECT expires_at FROM guest_access WHERE guest_username = ? AND expires_at > ? ORDER BY expires_at DESC LIMIT 1",
        username, time.Now(),
    ).Scan(&expiresAt)
    // Access only allowed if there is a non-expired grant
    return err == nil && time.Now().Before(expiresAt)
}

func incrementFailedLogin(username string) error {
	var failedAttempts int
	db.QueryRow("SELECT failed_login_attempts FROM users WHERE username = ?", username).Scan(&failedAttempts)
	failedAttempts++
	if failedAttempts >= MaxFailedLoginAttempts {
		lockUntil := time.Now().Add(AccountLockDuration)
		_, err := db.Exec("UPDATE users SET failed_login_attempts = ?, locked_until = ? WHERE username = ?", failedAttempts, lockUntil, username)
		LogEvent("account_locked", "Account locked", username, "warning")
		return err
	}
	_, err := db.Exec("UPDATE users SET failed_login_attempts = ? WHERE username = ?", failedAttempts, username)
	return err
}

func resetFailedLogin(username string) error {
	_, err := db.Exec("UPDATE users SET failed_login_attempts = 0, locked_until = NULL WHERE username = ?", username)
	return err
}

func AuthenticateUser(username, password string) (*User, error) {
	locked, err := isAccountLocked(username)
	if err != nil {
		return nil, errors.New("authentication error")
	}
	if locked {
		LogEvent("auth_fail", "Login to locked account", username, "warning")
		return nil, errors.New("account temporarily locked")
	}
	var user User
	var lastLogin sql.NullTime
	err = db.QueryRow("SELECT id, username, password_hash, role, is_active, last_login FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.IsActive, &lastLogin)
	if err != nil {
		LogEvent("auth_fail", "User not found", username, "warning")
		return nil, errors.New("invalid credentials")
	}
	if !user.IsActive {
		LogEvent("auth_fail", "Account disabled", username, "warning")
		return nil, errors.New("account disabled")
	}
	if user.Role == "technician" && !IsTechnicianAccessAllowed(db, user.Username) {
		LogEvent("auth_fail", "Technician access expired or not granted", user.Username, "warning")
		return nil, errors.New("technician access expired or not granted")
	}
	if !CheckPassword(user.PasswordHash, password) {
		incrementFailedLogin(username)
		LogEvent("auth_fail", "Invalid password", username, "warning")
		return nil, errors.New("invalid credentials")
	}
	resetFailedLogin(username)
	user.SessionToken = GenerateSessionToken()
	db.Exec("UPDATE users SET last_login = ?, session_token = ? WHERE username = ?", time.Now(), user.SessionToken, username)
	if lastLogin.Valid {
		user.LastLogin = lastLogin.Time
	}
	LogEvent("auth_success", "Login successful", username, "info")
	return &user, nil
}

func VerifySession(token string) (*User, error) {
	if token == "" {
		return nil, errors.New("no session token")
	}
	var user User
	err := db.QueryRow("SELECT id, username, role, is_active FROM users WHERE session_token = ? AND is_active = 1", token).Scan(&user.ID, &user.Username, &user.Role, &user.IsActive)
	if err != nil {
		return nil, errors.New("invalid session")
	}
	user.SessionToken = token
	return &user, nil
}

func LogoutUser(username string) error {
	_, err := db.Exec("UPDATE users SET session_token = NULL WHERE username = ?", username)
	if err != nil {
		return err
	}
	LogEvent("logout", "User logged out", username, "info")
	return nil
}

func GenerateSessionToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic("unable to generate session token")
	}
	return base64.URLEncoding.EncodeToString(b)
}

func GetUserByUsername(username string) (*User, error) {
	var user User
	err := db.QueryRow("SELECT id, username, role, is_active FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Role, &user.IsActive)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}
