package main

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// GeofenceStatus represents user presence relative to home
type GeofenceStatus string

const (
	StatusHome      GeofenceStatus = "home"
	StatusNearby    GeofenceStatus = "nearby"     // Within 5km
	StatusAway      GeofenceStatus = "away"       // Beyond 5km
	StatusUnknown   GeofenceStatus = "unknown"
)

// Location represents GPS coordinates (simulated)
type Location struct {
	Latitude  float64
	Longitude float64
	Timestamp time.Time
}

// GeofenceConfig stores geofencing settings
type GeofenceConfig struct {
	ID                    int
	IsEnabled             bool
	HomeLatitude          float64
	HomeLongitude         float64
	GeofenceRadius        float64 // in kilometers
	HomeTemp              float64
	AwayTemp              float64
	ComingHomeTemp        float64 // Pre-heating/cooling temp
	AutoAdjustEnabled     bool
	Owner                 string
	LastLocationUpdate    time.Time
	CurrentStatus         GeofenceStatus
	SimulatedLatitude     float64
	SimulatedLongitude    float64
}

// PresenceEvent tracks presence changes for analytics
type PresenceEvent struct {
	ID             int
	Username       string
	EventType      string // "arrived_home", "left_home", "approaching_home"
	PreviousStatus GeofenceStatus
	NewStatus      GeofenceStatus
	Distance       float64
	Timestamp      time.Time
}

// InitializeGeofencingTable creates necessary tables for geofencing
func InitializeGeofencingTable() error {
	createGeofenceTable := `CREATE TABLE IF NOT EXISTS geofence_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		is_enabled INTEGER DEFAULT 0,
		home_latitude REAL NOT NULL,
		home_longitude REAL NOT NULL,
		geofence_radius REAL DEFAULT 5.0 CHECK(geofence_radius > 0 AND geofence_radius <= 50),
		home_temp REAL DEFAULT 22.0 CHECK(home_temp >= 10 AND home_temp <= 35),
		away_temp REAL DEFAULT 18.0 CHECK(away_temp >= 10 AND away_temp <= 35),
		coming_home_temp REAL DEFAULT 21.0 CHECK(coming_home_temp >= 10 AND coming_home_temp <= 35),
		auto_adjust_enabled INTEGER DEFAULT 1,
		owner TEXT NOT NULL,
		last_location_update DATETIME DEFAULT CURRENT_TIMESTAMP,
		current_status TEXT DEFAULT 'unknown',
		simulated_latitude REAL,
		simulated_longitude REAL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	createPresenceEventsTable := `CREATE TABLE IF NOT EXISTS presence_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		event_type TEXT NOT NULL,
		previous_status TEXT,
		new_status TEXT NOT NULL,
		distance REAL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(createGeofenceTable)
	if err != nil {
		return fmt.Errorf("failed to create geofence_config table: %w", err)
	}

	_, err = db.Exec(createPresenceEventsTable)
	if err != nil {
		return fmt.Errorf("failed to create presence_events table: %w", err)
	}

	// Create indices for faster queries
	indices := []string{
		"CREATE INDEX IF NOT EXISTS idx_geofence_enabled ON geofence_config(is_enabled)",
		"CREATE INDEX IF NOT EXISTS idx_presence_events_timestamp ON presence_events(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_presence_events_username ON presence_events(username)",
	}

	for _, index := range indices {
		if _, err = db.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Initialize default config if none exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM geofence_config").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check geofence config: %w", err)
	}

	if count == 0 {
		// Default home location: Johns Hopkins University (for simulation)
		_, err = db.Exec(`
			INSERT INTO geofence_config (is_enabled, home_latitude, home_longitude, geofence_radius, 
				home_temp, away_temp, coming_home_temp, auto_adjust_enabled, owner, current_status,
				simulated_latitude, simulated_longitude)
			VALUES (0, 39.3299, -76.6205, 5.0, 22.0, 18.0, 21.0, 1, 'admin', 'home', 39.3299, -76.6205)`)
		if err != nil {
			return fmt.Errorf("failed to initialize geofence config: %w", err)
		}
		LogEvent("geofence_init", "Geofencing system initialized with default settings", "system", "info")
	}

	return nil
}

// EnableGeofencing enables the geofencing feature
func EnableGeofencing(user *User) error {
	if user.Role != "homeowner" {
		return errors.New("only homeowners can enable geofencing")
	}

	_, err := db.Exec(`
		UPDATE geofence_config 
		SET is_enabled = 1, updated_at = ?
		WHERE id = 1`, time.Now())
	
	if err != nil {
		return fmt.Errorf("failed to enable geofencing: %w", err)
	}

	LogEvent("geofence_enable", "Geofencing enabled", user.Username, "info")
	SendNotification(user.Username, "geofence", "Geofencing enabled. Temperature will auto-adjust based on your location.")

	return nil
}

