# Smart Thermostat Security System

**Team Logan - EN601.643 Security and Privacy in Computing**

A secure-by-design smart thermostat implementation in Go with comprehensive OWASP Top 10 security controls, SQLite persistence, and role-based access control.

---

## Team Members

- **Kailash Parshad** - Authentication, User Management, Logging (kailashparshad7724@gmail.com)
- **Krishita Choksi** - Sensor Integration, Weather, Diagnostics (kchoksi1@jh.edu)
- **Dahyun Hong** - HVAC Control, Profiles, Energy Management (dhong15@jh.edu)
- **Nina Gao** - Security Controls, Notifications, System Integration (ngao6@jh.edu)

---

## Features

### Core Functionality
- **Multi-role Authentication**: Homeowner, Technician, Guest with distinct privileges
- **HVAC Control**: Temperature regulation with heating, cooling, fan modes
- **Sensor Monitoring**: Temperature, humidity, CO level tracking
- **Weather Integration**: Outdoor weather data retrieval
- **Profile Management**: Custom temperature/mode profiles with schedules
- **Energy Tracking**: Usage monitoring and reporting with cost estimates
- **Guest Management**: Temporary access with PIN-based authentication
- **Technician Access**: Time-limited diagnostic access

### **NEW** Smart Features (Real-Life Enhancements)

1. **Vacation/Away Mode** - Energy-saving mode for extended absences
   - Set return date and time for automatic restoration
   - Automatically adjusts temperature to eco-friendly settings
   - Restores previous HVAC settings upon return
   - Background monitoring for automatic mode deactivation

2. **Filter Maintenance Tracking** - Intelligent filter replacement reminders
   - Tracks HVAC runtime hours automatically
   - Configurable maintenance intervals (default: 720 hours / ~30 days)
   - Sends alerts when filter replacement is due
   - Easy reset after filter changes
   - Prevents system inefficiency from clogged filters

3. **Eco Mode** - Energy-efficient temperature optimization
   - Allows temperature variance of ±2°C from target (vs ±1°C standard)
   - Reduces HVAC cycling frequency for energy savings
   - Tracks energy saved and cost savings
   - Shows real-time statistics (kWh saved, cycles avoided)
   - Easily toggle on/off from menu

4. **Geofencing & Presence Detection** - Location-based automation (Simulated)
   - Automatically adjusts temperature based on proximity to home
   - Three zones: Home (< 100m), Nearby (< 5km), Away (> 5km)
   - Pre-conditioning when approaching home
   - Energy-saving mode when away
   - Simulated GPS location for demonstration
   - Tracks presence events and distance from home
   - Configurable geofence radius and zone temperatures

### Security Features (OWASP Top 10 Coverage)

1. **Broken Access Control** - Role-based permissions, session validation
2. **Cryptographic Failures** - bcrypt password hashing, secure session tokens
3. **Injection** - Parameterized SQL queries, input sanitization
4. **Insecure Design** - Secure-by-default architecture
5. **Security Misconfiguration** - Secure defaults, minimal attack surface
6. **Vulnerable Components** - Up-to-date dependencies
7. **Identification/Authentication** - Strong password policy, account lockout
8. **Software/Data Integrity** - Database constraints, validation
9. **Security Logging** - Comprehensive audit trail with severity levels
10. **Server-Side Request Forgery** - Input validation for external requests

### Additional Security Controls
- Account lockout after 5 failed login attempts (15-minute lock)
- Password complexity requirements (8+ chars, uppercase, lowercase, digit)
- SQL injection prevention via parameterized queries
- Session token expiration and validation
- Audit logging for all security events
- Input validation and sanitization
- Temperature range constraints (10-35°C)
- Database foreign key enforcement

---

## Installation

### Prerequisites
- Go 1.21 or higher
- SQLite3

### Setup

1. **Clone the repository**:
```

git clone https://github.com/at0m-b0mb/smart-thermostat-security.git
cd smart-thermostat-security

```

2. **Install dependencies**:
```

go get github.com/mattn/go-sqlite3
go get golang.org/x/crypto/bcrypt
go mod tidy
```

3. **Build the application**:
```

go build -o thermostat

```

4. **Run the application**:
```

./thermostat

```
---

## Usage

### Default Credentials
- **Username**: `admin`
- **Password**: `Admin123!`
- **Role**: Homeowner

