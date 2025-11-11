package main

import (
	"errors"
	"fmt"
	"time"
)

type DiagnosticReport struct {
	Timestamp     time.Time
	SystemHealth  string
	SensorStatus  SensorStatus
	NetworkStatus bool
	Errors        []string
	Warnings      []string
}

func RunSystemDiagnostics(user *User) (DiagnosticReport, error) {
    if user.Role != "homeowner" && user.Role != "technician" {
        return DiagnosticReport{}, errors.New("access denied: only homeowners or technicians can run diagnostics")
    }
	LogEvent("diagnostics_start", "System diagnostics initiated", "system", "info")
	report := DiagnosticReport{
		Timestamp:     time.Now(),
		SystemHealth:  "Unknown",
		NetworkStatus: true,
		Errors:        []string{},
		Warnings:      []string{},
	}
	sensorStatus := GetSensorStatus()
	report.SensorStatus = sensorStatus
	if !sensorStatus.IsHealthy {
		report.Errors = append(report.Errors, "Sensor system unhealthy")
	}
	if sensorStatus.ErrorCount > 0 {
		report.Warnings = append(report.Warnings, fmt.Sprintf("Sensor errors: %d", sensorStatus.ErrorCount))
	}
	_, err := ReadTemperature()
	if err != nil {
		report.Errors = append(report.Errors, "Temperature sensor failed")
	}
	_, err = ReadHumidity()
	if err != nil {
		report.Errors = append(report.Errors, "Humidity sensor failed")
	}
	_, err = ReadCO()
	if err != nil {
		report.Errors = append(report.Errors, "CO sensor failed")
	}
	report.NetworkStatus = testNetworkConnectivity()
	if !report.NetworkStatus {
		report.Warnings = append(report.Warnings, "Network connectivity issue")
	}
	if len(report.Errors) == 0 {
		report.SystemHealth = "Healthy"
	} else if len(report.Errors) < 3 {
		report.SystemHealth = "Degraded"
	} else {
		report.SystemHealth = "Critical"
	}
	LogEvent("diagnostics_complete", "Health: "+report.SystemHealth, "system", "info")
	return report, nil
}

func testNetworkConnectivity() bool {
	return true
}

func CheckSensorHealth() error {
	status := GetSensorStatus()
	if !status.IsHealthy {
		return errors.New("sensor system unhealthy")
	}
	if time.Since(status.LastReading) > 5*time.Minute {
		return errors.New("sensor data stale")
	}
	return nil
}

func GenerateDiagnosticReport(report DiagnosticReport) string {
	output := "=== SYSTEM DIAGNOSTIC REPORT ===\n"
	output += fmt.Sprintf("Time: %s\n", report.Timestamp.Format(time.RFC3339))
	output += fmt.Sprintf("Overall Health: %s\n\n", report.SystemHealth)
	output += fmt.Sprintf("Sensor Status:\n")
	output += fmt.Sprintf("  Healthy: %v\n", report.SensorStatus.IsHealthy)
	output += fmt.Sprintf("  Last Reading: %s\n", report.SensorStatus.LastReading.Format("15:04:05"))
	output += fmt.Sprintf("  Error Count: %d\n\n", report.SensorStatus.ErrorCount)
	output += fmt.Sprintf("Network Status: %v\n\n", report.NetworkStatus)
	if len(report.Errors) > 0 {
		output += "ERRORS:\n"
		for _, err := range report.Errors {
			output += fmt.Sprintf("  - %s\n", err)
		}
		output += "\n"
	}
	if len(report.Warnings) > 0 {
		output += "WARNINGS:\n"
		for _, warn := range report.Warnings {
			output += fmt.Sprintf("  - %s\n", warn)
		}
	}
	return output
}
