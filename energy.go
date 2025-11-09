package main

import (
	"fmt"
	"time"
)

type EnergyStats struct {
	TotalKWH       float64
	TotalRuntime   int
	HeatingKWH     float64
	CoolingKWH     float64
	FanKWH         float64
	EstimatedCost  float64
	Period         string
}

func GetEnergyUsage(days int) (EnergyStats, error) {
	if days <= 0 {
		days = 7
	}
	cutoffDate := time.Now().AddDate(0, 0, -days)
	rows, err := db.Query("SELECT hvac_mode, runtime_minutes, estimated_kwh FROM energy_logs WHERE timestamp >= ?", cutoffDate)
	if err != nil {
		return EnergyStats{}, err
	}
	defer rows.Close()
	stats := EnergyStats{Period: fmt.Sprintf("Last %d days", days)}
	for rows.Next() {
		var mode string
		var runtime int
		var kwh float64
		if err := rows.Scan(&mode, &runtime, &kwh); err != nil {
			continue
		}
		stats.TotalKWH += kwh
		stats.TotalRuntime += runtime
		switch mode {
		case "heat":
			stats.HeatingKWH += kwh
		case "cool":
			stats.CoolingKWH += kwh
		case "fan":
			stats.FanKWH += kwh
		}
	}
	stats.EstimatedCost = stats.TotalKWH * 0.12
	return stats, nil
}

func GenerateEnergyReport(stats EnergyStats) string {
	output := "=== ENERGY USAGE REPORT ===\n"
	output += fmt.Sprintf("Period: %s\n\n", stats.Period)
	output += fmt.Sprintf("Total Energy Used: %.2f kWh\n", stats.TotalKWH)
	output += fmt.Sprintf("Total Runtime: %d minutes (%.1f hours)\n", stats.TotalRuntime, float64(stats.TotalRuntime)/60.0)
	output += fmt.Sprintf("\nBreakdown by Mode:\n")
	output += fmt.Sprintf("  Heating: %.2f kWh\n", stats.HeatingKWH)
	output += fmt.Sprintf("  Cooling: %.2f kWh\n", stats.CoolingKWH)
	output += fmt.Sprintf("  Fan: %.2f kWh\n", stats.FanKWH)
	output += fmt.Sprintf("\nEstimated Cost: $%.2f\n", stats.EstimatedCost)
	return output
}

func GetDailyEnergyUsage(date time.Time) (float64, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	var totalKWH float64
	err := db.QueryRow("SELECT COALESCE(SUM(estimated_kwh), 0) FROM energy_logs WHERE timestamp >= ? AND timestamp < ?", startOfDay, endOfDay).Scan(&totalKWH)
	if err != nil {
		return 0, err
	}
	return totalKWH, nil
}

func GetMonthlyEnergyUsage(year int, month time.Month) (float64, error) {
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)
	var totalKWH float64
	err := db.QueryRow("SELECT COALESCE(SUM(estimated_kwh), 0) FROM energy_logs WHERE timestamp >= ? AND timestamp < ?", startOfMonth, endOfMonth).Scan(&totalKWH)
	if err != nil {
		return 0, err
	}
	return totalKWH, nil
}

func TrackEnergyUsage(mode HVACMode, runtimeMinutes int) error {
	kwh := estimateEnergyUsage(mode, runtimeMinutes)
	_, err := db.Exec("INSERT INTO energy_logs (hvac_mode, runtime_minutes, estimated_kwh) VALUES (?, ?, ?)", mode, runtimeMinutes, kwh)
	if err != nil {
		return err
	}
	LogEvent("energy_track", fmt.Sprintf("Tracked %.2f kWh for %s mode", kwh, mode), "system", "info")
	return nil
}