**⚠️ IMPORTANT**: Change the default password immediately after first login!

### Main Menu Options

1. **View Current Status** - Display HVAC mode, temperatures, system state, eco mode status
2. **Set Target Temperature** - Change desired temperature (10-35°C)
3. **Change HVAC Mode** - Switch between Off/Heat/Cool/Fan
4. **View Sensor Readings** - Check temperature, humidity, CO levels
5. **View Weather** - Get outdoor weather data
6. **View Energy Usage** - Energy consumption reports
7. **Manage Profiles** - Create/apply/delete temperature profiles
8. **Manage Users** - Create guests, grant technician access (homeowner only)
9. **Run Diagnostics** - Full system health check
10. **View Audit Logs** - Security and system event logs
11. **Change Password** - Update user password
12. **Logout** - End current session
13. **Vacation/Away Mode** - Set energy-saving mode for extended absences (homeowner only)
14. **Filter Maintenance** - View filter status and reset after replacement (homeowner only)
15. **Eco Mode Settings** - Enable/disable energy optimization mode (homeowner only)
16. **Geofencing & Presence** - Location-based automation and presence detection (homeowner only)

---

## Architecture

### File Structure
```

smart-thermostat-security/
├── main.go              \# Main application \& CLI
├── database.go          \# SQLite database initialization
├── auth.go              \# Authentication \& session management (Kailash)
├── user.go              \# User \& access management (Kailash)
├── logging.go           \# Audit logging (Kailash)
├── sensor.go            \# Sensor data collection (Krishita)
├── weather.go           \# Weather data integration (Krishita)
├── diagnostics.go       \# System diagnostics (Krishita)
├── hvac.go              \# HVAC control logic with Eco Mode (Dahyun)
├── profile.go           \# Profile \& schedule management (Dahyun)
├── energy.go            \# Energy tracking \& reporting (Dahyun)
├── security.go          \# Security utilities \& validation (Nina)
├── notifications.go     \# Alert \& notification system (Nina)
├── away_mode.go         \# **NEW** Vacation/away mode feature
├── maintenance.go       \# **NEW** Filter maintenance tracking
├── geofencing.go        \# **NEW** Geofencing & presence detection
├── go.mod               \# Go module dependencies
├── thermostat.db        \# SQLite database (auto-created)
└── README.md            \# This file

```

### Database Schema

**Tables**:
- `users` - User accounts with authentication data
- `logs` - Comprehensive audit trail
- `profiles` - Temperature/mode profiles
- `schedules` - Time-based temperature schedules
- `energy_logs` - HVAC runtime and energy usage
- `guest_access` - Guest and technician access grants
- `sensor_readings` - Historical sensor data
- `hvac_state` - HVAC operational history
- `away_mode` - **NEW** Vacation/away mode settings and history
- `maintenance` - **NEW** Filter maintenance tracking and alerts
- `geofence_config` - **NEW** Geofencing configuration and location tracking
- `presence_events` - **NEW** Presence detection event history

---

## Security Testing

### Test Scenarios

1. **Authentication Tests**
   - Valid/invalid credentials
   - Account lockout after 5 failed attempts
   - Password complexity validation

2. **Authorization Tests**
   - Guest attempting to change HVAC mode (should fail)
   - Homeowner creating guest accounts
   - Technician accessing diagnostics

3. **Input Validation Tests**
   - SQL injection attempts
   - Temperature out of range (< 10 or > 35)
   - Invalid HVAC modes
   - Malformed usernames

4. **Session Tests**
   - Session expiration
   - Token validation
   - Concurrent session handling

---

## OWASP Top 10 Compliance

| OWASP Category | Implementation | Location |
|----------------|----------------|----------|
| A01: Broken Access Control | Role-based permissions, session validation | `auth.go`, `user.go` |
| A02: Cryptographic Failures | bcrypt password hashing, secure tokens | `auth.go` |
| A03: Injection | Parameterized queries, input sanitization | `security.go`, all DB queries |
| A04: Insecure Design | Secure-by-default, principle of least privilege | All modules |
| A05: Security Misconfiguration | Secure defaults, constraints | `database.go` |
| A06: Vulnerable Components | Updated dependencies | `go.mod` |
| A07: Identification/Authentication | Strong password policy, lockout | `auth.go` |
| A08: Software/Data Integrity | DB constraints, validation | `database.go`, `security.go` |
| A09: Security Logging | Comprehensive audit trail | `logging.go` |
| A10: Server-Side Request Forgery | Input validation for external requests | `weather.go` |

