package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type AwayMode struct {
	ID               int
	IsActive         bool
	StartTime        time.Time
	ReturnTime       time.Time
	AwayTemp         float64
	OriginalMode     string
	OriginalTemp     float64
	SetBy            string
	CreatedAt        time.Time
}

// InitializeAwayModeTable creates the away_mode table if it doesn't exist
func InitializeAwayModeTable() error {
	createAwayModeTable := `CREATE TABLE IF NOT EXISTS away_mode (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		is_active INTEGER DEFAULT 0,
		start_time DATETIME,
		return_time DATETIME,
		away_temp REAL CHECK(away_temp >= 10 AND away_temp <= 35),
		original_mode TEXT,
		original_temp REAL,
		set_by TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(createAwayModeTable)
	if err != nil {
		return fmt.Errorf("failed to create away_mode table: %w", err)
	}

	// Create index for faster queries
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_away_mode_active ON away_mode(is_active)")
	if err != nil {
		return fmt.Errorf("failed to create away_mode index: %w", err)
	}

	return nil
}

// SetAwayMode activates away mode with energy-saving settings
func SetAwayMode(returnTime time.Time, awayTemp float64, user *User) error {
	// Only homeowners can set away mode
	if user.Role != "homeowner" {
		return errors.New("only homeowners can set away mode")
	}

	// Validate away temperature
	if awayTemp < 10 || awayTemp > 35 {
		return errors.New("away temperature out of range (10-35°C)")
	}

	// Validate return time is in the future
	if returnTime.Before(time.Now()) {
		return errors.New("return time must be in the future")
	}

	// Check if away mode is already active
	var activeCount int
	err := db.QueryRow("SELECT COUNT(*) FROM away_mode WHERE is_active = 1").Scan(&activeCount)
	if err != nil {
		return fmt.Errorf("failed to check away mode status: %w", err)
	}
	if activeCount > 0 {
		return errors.New("away mode is already active, deactivate it first")
	}

	// Get current HVAC state
	hvacMutex.RLock()
	originalMode := string(hvacState.Mode)
	originalTemp := hvacState.TargetTemp
	hvacMutex.RUnlock()

	// Insert new away mode record
	_, err = db.Exec(`
		INSERT INTO away_mode (is_active, start_time, return_time, away_temp, original_mode, original_temp, set_by)
		VALUES (1, ?, ?, ?, ?, ?, ?)`,
		time.Now(), returnTime, awayTemp, originalMode, originalTemp, user.Username)
	if err != nil {
		return fmt.Errorf("failed to activate away mode: %w", err)
	}

	// Set HVAC to away temperature
	err = SetTargetTemperature(awayTemp, user)
	if err != nil {
		return fmt.Errorf("failed to set away temperature: %w", err)
	}

	LogEvent("away_mode_set", fmt.Sprintf("Away mode activated until %s with temp %.1f°C", 
		returnTime.Format("2006-01-02 15:04"), awayTemp), user.Username, "info")
	
	SendNotification(user.Username, "away_mode", 
		fmt.Sprintf("Away mode activated. Return time: %s", returnTime.Format("2006-01-02 15:04")))

	return nil
}

// DeactivateAwayMode restores previous settings and deactivates away mode
func DeactivateAwayMode(user *User) error {
	// Only homeowners can deactivate away mode
	if user.Role != "homeowner" {
		return errors.New("only homeowners can deactivate away mode")
	}

	// Get active away mode record
	var awayMode AwayMode
	err := db.QueryRow(`
		SELECT id, original_mode, original_temp, set_by, return_time
		FROM away_mode WHERE is_active = 1
		LIMIT 1`).Scan(&awayMode.ID, &awayMode.OriginalMode, &awayMode.OriginalTemp, 
			&awayMode.SetBy, &awayMode.ReturnTime)
	
	if err == sql.ErrNoRows {
		return errors.New("no active away mode found")
	}
	if err != nil {
		return fmt.Errorf("failed to retrieve away mode: %w", err)
	}

	// Restore original settings
	err = SetTargetTemperature(awayMode.OriginalTemp, user)
	if err != nil {
		return fmt.Errorf("failed to restore temperature: %w", err)
	}

	err = SetHVACMode(awayMode.OriginalMode, user)
	if err != nil {
		return fmt.Errorf("failed to restore HVAC mode: %w", err)
	}

	// Deactivate away mode
	_, err = db.Exec("UPDATE away_mode SET is_active = 0 WHERE id = ?", awayMode.ID)
	if err != nil {
		return fmt.Errorf("failed to deactivate away mode: %w", err)
	}

	LogEvent("away_mode_deactivate", "Away mode deactivated, settings restored", user.Username, "info")
	SendNotification(user.Username, "away_mode", "Welcome back! Previous settings restored.")

	return nil
}

// CheckAwayModeReturn checks if it's time to return from away mode and auto-deactivates
func CheckAwayModeReturn() error {
	var awayMode AwayMode
	err := db.QueryRow(`
		SELECT id, return_time, original_mode, original_temp, set_by
		FROM away_mode WHERE is_active = 1
		LIMIT 1`).Scan(&awayMode.ID, &awayMode.ReturnTime, 
			&awayMode.OriginalMode, &awayMode.OriginalTemp, &awayMode.SetBy)
	
	if err == sql.ErrNoRows {
		return nil // No active away mode
	}
	if err != nil {
		return fmt.Errorf("failed to check away mode: %w", err)
	}

	// Check if return time has passed
	if time.Now().After(awayMode.ReturnTime) {
		// Create a temporary user for system operations
		systemUser := &User{Username: awayMode.SetBy, Role: "homeowner"}
		
		// Restore settings
		SetTargetTemperature(awayMode.OriginalTemp, systemUser)
		SetHVACMode(awayMode.OriginalMode, systemUser)
		
		// Deactivate away mode
		_, err = db.Exec("UPDATE away_mode SET is_active = 0 WHERE id = ?", awayMode.ID)
		if err != nil {
			return fmt.Errorf("failed to auto-deactivate away mode: %w", err)
		}

		LogEvent("away_mode_auto_return", "Away mode auto-deactivated at return time", awayMode.SetBy, "info")
		SendNotification(awayMode.SetBy, "away_mode", "Welcome back! Away mode has been automatically deactivated.")
	}

	return nil
}

// GetAwayModeStatus returns the current away mode status
func GetAwayModeStatus() (*AwayMode, error) {
	var awayMode AwayMode
	err := db.QueryRow(`
		SELECT id, is_active, start_time, return_time, away_temp, original_mode, original_temp, set_by, created_at
		FROM away_mode WHERE is_active = 1
		LIMIT 1`).Scan(&awayMode.ID, &awayMode.IsActive, &awayMode.StartTime, 
			&awayMode.ReturnTime, &awayMode.AwayTemp, &awayMode.OriginalMode, 
			&awayMode.OriginalTemp, &awayMode.SetBy, &awayMode.CreatedAt)
	
	if err == sql.ErrNoRows {
		return nil, nil // No active away mode
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get away mode status: %w", err)
	}

	return &awayMode, nil
}

// DisplayAwayModeStatus formats away mode status for display
func DisplayAwayModeStatus(awayMode *AwayMode) string {
	if awayMode == nil {
		return "Away Mode: Inactive"
	}

	duration := time.Until(awayMode.ReturnTime)
	hoursUntilReturn := int(duration.Hours())
	minutesUntilReturn := int(duration.Minutes()) % 60

	return fmt.Sprintf(`Away Mode: ACTIVE
Start Time: %s
Return Time: %s
Time Until Return: %dh %dm
Away Temperature: %.1f°C
Original Mode: %s
Original Temperature: %.1f°C
Set By: %s`,
		awayMode.StartTime.Format("2006-01-02 15:04"),
		awayMode.ReturnTime.Format("2006-01-02 15:04"),
		hoursUntilReturn, minutesUntilReturn,
		awayMode.AwayTemp,
		awayMode.OriginalMode,
		awayMode.OriginalTemp,
		awayMode.SetBy)
}
