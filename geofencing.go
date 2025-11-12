package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// Location represents GPS coordinates
type Location struct {
	Latitude  float64
	Longitude float64
	Timestamp time.Time
}

// GeofenceConfig stores the geofence settings
type GeofenceConfig struct {
	ID              int
	Owner           string
	HomeLat         float64
	HomeLong        float64
	RadiusMeters    float64
	AwayModeTemp    float64
	HomeModeTemp    float64
	Enabled         bool
	CreatedAt       time.Time
	LastUpdated     time.Time
}

// PresenceStatus represents current user presence state
type PresenceStatus struct {
	IsHome          bool
	CurrentLocation Location
	DistanceFromHome float64
	LastUpdate      time.Time
	TriggerCount    int
}

var (
	geofenceMutex    sync.RWMutex
	presenceStatus   PresenceStatus
	geofenceConfig   *GeofenceConfig
	simulatedLocation Location
	geofenceEnabled  bool = false
)

// InitializeGeofencing sets up the geofencing system
func InitializeGeofencing() error {
	geofenceMutex.Lock()
	defer geofenceMutex.Unlock()

	// Initialize with default simulated home location (Baltimore, MD area)
	simulatedLocation = Location{
		Latitude:  39.2904,
		Longitude: -76.6122,
		Timestamp: time.Now(),
	}

	presenceStatus = PresenceStatus{
		IsHome:          true,
		CurrentLocation: simulatedLocation,
		DistanceFromHome: 0,
		LastUpdate:      time.Now(),
		TriggerCount:    0,
	}

	LogEvent("geofence_init", "Geofencing system initialized", "system", "info")
	return nil
}

// ConfigureGeofence sets up or updates geofence settings
func ConfigureGeofence(owner string, homeLat, homeLong, radiusMeters, awayTemp, homeTemp float64, user *User) error {
	if user.Role != "homeowner" {
		AuditSecurityEvent("geofence_unauthorized", "Non-homeowner attempted to configure geofence", user.Username)
		return errors.New("only homeowners can configure geofencing")
	}

	// Validate coordinates
	if homeLat < -90 || homeLat > 90 {
		return errors.New("invalid latitude (must be -90 to 90)")
	}
	if homeLong < -180 || homeLong > 180 {
		return errors.New("invalid longitude (must be -180 to 180)")
	}
	if radiusMeters <= 0 || radiusMeters > 10000 {
		return errors.New("invalid radius (must be 1-10000 meters)")
	}

	// Validate temperatures
	if err := ValidateTemperatureInput(awayTemp); err != nil {
		return fmt.Errorf("invalid away temperature: %w", err)
	}
	if err := ValidateTemperatureInput(homeTemp); err != nil {
		return fmt.Errorf("invalid home temperature: %w", err)
	}

	geofenceMutex.Lock()
	defer geofenceMutex.Unlock()

	// Create or update geofence config
	config := &GeofenceConfig{
		Owner:        owner,
		HomeLat:      homeLat,
		HomeLong:     homeLong,
		RadiusMeters: radiusMeters,
		AwayModeTemp: awayTemp,
		HomeModeTemp: homeTemp,
		Enabled:      true,
		CreatedAt:    time.Now(),
		LastUpdated:  time.Now(),
	}

	// Save to database
	result, err := db.Exec(`
		INSERT INTO geofence_config (owner, home_lat, home_long, radius_meters, away_temp, home_temp, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(owner) DO UPDATE SET
			home_lat = excluded.home_lat,
			home_long = excluded.home_long,
			radius_meters = excluded.radius_meters,
			away_temp = excluded.away_temp,
			home_temp = excluded.home_temp,
			enabled = excluded.enabled,
			last_updated = CURRENT_TIMESTAMP
	`, owner, homeLat, homeLong, radiusMeters, awayTemp, homeTemp, 1)

	if err != nil {
		return fmt.Errorf("failed to save geofence config: %w", err)
	}

	// Get the ID of the inserted/updated row
	id, _ := result.LastInsertId()
	config.ID = int(id)
	geofenceConfig = config
	geofenceEnabled = true

	LogEvent("geofence_configured", fmt.Sprintf("Geofence configured: radius=%.0fm, away=%.1f°C, home=%.1f°C", radiusMeters, awayTemp, homeTemp), owner, "info")
	return nil
}

