package main

import (
	"fmt"
	"time"
)

type Notification struct {
	ID        int
	Message   string
	Type      string
	Timestamp time.Time
	Username  string
	IsRead    bool
}

func SendNotification(username, notifType, message string) error {
	LogEvent("notification", message, username, "info")
	fmt.Printf("[NOTIFICATION] To: %s | Type: %s | Message: %s\n", username, notifType, message)
	return nil
}

func SendTemperatureAlert(username string, currentTemp, targetTemp float64) error {
	message := fmt.Sprintf("Temperature alert: Current %.1f°C, Target %.1f°C", currentTemp, targetTemp)
	return SendNotification(username, "temperature_alert", message)
}

func SendCOAlert(username string, coLevel float64) error {
	message := fmt.Sprintf("CRITICAL: Dangerous CO level detected: %.2f ppm", coLevel)
	LogEvent("co_alert", message, username, "critical")
	return SendNotification(username, "co_alert", message)
}

func SendSystemAlert(username, alertMessage string) error {
	return SendNotification(username, "system_alert", alertMessage)
}

func SendMaintenanceReminder(username string) error {
	message := "Maintenance reminder: Schedule system checkup"
	return SendNotification(username, "maintenance", message)
}

func SendEnergyUsageAlert(username string, usage float64, threshold float64) error {
	message := fmt.Sprintf("Energy usage alert: %.2f kWh (threshold: %.2f kWh)", usage, threshold)
	return SendNotification(username, "energy_alert", message)
}

func SendSecurityAlert(username, alertType, details string) error {
	message := fmt.Sprintf("Security Alert [%s]: %s", alertType, details)
	LogEvent("security_alert", message, username, "critical")
	return SendNotification(username, "security_alert", message)
}

func SendAccessGrantedNotification(username, grantedTo string) error {
	message := fmt.Sprintf("Access granted to %s", grantedTo)
	return SendNotification(username, "access_granted", message)
}

func SendAccessRevokedNotification(username, revokedFrom string) error {
	message := fmt.Sprintf("Access revoked from %s", revokedFrom)
	return SendNotification(username, "access_revoked", message)
}

func BroadcastSystemNotification(message string) error {
	users, err := ListAllUsers()
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.IsActive {
			SendNotification(user.Username, "system_broadcast", message)
		}
	}
	return nil
}
