package main

import (
	"errors"
	"time"
)

func CreateGuestAccount(homeowner, guestName, pin string) error {
	if len(guestName) < 3 || len(pin) < 4 {
		return errors.New("guest name or PIN too short")
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
	LogEvent("create_guest", "Guest created: "+guestUsername, homeowner, "info")
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

func GrantTechnicianAccess(homeowner, technician string, duration time.Duration) error {
	tech, err := GetUserByUsername(technician)
	if err != nil || tech.Role != "technician" {
		return errors.New("invalid technician")
	}
	expiresAt := time.Now().Add(duration)

	res, err := db.Exec(
		"UPDATE guest_access SET expires_at = ?, is_active = 1 WHERE guest_username = ? AND granted_by = ?",
		expiresAt, technician, homeowner,
	)
	if err != nil {
		return err
	}
	if ra, _ := res.RowsAffected(); ra == 0 {
		return errors.New("no existing grant found to update")
	}
	LogEvent("grant_tech", "Tech access extended until "+expiresAt.Format(time.RFC3339), homeowner, "info")
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

// ChangePassword changes the password for homeowners and technicians
// Guests CANNOT use this function - they must use ChangePIN instead
func ChangePassword(username, oldPassword, newPassword string) error {
	var passwordHash string
	var role string
	err := db.QueryRow("SELECT password_hash, role FROM users WHERE username = ?", username).Scan(&passwordHash, &role)
	if err != nil {
		return err
	}

	// Security check: prevent guests from changing passwords
	if role == "guest" {
		return errors.New("guests cannot change passwords, use ChangePIN instead")
	}

	if !CheckPassword(passwordHash, oldPassword) {
		return errors.New("incorrect old password")
	}

	if err := ValidatePassword(newPassword); err != nil {
		return err
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

// ChangePIN changes the PIN for guest accounts only
func ChangePIN(username, oldPIN, newPIN string) error {
	var passwordHash string
	var role string
	err := db.QueryRow("SELECT password_hash, role FROM users WHERE username = ?", username).Scan(&passwordHash, &role)
	if err != nil {
		return err
	}

	// Security check: only guests can change PINs
	if role != "guest" {
		return errors.New("only guests can change PINs, use ChangePassword instead")
	}

	if !CheckPassword(passwordHash, oldPIN) {
		return errors.New("incorrect old PIN")
	}

	// Validate PIN format (must be numeric, minimum 4 digits)
	if err := ValidatePin(newPIN); err != nil {
		return err
	}

	newHash, err := HashPassword(newPIN)
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE users SET password_hash = ? WHERE username = ?", newHash, username)
	if err != nil {
		return err
	}

	LogEvent("pin_change", "PIN changed", username, "info")
	return nil