// GetGeofenceConfig retrieves the current geofence configuration
func GetGeofenceConfig(owner string, user *User) (*GeofenceConfig, error) {
	if user.Role != "homeowner" && user.Username != owner {
		return nil, errors.New("unauthorized to view geofence config")
	}

	var config GeofenceConfig
	err := db.QueryRow(`
		SELECT id, owner, home_lat, home_long, radius_meters, away_temp, home_temp, enabled, created_at, last_updated
		FROM geofence_config WHERE owner = ?
	`, owner).Scan(&config.ID, &config.Owner, &config.HomeLat, &config.HomeLong, &config.RadiusMeters,
		&config.AwayModeTemp, &config.HomeModeTemp, &config.Enabled, &config.CreatedAt, &config.LastUpdated)

	if err != nil {
		return nil, fmt.Errorf("no geofence configured: %w", err)
	}

	return &config, nil
}

// ToggleGeofence enables or disables geofencing
func ToggleGeofence(owner string, enabled bool, user *User) error {
	if user.Role != "homeowner" {
		AuditSecurityEvent("geofence_unauthorized", "Non-homeowner attempted to toggle geofence", user.Username)
		return errors.New("only homeowners can toggle geofencing")
	}

	geofenceMutex.Lock()
	defer geofenceMutex.Unlock()

	_, err := db.Exec("UPDATE geofence_config SET enabled = ?, last_updated = CURRENT_TIMESTAMP WHERE owner = ?", enabled, owner)
	if err != nil {
		return err
	}

	geofenceEnabled = enabled
	if geofenceConfig != nil {
		geofenceConfig.Enabled = enabled
	}

	status := "enabled"
	if !enabled {
		status = "disabled"
	}
	LogEvent("geofence_toggle", fmt.Sprintf("Geofencing %s", status), owner, "info")
	return nil
}

// SimulateLocation updates the simulated user location
func SimulateLocation(lat, long float64) error {
	// Validate coordinates
	if lat < -90 || lat > 90 {
		return errors.New("invalid latitude (must be -90 to 90)")
	}
	if long < -180 || long > 180 {
		return errors.New("invalid longitude (must be -180 to 180)")
	}

	geofenceMutex.Lock()
	defer geofenceMutex.Unlock()

	simulatedLocation = Location{
		Latitude:  lat,
		Longitude: long,
		Timestamp: time.Now(),
	}

	// Log location update
	_, err := db.Exec("INSERT INTO location_logs (latitude, longitude) VALUES (?, ?)", lat, long)
	if err != nil {
		return fmt.Errorf("failed to log location: %w", err)
	}

	return nil
}

// GetPresenceStatus returns the current presence detection status
func GetPresenceStatus() PresenceStatus {
	geofenceMutex.RLock()
	defer geofenceMutex.RUnlock()
	return presenceStatus
}

