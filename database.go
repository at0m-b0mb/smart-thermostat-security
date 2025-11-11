package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

var db *sql.DB

func InitializeDatabase() error {
	var err error
	db, err = sql.Open("sqlite3", "./thermostat.db?_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	createUsersTable := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL CHECK(role IN ('homeowner', 'technician', 'guest')),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_login DATETIME,
		session_token TEXT,
		is_active INTEGER DEFAULT 1,
		failed_login_attempts INTEGER DEFAULT 0,
		locked_until DATETIME
	);`

	createLogsTable := `CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		event_type TEXT NOT NULL,
		details TEXT,
		username TEXT,
		severity TEXT DEFAULT 'info'
	);`

	createProfilesTable := `CREATE TABLE IF NOT EXISTS profiles (
	    id INTEGER PRIMARY KEY AUTOINCREMENT,
	    profile_name TEXT UNIQUE NOT NULL,
	    target_temp REAL NOT NULL CHECK(target_temp >= 10 AND target_temp <= 35),
	    hvac_mode TEXT NOT NULL CHECK(hvac_mode IN ('off', 'heat', 'cool', 'fan')),
	    owner TEXT NOT NULL,
	    guest_accessible INTEGER DEFAULT 0, 
	    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	createSchedulesTable := `CREATE TABLE IF NOT EXISTS schedules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		profile_id INTEGER NOT NULL,
		day_of_week INTEGER NOT NULL CHECK(day_of_week >= 0 AND day_of_week <= 6),
		start_time TEXT NOT NULL,
		end_time TEXT NOT NULL,
		target_temp REAL NOT NULL CHECK(target_temp >= 10 AND target_temp <= 35),
		FOREIGN KEY(profile_id) REFERENCES profiles(id) ON DELETE CASCADE
	);`

	createEnergyTable := `CREATE TABLE IF NOT EXISTS energy_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		hvac_mode TEXT NOT NULL,
		runtime_minutes INTEGER NOT NULL CHECK(runtime_minutes >= 0),
		estimated_kwh REAL NOT NULL CHECK(estimated_kwh >= 0)
	);`

	createGuestAccessTable := `CREATE TABLE IF NOT EXISTS guest_access (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guest_username TEXT NOT NULL,
		granted_by TEXT NOT NULL,
		granted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME,
		is_active INTEGER DEFAULT 1
	);`

	createSensorTable := `CREATE TABLE IF NOT EXISTS sensor_readings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		temperature REAL NOT NULL,
		humidity REAL,
		co_level REAL,
		sensor_status TEXT DEFAULT 'healthy'
	);`

	createHVACStateTable := `CREATE TABLE IF NOT EXISTS hvac_state (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		mode TEXT NOT NULL,
		target_temp REAL,
		current_temp REAL,
		is_running INTEGER DEFAULT 0
	);`

	tables := []string{
		createUsersTable, createLogsTable, createProfilesTable,
		createSchedulesTable, createEnergyTable, createGuestAccessTable,
		createSensorTable, createHVACStateTable,
	}

	for _, table := range tables {
		if _, err = db.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	indices := []string{
		"CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_users_session ON users(session_token)",
		"CREATE INDEX IF NOT EXISTS idx_energy_timestamp ON energy_logs(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_sensor_timestamp ON sensor_readings(timestamp)",
	}

	for _, index := range indices {
		if _, err = db.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	if err = createDefaultUser(); err != nil {
		return err
	}

	LogEvent("system", "Database initialized", "system", "info")
	return nil
}

func createDefaultUser() error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE role='homeowner'").Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		err = RegisterUser("admin", "Admin123!", "homeowner")
		if err != nil {
			return err
		}
		LogEvent("system", "Default admin created (username: admin, password: Admin123!)", "admin", "info")
	}
	return nil
}

func CloseDatabase() {
	if db != nil {
		db.Close()
	}
}

func CleanOldLogs(daysToKeep int) error {
	cutoffDate := time.Now().AddDate(0, 0, -daysToKeep)
	_, err := db.Exec("DELETE FROM logs WHERE timestamp < ?", cutoffDate)
	return err
}