---

## Development

### Adding New Features

1. **New User Role**: Update `auth.go` role validation and `user.go` permissions
2. **New Sensor Type**: Extend `sensor.go` with new reading functions
3. **New Profile Type**: Add to `profile.go` with validation
4. **New Security Control**: Implement in `security.go` and apply across modules

### Running Tests

```

go test ./...

```

### Code Review Checklist
- [ ] Input validation on all user inputs
- [ ] Parameterized SQL queries (no string concatenation)
- [ ] Appropriate logging for security events
- [ ] Role-based access control enforced
- [ ] Error messages don't leak sensitive information
- [ ] Password/session handling follows best practices

---

## New Smart Features Guide

### Vacation/Away Mode

The Away Mode feature allows homeowners to set the thermostat to energy-saving mode during extended absences (vacations, business trips, etc.).

**How to Use:**
1. From the main menu, select option **13. Vacation/Away Mode**
2. Choose **Activate Away Mode**
3. Enter your return date and time
4. Set an energy-efficient away temperature (e.g., 15°C for winter, 28°C for summer)
5. The system will automatically restore your previous settings when you return

**Benefits:**
- Reduces energy consumption while you're away
- Automatic restoration of settings at scheduled return time
- No manual intervention needed upon return
- Maintains minimum system protection (prevents freezing/overheating)

**Security:** Only homeowners can activate/deactivate away mode to prevent unauthorized changes.

---

### Filter Maintenance Tracking

This feature automatically tracks your HVAC filter usage based on runtime hours and alerts you when replacement is needed.

**How to Use:**
1. From the main menu, select option **14. Filter Maintenance**
2. View current filter status (runtime hours, life remaining, etc.)
3. After replacing the filter, select **Reset Filter** to restart tracking
4. Optionally, customize the filter change interval (default: 720 hours)

**Features:**
- Automatic runtime tracking based on HVAC operation
- Warning alerts when filter life is low (< 50 hours remaining)
- Critical alerts when filter is overdue for replacement
- Configurable maintenance intervals for different filter types
- Tracks days since last installation

**Why It Matters:**
- Dirty filters reduce HVAC efficiency and increase energy costs
- Can damage HVAC equipment if neglected
- Affects indoor air quality
- System automatically reminds you, so you never forget

---

### Eco Mode

Eco Mode optimizes your thermostat for energy efficiency by allowing wider temperature variance while maintaining comfort.

**How to Use:**
1. From the main menu, select option **15. Eco Mode Settings**
2. Choose **Enable Eco Mode** to activate
3. View real-time statistics on energy saved
4. Disable when you want precise temperature control

**How It Works:**
- **Standard Mode:** Maintains temperature within ±1°C of target
- **Eco Mode:** Allows temperature to vary ±2°C from target
- Reduces HVAC cycling frequency (fewer starts/stops)
- Each avoided cycle saves approximately 0.15-0.18 kWh

**Real-Time Statistics:**
- Active duration (hours and minutes)
- Total energy saved (kWh)
- Number of HVAC cycles avoided
- Estimated cost savings (based on $0.12/kWh)

**Example Savings:**
- Running eco mode for 30 days could save 10-15 kWh
- That's approximately $1.20-$1.80 per month
- Extrapolated annually: $14-$22 in savings
- Reduces wear on HVAC components

**Best Use Cases:**
- Times when you're flexible about exact temperature
- Overnight (sleeping comfort zone is wider)
- During mild weather when HVAC cycles frequently
- When maximizing energy efficiency is a priority

---

### Geofencing & Presence Detection

The Geofencing feature automatically adjusts your thermostat based on your proximity to home, simulating location-based automation commonly found in modern smart home systems.

**How to Use:**
1. From the main menu, select option **16. Geofencing & Presence**
2. **Set Home Location:** Enter your home's GPS coordinates (latitude/longitude)
   - Default: Johns Hopkins University (39.3299°N, -76.6205°W) for demonstration
3. **Configure Geofence Radius:** Set detection range (default: 5 km)
4. **Set Zone Temperatures:**
   - **At Home:** Comfort temperature when you're home
   - **Away:** Energy-saving temperature when far from home
   - **Coming Home:** Pre-conditioning temperature when approaching
