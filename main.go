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
	go sessionCleanupLoop()
	go awayModeCheckLoop()
	go maintenanceCheckLoop()

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

func sessionCleanupLoop() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		if err := CleanExpiredSessions(); err != nil {
			LogEvent("cleanup_error", "Session cleanup failed: "+err.Error(), "system", "warning")
		}
	}
}

func awayModeCheckLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		if err := CheckAwayModeReturn(); err != nil {
			LogEvent("away_mode_error", "Away mode check failed: "+err.Error(), "system", "warning")
		}
	}
}

func maintenanceCheckLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if err := CheckAndUpdateMaintenance(); err != nil {
			LogEvent("maintenance_error", "Maintenance check failed: "+err.Error(), "system", "warning")
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
	
	if currentUser.Role == "homeowner" || currentUser.Role == "technician" {
  		fmt.Println("2.  Set Target Temperature")    
	}
  
    fmt.Println("3.  Change HVAC Mode")
    fmt.Println("4.  View Sensor Readings")
    fmt.Println("5.  View Weather")

    // Homeowner and technician can view energy usage
    if currentUser.Role == "homeowner" || currentUser.Role == "technician" {
        fmt.Println("6.  View Energy Usage")
        fmt.Println("7.  Manage Profiles")
        fmt.Println("8.  Manage Users")
        fmt.Println("9.  Run Diagnostics")
    } else {
        // For guests, allow only these:
        fmt.Println("6.  Manage Profiles")
    }

    // Only homeowner can view audit logs and manage advanced features
    if currentUser.Role == "homeowner" {
        fmt.Println("10. View Audit Logs")
        fmt.Println("13. Vacation/Away Mode")
        fmt.Println("14. Filter Maintenance")
        fmt.Println("15. Eco Mode Settings")
    }

    fmt.Println("11. Change Password")
    fmt.Println("12. Logout")
    fmt.Println("0.  Exit")
}

