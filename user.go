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
    
    // Register the guest account in the users table
    err := RegisterGuestUser(guestUsername, pin)
    if err != nil {
        return err
    }

    // Link guest to the creator (homeowner/technician)
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

    // Add the technician to the users table
    err := RegisterUser(techName, password, "technician")
    if err != nil {
        return err
    }

    // Insert technician into guest_access for tracking (no expiration)
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

    // Ensure technician exists and has valid role
    tech, err := GetUserByUsername(technician)
    if err != nil || tech.Role != "technician" {
        return errors.New("invalid technician")
    }

    expiresAt := time.Now().Add(duration)
    
    // Update technician’s access duration in guest_access table
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

    // Fetch target user from the database
    targetUser, err := GetUserByUsername(username)
    if err != nil {
        return errors.New("user not found")
    }

    // SECURITY: Restrict technician revocation rights to their own guests
    if revokerRole == "technician" {
        if targetUser.Role != "guest" {
            return errors.New("technicians can only revoke guest accounts")
        }
        var grantedBy string
        err := db.QueryRow("SELECT granted_by FROM guest_access WHERE guest_username = ?", username).Scan(&grantedBy)
        if err != nil || (grantedBy != revokerUsername && !isHomeownerOfTechnician(grantedBy, revokerUsername)) {
            return errors.New("you do not have permission to revoke this guest")
        }
    }

    // Perform deactivation in both users and guest_access tables
    _, err = db.Exec("UPDATE users SET is_active = 0, session_token = NULL, session_expires_at = NULL WHERE username = ?", username)
    if err != nil {
        return err
    }

    db.Exec("UPDATE guest_access SET is_active = 0 WHERE guest_username = ?", username)
    LogEvent("revoke_access", "Access revoked", username, "info")
    return nil
}
