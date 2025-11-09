package main

import (
	"fmt"
	"time"
)

type LogEntry struct {
	ID        int
	Timestamp time.Time
	EventType string
	Details   string
	Username  string
	Severity  string
}

func LogEvent(eventType, details, username, severity string) {
	if db == nil {
		fmt.Printf("[%s] %s: %s (%s)\n", time.Now().Format(time.RFC3339), eventType, details, username)
		return
	}
	_, err := db.Exec("INSERT INTO logs (event_type, details, username, severity) VALUES (?, ?, ?, ?)", eventType, details, username, severity)
	if err != nil {
		fmt.Printf("Error logging: %v\n", err)
	}
	fmt.Printf("[%s] %s: %s (%s)\n", time.Now().Format(time.RFC3339), eventType, details, username)
}

func ViewAuditTrail(limit int) ([]LogEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := db.Query("SELECT id, timestamp, event_type, details, username, severity FROM logs ORDER BY timestamp DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	logs := []LogEntry{}
	for rows.Next() {
		var log LogEntry
		if err := rows.Scan(&log.ID, &log.Timestamp, &log.EventType, &log.Details, &log.Username, &log.Severity); err != nil {
			continue
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func ViewAuditTrailByUser(username string, limit int) ([]LogEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := db.Query("SELECT id, timestamp, event_type, details, username, severity FROM logs WHERE username = ? ORDER BY timestamp DESC LIMIT ?", username, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	logs := []LogEntry{}
	for rows.Next() {
		var log LogEntry
		if err := rows.Scan(&log.ID, &log.Timestamp, &log.EventType, &log.Details, &log.Username, &log.Severity); err != nil {
			continue
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func GetSecurityAlerts() ([]LogEntry, error) {
	rows, err := db.Query("SELECT id, timestamp, event_type, details, username, severity FROM logs WHERE severity IN ('warning', 'critical') ORDER BY timestamp DESC LIMIT 50")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	logs := []LogEntry{}
	for rows.Next() {
		var log LogEntry
		if err := rows.Scan(&log.ID, &log.Timestamp, &log.EventType, &log.Details, &log.Username, &log.Severity); err != nil {
			continue
		}
		logs = append(logs, log)
	}
	return logs, nil
}
