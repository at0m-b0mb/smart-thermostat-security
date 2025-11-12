package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type MaintenanceRecord struct {
	ID                    int
	FilterInstallDate     time.Time
	FilterRuntimeHours    float64
	FilterChangeInterval  float64 // Hours until filter needs changing
	LastMaintenanceDate   time.Time
	NextMaintenanceDate   time.Time
	MaintenanceAlertSent  bool
}

// InitializeMaintenanceTable creates the maintenance table if it doesn't exist
func InitializeMaintenanceTable() error {
	createMaintenanceTable := `CREATE TABLE IF NOT EXISTS maintenance (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		filter_install_date DATETIME DEFAULT CURRENT_TIMESTAMP,
		filter_runtime_hours REAL DEFAULT 0,
		filter_change_interval REAL DEFAULT 720.0,
		last_maintenance_date DATETIME,
		next_maintenance_date DATETIME,
		maintenance_alert_sent INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(createMaintenanceTable)
	if err != nil {
		return fmt.Errorf("failed to create maintenance table: %w", err)
	}

	// Initialize with default record if none exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM maintenance").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check maintenance records: %w", err)
	}

	if count == 0 {
		_, err = db.Exec(`
			INSERT INTO maintenance (filter_install_date, filter_runtime_hours, filter_change_interval, last_maintenance_date, next_maintenance_date)
			VALUES (?, 0, 720.0, ?, ?)`,
			time.Now(), time.Now(), time.Now().AddDate(0, 0, 30))
		if err != nil {
			return fmt.Errorf("failed to initialize maintenance record: %w", err)
		}
		LogEvent("maintenance_init", "Maintenance tracking initialized", "system", "info")
	}

	return nil
}

// UpdateFilterRuntime adds runtime hours to the filter usage counter
func UpdateFilterRuntime(additionalHours float64) error {
	if additionalHours < 0 {
		return errors.New("runtime hours cannot be negative")
	}

	_, err := db.Exec(`
		UPDATE maintenance 
		SET filter_runtime_hours = filter_runtime_hours + ?,
		    updated_at = ?
		WHERE id = 1`,
		additionalHours, time.Now())
	
	if err != nil {
		return fmt.Errorf("failed to update filter runtime: %w", err)
	}

	// Check if maintenance is due
	checkMaintenanceDue()

	return nil
}

// ResetFilter resets the filter tracking after a filter change
func ResetFilter(user *User) error {
	// Only homeowners and technicians can reset filter
	if user.Role != "homeowner" && user.Role != "technician" {
		return errors.New("insufficient permissions to reset filter")
	}

	now := time.Now()
	nextMaintenance := now.AddDate(0, 0, 30) // 30 days from now

	_, err := db.Exec(`
		UPDATE maintenance 
		SET filter_install_date = ?,
		    filter_runtime_hours = 0,
		    last_maintenance_date = ?,
		    next_maintenance_date = ?,
		    maintenance_alert_sent = 0,
		    updated_at = ?
		WHERE id = 1`,
		now, now, nextMaintenance, now)
	
	if err != nil {
		return fmt.Errorf("failed to reset filter: %w", err)
	}

	LogEvent("filter_reset", "Filter replaced and tracking reset", user.Username, "info")
	SendNotification(user.Username, "maintenance", "Filter maintenance recorded. Next maintenance due: "+nextMaintenance.Format("2006-01-02"))

	return nil
}

// GetMaintenanceStatus retrieves the current maintenance status
func GetMaintenanceStatus() (*MaintenanceRecord, error) {
	var record MaintenanceRecord
	err := db.QueryRow(`
		SELECT id, filter_install_date, filter_runtime_hours, filter_change_interval,
		       last_maintenance_date, next_maintenance_date, maintenance_alert_sent
		FROM maintenance WHERE id = 1`).Scan(
		&record.ID, &record.FilterInstallDate, &record.FilterRuntimeHours,
		&record.FilterChangeInterval, &record.LastMaintenanceDate,
		&record.NextMaintenanceDate, &record.MaintenanceAlertSent)
	
	if err == sql.ErrNoRows {
		return nil, errors.New("no maintenance record found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get maintenance status: %w", err)
	}

	return &record, nil
}

// SetFilterChangeInterval sets a custom filter change interval in hours
func SetFilterChangeInterval(hours float64, user *User) error {
	// Only homeowners can change the interval
	if user.Role != "homeowner" {
		return errors.New("only homeowners can set filter change interval")
	}

	if hours < 100 || hours > 2000 {
		return errors.New("filter change interval must be between 100 and 2000 hours")
	}

	_, err := db.Exec(`
		UPDATE maintenance 
		SET filter_change_interval = ?,
		    updated_at = ?
		WHERE id = 1`,
		hours, time.Now())
	
	if err != nil {
		return fmt.Errorf("failed to set filter change interval: %w", err)
	}

	LogEvent("filter_interval_set", fmt.Sprintf("Filter change interval set to %.0f hours", hours), user.Username, "info")

	return nil
}

// checkMaintenanceDue checks if maintenance is due and sends alerts
func checkMaintenanceDue() error {
	var record MaintenanceRecord
	err := db.QueryRow(`
		SELECT id, filter_runtime_hours, filter_change_interval, maintenance_alert_sent
		FROM maintenance WHERE id = 1`).Scan(
		&record.ID, &record.FilterRuntimeHours, &record.FilterChangeInterval, &record.MaintenanceAlertSent)
	
	if err != nil {
		return err
	}

	// Check if filter needs replacement
	hoursRemaining := record.FilterChangeInterval - record.FilterRuntimeHours
	
	// Send alert if less than 50 hours remaining and alert not sent
	if hoursRemaining < 50 && !record.MaintenanceAlertSent {
		// Get homeowner username for notification
		var homeowner string
		err = db.QueryRow("SELECT username FROM users WHERE role = 'homeowner' LIMIT 1").Scan(&homeowner)
		if err == nil {
			SendMaintenanceAlert(homeowner, hoursRemaining)
			// Mark alert as sent
			db.Exec("UPDATE maintenance SET maintenance_alert_sent = 1 WHERE id = 1")
		}
	}

	// Send critical alert if overdue
	if hoursRemaining <= 0 && !record.MaintenanceAlertSent {
		var homeowner string
		err = db.QueryRow("SELECT username FROM users WHERE role = 'homeowner' LIMIT 1").Scan(&homeowner)
		if err == nil {
			SendMaintenanceCriticalAlert(homeowner, -hoursRemaining)
			db.Exec("UPDATE maintenance SET maintenance_alert_sent = 1 WHERE id = 1")
		}
	}

	return nil
}

// SendMaintenanceAlert sends a filter maintenance reminder
func SendMaintenanceAlert(username string, hoursRemaining float64) error {
	message := fmt.Sprintf("Filter maintenance due soon! Approximately %.0f hours of runtime remaining.", hoursRemaining)
	LogEvent("maintenance_alert", message, username, "warning")
	return SendNotification(username, "maintenance_reminder", message)
}

// SendMaintenanceCriticalAlert sends a critical filter replacement alert
func SendMaintenanceCriticalAlert(username string, hoursOverdue float64) error {
	message := fmt.Sprintf("CRITICAL: Filter replacement overdue by %.0f hours! Replace immediately.", hoursOverdue)
	LogEvent("maintenance_critical", message, username, "critical")
	return SendNotification(username, "maintenance_critical", message)
}

// DisplayMaintenanceStatus formats maintenance status for display
func DisplayMaintenanceStatus(record *MaintenanceRecord) string {
	if record == nil {
		return "Maintenance: No data available"
	}

	hoursRemaining := record.FilterChangeInterval - record.FilterRuntimeHours
	percentUsed := (record.FilterRuntimeHours / record.FilterChangeInterval) * 100

	status := "OK"
	if hoursRemaining < 50 {
		status = "WARNING - Replacement Soon"
	}
	if hoursRemaining <= 0 {
		status = "CRITICAL - Overdue!"
	}

	daysSinceInstall := int(time.Since(record.FilterInstallDate).Hours() / 24)
	daysUntilNext := int(time.Until(record.NextMaintenanceDate).Hours() / 24)

	return fmt.Sprintf(`Filter Maintenance Status
====================================
Status: %s
Filter Install Date: %s
Days Since Install: %d
Runtime Hours: %.1f / %.0f hours
Filter Life Used: %.1f%%
Hours Remaining: %.1f
Last Maintenance: %s
Next Maintenance Due: %s (in %d days)`,
		status,
		record.FilterInstallDate.Format("2006-01-02"),
		daysSinceInstall,
		record.FilterRuntimeHours, record.FilterChangeInterval,
		percentUsed,
		hoursRemaining,
		record.LastMaintenanceDate.Format("2006-01-02"),
		record.NextMaintenanceDate.Format("2006-01-02"),
		daysUntilNext)
}

// CheckAndUpdateMaintenance is called periodically to update maintenance tracking
func CheckAndUpdateMaintenance() error {
	// Get current HVAC state
	hvacMutex.RLock()
	isRunning := hvacState.IsRunning
	hvacMutex.RUnlock()

	// If HVAC is running, add runtime (called every 30 seconds, so 0.0083 hours)
	if isRunning {
		err := UpdateFilterRuntime(0.0083)
		if err != nil {
			return err
		}
	}

	return checkMaintenanceDue()
}
