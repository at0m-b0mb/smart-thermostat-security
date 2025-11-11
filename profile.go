package main

import (
	"database/sql"
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
	GuestAccessible int   
}

type Schedule struct {
	ID         int
	ProfileID  int
	DayOfWeek  int
	StartTime  string
	EndTime    string
	TargetTemp float64
}

func CreateProfile(profileName string, targetTemp float64, hvacMode, owner string, user *User, guestAccessible int) error {
	if user.Role != "homeowner" && user.Role != "technician" {
	return errors.New("only homeowners or technicians can create a profile")
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
    // Secure check: only allow guestAccessible to be 0 or 1
    if guestAccessible != 0 && guestAccessible != 1 {
        return errors.New("invalid guest accessible flag (must be 0 or 1)")
    }

    _, err := db.Exec(
        "INSERT INTO profiles (profile_name, target_temp, hvac_mode, owner, guest_accessible) VALUES (?, ?, ?, ?, ?)",
        profileName, targetTemp, hvacMode, owner, guestAccessible,
    )
    if err != nil {
        return errors.New("profile already exists or database error")
    }
    LogEvent("profile_create", "Profile created: " + profileName, owner, "info")
    return nil
}

func GetProfile(profileName string) (*Profile, error) {
	var profile Profile
	err := db.QueryRow("SELECT id, profile_name, target_temp, hvac_mode, owner, guest_accessible, created_at FROM profiles WHERE profile_name = ?", profileName).
    Scan(&profile.ID, &profile.Name, &profile.TargetTemp, &profile.HVACMode, &profile.Owner, &profile.GuestAccessible, &profile.CreatedAt)
	if err != nil {
		return nil, errors.New("cannot apply this profile")
	}
	return &profile, nil
}

func ListProfiles(owner string, user *User) ([]Profile, error) {
    var rows *sql.Rows
    var err error

    if user.Role == "guest" {
        // Guests: only see guest-accessible profiles
        rows, err = db.Query("SELECT id, profile_name, target_temp, hvac_mode, owner, guest_accessible, created_at FROM profiles WHERE guest_accessible = 1")
    } else if user.Role == "technician" {
        // Technicians: see profiles they own OR guest-accessible profiles
        rows, err = db.Query("SELECT id, profile_name, target_temp, hvac_mode, owner, guest_accessible, created_at FROM profiles WHERE owner = ? OR guest_accessible = 1", user.Username)
    } else {
        // Homeowner/Admin: see all profiles created by owner
        rows, err = db.Query("SELECT id, profile_name, target_temp, hvac_mode, owner, guest_accessible, created_at FROM profiles")
    }
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    profiles := []Profile{}
    for rows.Next() {
        var p Profile
        if err := rows.Scan(&p.ID, &p.Name, &p.TargetTemp, &p.HVACMode, &p.Owner, &p.GuestAccessible, &p.CreatedAt); err != nil {
            continue
        }
        profiles = append(profiles, p)
    }
    return profiles, nil
}

func ApplyProfile(profileName string, user *User) error {
    profile, err := GetProfile(profileName)
    if err != nil {
        return err
    }
    // Guests: only apply if guest-accessible
    if user.Role == "guest" {
        if profile.GuestAccessible != 1 {
            return errors.New("cannot apply this profile")
        }
    }
    // Technicians: only apply if guest-accessible or their own profile
    if user.Role == "technician" {
        if profile.Owner != user.Username && profile.GuestAccessible != 1 {
            return errors.New("technician cannot apply this profile")
        }
    }
    // Homeowner/Admin: can apply any profile

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

func DeleteProfile(profileName, user string, role string) error {
    var result sql.Result
    var err error
    if role == "homeowner" || role == "admin" {
        // Homeowner or admin: delete any profile matches
        result, err = db.Exec("DELETE FROM profiles WHERE profile_name = ?", profileName)
    } else if role == "technician" {
        // Technician: delete if guest_accessible = 1
        result, err = db.Exec("DELETE FROM profiles WHERE profile_name = ? AND guest_accessible = 1", profileName)
    } else {
        return errors.New("unauthorized")
    }
	LogEvent("profile_delete", "Profile deleted: "+profileName, owner, "info")
	return nil
}

func AddSchedule(profileID, dayOfWeek int, startTime, endTime string, targetTemp float64, user *User) error {
    if user.Role != "homeowner" && user.Role != "technician" {
        return errors.New("only homeowners or technicians can add a schedule")
    }
    if dayOfWeek < 0 || dayOfWeek > 6 {
        return errors.New("invalid day of week")
    }
    if targetTemp < 10 || targetTemp > 35 {
        return errors.New("temperature out of range")
    }
    _, err := db.Exec("INSERT INTO schedules (profile_id, day_of_week, start_time, end_time, target_temp) VALUES (?, ?, ?, ?, ?)", profileID, dayOfWeek, startTime, endTime, targetTemp)
    if err != nil {
        return errors.New("failed to add schedule")
    }
    LogEvent("schedule_add", fmt.Sprintf("Schedule added for profile %d", profileID), "system", "info")
    return nil
}


func GetSchedules(profileID int, user *User) ([]Schedule, error) {
    if user.Role != "homeowner" && user.Role != "technician" {
        return nil, errors.New("permission denied")
    }
    rows, err := db.Query("SELECT day_of_week, start_time, end_time, target_temp FROM schedules WHERE profile_id = ?", profileID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var schedules []Schedule
    for rows.Next() {
        var s Schedule
        err := rows.Scan(&s.DayOfWeek, &s.StartTime, &s.EndTime, &s.TargetTemp)
        if err != nil {
            return nil, err
        }
        schedules = append(schedules, s)
    }
    return schedules, nil
}

