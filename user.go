package main

import (
	"errors"
	"time"
)

// CreateGuestAccount - Both homeowners and technicians can create guest accounts
func CreateGuestAccount(creator, guestName, pin string, creatorRole string) error {
	// Only homeowners and technicians can create guests
	if creatorRole != "homeowner" && creatorRole != "technician" {
		return errors.New("only homeowners or technicians can create guest accounts")
	}

	if len(guestName) < 3 || len(pin) < 4 {
		return errors.New("guest name or PIN too short")
	}

	guestUsername := creator + "_guest_" + guestName
	err := RegisterGuestUser(guestUsername, pin)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO guest_access (guest_username, granted_by) VALUES (?, ?)", guestUsername, creator)
	if err != nil {
		return err
	}

	LogEvent("create_guest", "Guest created: "+guestUsername, creator, "info")
	return nil
}

// CreateTechnicianAccount - ONLY homeowners can create technician accounts
func CreateTechnicianAccount(homeowner, techName, password string, creatorRole string) error {
	// SECURITY: Only homeowners can create technician accounts
	if creatorRole != "homeowner" {
		return errors.New("only homeowners can create technician accounts")
	}

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

// GrantTechnicianAccess - ONLY homeowners can grant/extend technician access
func GrantTechnicianAccess(homeowner, technician string, duration time.Duration, granterRole string) error {
	// SECURITY: Only homeowners can grant technician access
	if granterRole != "homeowner" {
		return errors.New("only homeowners can grant technician access")
	}

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

// RevokeAccess - Both homeowners and technicians can revoke access, but with restrictions
func RevokeAccess(username string, revokerUsername string, revokerRole string) error {
	// SECURITY: Check permissions based on role
	if revokerRole != "homeowner" && revokerRole != "technician" {
		return errors.New("only homeowners or technicians can revoke access")
	}

	// Get the user being revoked
	targetUser, err := GetUserByUsername(username)
	if err != nil {
		return errors.New("user not found")
	}

	// SECURITY: Technicians can only revoke guest accounts they manage or that were granted by their homeowner
	// Technicians CANNOT revoke other technicians or homeowners
	if revokerRole == "technician" {
		if targetUser.Role != "guest" {
			return errors.New("technicians can only revoke guest accounts")
		}

		// Verify the guest was granted by this technician or their homeowner
		var grantedBy string
		err := db.QueryRow("SELECT granted_by FROM guest_access WHERE guest_username = ?", username).Scan(&grantedBy)
		if err != nil || (grantedBy != revokerUsername && !isHomeownerOfTechnician(grantedBy, revokerUsername)) {
			return errors.New("you do not have permission to revoke this guest")
		}
	}

	// Homeowners can revoke anyone in their system
	// Execute the revocation
	_, err = db.Exec("UPDATE users SET is_active = 0, session_token = NULL, session_expires_at = NULL WHERE username = ?", username)
	if err != nil {
		return err
	}

	db.Exec("UPDATE guest_access SET is_active = 0 WHERE guest_username = ?", username)
	LogEvent("revoke_access", "Access revoked", username, "info")
	return nil
}

// Helper function to check if a homeowner manages a technician
func isHomeownerOfTechnician(homeowner, technician string) bool {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM guest_access WHERE guest_username = ? AND granted_by = ?", technician, homeowner).Scan(&count)
	return count > 0
}

// ListAllUsers - ONLY homeowners can list all users
// Technicians and guests cannot view the user list for security/privacy
func ListAllUsers(requesterRole string) ([]User, error) {
	// SECURITY: Only homeowners can list all users
	if requesterRole != "homeowner" {
		return nil, errors.New("only homeowners can view the user list")
	}

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

// DeleteUser permanently deletes a user from the system.
// Only homeowners can delete a user—and never themselves!
func DeleteUser(requester, usernameToDelete, requesterRole string) error {
	// SECURITY: Only homeowners can delete users
	if requesterRole != "homeowner" {
		return errors.New("only homeowners can delete users")
	}

	// Prevent deletion of the requesting homeowner’s own account
	if usernameToDelete == requester {
		return errors.New("cannot delete your own (homeowner) account")
	}

	// Check if user exists
	targetUser, err := GetUserByUsername(usernameToDelete)
	if err != nil {
		return errors.New("user not found")
	}

	// Prevent homeowners from deleting other homeowners
	if targetUser.Role == "homeowner" {
		return errors.New("cannot delete other homeowner accounts")
	}

	// Real delete from users table
	_, err = db.Exec("DELETE FROM users WHERE username = ?", usernameToDelete)
	if err != nil {
		return err
	}

	// Clean up guest_access (if any)
	db.Exec("DELETE FROM guest_access WHERE guest_username = ?", usernameToDelete)
	LogEvent("delete_user", "Permanently deleted user: "+usernameToDelete, requester, "warning")
	return nil
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
}