func handleMenuChoice(choice string, reader *bufio.Reader) {
    switch choice {
    case "1":
        viewCurrentStatus()
    case "2":
		if currentUser.Role == "homeowner" || currentUser.Role == "technician" {
        	setTargetTemperature(reader)
        } else {
            fmt.Println("Invalid choice")
        }
    case "3":
        changeHVACMode(reader)
    case "4":
        viewSensorReadings()
    case "5":
        viewWeather(reader)

    // Homeowner or technician only
    case "6":
        if currentUser.Role == "homeowner" || currentUser.Role == "technician" {
            viewEnergyUsage(reader)
        } else if currentUser.Role == "guest" {
            manageProfiles(reader, currentUser)
        } else {
            fmt.Println("Invalid choice")
        }

    case "7":
        if currentUser.Role == "homeowner" || currentUser.Role == "technician" {
            manageProfiles(reader, currentUser)
        } else {
            fmt.Println("Invalid choice")
        }
    case "8":
        if currentUser.Role == "homeowner" || currentUser.Role == "technician" {
            manageUsers(reader)
        } else {
            fmt.Println("Invalid choice")
        }
    case "9":
        if currentUser.Role == "homeowner" || currentUser.Role == "technician" {
            runDiagnostics()
        } else {
            fmt.Println("Invalid choice")
        }
    case "10":
        if currentUser.Role == "homeowner" {
            viewAuditLogs()
        } else {
            fmt.Println("Invalid choice")
        }
    case "11":
        changePasswordCLI(reader)
    case "12":
        logout()
    case "13":
        if currentUser.Role == "homeowner" {
            manageAwayMode(reader)
        } else {
            fmt.Println("Invalid choice")
        }
    case "14":
        if currentUser.Role == "homeowner" {
            manageFilterMaintenance(reader)
        } else {
            fmt.Println("Invalid choice")
        }
    case "15":
        if currentUser.Role == "homeowner" {
            manageEcoMode(reader)
        } else {
            fmt.Println("Invalid choice")
        }
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
	fmt.Printf("Eco Mode: %v\n", status.EcoMode)
	fmt.Printf("Last Update: %s\n", status.LastUpdate.Format(time.RFC3339))
	
	// Show away mode status if homeowner
	if currentUser.Role == "homeowner" {
		awayStatus, err := GetAwayModeStatus()
		if err == nil && awayStatus != nil {
			fmt.Printf("\nAway Mode: ACTIVE (Return: %s)\n", awayStatus.ReturnTime.Format("2006-01-02 15:04"))
		}
	}
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
		
	// Validate temperature using security.go function
	if err := ValidateTemperatureInput(temp); err != nil {
		fmt.Printf("Security validation failed: %v\n", err)
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

func manageProfiles(reader *bufio.Reader, currentUser *User) {
    for {
        fmt.Println("\n=== PROFILE MANAGEMENT ===")
        fmt.Println("1. List Profiles")
        fmt.Println("2. Apply Profile")
        if currentUser.Role == "homeowner" || currentUser.Role == "technician" {
            fmt.Println("3. Create Profile")
            fmt.Println("4. Delete Profile")
            fmt.Println("5. Add Schedule")
			fmt.Println("6. View Schedules")

        }
        fmt.Println("0. Back to Main Menu")
        fmt.Print("Enter choice: ")

        choice, _ := reader.ReadString('\n')
        choice = strings.TrimSpace(choice)

        switch choice {
        case "1":
            listProfiles(reader)
        case "2":
            applyProfile(reader)
        case "3":
            if currentUser.Role != "homeowner" && currentUser.Role != "technician" {
                fmt.Println("Invalid choice")
                continue
            }
            createProfile(reader)
        case "4":
            if currentUser.Role != "homeowner" && currentUser.Role != "technician" {
                fmt.Println("Invalid choice")
                continue
            }
            deleteProfile(reader)
        case "5":
            if currentUser.Role != "homeowner" && currentUser.Role != "technician" {
                fmt.Println("Invalid choice")
                continue
            }
            addSchedule(reader)
        case "6":
            if currentUser.Role != "homeowner" && currentUser.Role != "technician" {
                fmt.Println("Invalid choice")
                continue
            }
            viewSchedules(reader)
        case "0":
            return
        default:
            fmt.Println("Invalid choice")
        }
    }
}

func listProfiles(reader *bufio.Reader) {
	profiles, err := ListProfiles(currentUser.Username, currentUser)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Name    | Temp (°C) | Mode | Guest Accessible")
	fmt.Println("-----------------------------------------------")
	for _, p := range profiles {
	    guest := "No"
	    if p.GuestAccessible == 1 {
	        guest = "Yes"
	    }
	    fmt.Printf("%-8s | %9.1f | %-4s | %-15s\n", p.Name, p.TargetTemp, p.HVACMode, guest)
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

    // FIX: Add this section to get HVAC mode!
    fmt.Print("HVAC mode (off/heat/cool/fan): ")
    mode, _ := reader.ReadString('\n')
    mode = strings.TrimSpace(mode)

    // New: prompt for guest accessibility
    fmt.Print("Allow guests to view/apply this profile? (yes/no): ")
    guestInput, _ := reader.ReadString('\n')
    guestInput = strings.TrimSpace(strings.ToLower(guestInput))
    guestAccessible := 0
    if guestInput == "yes" || guestInput == "y" {
        guestAccessible = 1
    }

    // Call CreateProfile with the new parameter
    if err := CreateProfile(name, temp, mode, currentUser.Username, currentUser, guestAccessible); err != nil {
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

	if err := DeleteProfile(name, currentUser.Username, currentUser.Role); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Profile deleted successfully")
}

func addSchedule(reader *bufio.Reader) {
    fmt.Print("Profile ID: ")
    profileIDStr, _ := reader.ReadString('\n')
    profileID, _ := strconv.Atoi(strings.TrimSpace(profileIDStr))
    fmt.Print("Day of Week (0=Sun, 6=Sat): ")
    dayStr, _ := reader.ReadString('\n')
    dayOfWeek, _ := strconv.Atoi(strings.TrimSpace(dayStr))
    fmt.Print("Start Time (HH:MM): ")
    startTime, _ := reader.ReadString('\n')
    startTime = strings.TrimSpace(startTime)
    fmt.Print("End Time (HH:MM): ")
    endTime, _ := reader.ReadString('\n')
    endTime = strings.TrimSpace(endTime)
    fmt.Print("Target Temperature (Celsius): ")
    targetStr, _ := reader.ReadString('\n')
    targetTemp, _ := strconv.ParseFloat(strings.TrimSpace(targetStr), 64)

    err := AddSchedule(profileID, dayOfWeek, startTime, endTime, targetTemp, currentUser)
    if err != nil {
        fmt.Printf("Error adding schedule: %v\n", err)
    } else {
        fmt.Println("Schedule added successfully.")
    }
}

func viewSchedules(reader *bufio.Reader) {
    fmt.Print("Profile ID: ")
    profileIDStr, _ := reader.ReadString('\n')
    profileID, _ := strconv.Atoi(strings.TrimSpace(profileIDStr))

    schedules, err := GetSchedules(profileID, currentUser)
    if err != nil {
        fmt.Printf("Error retrieving schedules: %v\n", err)
        return
    }
    if len(schedules) == 0 {
        fmt.Println("No schedules found for this profile.")
        return
    }
    fmt.Println("Schedules for this profile:")
    for _, s := range schedules {
        fmt.Printf("Day %d: %s - %s, Target: %.1f°C\n", s.DayOfWeek, s.StartTime, s.EndTime, s.TargetTemp)
    }
}


func manageUsers(reader *bufio.Reader) {
    if currentUser.Role != "homeowner" && currentUser.Role != "technician" {
        fmt.Println("Only homeowners or technicians can manage users")
        return
    }
    
    for {
        fmt.Println("\n=== USER MANAGEMENT ===")
        fmt.Println("1. Create Guest Account")
        
        // Only show "Create Technician" option to homeowners
        if currentUser.Role == "homeowner" {
            fmt.Println("2. Create Technician Account")
            fmt.Println("3. Grant/Extend Technician Access")
        }
        
        fmt.Println("4. Revoke User Access")
        
        // Only show "List All Users" to homeowners
        if currentUser.Role == "homeowner" {
            fmt.Println("5. List All Users")
			fmt.Println("6. Permanently Delete User")
        }
        
        fmt.Println("0. Back to Main Menu")
        fmt.Print("Enter choice: ")
        
        choice, _ := reader.ReadString('\n')
        choice = strings.TrimSpace(choice)
        
        switch choice {
        case "1":
            // Create guest account - both homeowner and technician allowed
            fmt.Print("Guest name: ")
            guestName, _ := reader.ReadString('\n')
            guestName = strings.TrimSpace(guestName)
            
            fmt.Print("PIN (minimum 4 digits): ")
            pin, _ := reader.ReadString('\n')
            pin = strings.TrimSpace(pin)
            
            err := CreateGuestAccount(currentUser.Username, guestName, pin, currentUser.Role)
            if err != nil {
                fmt.Printf("Error: %v\n", err)
            } else {
                fmt.Println("Guest account created successfully")
            }
            
        case "2":
            // Create technician account - only homeowners
            if currentUser.Role != "homeowner" {
                fmt.Println("Invalid choice")
                continue
            }
            
            fmt.Print("Technician username: ")
            techName, _ := reader.ReadString('\n')
            techName = strings.TrimSpace(techName)
            
            fmt.Print("Password (minimum 4 characters): ")
            password, _ := reader.ReadString('\n')
            password = strings.TrimSpace(password)
            
            err := CreateTechnicianAccount(currentUser.Username, techName, password, currentUser.Role)
            if err != nil {
                fmt.Printf("Error: %v\n", err)
            } else {
                fmt.Println("Technician account created successfully")
            }
            
        case "3":
            // Grant technician access - only homeowners
            if currentUser.Role != "homeowner" {
                fmt.Println("Invalid choice")
                continue
            }
            
            fmt.Print("Technician username: ")
            techName, _ := reader.ReadString('\n')
            techName = strings.TrimSpace(techName)
            
            fmt.Print("Duration in hours: ")
            durationStr, _ := reader.ReadString('\n')
            durationStr = strings.TrimSpace(durationStr)
            hours, err := strconv.Atoi(durationStr)
            if err != nil {
                fmt.Println("Invalid duration")
                continue
            }
            duration := time.Duration(hours) * time.Hour
            
            err = GrantTechnicianAccess(currentUser.Username, techName, duration, currentUser.Role)
            if err != nil {
                fmt.Printf("Error: %v\n", err)
            } else {
                fmt.Println("Technician access granted successfully")
            }
            
        case "4":
            // Revoke user access - both homeowner and technician (with restrictions)
            fmt.Print("Username to revoke: ")
            username, _ := reader.ReadString('\n')
            username = strings.TrimSpace(username)
            
            err := RevokeAccess(username, currentUser.Username, currentUser.Role)
            if err != nil {
                fmt.Printf("Error: %v\n", err)
            } else {
                fmt.Println("User access revoked successfully")
            }
            
        case "5":
            // List all users - only homeowners
            if currentUser.Role != "homeowner" {
                fmt.Println("Invalid choice")
                continue
            }
            
            users, err := ListAllUsers(currentUser.Role)
            if err != nil {
                fmt.Printf("Error: %v\n", err)
                continue
            }
            
            fmt.Println("\n=== ALL USERS ===")
            for _, u := range users {
                status := "Active"
                if !u.IsActive {
                    status = "Inactive"
                }
                fmt.Printf("ID: %d | Username: %s | Role: %s | Status: %s\n", u.ID, u.Username, u.Role, status)
            }

		case "6":
			// List all users - only homeowners
            if currentUser.Role != "homeowner" {
                fmt.Println("Invalid choice")
                continue
            }
			deleteUser(reader)
        case "0":
            return
            
        default:
            fmt.Println("Invalid choice")
        }
    }
}


// Helper function to list users
func listUsers() {
    users, err := ListAllUsers(currentUser.Role)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Println("\n=== ALL USERS ===")
    for _, u := range users {
        status := "Active"
        if !u.IsActive {
            status = "Inactive"
        }
        fmt.Printf("ID: %d | Username: %s | Role: %s | Status: %s\n", u.ID, u.Username, u.Role, status)
    }
}


func createGuest(reader *bufio.Reader) {
	fmt.Print("Guest name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Guest PIN (min 4): ")
	pin, _ := reader.ReadString('\n')
	pin = strings.TrimSpace(pin)

	if err := CreateGuestAccount(currentUser.Username, name, pin, currentUser.Role); err != nil {
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
    if err := CreateTechnicianAccount(currentUser.Username, name, password, currentUser.Role); err != nil {
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

	if err := GrantTechnicianAccess(currentUser.Username, tech, time.Duration(hours)*time.Hour, currentUser.Role); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Technician access granted")
}

func revokeUserAccess(reader *bufio.Reader) {
	fmt.Print("Username to revoke: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	if err := RevokeAccess(username, currentUser.Username, currentUser.Role); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Access revoked for %s\n", username)
}

func deleteUser(reader *bufio.Reader) {
    fmt.Print("Username to permanently delete: ")
    username, _ := reader.ReadString('\n')
    username = strings.TrimSpace(username)
    err := DeleteUser(currentUser.Username, username, currentUser.Role)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("User %s permanently deleted.\n", username)
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

func manageAwayMode(reader *bufio.Reader) {
	for {
		fmt.Println("\n=== VACATION/AWAY MODE ===")
		
		// Check current status
		awayStatus, err := GetAwayModeStatus()
		if err == nil && awayStatus != nil {
			fmt.Println(DisplayAwayModeStatus(awayStatus))
		} else {
			fmt.Println("Away Mode: Inactive")
		}
		
		fmt.Println("\n1. Activate Away Mode")
		fmt.Println("2. Deactivate Away Mode")
		fmt.Println("0. Back to Main Menu")
		fmt.Print("Choice: ")
		
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		
		switch choice {
		case "1":
			fmt.Print("Return date (YYYY-MM-DD): ")
			dateStr, _ := reader.ReadString('\n')
			dateStr = strings.TrimSpace(dateStr)
			
			fmt.Print("Return time (HH:MM): ")
			timeStr, _ := reader.ReadString('\n')
			timeStr = strings.TrimSpace(timeStr)
			
			returnTime, err := time.Parse("2006-01-02 15:04", dateStr+" "+timeStr)
			if err != nil {
				fmt.Printf("Invalid date/time format: %v\n", err)
				continue
			}
			
			fmt.Print("Away temperature (10-35°C): ")
			tempStr, _ := reader.ReadString('\n')
			tempStr = strings.TrimSpace(tempStr)
			awayTemp, err := strconv.ParseFloat(tempStr, 64)
			if err != nil {
				fmt.Println("Invalid temperature")
				continue
			}
			
			err = SetAwayMode(returnTime, awayTemp, currentUser)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Away mode activated successfully!")
			}
			
		case "2":
			err := DeactivateAwayMode(currentUser)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Away mode deactivated. Settings restored.")
			}
			
		case "0":
			return
			
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func manageFilterMaintenance(reader *bufio.Reader) {
	for {
		fmt.Println("\n=== FILTER MAINTENANCE ===")
		
		// Display current maintenance status
		status, err := GetMaintenanceStatus()
		if err != nil {
			fmt.Printf("Error getting maintenance status: %v\n", err)
		} else {
			fmt.Println(DisplayMaintenanceStatus(status))
		}
		
		fmt.Println("\n1. Reset Filter (After Replacement)")
		fmt.Println("2. Set Filter Change Interval")
		fmt.Println("0. Back to Main Menu")
		fmt.Print("Choice: ")
		
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		
		switch choice {
		case "1":
			err := ResetFilter(currentUser)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Filter maintenance reset successfully!")
			}
			
		case "2":
			fmt.Print("Filter change interval (hours, 100-2000): ")
			hoursStr, _ := reader.ReadString('\n')
			hoursStr = strings.TrimSpace(hoursStr)
			hours, err := strconv.ParseFloat(hoursStr, 64)
			if err != nil {
				fmt.Println("Invalid number")
				continue
			}
			
			err = SetFilterChangeInterval(hours, currentUser)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Filter change interval set to %.0f hours\n", hours)
			}
			
		case "0":
			return
			
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func manageEcoMode(reader *bufio.Reader) {
	for {
		fmt.Println("\n=== ECO MODE SETTINGS ===")
		
		// Display current eco mode status
		isEco, _ := GetEcoModeStatus()
		fmt.Println(DisplayEcoModeStatus())
		
		fmt.Println("\n1. Enable Eco Mode")
		fmt.Println("2. Disable Eco Mode")
		fmt.Println("0. Back to Main Menu")
		fmt.Print("Choice: ")
		
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		
		switch choice {
		case "1":
			if isEco {
				fmt.Println("Eco mode is already enabled")
			} else {
				err := SetEcoMode(true, currentUser)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Println("Eco mode enabled! System will optimize for energy savings.")
				}
			}
			
		case "2":
			if !isEco {
				fmt.Println("Eco mode is already disabled")
			} else {
				err := SetEcoMode(false, currentUser)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Println("Eco mode disabled. Returning to standard operation.")
				}
			}
			
		case "0":
			return
			
		default:
			fmt.Println("Invalid choice")
		}
	}
}