// DisableGeofencing disables the geofencing feature
func DisableGeofencing(user *User) error {
	if user.Role != "homeowner" {
		return errors.New("only homeowners can disable geofencing")
	}

	_, err := db.Exec(`
		UPDATE geofence_config 
		SET is_enabled = 0, updated_at = ?
		WHERE id = 1`, time.Now())
	
	if err != nil {
		return fmt.Errorf("failed to disable geofencing: %w", err)
	}

	LogEvent("geofence_disable", "Geofencing disabled", user.Username, "info")
	SendNotification(user.Username, "geofence", "Geofencing disabled. Manual temperature control resumed.")

	return nil
}

// SetHomeLocation sets the home location coordinates
func SetHomeLocation(latitude, longitude float64, user *User) error {
	if user.Role != "homeowner" {
		return errors.New("only homeowners can set home location")
	}

	// Validate coordinates
	if latitude < -90 || latitude > 90 {
		return errors.New("invalid latitude (must be between -90 and 90)")
	}
	if longitude < -180 || longitude > 180 {
		return errors.New("invalid longitude (must be between -180 and 180)")
	}

	_, err := db.Exec(`
		UPDATE geofence_config 
		SET home_latitude = ?, home_longitude = ?, 
		    simulated_latitude = ?, simulated_longitude = ?,
		    updated_at = ?
		WHERE id = 1`, latitude, longitude, latitude, longitude, time.Now())
	
	if err != nil {
		return fmt.Errorf("failed to set home location: %w", err)
	}

	LogEvent("geofence_location_set", fmt.Sprintf("Home location set to %.4f, %.4f", latitude, longitude), user.Username, "info")

	return nil
}

// SetGeofenceTemperatures sets the temperature preferences for different zones
func SetGeofenceTemperatures(homeTemp, awayTemp, comingHomeTemp float64, user *User) error {
	if user.Role != "homeowner" {
		return errors.New("only homeowners can set geofence temperatures")
	}

	// Validate temperatures
	temps := []float64{homeTemp, awayTemp, comingHomeTemp}
	for _, temp := range temps {
		if temp < 10 || temp > 35 {
			return errors.New("temperature out of range (10-35°C)")
		}
	}

	_, err := db.Exec(`
		UPDATE geofence_config 
		SET home_temp = ?, away_temp = ?, coming_home_temp = ?, updated_at = ?
		WHERE id = 1`, homeTemp, awayTemp, comingHomeTemp, time.Now())
	
	if err != nil {
		return fmt.Errorf("failed to set geofence temperatures: %w", err)
	}

	LogEvent("geofence_temps_set", fmt.Sprintf("Geofence temps: Home=%.1f°C, Away=%.1f°C, Coming=%.1f°C", 
		homeTemp, awayTemp, comingHomeTemp), user.Username, "info")

	return nil
}

// SetGeofenceRadius sets the geofence radius in kilometers
func SetGeofenceRadius(radius float64, user *User) error {
	if user.Role != "homeowner" {
		return errors.New("only homeowners can set geofence radius")
	}

	if radius <= 0 || radius > 50 {
		return errors.New("geofence radius must be between 0 and 50 km")
	}

	_, err := db.Exec(`
		UPDATE geofence_config 
		SET geofence_radius = ?, updated_at = ?
		WHERE id = 1`, radius, time.Now())
	
	if err != nil {
		return fmt.Errorf("failed to set geofence radius: %w", err)
	}

	LogEvent("geofence_radius_set", fmt.Sprintf("Geofence radius set to %.1f km", radius), user.Username, "info")

	return nil
}

// SimulateLocationUpdate simulates a location update (for demonstration)
func SimulateLocationUpdate(latitude, longitude float64, user *User) error {
	if user.Role != "homeowner" {
		return errors.New("only homeowners can simulate location")
	}

	// Validate coordinates
	if latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 {
		return errors.New("invalid coordinates")
	}

	_, err := db.Exec(`
		UPDATE geofence_config 
		SET simulated_latitude = ?, simulated_longitude = ?, 
		    last_location_update = ?, updated_at = ?
		WHERE id = 1`, latitude, longitude, time.Now(), time.Now())
	
	if err != nil {
		return fmt.Errorf("failed to update simulated location: %w", err)
	}

	// Trigger geofence check
	CheckGeofenceStatus()

	return nil
}

