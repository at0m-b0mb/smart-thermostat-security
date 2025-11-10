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
    rand.Seed(time.Now().UnixNano())
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
    defer sensorMutex.RUnlock()
    if !sensorHealth {
        return 0, errors.New("sensor malfunction")
    }
    temp := 18.0 + rand.Float64()*10.0
    if temp < -50 || temp > 100 {
        errorCount++
        LogEvent("sensor_error", "Temperature out of range", "system", "warning")
        return 0, errors.New("invalid temperature")
    }
    // Only update timestamp and lastReading with a LOCK
    sensorMutex.RUnlock()
    sensorMutex.Lock()
    lastReading.Temperature = temp
    lastReading.Timestamp = time.Now()
    sensorMutex.Unlock()
    sensorMutex.RLock()
    db.Exec("INSERT INTO sensor_readings (temperature, humidity, co_level) VALUES (?, ?, ?)", temp, lastReading.Humidity, lastReading.CO)
    return temp, nil
}

func ReadHumidity() (float64, error) {
    sensorMutex.RLock()
    defer sensorMutex.RUnlock()
    if !sensorHealth {
        return 0, errors.New("sensor malfunction")
    }
    humidity := 30.0 + rand.Float64()*40.0
    if humidity < 0 || humidity > 100 {
        errorCount++
        LogEvent("sensor_error", "Humidity out of range", "system", "warning")
        return 0, errors.New("invalid humidity")
    }
    // Only update timestamp and lastReading with a LOCK
    sensorMutex.RUnlock()
    sensorMutex.Lock()
    lastReading.Humidity = humidity
    lastReading.Timestamp = time.Now()
    sensorMutex.Unlock()
    sensorMutex.RLock()
    return humidity, nil
}

func ReadCO() (float64, error) {
    sensorMutex.RLock()
    defer sensorMutex.RUnlock()
    if !sensorHealth {
        return 0, errors.New("sensor malfunction")
    }
    co := rand.Float64() * 10.0
    if co < 0 || co > 1000 {
        errorCount++
        LogEvent("sensor_error", "CO out of range", "system", "warning")
        return 0, errors.New("invalid CO")
    }
    if co > 50 {
        LogEvent("co_alert", "Dangerous CO level detected", "system", "critical")
    }
    // Only update timestamp and lastReading with a LOCK
    sensorMutex.RUnlock()
    sensorMutex.Lock()
    lastReading.CO = co
    lastReading.Timestamp = time.Now()
    sensorMutex.Unlock()
    sensorMutex.RLock()
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
