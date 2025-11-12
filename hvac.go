package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type HVACMode string

const (
	ModeOff  HVACMode = "off"
	ModeHeat HVACMode = "heat"
	ModeCool HVACMode = "cool"
	ModeFan  HVACMode = "fan"
)

type HVACState struct {
	Mode        HVACMode
	TargetTemp  float64
	CurrentTemp float64
	IsRunning   bool
	LastUpdate  time.Time
}

var (
	hvacMutex     sync.RWMutex
	hvacState     HVACState
	startTime     time.Time
	lastEnergyLog time.Time
)

func InitializeHVAC() error {
	hvacMutex.Lock()
	defer hvacMutex.Unlock()
	hvacState = HVACState{
		Mode:        ModeOff,
		TargetTemp:  22.0,
		CurrentTemp: 20.0,
		IsRunning:   false,
		LastUpdate:  time.Now(),
	}
	LogEvent("hvac_init", "HVAC system initialized", "system", "info")
	return nil
}

func SetHVACMode(mode string, user *User) error {
	hvacMutex.Lock()
	defer hvacMutex.Unlock()

	// Sanitize mode input
	mode = SanitizeInput(mode)
	// --- Guest restriction removed ---
	hvacMode := HVACMode(mode)
	if hvacMode != ModeOff && hvacMode != ModeHeat && hvacMode != ModeCool && hvacMode != ModeFan {
		return errors.New("invalid HVAC mode")
	}
	oldMode := hvacState.Mode
	hvacState.Mode = hvacMode
	hvacState.LastUpdate = time.Now()
	if hvacMode == ModeOff {
		hvacState.IsRunning = false
	}
	db.Exec("INSERT INTO hvac_state (mode, target_temp, current_temp, is_running) VALUES (?, ?, ?, ?)",
		mode, hvacState.TargetTemp, hvacState.CurrentTemp, hvacState.IsRunning)
	LogEvent("hvac_mode_change", fmt.Sprintf("Mode changed from %s to %s", oldMode, hvacMode), user.Username, "info")
	return nil
}

func SetTargetTemperature(temp float64, user *User) error {
	hvacMutex.Lock()
	defer hvacMutex.Unlock()
	// Validate temperature using security.go function
	if err := ValidateTemperatureInput(temp); err != nil {
		AuditSecurityEvent("invalid_temp", fmt.Sprintf("Invalid temperature attempted: %.1f", temp), user.Username)
		return err
	}
	oldTemp := hvacState.TargetTemp
	hvacState.TargetTemp = temp
	hvacState.LastUpdate = time.Now()
	db.Exec("INSERT INTO hvac_state (mode, target_temp, current_temp, is_running) VALUES (?, ?, ?, ?)",
		hvacState.Mode, temp, hvacState.CurrentTemp, hvacState.IsRunning)
	LogEvent("hvac_temp_change", fmt.Sprintf("Target temp changed from %.1f to %.1f", oldTemp, temp), user.Username, "info")
	return nil
}

func GetHVACStatus() HVACState {
	hvacMutex.RLock()
	defer hvacMutex.RUnlock()
	return hvacState
}

func UpdateHVACLogic() error {
	hvacMutex.Lock()
	defer hvacMutex.Unlock()
	currentTemp, err := ReadTemperature()
	if err != nil {
		return err
	}
	hvacState.CurrentTemp = currentTemp
	if hvacState.Mode == ModeOff {
		if hvacState.IsRunning {
			logRuntime()
			hvacState.IsRunning = false
		}
		return nil
	}
	if hvacState.Mode == ModeHeat {
		if currentTemp < hvacState.TargetTemp-1.0 {
			if !hvacState.IsRunning {
				hvacState.IsRunning = true
				startTime = time.Now()
				LogEvent("hvac_start", "Heating started", "system", "info")
			} else {
				// Log periodic energy for long-running operations
				logPeriodicRuntime()
			}
		} else if currentTemp > hvacState.TargetTemp+0.5 {
			if hvacState.IsRunning {
				logRuntime()
				hvacState.IsRunning = false
				LogEvent("hvac_stop", "Heating stopped", "system", "info")
			}
		}
	} else if hvacState.Mode == ModeCool {
		if currentTemp > hvacState.TargetTemp+1.0 {
			if !hvacState.IsRunning {
				hvacState.IsRunning = true
				startTime = time.Now()
				LogEvent("hvac_start", "Cooling started", "system", "info")
			} else {
				// Log periodic energy for long-running operations
				logPeriodicRuntime()
			}
		} else if currentTemp < hvacState.TargetTemp-0.5 {
			if hvacState.IsRunning {
				logRuntime()
				hvacState.IsRunning = false
				LogEvent("hvac_stop", "Cooling stopped", "system", "info")
			}
		}
	} else if hvacState.Mode == ModeFan {
		if !hvacState.IsRunning {
			hvacState.IsRunning = true
			startTime = time.Now()
			LogEvent("hvac_start", "Fan started", "system", "info")
		} else {
			// Log periodic energy for fan mode
			logPeriodicRuntime()
		}
	}
	hvacState.LastUpdate = time.Now()
	return nil
}

func logRuntime() {
	if !startTime.IsZero() {
		runtime := int(time.Since(startTime).Minutes())
		if runtime > 0 {
			kwh := estimateEnergyUsage(hvacState.Mode, runtime)
			db.Exec("INSERT INTO energy_logs (hvac_mode, runtime_minutes, estimated_kwh) VALUES (?, ?, ?)",
				hvacState.Mode, runtime, kwh)
			LogEvent("energy_track", fmt.Sprintf("Tracked %.2f kWh for %s mode (%d minutes)", kwh, hvacState.Mode, runtime), "system", "info")
		}
		startTime = time.Time{}     // Reset startTime
		lastEnergyLog = time.Time{} // Reset last log time
	}
}

func logPeriodicRuntime() {
	if !startTime.IsZero() {
		// Only log every 2 minutes to avoid too many small entries
		timeSinceLastLog := time.Since(lastEnergyLog)
		if lastEnergyLog.IsZero() || timeSinceLastLog >= 2*time.Minute {
			runtime := int(time.Since(startTime).Minutes())
			if runtime > 0 {
				kwh := estimateEnergyUsage(hvacState.Mode, runtime)
				db.Exec("INSERT INTO energy_logs (hvac_mode, runtime_minutes, estimated_kwh) VALUES (?, ?, ?)",
					hvacState.Mode, runtime, kwh)
				LogEvent("energy_track", fmt.Sprintf("Tracked %.2f kWh for %s mode (%d minutes)", kwh, hvacState.Mode, runtime), "system", "info")
				// Reset startTime to track next period
				startTime = time.Now()
				lastEnergyLog = time.Now()
			}
		}
	}
}

func estimateEnergyUsage(mode HVACMode, runtimeMinutes int) float64 {
	kwhPerHour := 0.0
	switch mode {
	case ModeHeat:
		kwhPerHour = 2.5
	case ModeCool:
		kwhPerHour = 3.0
	case ModeFan:
		kwhPerHour = 0.5
	}
	return kwhPerHour * (float64(runtimeMinutes) / 60.0)
}