// CheckPresenceAndAdjustHVAC checks if user is within geofence and adjusts HVAC accordingly
func CheckPresenceAndAdjustHVAC() error {
	geofenceMutex.Lock()
	defer geofenceMutex.Unlock()

	if geofenceConfig == nil || !geofenceEnabled {
		return nil // Geofencing not configured or disabled
	}

	// Calculate distance from home
	distance := calculateDistance(
		simulatedLocation.Latitude, simulatedLocation.Longitude,
		geofenceConfig.HomeLat, geofenceConfig.HomeLong,
	)

	wasHome := presenceStatus.IsHome
	isHome := distance <= geofenceConfig.RadiusMeters

	// Update presence status
	presenceStatus.CurrentLocation = simulatedLocation
	presenceStatus.DistanceFromHome = distance
	presenceStatus.IsHome = isHome
	presenceStatus.LastUpdate = time.Now()

	// If presence state changed, adjust HVAC and log event
	if wasHome != isHome {
		presenceStatus.TriggerCount++

		if isHome {
			// User arrived home - set to home mode temperature
			LogEvent("geofence_trigger", fmt.Sprintf("User arrived home (%.0fm from home)", distance), geofenceConfig.Owner, "info")
			
			// Create a system user for automatic HVAC changes
			systemUser := &User{Username: "system", Role: "homeowner"}
			if err := SetTargetTemperature(geofenceConfig.HomeModeTemp, systemUser); err != nil {
				return fmt.Errorf("failed to adjust temperature: %w", err)
			}
			
			SendNotification(geofenceConfig.Owner, "presence_detected", 
				fmt.Sprintf("Welcome home! Temperature set to %.1f°C", geofenceConfig.HomeModeTemp))
		} else {
			// User left home - set to away mode temperature
			LogEvent("geofence_trigger", fmt.Sprintf("User left home (%.0fm from home)", distance), geofenceConfig.Owner, "info")
			
			systemUser := &User{Username: "system", Role: "homeowner"}
			if err := SetTargetTemperature(geofenceConfig.AwayModeTemp, systemUser); err != nil {
				return fmt.Errorf("failed to adjust temperature: %w", err)
			}
			
			SendNotification(geofenceConfig.Owner, "presence_lost", 
				fmt.Sprintf("Away mode activated. Temperature set to %.1f°C", geofenceConfig.AwayModeTemp))
		}

		// Log the presence change
		_, err := db.Exec(`
			INSERT INTO presence_logs (owner, is_home, distance_meters, latitude, longitude)
			VALUES (?, ?, ?, ?, ?)
		`, geofenceConfig.Owner, isHome, distance, simulatedLocation.Latitude, simulatedLocation.Longitude)
		
		if err != nil {
			LogEvent("geofence_error", "Failed to log presence change: "+err.Error(), "system", "warning")
		}
	}

	return nil
}

// calculateDistance computes the distance between two GPS coordinates using the Haversine formula
// Returns distance in meters
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // Earth radius in meters

	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// SimulateMovement simulates random user movement (for testing)
func SimulateMovement() error {
	geofenceMutex.RLock()
	if geofenceConfig == nil {
		geofenceMutex.RUnlock()
		return errors.New("geofence not configured")
	}
	homeLat := geofenceConfig.HomeLat
	homeLong := geofenceConfig.HomeLong
	geofenceMutex.RUnlock()

	// Randomly move within +/- 2km from home
	// 1 degree of latitude is approximately 111km
	// 1 degree of longitude varies by latitude but approximately 111km * cos(latitude)
	latOffset := (rand.Float64()*2 - 1) * 0.018  // ~2km range
	longOffset := (rand.Float64()*2 - 1) * 0.018

	newLat := homeLat + latOffset
	newLong := homeLong + longOffset

	// Ensure coordinates are valid
	if newLat < -90 {
		newLat = -90
	} else if newLat > 90 {
		newLat = 90
	}
	if newLong < -180 {
		newLong = -180
	} else if newLong > 180 {
		newLong = 180
	}

	return SimulateLocation(newLat, newLong)
}

// GetRecentPresenceLogs retrieves recent presence detection events
func GetRecentPresenceLogs(owner string, limit int, user *User) ([]PresenceLog, error) {
	if user.Role != "homeowner" && user.Username != owner {
		return nil, errors.New("unauthorized to view presence logs")
	}

	rows, err := db.Query(`
		SELECT id, owner, is_home, distance_meters, latitude, longitude, timestamp
		FROM presence_logs
		WHERE owner = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`, owner, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []PresenceLog
	for rows.Next() {
		var log PresenceLog
		err := rows.Scan(&log.ID, &log.Owner, &log.IsHome, &log.Distance, &log.Latitude, &log.Longitude, &log.Timestamp)
		if err != nil {
			continue
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// PresenceLog represents a logged presence detection event
type PresenceLog struct {
	ID        int
	Owner     string
	IsHome    bool
	Distance  float64
	Latitude  float64
	Longitude float64
	Timestamp time.Time
}

// GeofenceMonitorLoop continuously monitors location and adjusts HVAC
func GeofenceMonitorLoop() {
	ticker := time.NewTicker(45 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		geofenceMutex.RLock()
		enabled := geofenceEnabled && geofenceConfig != nil
		geofenceMutex.RUnlock()
		
		if enabled {
			// Simulate some movement
			if rand.Float64() < 0.3 { // 30% chance to move
				SimulateMovement()
			}
			
			// Check presence and adjust HVAC
			if err := CheckPresenceAndAdjustHVAC(); err != nil {
				LogEvent("geofence_error", "Presence check failed: "+err.Error(), "system", "warning")
			}
		}
	}
}
