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
    EcoMode     bool
}

type EcoModeStats struct {
    EnergySaved   float64
    CyclesSaved   int
    ActiveSince   time.Time
}

var (
    hvacMutex sync.RWMutex
    hvacState HVACState
    startTime time.Time
    ecoStats  EcoModeStats
)

func InitializeHVAC() error {
    hvacMutex.Lock()
    defer hvacMutex.Unlock()
    hvacState = HVACState{
        Mode:       ModeOff,
        TargetTemp: 22.0,
        CurrentTemp: 20.0,
        IsRunning:   false,
        LastUpdate:  time.Now(),
        EcoMode:    false,
    }
    ecoStats = EcoModeStats{
        EnergySaved: 0,
        CyclesSaved: 0,
        ActiveSince: time.Time{},
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
    return UpdateHVACLogicWithEco()
}

func logRuntime() {
    if !startTime.IsZero() {
        runtime := int(time.Since(startTime).Minutes())
        kwh := estimateEnergyUsage(hvacState.Mode, runtime)
        db.Exec("INSERT INTO energy_logs (hvac_mode, runtime_minutes, estimated_kwh) VALUES (?, ?, ?)",
            hvacState.Mode, runtime, kwh)
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

// SetEcoMode enables or disables eco mode for energy efficiency
func SetEcoMode(enable bool, user *User) error {
    hvacMutex.Lock()
    defer hvacMutex.Unlock()
    
    // Only homeowners can toggle eco mode
    if user.Role != "homeowner" {
        return errors.New("only homeowners can change eco mode")
    }
    
    hvacState.EcoMode = enable
    
    if enable {
        ecoStats.ActiveSince = time.Now()
        ecoStats.EnergySaved = 0
        ecoStats.CyclesSaved = 0
        LogEvent("eco_mode_enable", "Eco mode enabled", user.Username, "info")
        SendNotification(user.Username, "eco_mode", "Eco mode enabled. Temperature tolerance increased for energy savings.")
    } else {
        LogEvent("eco_mode_disable", fmt.Sprintf("Eco mode disabled. Saved %.2f kWh, %d cycles avoided", 
            ecoStats.EnergySaved, ecoStats.CyclesSaved), user.Username, "info")
        SendNotification(user.Username, "eco_mode", 
            fmt.Sprintf("Eco mode disabled. Energy saved: %.2f kWh", ecoStats.EnergySaved))
    }
    
    return nil
}

// GetEcoModeStatus returns whether eco mode is active
func GetEcoModeStatus() (bool, EcoModeStats) {
    hvacMutex.RLock()
    defer hvacMutex.RUnlock()
    return hvacState.EcoMode, ecoStats
}

// UpdateHVACLogicWithEco updates HVAC logic considering eco mode
func UpdateHVACLogicWithEco() error {
    hvacMutex.Lock()
    defer hvacMutex.Unlock()
    
    currentTemp, err := ReadTemperature()
    if err != nil {
        return err
    }
    hvacState.CurrentTemp = currentTemp
    
    if hvacState.Mode == ModeOff {
        hvacState.IsRunning = false
        return nil
    }
    
    // Determine temperature thresholds based on eco mode
    heatThreshold := 1.0
    coolThreshold := 1.0
    heatStopThreshold := 0.5
    coolStopThreshold := 0.5
    
    if hvacState.EcoMode {
        // In eco mode, allow wider temperature variance (±2°C)
        heatThreshold = 2.0
        coolThreshold = 2.0
        heatStopThreshold = 1.0
        coolStopThreshold = 1.0
    }
    
    if hvacState.Mode == ModeHeat {
        if currentTemp < hvacState.TargetTemp-heatThreshold {
            if !hvacState.IsRunning {
                hvacState.IsRunning = true
                startTime = time.Now()
                LogEvent("hvac_start", "Heating started", "system", "info")
            }
        } else if currentTemp > hvacState.TargetTemp+heatStopThreshold {
            if hvacState.IsRunning {
                logRuntime()
                hvacState.IsRunning = false
                LogEvent("hvac_stop", "Heating stopped", "system", "info")
                
                // Track eco savings if in eco mode
                if hvacState.EcoMode {
                    ecoStats.CyclesSaved++
                    ecoStats.EnergySaved += 0.15 // Estimate ~0.15 kWh saved per avoided cycle
                }
            }
        }
    } else if hvacState.Mode == ModeCool {
        if currentTemp > hvacState.TargetTemp+coolThreshold {
            if !hvacState.IsRunning {
                hvacState.IsRunning = true
                startTime = time.Now()
                LogEvent("hvac_start", "Cooling started", "system", "info")
            }
        } else if currentTemp < hvacState.TargetTemp-coolStopThreshold {
            if hvacState.IsRunning {
                logRuntime()
                hvacState.IsRunning = false
                LogEvent("hvac_stop", "Cooling stopped", "system", "info")
                
                // Track eco savings if in eco mode
                if hvacState.EcoMode {
                    ecoStats.CyclesSaved++
                    ecoStats.EnergySaved += 0.18 // Estimate ~0.18 kWh saved per avoided cycle
                }
            }
        }
    } else if hvacState.Mode == ModeFan {
        hvacState.IsRunning = true
    }
    
    hvacState.LastUpdate = time.Now()
    return nil
}

// DisplayEcoModeStatus formats eco mode status for display
func DisplayEcoModeStatus() string {
    isEco, stats := GetEcoModeStatus()
    
    if !isEco {
        return "Eco Mode: Disabled"
    }
    
    duration := time.Since(stats.ActiveSince)
    hours := int(duration.Hours())
    minutes := int(duration.Minutes()) % 60
    
    return fmt.Sprintf(`Eco Mode: ENABLED
Active Duration: %dh %dm
Energy Saved: %.2f kWh
HVAC Cycles Avoided: %d
Estimated Cost Savings: $%.2f`,
        hours, minutes,
        stats.EnergySaved,
        stats.CyclesSaved,
        stats.EnergySaved * 0.12) // Assuming $0.12 per kWh
}
