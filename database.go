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
		session_expires_at DATETIME,
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

	createGeofenceConfigTable := `CREATE TABLE IF NOT EXISTS geofence_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		owner TEXT UNIQUE NOT NULL,
		home_lat REAL NOT NULL CHECK(home_lat >= -90 AND home_lat <= 90),
		home_long REAL NOT NULL CHECK(home_long >= -180 AND home_long <= 180),
		radius_meters REAL NOT NULL CHECK(radius_meters > 0 AND radius_meters <= 10000),
		away_temp REAL NOT NULL CHECK(away_temp >= 10 AND away_temp <= 35),
		home_temp REAL NOT NULL CHECK(home_temp >= 10 AND home_temp <= 35),
		enabled INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_updated DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	createLocationLogsTable := `CREATE TABLE IF NOT EXISTS location_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		latitude REAL NOT NULL CHECK(latitude >= -90 AND latitude <= 90),
		longitude REAL NOT NULL CHECK(longitude >= -180 AND longitude <= 180),
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	createPresenceLogsTable := `CREATE TABLE IF NOT EXISTS presence_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		owner TEXT NOT NULL,
		is_home INTEGER NOT NULL,
		distance_meters REAL NOT NULL,
		latitude REAL NOT NULL,
		longitude REAL NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	tables := []string{
		createUsersTable, createLogsTable, createProfilesTable,
		createSchedulesTable, createEnergyTable, createGuestAccessTable,
		createSensorTable, createHVACStateTable, createGeofenceConfigTable,
		createLocationLogsTable, createPresenceLogsTable,
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
		"CREATE INDEX IF NOT EXISTS idx_location_timestamp ON location_logs(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_presence_timestamp ON presence_logs(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_presence_owner ON presence_logs(owner)",
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

func CleanExpiredSessions() error {
	result, err := db.Exec("UPDATE users SET session_token = NULL, session_expires_at = NULL WHERE session_expires_at < ?", time.Now())
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows > 0 {
		LogEvent("session_cleanup", fmt.Sprintf("Cleaned up %d expired sessions", rows), "system", "info")
	}
	return nil
}
