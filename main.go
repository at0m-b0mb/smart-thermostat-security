package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var currentUser *User

func main() {
	fmt.Println("==============================================")
	fmt.Println("   SMART THERMOSTAT SECURITY SYSTEM v1.0")
	fmt.Println("   Secure-by-Design IoT Thermostat")
	fmt.Println("   Team Logan - EN601.643")
	fmt.Println("==============================================")
	fmt.Println()

	// Initialize database
	if err := InitializeDatabase(); err != nil {
		fmt.Printf("FATAL: Database initialization failed: %v\n", err)
		os.Exit(1)
	}
	defer CloseDatabase()

	// Initialize sensors
	if err := InitializeSensors(); err != nil {
		fmt.Printf("ERROR: Sensor initialization failed: %v\n", err)
	}

	// Initialize HVAC
	if err := InitializeHVAC(); err != nil {
		fmt.Printf("ERROR: HVAC initialization failed: %v\n", err)
	}

	// Setup graceful shutdown
	setupGracefulShutdown()

	// Start background tasks
	go hvacControlLoop()
	go sensorMonitorLoop()

	// Main CLI loop
	runCLI()
}

func setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\n\nShutting down gracefully...")
		CloseDatabase()
		os.Exit(0)
	}()
}

func hvacControlLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if err := UpdateHVACLogic(); err != nil {
			LogEvent("hvac_error", "HVAC update failed: "+err.Error(), "system", "warning")
		}
	}
}

func sensorMonitorLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if _, err := ReadAllSensors(); err != nil {
			LogEvent("sensor_error", "Sensor read failed: "+err.Error(), "system", "warning")
		}
	}
}