// CalculateDistance calculates distance between two points using Haversine formula (in km)
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371.0 // Earth's radius in kilometers

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// CheckGeofenceStatus checks current location and adjusts temperature accordingly
func CheckGeofenceStatus() error {
	var config GeofenceConfig
	err := db.QueryRow(`
		SELECT id, is_enabled, home_latitude, home_longitude, geofence_radius,
		       home_temp, away_temp, coming_home_temp, auto_adjust_enabled, owner,
		       current_status, simulated_latitude, simulated_longitude
		FROM geofence_config WHERE id = 1`).Scan(
		&config.ID, &config.IsEnabled, &config.HomeLatitude, &config.HomeLongitude,
		&config.GeofenceRadius, &config.HomeTemp, &config.AwayTemp, &config.ComingHomeTemp,
		&config.AutoAdjustEnabled, &config.Owner, &config.CurrentStatus,
		&config.SimulatedLatitude, &config.SimulatedLongitude)
	
	if err == sql.ErrNoRows {
		return nil // No config, nothing to do
	}
	if err != nil {
		return fmt.Errorf("failed to get geofence config: %w", err)
	}

	// Skip if geofencing is disabled
	if !config.IsEnabled {
		return nil
	}

	// Calculate distance from home
	distance := CalculateDistance(
		config.HomeLatitude, config.HomeLongitude,
		config.SimulatedLatitude, config.SimulatedLongitude)

	// Determine new status
	var newStatus GeofenceStatus
	if distance <= 0.1 { // Within 100 meters
		newStatus = StatusHome
	} else if distance <= config.GeofenceRadius {
		newStatus = StatusNearby
	} else {
		newStatus = StatusAway
	}

	previousStatus := GeofenceStatus(config.CurrentStatus)

	// Handle status change
	if newStatus != previousStatus && config.AutoAdjustEnabled {
		handlePresenceChange(config, previousStatus, newStatus, distance)
		
		// Update current status
		_, err = db.Exec(`
			UPDATE geofence_config 
			SET current_status = ?, updated_at = ?
			WHERE id = 1`, string(newStatus), time.Now())
		
		if err != nil {
			return fmt.Errorf("failed to update geofence status: %w", err)
		}
	}

	return nil
}

// handlePresenceChange adjusts temperature based on presence change
func handlePresenceChange(config GeofenceConfig, previousStatus, newStatus GeofenceStatus, distance float64) {
	// Create system user for automated actions
	systemUser := &User{Username: config.Owner, Role: "homeowner"}

	var eventType string
	var targetTemp float64

	switch newStatus {
	case StatusHome:
		eventType = "arrived_home"
		targetTemp = config.HomeTemp
		SendNotification(config.Owner, "geofence", "Welcome home! Setting temperature to home comfort level.")
		
	case StatusNearby:
		if previousStatus == StatusAway {
			eventType = "approaching_home"
			targetTemp = config.ComingHomeTemp
			SendNotification(config.Owner, "geofence", 
				fmt.Sprintf("You're nearby (%.1f km away). Pre-conditioning to %.1f°C.", distance, targetTemp))
		} else if previousStatus == StatusHome {
			eventType = "left_home_nearby"
			targetTemp = config.AwayTemp
			SendNotification(config.Owner, "geofence", "You've left home. Switching to away mode.")
		}
		
	case StatusAway:
		eventType = "left_home"
		targetTemp = config.AwayTemp
		SendNotification(config.Owner, "geofence", 
			fmt.Sprintf("You're away (%.1f km from home). Energy-saving mode activated.", distance))
	}

	// Adjust temperature if we have a target
	if targetTemp > 0 {
		err := SetTargetTemperature(targetTemp, systemUser)
		if err != nil {
			LogEvent("geofence_error", "Failed to adjust temperature: "+err.Error(), config.Owner, "warning")
		} else {
			LogEvent("geofence_auto_adjust", 
				fmt.Sprintf("Temperature auto-adjusted to %.1f°C (%s)", targetTemp, eventType), 
				config.Owner, "info")
		}
	}

	// Record presence event
	_, err := db.Exec(`
		INSERT INTO presence_events (username, event_type, previous_status, new_status, distance)
		VALUES (?, ?, ?, ?, ?)`,
		config.Owner, eventType, string(previousStatus), string(newStatus), distance)
	
	if err != nil {
		LogEvent("geofence_error", "Failed to record presence event: "+err.Error(), config.Owner, "warning")
	}
}

