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
