package main

import (
    "errors"
    "math/rand"
    "sync"
    "time"
)

type SensorReading struct {
    Temperature float64
    Humidity    float64
    CO          float64
    Timestamp   time.Time
}

type SensorStatus struct {
    IsHealthy   bool
    LastReading time.Time
    ErrorCount  int
    SensorType  string
}

var (
    sensorMutex  sync.RWMutex
    lastReading  SensorReading
    sensorHealth = true
    errorCount   = 0
)

func InitializeSensors() error {
    sensorMutex.Lock()
    defer sensorMutex.Unlock()
    // Note: math/rand is automatically seeded in Go 1.20+
    // No need for rand.Seed() anymore
    lastReading = SensorReading{
        Temperature: 20.0,
        Humidity:    50.0,
        CO:          0.0,
        Timestamp:   time.Now(),
    }
    LogEvent("sensor_init", "Sensors initialized", "system", "info")
    return nil
}

func ReadTemperature() (float64, error) {
    sensorMutex.RLock()
    if !sensorHealth {
        sensorMutex.RUnlock()
        return 0, errors.New("sensor malfunction")
    }
    sensorMutex.RUnlock()
    
    temp := 18.0 + rand.Float64()*10.0
    if temp < -50 || temp > 100 {
        sensorMutex.Lock()
        errorCount++
        sensorMutex.Unlock()
        LogEvent("sensor_error", "Temperature out of range", "system", "warning")
        return 0, errors.New("invalid temperature")
    }
    
    sensorMutex.Lock()
    lastReading.Temperature = temp
    lastReading.Timestamp = time.Now()
    humidity := lastReading.Humidity
    co := lastReading.CO
    sensorMutex.Unlock()
    
    db.Exec("INSERT INTO sensor_readings (temperature, humidity, co_level) VALUES (?, ?, ?)", temp, humidity, co)
    return temp, nil
}

func ReadHumidity() (float64, error) {
    sensorMutex.RLock()
    if !sensorHealth {
        sensorMutex.RUnlock()
        return 0, errors.New("sensor malfunction")
    }
    sensorMutex.RUnlock()
    
    humidity := 30.0 + rand.Float64()*40.0
    if humidity < 0 || humidity > 100 {
        sensorMutex.Lock()
        errorCount++
        sensorMutex.Unlock()
        LogEvent("sensor_error", "Humidity out of range", "system", "warning")
        return 0, errors.New("invalid humidity")
    }
    
    sensorMutex.Lock()
    lastReading.Humidity = humidity
    lastReading.Timestamp = time.Now()
    sensorMutex.Unlock()
    
    return humidity, nil
}

func ReadCO() (float64, error) {
    sensorMutex.RLock()
    if !sensorHealth {
        sensorMutex.RUnlock()
        return 0, errors.New("sensor malfunction")
    }
    sensorMutex.RUnlock()
    
    co := rand.Float64() * 10.0
    if co < 0 || co > 1000 {
        sensorMutex.Lock()
        errorCount++
        sensorMutex.Unlock()
        LogEvent("sensor_error", "CO out of range", "system", "warning")
        return 0, errors.New("invalid CO")
    }
    if co > 50 {
        LogEvent("co_alert", "Dangerous CO level detected", "system", "critical")
    }
    
    sensorMutex.Lock()
    lastReading.CO = co
    lastReading.Timestamp = time.Now()
    sensorMutex.Unlock()
    
    return co, nil
}

func ReadAllSensors() (SensorReading, error) {
    // No sensorMutex.Lock() here. Each function handles its own lock.
    temp, err1 := ReadTemperature()
    humidity, err2 := ReadHumidity()
    co, err3 := ReadCO()
    if err1 != nil || err2 != nil || err3 != nil {
        return SensorReading{}, errors.New("sensor read failed")
    }
    reading := SensorReading{
        Temperature: temp,
        Humidity:    humidity,
        CO:          co,
        Timestamp:   time.Now(),
    }
    // Optionally update lastReading atomically here, if needed:
    sensorMutex.Lock()
    lastReading = reading
    sensorMutex.Unlock()
    return reading, nil
}

func GetSensorStatus() SensorStatus {
    sensorMutex.RLock()
    defer sensorMutex.RUnlock()
    return SensorStatus{
        IsHealthy:   sensorHealth,
        LastReading: lastReading.Timestamp,
        ErrorCount:  errorCount,
        SensorType:  "Temperature/Humidity/CO",
    }
}

func SimulateSensorFailure() {
    sensorMutex.Lock()
    defer sensorMutex.Unlock()
    sensorHealth = false
    LogEvent("sensor_failure", "Sensor failure simulated", "system", "warning")
}

func ResetSensor() error {
    sensorMutex.Lock()
    defer sensorMutex.Unlock()
    sensorHealth = true
    errorCount = 0
    LogEvent("sensor_reset", "Sensor system reset", "system", "info")
    return nil
}
