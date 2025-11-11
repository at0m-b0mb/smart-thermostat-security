package main

import (
	"errors"
	"fmt"
	"time"
)

type Profile struct {
	ID         int
	Name       string
	TargetTemp float64
	HVACMode   string
	Owner      string
	CreatedAt  time.Time
}

type Schedule struct {
	ID         int
	ProfileID  int
	DayOfWeek  int
	StartTime  string
	EndTime    string
	TargetTemp float64
}

// CreateProfile creates a profile with role-based permission checks
func CreateProfile(requestor *User, profileName string, targetTemp float64, hvacMode string) error {
	// Only homeowners (and approved technicians) can create profiles
	if requestor.Role != "homeowner" && requestor.Role != "technician" {
		return errors.New("profile creation denied for guests")
	}
	if requestor.Role == "technician" && !IsTechnicianAccessAllowed(db, requestor.Username) {
		return errors.New("technician not approved by homeowner")
	}
	
	if len(profileName) < 2 || len(profileName) > 50 {
		return errors.New("invalid profile name length")
	}
	
	if targetTemp < 10 || targetTemp > 35 {
		return errors.New("temperature out of range")
	}
	
	if hvacMode != "off" && hvacMode != "heat" && hvacMode != "cool" && hvacMode != "fan" {
		return errors.New("invalid HVAC mode")
	}
	
	_, err := db.Exec("INSERT INTO profiles (profile_name, target_temp, hvac_mode, owner) VALUES (?, ?, ?, ?)",
		profileName, targetTemp, hvacMode, requestor.Username)
	if err != nil {
		return errors.New("profile already exists")
	}
	
	LogEvent("profile_create", "Profile created: "+profileName, requestor.Username, "info")
	return nil
}

func GetProfile(profileName string) (*Profile, error) {
	var profile Profile
	err := db.QueryRow("SELECT id, profile_name, target_temp, hvac_mode, owner, created_at FROM profiles WHERE profile_name = ?",
		profileName).Scan(&profile.ID, &profile.Name, &profile.TargetTemp, &profile.HVACMode, &profile.Owner, &profile.CreatedAt)
	if err != nil {
		return nil, errors.New("profile not found")
	}
	return &profile, nil
}

func ListProfiles(owner string) ([]Profile, error) {
	rows, err := db.Query("SELECT id, profile_name, target_temp, hvac_mode, owner, created_at FROM profiles WHERE owner = ?", owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	profiles := []Profile{}
	for rows.Next() {
		var p Profile
		if err := rows.Scan(&p.ID, &p.Name, &p.TargetTemp, &p.HVACMode, &p.Owner, &p.CreatedAt); err != nil {
			continue
		}
		profiles = append(profiles, p)
	}
	return profiles, nil
}

// ApplyProfile applies a profile with role-based access control
func ApplyProfile(profileName string, user *User) error {
	profile, err := GetProfile(profileName)
	if err != nil {
		return err
	}
	
	// Guests can only apply their own profiles
	if user.Role == "guest" && profile.Owner != user.Username {
		return errors.New("guests can only apply their own profiles")
	}
	
	// Homeowners and technicians (with access) can apply any profile
	if user.Role == "technician" && !IsTechnicianAccessAllowed(db, user.Username) {
		return errors.New("technician access expired or not granted")
	}
	
	err = SetHVACMode(profile.HVACMode, user)
	if err != nil {
		return err
	}
	
	err = SetTargetTemperature(profile.TargetTemp, user)
	if err != nil {
		return err
	}
	
	LogEvent("profile_apply", "Profile applied: "+profileName, user.Username, "info")
	return nil
}

// DeleteProfile deletes a profile with ownership validation
func DeleteProfile(requestor *User, profileName string) error {
	// Only homeowners and technicians can delete profiles
	if requestor.Role == "guest" {
		return errors.New("guests cannot delete profiles")
	}
	if requestor.Role == "technician" && !IsTechnicianAccessAllowed(db, requestor.Username) {
		return errors.New("technician access expired or not granted")
	}
	
	result, err := db.Exec("DELETE FROM profiles WHERE profile_name = ? AND owner = ?", profileName, requestor.Username)
	if err != nil {
		return err
	}
	
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("profile not found or unauthorized")
	}
	
	LogEvent("profile_delete", "Profile deleted: "+profileName, requestor.Username, "info")
	return nil
}

func AddSchedule(profileID, dayOfWeek int, startTime, endTime string, targetTemp float64) error {
	if dayOfWeek < 0 || dayOfWeek > 6 {
		return errors.New("invalid day of week")
	}
	
	if targetTemp < 10 || targetTemp > 35 {
		return errors.New("temperature out of range")
	}
	
	_, err := db.Exec("INSERT INTO schedules (profile_id, day_of_week, start_time, end_time, target_temp) VALUES (?, ?, ?, ?, ?)",
		profileID, dayOfWeek, startTime, endTime, targetTemp)
	if err != nil {
		return errors.New("failed to add schedule")
	}
	
	LogEvent("schedule_add", fmt.Sprintf("Schedule added for profile %d", profileID), "system", "info")
	return nil
}

func GetSchedules(profileID int) ([]Schedule, error) {
	rows, err := db.Query("SELECT id, profile_id, day_of_week, start_time, end_time, target_temp FROM schedules WHERE profile_id = ?", profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	schedules := []Schedule{}
	for rows.Next() {
		var s Schedule
		if err := rows.Scan(&s.ID, &s.ProfileID, &s.DayOfWeek, &s.StartTime, &s.EndTime, &s.TargetTemp); err != nil {
			continue
		}
		schedules = append(schedules, s)
	}
	return schedules, nil
}