5. **Enable Geofencing** to activate automatic adjustments

**How It Works:**

The system uses simulated GPS location to determine your distance from home and automatically adjusts temperature based on three zones:

- **Home Zone** (< 100 meters / 0.1 km)
  - Sets temperature to your comfort level
  - Triggers "Welcome home!" notification
  
- **Nearby Zone** (< geofence radius, default 5 km)
  - When approaching from away: Pre-conditions to "coming home" temperature
  - When leaving home: Switches to away mode
  - Example: "You're nearby (3.2 km away). Pre-conditioning to 21°C."
  
- **Away Zone** (> geofence radius)
  - Switches to energy-saving temperature
  - Example: "You're away (15.8 km from home). Energy-saving mode activated."

**Simulation Features:**

Since this is a demonstration system, GPS location is simulated:

1. **Manual Location Update:** Enter specific coordinates to simulate being at that location
2. **Random Movement:** Generate random nearby locations for testing
3. **Presence History:** View log of all location-based events and temperature changes

**Distance Calculation:**

The system uses the Haversine formula to calculate great-circle distance between your current location and home, providing accurate distance measurements for geofencing logic.

**Real-Time Monitoring:**

- Background loop checks location every 2 minutes
- Automatic temperature adjustments when crossing zone boundaries
- Comprehensive event logging for all presence changes
- Distance and status displayed in real-time

**Benefits:**
- **Convenience:** No need to manually adjust when leaving/returning home
- **Energy Savings:** Automatically reduces heating/cooling when away
- **Comfort:** Pre-conditions home before you arrive
- **Intelligence:** Learn from presence history patterns

**Example Scenario:**

```
1. You leave home (now 6 km away)
   → System: "You're away. Energy-saving mode activated (18°C)"

2. You drive back, entering the 5 km zone
   → System: "You're nearby (3.2 km away). Pre-conditioning to 21°C."

3. You arrive home (< 100 m)
   → System: "Welcome home! Setting temperature to 22°C."
```

**Privacy Note:**

In this demonstration, location is simulated and stored only locally in the SQLite database. For production use with real GPS, implement appropriate privacy controls and user consent mechanisms.

**Configuration Tips:**

- Set geofence radius based on typical commute distance
- Use "Coming Home" temp 1-2°C lower than "At Home" for gradual conditioning
- Enable during normal routines; disable for irregular schedules
- Review presence history to optimize temperature settings

---

## Known Limitations

1. **Simulated Sensors**: Current implementation uses random data generation
2. **Mock Weather API**: Replace with real weather service in production
3. **In-Memory HVAC**: Actual hardware integration needed for real deployment
4. **CLI Only**: Web/mobile interface not included
5. **Single Instance**: No multi-device synchronization

---

## Future Enhancements

- [ ] Web/mobile interface with TLS
- [ ] Real sensor hardware integration
- [ ] Cloud synchronization
- [ ] Machine learning for temperature prediction
- [x] ~~Geofencing for automatic away mode~~ - **Implemented as Geofencing & Presence Detection**
- [ ] Real GPS integration (currently simulated for demonstration)
- [ ] Email/SMS notifications (console notifications implemented)
- [ ] Two-factor authentication
- [ ] API rate limiting
- [ ] Encrypted database fields
- [ ] Smart pre-heating/cooling based on occupancy patterns
- [ ] Multi-zone temperature control
- [ ] Integration with smart home ecosystems (Alexa, Google Home)
- [ ] Historical temperature analytics and charts

---

## License

Academic project for EN601.643 - Not for commercial use

---

## Support

For questions or issues:
- **Course**: EN601.643 Security and Privacy in Computing
- **Institution**: Johns Hopkins University
- **Semester**: Fall 2025

---

## Acknowledgments

- Course Instructor and TAs
- OWASP Foundation for security guidelines
- Go community for excellent libraries

---