// GetGeofenceConfig retrieves current geofencing configuration
func GetGeofenceConfig() (*GeofenceConfig, error) {
	var config GeofenceConfig
	err := db.QueryRow(`
		SELECT id, is_enabled, home_latitude, home_longitude, geofence_radius,
		       home_temp, away_temp, coming_home_temp, auto_adjust_enabled, owner,
		       last_location_update, current_status, simulated_latitude, simulated_longitude
		FROM geofence_config WHERE id = 1`).Scan(
		&config.ID, &config.IsEnabled, &config.HomeLatitude, &config.HomeLongitude,
		&config.GeofenceRadius, &config.HomeTemp, &config.AwayTemp, &config.ComingHomeTemp,
		&config.AutoAdjustEnabled, &config.Owner, &config.LastLocationUpdate,
		&config.CurrentStatus, &config.SimulatedLatitude, &config.SimulatedLongitude)
	
	if err == sql.ErrNoRows {
		return nil, errors.New("geofence config not initialized")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get geofence config: %w", err)
	}

	return &config, nil
}

// GetPresenceHistory retrieves recent presence events
func GetPresenceHistory(limit int) ([]PresenceEvent, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := db.Query(`
		SELECT id, username, event_type, previous_status, new_status, distance, timestamp
		FROM presence_events
		ORDER BY timestamp DESC
		LIMIT ?`, limit)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get presence history: %w", err)
	}
	defer rows.Close()

	var events []PresenceEvent
	for rows.Next() {
		var event PresenceEvent
		var prevStatus sql.NullString
		err := rows.Scan(&event.ID, &event.Username, &event.EventType,
			&prevStatus, &event.NewStatus, &event.Distance, &event.Timestamp)
		if err != nil {
			continue
		}
		if prevStatus.Valid {
			event.PreviousStatus = GeofenceStatus(prevStatus.String)
		}
		events = append(events, event)
	}

	return events, nil
}

// SimulateRandomMovement simulates random movement for demonstration
func SimulateRandomMovement(user *User) error {
	if user.Role != "homeowner" {
		return errors.New("only homeowners can simulate movement")
	}

	config, err := GetGeofenceConfig()
	if err != nil {
		return err
	}

	// Generate random movement within ±0.1 degrees (~11km max)
	randomLat := config.HomeLatitude + (rand.Float64()-0.5)*0.2
	randomLon := config.HomeLongitude + (rand.Float64()-0.5)*0.2

	return SimulateLocationUpdate(randomLat, randomLon, user)
}

// DisplayGeofenceStatus formats geofence status for display
func DisplayGeofenceStatus(config *GeofenceConfig) string {
	if config == nil {
		return "Geofencing: Not Configured"
	}

	status := "Disabled"
	if config.IsEnabled {
		status = "Enabled"
	}

	autoAdjust := "No"
	if config.AutoAdjustEnabled {
		autoAdjust = "Yes"
	}

	distance := CalculateDistance(
		config.HomeLatitude, config.HomeLongitude,
		config.SimulatedLatitude, config.SimulatedLongitude)

	return fmt.Sprintf(`Geofencing & Presence Detection
=====================================
Status: %s
Current Presence: %s
Distance from Home: %.2f km

Home Location: %.4f°N, %.4f°W
Geofence Radius: %.1f km

Temperature Settings:
  - At Home: %.1f°C
  - Away: %.1f°C
  - Coming Home: %.1f°C

Auto-Adjust: %s
Last Update: %s

Simulated Location: %.4f°N, %.4f°W`,
		status,
		string(config.CurrentStatus),
		distance,
		config.HomeLatitude, config.HomeLongitude,
		config.GeofenceRadius,
		config.HomeTemp,
		config.AwayTemp,
		config.ComingHomeTemp,
		autoAdjust,
		config.LastLocationUpdate.Format("2006-01-02 15:04:05"),
		config.SimulatedLatitude, config.SimulatedLongitude)
}

// DisplayPresenceHistory formats presence event history for display
func DisplayPresenceHistory(events []PresenceEvent) string {
	if len(events) == 0 {
		return "No presence events recorded yet."
	}

	result := "Recent Presence Events\n"
	result += "=====================================================\n"
	for _, event := range events {
		result += fmt.Sprintf("[%s] %s: %s -> %s (%.2f km)\n",
			event.Timestamp.Format("2006-01-02 15:04"),
			event.EventType,
			event.PreviousStatus,
			event.NewStatus,
			event.Distance)
	}
	return result
}
