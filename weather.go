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

// package main

// import (
//     "encoding/json"         // For decoding JSON from the API
//     "errors"                // For meaningful error returns
//     "fmt"                   // For string formatting
//     "io/ioutil"             // For reading HTTP response body
//     "net/http"              // For HTTP requests
//     "time"                  // For timestamps and cache expiration
// )

// // WeatherData holds the weather information retrieved from the API
// type WeatherData struct {
//     Temperature float64   // Temperature in Celsius
//     Humidity    float64   // Humidity as a percentage
//     Conditions  string    // Weather condition description
//     Location    string    // Name of the location
//     Timestamp   time.Time // Time when the weather was fetched
// }

// // Cached weather and timing
// var cachedWeather WeatherData
// var lastFetch time.Time
// var cacheDuration = 10 * time.Minute

// // Substitute this with your actual OpenWeatherMap API key
// const openWeatherMapAPIKey = "YOUR_API_KEY_HERE"

// // GetOutdoorWeather fetches real weather for a given location, returns WeatherData struct
// func GetOutdoorWeather(location string) (WeatherData, error) {
//     // Check cache for recent data for the same location
//     if time.Since(lastFetch) < cacheDuration && cachedWeather.Location == location {
//         LogEvent("weather_cache", "Weather served from cache", "system", "info")
//         return cachedWeather, nil
//     }
//     // Input validation: basic sanity for location string
//     if len(location) < 2 || len(location) > 100 {
//         return WeatherData{}, errors.New("invalid location name")
//     }
//     // Compose the API request URL (metric units)
//     url := fmt.Sprintf(
//         "https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric",
//         location, openWeatherMapAPIKey)
//     // Make HTTP GET request
//     resp, err := http.Get(url)
//     if err != nil {
//         return WeatherData{}, fmt.Errorf("failed to fetch weather data: %v", err)
//     }
//     defer resp.Body.Close()
//     if resp.StatusCode != http.StatusOK {
//         return WeatherData{}, fmt.Errorf("weather API error: status code %d", resp.StatusCode)
//     }
//     // Read response body
//     bodyBytes, err := ioutil.ReadAll(resp.Body)
//     if err != nil {
//         return WeatherData{}, fmt.Errorf("failed to read weather response: %v", err)
//     }
//     // Minimal struct to parse only needed fields from API JSON
//     var apiResp struct {
//         Main struct {
//             Temp     float64 `json:"temp"`
//             Humidity float64 `json:"humidity"`
//         } `json:"main"`
//         Weather []struct {
//             Description string `json:"description"`
//         } `json:"weather"`
//         Name string `json:"name"`
//     }
//     if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
//         return WeatherData{}, fmt.Errorf("failed to parse weather JSON: %v", err)
//     }
//     // Build result struct with parsed data
//     weather := WeatherData{
//         Temperature: apiResp.Main.Temp,
//         Humidity:    apiResp.Main.Humidity,
//         Conditions:  "",
//         Location:    apiResp.Name,
//         Timestamp:   time.Now(),
//     }
//     if len(apiResp.Weather) > 0 {
//         weather.Conditions = apiResp.Weather[0].Description
//     } else {
//         weather.Conditions = "Unknown"
//     }
//     // Update and record cache
//     cachedWeather = weather
//     lastFetch = time.Now()
//     LogEvent("weather_fetch", "Real weather fetched for "+location, "system", "info")
//     return weather, nil
// }

// // DisplayWeather returns a formatted string for displaying weather info
// func DisplayWeather(weather WeatherData) string {
//     return fmt.Sprintf(
//         "Location: %s\nTemperature: %.1f°C\nHumidity: %.1f%%\nConditions: %s\nUpdated: %s",
//         weather.Location,
//         weather.Temperature,
//         weather.Humidity,
//         weather.Conditions,
//         weather.Timestamp.Format("15:04:05"))
// }

// // LogEvent is part of your existing logging.go (called here for audit)
// func LogEvent(eventType, details, username, severity string) {
//     // Assume this logs to the audit trail
// }
