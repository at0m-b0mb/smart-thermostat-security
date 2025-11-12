package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type WeatherData struct {
	Temperature float64
	Humidity    float64
	Conditions  string
	Location    string
	Timestamp   time.Time
}

var cachedWeather WeatherData
var lastFetch time.Time
var cacheDuration = 10 * time.Minute

func GetOutdoorWeather(location string) (WeatherData, error) {
	if time.Since(lastFetch) < cacheDuration && cachedWeather.Location == location {
		LogEvent("weather_cache", "Weather from cache", "system", "info")
		return cachedWeather, nil
	}
	if len(location) < 2 || len(location) > 100 {
		return WeatherData{}, errors.New("invalid location")
	}
	weather := WeatherData{
		Temperature: 15.0 + float64(time.Now().Hour())/2 + rand.Float64()*5,
		Humidity:    60.0 + rand.Float64()*20,
		Conditions:  getRandomCondition(),
		Location:    location,
		Timestamp:   time.Now(),
	}
	cachedWeather = weather
	lastFetch = time.Now()
	LogEvent("weather_fetch", "Weather fetched for "+location, "system", "info")
	return weather, nil
}

func getRandomCondition() string {
	conditions := []string{"Clear", "Cloudy", "Rainy", "Sunny", "Partly Cloudy"}
	return conditions[rand.Intn(len(conditions))]
}

func DisplayWeather(weather WeatherData) string {
	return "Location: " + weather.Location + "\nTemperature: " + formatFloat(weather.Temperature) + "°C\nHumidity: " + formatFloat(weather.Humidity) + "%\nConditions: " + weather.Conditions + "\nUpdated: " + weather.Timestamp.Format("15:04:05")
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.1f", f)
}