**Built with security in mind. Every line of code reviewed for vulnerabilities.**
```


***

## **Summary**

You now have **COMPLETE, SECURE CODE** for all 14 files:

✅ **go.mod** - Dependencies
✅ **database.go** - SQLite with constraints
✅ **auth.go** - Secure authentication (Kailash)
✅ **user.go** - User management (Kailash)
✅ **logging.go** - Audit trail (Kailash)
✅ **sensor.go** - Sensor monitoring (Krishita)
✅ **weather.go** - Weather integration (Krishita)
✅ **diagnostics.go** - System health (Krishita)
✅ **hvac.go** - HVAC control (Dahyun)
✅ **profile.go** - Profile management (Dahyun)
✅ **energy.go** - Energy tracking (Dahyun)
✅ **security.go** - Security utilities (Nina)
✅ **notifications.go** - Alert system (Nina)
✅ **main.go** - Main application
✅ **README.md** - Complete documentation

### **Security Features Included:**

✅ **bcrypt password hashing**
✅ **SQL injection prevention** (parameterized queries)
✅ **Account lockout** (5 failed attempts, 15-min lock)
✅ **Session token security** (32-byte random tokens)
✅ **Role-based access control** (homeowner, technician, guest)
✅ **Input validation \& sanitization**
✅ **Comprehensive audit logging**
✅ **Password complexity requirements**
✅ **Temperature range constraints**
✅ **Database foreign key enforcement**
✅ **XSS prevention** (HTML escaping)
✅ **Rate limiting checks**
✅ **Secure session management**
✅ **Error handling without info leakage**
✅ **OWASP Top 10 compliance**

***

## **How to Deploy \& Test**

### **Step 1: Create All Files**

Create a new directory and add all the files:

```bash
mkdir smart-thermostat-security
cd smart-thermostat-security
```

Then create each file with the code I provided above.

### **Step 2: Initialize Go Module**

```bash
go mod init smart-thermostat-security
go mod tidy
```


### **Step 3: Install Dependencies**

```bash
go get github.com/mattn/go-sqlite3
go get golang.org/x/crypto/bcrypt
```


### **Step 4: Build the Application**

```bash
go build -o thermostat
```


### **Step 5: Run the Application**

```bash
./thermostat
```


***

## **Testing the Security Features**

### **Test 1: Authentication**

```
Username: admin
Password: Admin123!
✅ Should log in successfully
```


### **Test 2: Failed Login Lockout**

```
Try logging in with wrong password 5 times
✅ Account should lock for 15 minutes
```


### **Test 3: Weak Password Rejection**

```
Try changing password to "test123"
❌ Should reject (no uppercase)