func runCLI() {
	reader := bufio.NewReader(os.Stdin)

	for {
		if currentUser == nil {
			fmt.Println("\n--- LOGIN REQUIRED ---")
			fmt.Print("Username: ")
			username, _ := reader.ReadString('\n')
			username = strings.TrimSpace(username)

			fmt.Print("Password: ")
			password, _ := reader.ReadString('\n')
			password = strings.TrimSpace(password)

			user, err := AuthenticateUser(username, password)
			if err != nil {
				fmt.Printf("Login failed: %v\n", err)
				continue
			}
			currentUser = user
			fmt.Printf("\nWelcome, %s! (Role: %s)\n", currentUser.Username, currentUser.Role)
		}

		displayMenu()
		fmt.Print("\nEnter choice: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		handleMenuChoice(choice, reader)
	}
}

func displayMenu() {
	fmt.Println("\n=== MAIN MENU ===")
	fmt.Println("1.  View Current Status")
	fmt.Println("2.  Set Target Temperature")
	fmt.Println("3.  Change HVAC Mode")
	fmt.Println("4.  View Sensor Readings")
	fmt.Println("5.  View Weather")
	fmt.Println("6.  View Energy Usage")
	fmt.Println("7.  Manage Profiles")
	fmt.Println("8.  Manage Users")
	fmt.Println("9.  Run Diagnostics")
	fmt.Println("10. View Audit Logs")
	fmt.Println("11. Change Password")
	fmt.Println("12. Logout")
	fmt.Println("0.  Exit")
}

func handleMenuChoice(choice string, reader *bufio.Reader) {
	switch choice {
	case "1":
		viewCurrentStatus()
	case "2":
		setTargetTemperature(reader)
	case "3":
		changeHVACMode(reader)
	case "4":
		viewSensorReadings()
	case "5":
		viewWeather(reader)
	case "6":
		viewEnergyUsage(reader)
	case "7":
		manageProfiles(reader)
	case "8":
		manageUsers(reader)
	case "9":
		runDiagnostics()
	case "10":
		viewAuditLogs()
	case "11":
		changePasswordCLI(reader)
	case "12":
		logout()
	case "0":
		fmt.Println("Goodbye!")
		CloseDatabase()
		os.Exit(0)
	default:
		fmt.Println("Invalid choice")
	}
}

func viewCurrentStatus() {
	fmt.Println("\n=== CURRENT SYSTEM STATUS ===")
	status := GetHVACStatus()
	fmt.Printf("HVAC Mode: %s\n", status.Mode)
	fmt.Printf("Target Temperature: %.1f°C\n", status.TargetTemp)
	fmt.Printf("Current Temperature: %.1f°C\n", status.CurrentTemp)
	fmt.Printf("System Running: %v\n", status.IsRunning)
	fmt.Printf("Last Update: %s\n", status.LastUpdate.Format(time.RFC3339))
}

func setTargetTemperature(reader *bufio.Reader) {
	fmt.Print("Enter target temperature (10-35°C): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	temp, err := strconv.ParseFloat(input, 64)
	if err != nil {
		fmt.Println("Invalid temperature")
		return
	}
	if err := SetTargetTemperature(temp, currentUser); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Target temperature set to %.1f°C\n", temp)
}

func changeHVACMode(reader *bufio.Reader) {
	fmt.Println("Select HVAC Mode:")
	fmt.Println("1. Off")
	fmt.Println("2. Heat")
	fmt.Println("3. Cool")
	fmt.Println("4. Fan")
	fmt.Print("Choice: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	modes := map[string]string{"1": "off", "2": "heat", "3": "cool", "4": "fan"}
	mode, ok := modes[input]
	if !ok {
		fmt.Println("Invalid mode")
		return
	}

	if err := SetHVACMode(mode, currentUser); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("HVAC mode set to %s\n", mode)
}

func viewSensorReadings() {
	fmt.Println("\n=== SENSOR READINGS ===")
	reading, err := ReadAllSensors()
	if err != nil {
		fmt.Printf("Error reading sensors: %v\n", err)
		return
	}
	fmt.Printf("Temperature: %.1f°C\n", reading.Temperature)
	fmt.Printf("Humidity: %.1f%%\n", reading.Humidity)
	fmt.Printf("CO Level: %.2f ppm\n", reading.CO)
	fmt.Printf("Timestamp: %s\n", reading.Timestamp.Format(time.RFC3339))
}

func viewWeather(reader *bufio.Reader) {
	fmt.Print("Enter location: ")
	location, _ := reader.ReadString('\n')
	location = strings.TrimSpace(location)
	weather, err := GetOutdoorWeather(location)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("\n=== OUTDOOR WEATHER ===")
	fmt.Println(DisplayWeather(weather))
}

func viewEnergyUsage(reader *bufio.Reader) {
	fmt.Print("Enter number of days (default 7): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	days := 7
	if input != "" {
		if d, err := strconv.Atoi(input); err == nil {
			days = d
		}
	}
	stats, err := GetEnergyUsage(days)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("\n" + GenerateEnergyReport(stats))
}

func manageProfiles(reader *bufio.Reader) {
	fmt.Println("\n=== PROFILE MANAGEMENT ===")
	fmt.Println("1. List Profiles")
	fmt.Println("2. Create Profile")
	fmt.Println("3. Apply Profile")
	fmt.Println("4. Delete Profile")
	fmt.Print("Choice: ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		profiles, err := ListProfiles(currentUser.Username)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Println("\nYour Profiles:")
		for _, p := range profiles {
			fmt.Printf("- %s (Temp: %.1f°C, Mode: %s)\n", p.Name, p.TargetTemp, p.HVACMode)
		}
	case "2":
		createProfile(reader)
	case "3":
		applyProfile(reader)
	case "4":
		deleteProfile(reader)
	}
}

func createProfile(reader *bufio.Reader) {
	fmt.Print("Profile name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Target temperature (10-35°C): ")
	tempStr, _ := reader.ReadString('\n')
	temp, err := strconv.ParseFloat(strings.TrimSpace(tempStr), 64)
	if err != nil {
		fmt.Println("Invalid temperature")
		return
	}

	fmt.Print("HVAC mode (off/heat/cool/fan): ")
	mode, _ := reader.ReadString('\n')
	mode = strings.TrimSpace(mode)

	if err := CreateProfile(name, temp, mode, currentUser.Username); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Profile created successfully")
}

func applyProfile(reader *bufio.Reader) {
	fmt.Print("Profile name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	if err := ApplyProfile(name, currentUser); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Profile applied successfully")
}

func deleteProfile(reader *bufio.Reader) {
	fmt.Print("Profile name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	if err := DeleteProfile(name, currentUser.Username); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Profile deleted successfully")
}

func manageUsers(reader *bufio.Reader) {
	if currentUser.Role != "homeowner" {
		fmt.Println("Only homeowners can manage users")
		return
	}

	fmt.Println("\n=== USER MANAGEMENT ===")
	fmt.Println("1. List Users")
	fmt.Println("2. Create Guest")
	fmt.Println("3. Create Technician")
	fmt.Println("4. Grant Technician Access")
	fmt.Println("5. Revoke Access")
	
	fmt.Print("Choice: ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		listUsers()
	case "2":
		createGuest(reader)
	case "3":
		createTechnician(reader) 
	case "4":
		grantTechAccess(reader)
	case "5":
		revokeUserAccess(reader)
	}
}

func listUsers() {
	users, err := ListAllUsers()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("\nAll Users:")
	for _, u := range users {
		status := "Active"
		if !u.IsActive {
			status = "Inactive"
		}
		fmt.Printf("- %s (Role: %s, Status: %s)\n", u.Username, u.Role, status)
	}
}

func createGuest(reader *bufio.Reader) {
	fmt.Print("Guest name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Guest PIN (min 4): ")
	pin, _ := reader.ReadString('\n')
	pin = strings.TrimSpace(pin)

	if err := CreateGuestAccount(currentUser.Username, name, pin); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Guest account created: %s_guest_%s\n", currentUser.Username, name)
}

func createTechnician(reader *bufio.Reader) {
    fmt.Print("Technician name: ")
    name, _ := reader.ReadString('\n')
    name = strings.TrimSpace(name)
    fmt.Print("Technician password (min 8 chars): ")
    password, _ := reader.ReadString('\n')
    password = strings.TrimSpace(password)
    if err := CreateTechnicianAccount(currentUser.Username, name, password); err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("Technician account created: %s\n", name)
}

func grantTechAccess(reader *bufio.Reader) {
	fmt.Print("Technician username: ")
	tech, _ := reader.ReadString('\n')
	tech = strings.TrimSpace(tech)

	fmt.Print("Duration in hours: ")
	durationStr, _ := reader.ReadString('\n')
	hours, err := strconv.Atoi(strings.TrimSpace(durationStr))
	if err != nil {
		fmt.Println("Invalid duration")
		return
	}

	if err := GrantTechnicianAccess(currentUser.Username, tech, time.Duration(hours)*time.Hour); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Technician access granted")
}

func revokeUserAccess(reader *bufio.Reader) {
	fmt.Print("Username to revoke: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	if err := RevokeAccess(username); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Access revoked for %s\n", username)
}

func runDiagnostics() {
	fmt.Println("\n=== RUNNING SYSTEM DIAGNOSTICS ===")
	report, err := RunSystemDiagnostics(currentUser)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println(GenerateDiagnosticReport(report))
}

func viewAuditLogs() {
	if currentUser.Role != "homeowner" && currentUser.Role != "technician" {
		fmt.Println("Insufficient permissions")
		return
	}
	fmt.Println("\n=== AUDIT LOGS (Last 20 entries) ===")
	logs, err := ViewAuditTrail(20)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	for _, log := range logs {
		fmt.Printf("[%s] %s - %s (%s) [%s]\n",
			log.Timestamp.Format("2006-01-02 15:04:05"),
			log.EventType,
			log.Details,
			log.Username,
			log.Severity)
	}
}

func changePasswordCLI(reader *bufio.Reader) {
	// Check if user is a guest - they use PINs, not passwords
	if currentUser.Role == "guest" {
		// Guest PIN change flow
		fmt.Print("Current PIN: ")
		oldPIN, _ := reader.ReadString('\n')
		oldPIN = strings.TrimSpace(oldPIN)

		fmt.Print("New PIN (numeric, min 4 digits): ")
		newPIN, _ := reader.ReadString('\n')
		newPIN = strings.TrimSpace(newPIN)

		fmt.Print("Confirm new PIN: ")
		confirmPIN, _ := reader.ReadString('\n')
		confirmPIN = strings.TrimSpace(confirmPIN)

		if newPIN != confirmPIN {
			fmt.Println("PINs do not match")
			return
		}

		if err := ChangePIN(currentUser.Username, oldPIN, newPIN); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Println("PIN changed successfully")
	} else {
		// Homeowner/Technician password change flow
		fmt.Print("Current password: ")
		oldPass, _ := reader.ReadString('\n')
		oldPass = strings.TrimSpace(oldPass)

		fmt.Print("New password: ")
		newPass, _ := reader.ReadString('\n')
		newPass = strings.TrimSpace(newPass)

		fmt.Print("Confirm new password: ")
		confirmPass, _ := reader.ReadString('\n')
		confirmPass = strings.TrimSpace(confirmPass)

		if newPass != confirmPass {
			fmt.Println("Passwords do not match")
			return
		}

		if err := ChangePassword(currentUser.Username, oldPass, newPass); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Println("Password changed successfully")
	}
}
func logout() {
	if currentUser != nil {
		LogoutUser(currentUser.Username)
		fmt.Printf("Goodbye, %s!\n", currentUser.Username)
		currentUser = nil
	}
}
