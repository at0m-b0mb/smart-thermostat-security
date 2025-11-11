package main

import (
	"errors"
	"time"
)

// CreateGuestAccount creates a guest account with role-based permission checks
func CreateGuestAccount(requestor *User, guestName, pin string) error {
	// Only homeowners (or technician with approval) can create guests
	if requestor.Role == "guest" {
		return errors.New("guests cannot create accounts")
	}
	if requestor.Role == "technician" && !IsTechnicianAccessAllowed(db, requestor.Username) {
		return errors.New("technician not authorized by homeowner")
	}
	
	if len(guestName) < 3 || len(pin) < 4 {
		return errors.New("guest name or PIN too short")
	}
	
	// Determine the homeowner username
	homeowner := requestor.Username
	if requestor.Role == "technician" {
		// Get the homeowner who granted this technician access
		var grantedBy string
		err := db.QueryRow("SELECT granted_by FROM guest_access WHERE guest_username = ? AND is_active = 1", requestor.Username).Scan(&grantedBy)
		if err != nil {
			return errors.New("cannot determine homeowner for technician")
		}
		homeowner = grantedBy
	}
	
	guestUsername := homeowner + "_guest_" + guestName
	err := RegisterGuestUser(guestUsername, pin)
	if err != nil {
		return err
	}
	
	_, err = db.Exec("INSERT INTO guest_access (guest_username, granted_by) VALUES (?, ?)", guestUsername, homeowner)
	if err != nil {
		return err
	}
	
	LogEvent("create_guest", "Guest created: "+guestUsername, requestor.Username, "info")
	return nil
}

func CreateTechnicianAccount(homeowner, techName, password string) error {
	if len(techName) < 3 || len(password) < 4 {
		return errors.New("technician name or password too short")
	}
	
	err := RegisterUser(techName, password, "technician")
	if err != nil {
		return err
	}
	
	expiresAt := "NULL"
	_, err = db.Exec("INSERT INTO guest_access (guest_username, granted_by, expires_at) VALUES (?, ?, ?)", techName, homeowner, expiresAt)
	if err != nil {
		return err
	}
	
	LogEvent("create_technician", "Technician created: "+techName, homeowner, "info")
	return nil
}

// GrantTechnicianAccess grants time-limited access to a technician (homeowner only)
func GrantTechnicianAccess(requestor *User, technician string, duration time.Duration) error {
	if requestor.Role != "homeowner" {
		return errors.New("only homeowners can grant technician access")
	}
	
	tech, err := GetUserByUsername(technician)
	if err != nil || tech.Role != "technician" {
		return errors.New("invalid technician")
	}
	
	expiresAt := time.Now().Add(duration)
	res, err := db.Exec(
		"UPDATE guest_access SET expires_at = ?, is_active = 1 WHERE guest_username = ? AND granted_by = ?",
		expiresAt, technician, requestor.Username,
	)
	if err != nil {
		return err
	}
	
	if ra, _ := res.RowsAffected(); ra == 0 {
		return errors.New("no existing grant found to update")
	}
	
	LogEvent("grant_tech", "Tech access extended until "+expiresAt.Format(time.RFC3339), requestor.Username, "info")
	return nil
}

func RevokeAccess(username string) error {
	_, err := db.Exec("UPDATE users SET is_active = 0, session_token = NULL WHERE username = ?", username)
	if err != nil {
		return err
	}
	
	db.Exec("UPDATE guest_access SET is_active = 0 WHERE guest_username = ?", username)
	LogEvent("revoke_access", "Access revoked", username, "info")
	return nil
}

func ListAllUsers() ([]User, error) {
	rows, err := db.Query("SELECT id, username, role, is_active FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	users := []User{}
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Role, &user.IsActive); err != nil {
			continue
		}
		users = append(users, user)
	}
	return users, nil
}

func ChangePassword(username, oldPassword, newPassword string) error {
	var passwordHash, role string
	err := db.QueryRow("SELECT password_hash, role FROM users WHERE username = ?", username).Scan(&passwordHash, &role)
	if err != nil {
		return err
	}
	
	if !CheckPassword(passwordHash, oldPassword) {
		return errors.New("incorrect old password")
	}
	
	if role == "guest" {
		// For guests, enforce PIN validation instead of password
		if err := ValidatePin(newPassword); err != nil {
			return err
		}
	} else {
		// For homeowners/technicians, enforce password rules
		if err := ValidatePassword(newPassword); err != nil {
			return err
		}
	}
	
	newHash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	
	_, err = db.Exec("UPDATE users SET password_hash = ? WHERE username = ?", newHash, username)
	if err != nil {
		return err
	}
	
	LogEvent("password_change", "Password changed", username, "info")
	return nil
}