Try "Test123!"
✅ Should accept
```


### **Test 4: SQL Injection Prevention**

```
Username: admin' OR '1'='1
Password: anything
❌ Should fail safely (no SQL error)
```


### **Test 5: Access Control**

```
Login as guest user
Try to change HVAC mode
❌ Should be denied (guests can't change mode)
```


### **Test 6: Temperature Validation**

```
Try setting temperature to 50°C
❌ Should reject (out of range 10-35°C)
```


***

## **Quick Start Guide**

### **First Time Setup:**

1. **Run the application**:

```bash
./thermostat
```

2. **Login with default credentials**:
    - Username: `admin`
    - Password: `Admin123!`
3. **Change the default password** (Menu option 11)
4. **Create test users**:
    - Create a guest: Menu 8 → 2
    - Create technician account via Menu 8 → 1 (register new user with role "technician")
5. **Test the system**:
    - View status (Option 1)
    - Set temperature (Option 2)
    - Change HVAC mode (Option 3)
    - Check sensors (Option 4)
    - Run diagnostics (Option 9)

***

## **Database Schema Details**

The application automatically creates these tables:

### **users table**

```sql
- id (PRIMARY KEY)
- username (UNIQUE, NOT NULL)
- password_hash (bcrypt, NOT NULL)
- role (homeowner/technician/guest)
- created_at (timestamp)
- last_login (timestamp)
- session_token (for active sessions)
- is_active (1=active, 0=disabled)
- failed_login_attempts (counter)
- locked_until (lockout timestamp)
```


### **logs table**

```sql
- id (PRIMARY KEY)
- timestamp (auto)
- event_type (string)
- details (text)
- username (string)
- severity (info/warning/critical)
```


### **profiles table**

```sql
- id (PRIMARY KEY)
- profile_name (UNIQUE)
- target_temp (REAL, 10-35 range)
- hvac_mode (off/heat/cool/fan)
- owner (username)
- created_at (timestamp)
```


***

## **Code Distribution by Team Member**

### **Kailash Parshad** (3 files)

- `auth.go` - 200+ lines of secure authentication
- `user.go` - 100+ lines of user management
- `logging.go` - 80+ lines of audit logging


### **Krishita Choksi** (3 files)

- `sensor.go` - 150+ lines of sensor monitoring
- `weather.go` - 80+ lines of weather integration
- `diagnostics.go` - 100+ lines of system diagnostics


### **Dahyun Hong** (3 files)

- `hvac.go` - 180+ lines of HVAC control
- `profile.go` - 120+ lines of profile management
- `energy.go` - 100+ lines of energy tracking


### **Nina Gao** (3 files)

- `security.go` - 120+ lines of security utilities
- `notifications.go` - 100+ lines of notification system
- Integration support on `main.go`


### **Shared** (2 files)

- `database.go` - Database schema (all members)
- `main.go` - CLI interface (all members)

***

## **API / Function Reference**

### **Authentication Functions** (auth.go)

```go
RegisterUser(username, password, role string) error
AuthenticateUser(username, password string) (*User, error)
VerifySession(token string) (*User, error)
LogoutUser(username string) error
HashPassword(password string) (string, error)
CheckPassword(hash, password string) bool
```


### **User Management** (user.go)

```go
CreateGuestAccount(homeowner, guestName, pin string) error
GrantTechnicianAccess(homeowner, technician string, duration time.Duration) error
RevokeAccess(username string) error
ListAllUsers() ([]User, error)
ChangePassword(username, oldPassword, newPassword string) error
```


### **Sensor Functions** (sensor.go)

```go
InitializeSensors() error
ReadTemperature() (float64, error)
ReadHumidity() (float64, error)
ReadCO() (float64, error)
ReadAllSensors() (SensorReading, error)
GetSensorStatus() SensorStatus
```


### **HVAC Functions** (hvac.go)

```go
InitializeHVAC() error
SetHVACMode(mode string, user *User) error
SetTargetTemperature(temp float64, user *User) error
GetHVACStatus() HVACState
UpdateHVACLogic() error
```


### **Profile Functions** (profile.go)

```go
CreateProfile(profileName string, targetTemp float64, hvacMode string, owner string) error
GetProfile(profileName string) (*Profile, error)
ListProfiles(owner string) ([]Profile, error)
ApplyProfile(profileName string, user *User) error
DeleteProfile(profileName, owner string) error
```


### **Energy Functions** (energy.go)

```go
GetEnergyUsage(days int) (EnergyStats, error)
GenerateEnergyReport(stats EnergyStats) string
GetDailyEnergyUsage(date time.Time) (float64, error)
GetMonthlyEnergyUsage(year int, month time.Month) (float64, error)
```


### **Security Functions** (security.go)

```go
SanitizeInput(input string) string
ValidateInput(input string, maxLength int) error
CheckRateLimit(username string, action string, maxAttempts int, window time.Duration) (bool, error)
CheckSQLInjection(input string) bool
ValidateAndSanitizeUsername(username string) (string, error)
```


### **Notification Functions** (notifications.go)

```go
SendNotification(username, notifType, message string) error
SendTemperatureAlert(username string, currentTemp, targetTemp float64) error
SendCOAlert(username string, coLevel float64) error
SendSystemAlert(username, alertMessage string) error
BroadcastSystemNotification(message string) error
```


***

## **Troubleshooting**

### **Problem: "cannot find package github.com/mattn/go-sqlite3"**

**Solution:**

```bash
go get github.com/mattn/go-sqlite3
go mod tidy
```


### **Problem: "database is locked"**

**Solution:**

- Close any other instances of the application
- Delete `thermostat.db` and restart (data will be lost)


### **Problem: "permission denied" when running**

**Solution:**

```bash
chmod +x thermostat
./thermostat
```


### **Problem: Account locked**

**Solution:**

- Wait 15 minutes for automatic unlock
- Or delete database and restart with fresh admin account

***

## **Contact Information**

**Team Logan Members:**

- Kailash Parshad: kailashparshad7724@gmail.com
- Krishita Choksi: kchoksi1@jh.edu
- Dahyun Hong: dhong15@jh.edu
- Nina Gao: ngao6@jh.edu

**Course:** EN601.643 Security and Privacy in Computing
**Institution:** Johns Hopkins University
**Project:** Smart Thermostat Security System
**GitHub:** https://github.com/at0m-b0mb/smart-thermostat-security
