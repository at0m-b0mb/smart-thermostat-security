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

## Table of Contents

1. [Overview](#overview)
2. [User Roles and Permissions](#user-roles-and-permissions)
3. [Installation](#installation)
4. [Quick Start Guide](#quick-start-guide)
5. [Features by User Role](#features-by-user-role)
6. [Security Features](#security-features)
7. [Architecture](#architecture)
8. [Testing](#testing)
9. [Troubleshooting](#troubleshooting)

---

## Overview

This smart thermostat system implements **role-based access control (RBAC)** with three distinct user types: **Homeowner**, **Technician**, and **Guest**. Each role has specific permissions designed to maintain security while providing appropriate access levels.

### Key Principles
- ✅ **Least Privilege**: Users only have access to features necessary for their role
- ✅ **Secure by Default**: All actions require authentication and authorization
- ✅ **Audit Trail**: Every action is logged for security and accountability
- ✅ **Defense in Depth**: Multiple layers of security controls protect the system

---

## User Roles and Permissions

### 🏠 Homeowner (Full Access)

**Description**: The primary account owner with complete system control.

**What Homeowners CAN Do:**
- ✅ View current system status and all sensor readings
- ✅ Set and change target temperature (10-35°C)
- ✅ Change HVAC mode (Off, Heat, Cool, Fan)
- ✅ View outdoor weather information
- ✅ Monitor energy usage and view reports
- ✅ Create, view, apply, and delete temperature profiles
- ✅ Mark profiles as guest-accessible or private
- ✅ Add and view schedules for automated temperature control
- ✅ Create guest accounts (PIN-based access)
- ✅ Create technician accounts
- ✅ Grant and extend time-limited technician access
- ✅ Revoke access for any user (guests or technicians)
- ✅ View complete list of all users in the system
- ✅ Permanently delete user accounts (except other homeowners)
- ✅ Run system diagnostics
- ✅ View complete audit logs
- ✅ Change their own password

**What Homeowners CANNOT Do:**
- ❌ Delete their own homeowner account
- ❌ Delete other homeowner accounts

**Use Cases:**
- Daily temperature adjustments and HVAC control
- Creating profiles for different times/seasons (e.g., "Sleep Mode", "Away Mode")
- Monitoring energy consumption and costs
- Managing family member guest access with PINs
- Granting time-limited access to HVAC service technicians
- Reviewing security logs for unusual activity

---

### 🔧 Technician (Diagnostic Access)

**Description**: Service professionals with time-limited access for maintenance and diagnostics.

**What Technicians CAN Do:**
- ✅ View current system status and all sensor readings
- ✅ Set and change target temperature (10-35°C)
- ✅ Change HVAC mode (Off, Heat, Cool, Fan)
- ✅ View outdoor weather information
- ✅ Monitor energy usage and view reports
- ✅ View profiles they created OR profiles marked as guest-accessible
- ✅ Create new temperature profiles (marked as their own)
- ✅ Apply profiles they own or guest-accessible profiles
- ✅ Delete profiles they own or guest-accessible profiles
- ✅ Add and view schedules for profiles they manage
- ✅ Create guest accounts (with homeowner permission model)
- ✅ Revoke access for guest accounts they created
- ✅ Run system diagnostics
- ✅ Change their own password

**What Technicians CANNOT Do:**
- ❌ Create other technician accounts
- ❌ Grant or extend technician access
- ❌ View the complete user list
- ❌ Revoke access for other technicians
- ❌ Delete technician or homeowner accounts
- ❌ View audit logs
- ❌ Access profiles owned by homeowner (unless marked guest-accessible)
- ❌ Delete homeowner-owned private profiles

**Access Requirements:**
- ⏰ Technician access is **time-limited** and must be granted by a homeowner
- ⏰ Once the granted time expires, technicians cannot log in until access is extended
- ⏰ Homeowners can revoke technician access at any time

**Use Cases:**
- HVAC system maintenance and troubleshooting
- Running diagnostic tests to identify system issues
- Temporary adjustments during service calls
- Creating service profiles for testing purposes
- Limited-time access for contractors or service companies

---

### 👥 Guest (View-Only + Limited Control)

**Description**: Family members or temporary visitors with limited, safe access.

**What Guests CAN Do:**
- ✅ View current system status and all sensor readings
- ✅ Change HVAC mode (Off, Heat, Cool, Fan)
- ✅ View outdoor weather information
- ✅ View temperature profiles marked as "guest-accessible"
- ✅ Apply guest-accessible profiles (like "Away Mode" or "Sleep Mode")
- ✅ Change their own PIN (4+ digits, numeric only)

**What Guests CANNOT Do:**
- ❌ Set or change target temperature directly
- ❌ View energy usage reports
- ❌ Create new temperature profiles
- ❌ Delete any profiles
- ❌ Add or view schedules
- ❌ Create or manage other user accounts
- ❌ Grant or revoke access
- ❌ View the user list
- ❌ Run system diagnostics
- ❌ View audit logs
- ❌ Access private (non-guest-accessible) profiles

**Authentication:**
- 🔐 Guests use **PIN-based authentication** (minimum 4 digits, numeric only)
- 🔐 Guest accounts are created by homeowners or technicians
- 🔐 Guest usernames follow format: `{creator}_guest_{guestname}`

**Use Cases:**
- Family members who need basic temperature control
- House sitters who need to adjust comfort settings
- Temporary visitors staying at the home
- Children who should have supervised access
- Pet sitters who need to ensure home comfort

**Important Security Notes for Guests:**
- Guests can change HVAC mode but **cannot set specific temperatures**
- This prevents accidentally setting extreme temperatures
- Guests can only use pre-approved profiles created by homeowners
- All guest actions are logged in the audit trail

---

## Comparison Table: What Each Role Can Do

| Feature | Homeowner | Technician | Guest |
|---------|-----------|------------|-------|
| View Status | ✅ | ✅ | ✅ |
| Set Target Temperature | ✅ | ✅ | ❌ |
| Change HVAC Mode | ✅ | ✅ | ✅ |
| View Sensors | ✅ | ✅ | ✅ |
| View Weather | ✅ | ✅ | ✅ |
| View Energy Usage | ✅ | ✅ | ❌ |
| Create Profiles | ✅ | ✅ | ❌ |
| View All Profiles | ✅ | Own + Guest | Guest Only |
| Apply Profiles | ✅ | Own + Guest | Guest Only |
| Delete Profiles | ✅ | Own + Guest | ❌ |
| Manage Schedules | ✅ | ✅ | ❌ |
| Create Guests | ✅ | ✅ | ❌ |
| Create Technicians | ✅ | ❌ | ❌ |
| Grant Tech Access | ✅ | ❌ | ❌ |
| Revoke Access | ✅ | Guests Only | ❌ |
| List All Users | ✅ | ❌ | ❌ |
| Delete Users | ✅ | ❌ | ❌ |
| Run Diagnostics | ✅ | ✅ | ❌ |
| View Audit Logs | ✅ | ❌ | ❌ |
| Time-Limited Access | ❌ | ✅ | ❌ |

---

## Core Features

### Multi-role Authentication
- **Three distinct user roles**: Homeowner, Technician, Guest
- **Strong authentication**: bcrypt password hashing, PIN-based guest access
- **Session management**: Secure tokens with 24-hour expiration
- **Account protection**: Lockout after 5 failed attempts (15 minutes)

### HVAC Control
- **Temperature regulation**: Heating, cooling, fan modes with safety limits
- **Smart logic**: Automatic on/off based on current vs target temperature
- **Real-time monitoring**: Continuous temperature tracking
- **Energy efficiency**: Runtime tracking for all HVAC operations

### Sensor Monitoring
- **Temperature tracking**: Real-time indoor temperature readings
- **Humidity monitoring**: Indoor humidity percentage
- **CO level detection**: Carbon monoxide safety monitoring
- **Historical data**: All sensor readings stored for analysis

### Profile Management
- **Custom profiles**: Save preferred temperature and HVAC mode combinations
- **Guest accessibility**: Mark profiles as accessible to guests
- **Scheduled automation**: Time-based profile activation
- **Owner tracking**: Profiles tracked by creator for access control

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
- PIN requirements for guests (4+ digits, numeric only)
- SQL injection prevention via parameterized queries
- Session token expiration (24 hours) and validation
- Audit logging for all security events
- Input validation and sanitization
- Temperature range constraints (10-35°C)
- Database foreign key enforcement

---

## Installation

### Prerequisites
- **Go 1.21 or higher**
- **SQLite3** (automatically included with Go dependencies)
- **Unix-like OS** (Linux, macOS) or Windows

### Setup Steps

1. **Clone the repository**:
```bash
git clone https://github.com/at0m-b0mb/smart-thermostat-security.git
cd smart-thermostat-security
```

2. **Install dependencies**:
```bash
go get github.com/mattn/go-sqlite3
go get golang.org/x/crypto/bcrypt
go mod tidy
```

3. **Build the application**:
```bash
go build -o thermostat
```

4. **Run the application**:
```bash
./thermostat
```

5. **First-time setup**:
   - Login with default credentials (see below)
   - **IMMEDIATELY change the default password** via menu option 11
   - The database `thermostat.db` will be created automatically

---

## Quick Start Guide

### Step 1: First Login (Homeowner)

**Default Homeowner Credentials:**
- **Username**: `admin`
- **Password**: `Admin123!`
- **Role**: Homeowner

**⚠️ CRITICAL SECURITY STEP**: Change the default password immediately!

```
1. Run: ./thermostat
2. Login with admin/Admin123!
3. Select option 11: Change Password
4. Enter current password: Admin123!
5. Enter new strong password (8+ chars, uppercase, lowercase, digit)
6. Confirm new password
```

### Step 2: Create User Accounts

#### Creating a Guest Account (Homeowner or Technician)
```
1. Select option 8: Manage Users
2. Select option 1: Create Guest Account
3. Enter guest name (e.g., "john")
4. Enter PIN (minimum 4 digits, e.g., "1234")
5. Guest account created as: admin_guest_john
```

**Guest Login:**
- Username: `admin_guest_john`
- PIN: `1234` (or whatever PIN you set)

#### Creating a Technician Account (Homeowner Only)
```
1. Select option 8: Manage Users
2. Select option 2: Create Technician Account
3. Enter technician username (e.g., "hvac_tech")
4. Enter password (minimum 8 characters)
5. Select option 3: Grant/Extend Technician Access
6. Enter technician username: hvac_tech
7. Enter duration in hours (e.g., 24 for 24 hours)
```

**Technician Login:**
- Username: `hvac_tech`
- Password: (as set during creation)
- Access expires after the granted duration

### Step 3: Basic Operations

#### For Homeowners:
```
View Status        → Option 1
Set Temperature    → Option 2 (enter 10-35°C)
Change Mode        → Option 3 (Off/Heat/Cool/Fan)
Create Profile     → Option 7 → 3 (set name, temp, mode, guest-accessible)
View Energy        → Option 6 (enter number of days)
View Audit Logs    → Option 10
```

#### For Technicians:
```
View Status        → Option 1
Run Diagnostics    → Option 9
Set Temperature    → Option 2
Create Profile     → Option 7 → 3 (your own profiles)
View Energy        → Option 6
```

#### For Guests:
```
View Status        → Option 1
Change Mode        → Option 3 (Off/Heat/Cool/Fan only)
View Profiles      → Option 6 → 1 (guest-accessible only)
Apply Profile      → Option 6 → 2 (choose from guest-accessible)
Change PIN         → Option 11
```

---

## Features by User Role

### Homeowner Complete Menu

When logged in as a homeowner, you see:

```
=== MAIN MENU ===
1.  View Current Status
2.  Set Target Temperature
3.  Change HVAC Mode
4.  View Sensor Readings
5.  View Weather
6.  View Energy Usage
7.  Manage Profiles
8.  Manage Users
9.  Run Diagnostics
10. View Audit Logs
11. Change Password
12. Logout
0.  Exit
```

**Profile Management Submenu (Option 7):**
```
1. List Profiles
2. Apply Profile
3. Create Profile
4. Delete Profile
5. Add Schedule
6. View Schedules
0. Back to Main Menu
```

**User Management Submenu (Option 8):**
```
1. Create Guest Account
2. Create Technician Account
3. Grant/Extend Technician Access
4. Revoke User Access
5. List All Users
6. Permanently Delete User
0. Back to Main Menu
```

### Technician Menu

When logged in as a technician, you see:

```
=== MAIN MENU ===
1.  View Current Status
2.  Set Target Temperature
3.  Change HVAC Mode
4.  View Sensor Readings
5.  View Weather
6.  View Energy Usage
7.  Manage Profiles
8.  Manage Users
9.  Run Diagnostics
11. Change Password
12. Logout
0.  Exit
```

**Profile Management Submenu (Option 7):**
```
1. List Profiles (your own + guest-accessible)
2. Apply Profile (your own + guest-accessible)
3. Create Profile (creates as your own)
4. Delete Profile (your own + guest-accessible)
5. Add Schedule (your own profiles)
6. View Schedules (your own profiles)
0. Back to Main Menu
```

**User Management Submenu (Option 8):**
```
1. Create Guest Account
4. Revoke User Access (guests only)
0. Back to Main Menu
```

### Guest Menu

When logged in as a guest, you see:

```
=== MAIN MENU ===
1.  View Current Status
3.  Change HVAC Mode
4.  View Sensor Readings
5.  View Weather
6.  Manage Profiles
11. Change Password (PIN)
12. Logout
0.  Exit
```

**Profile Management Submenu (Option 6):**
```
1. List Profiles (guest-accessible only)
2. Apply Profile (guest-accessible only)
0. Back to Main Menu
```

---

## Common Use Cases

### Use Case 1: Daily Temperature Adjustment (Homeowner)
```
Scenario: Homeowner wants to set temperature to 22°C in heating mode
1. Login as homeowner
2. Option 3: Change HVAC Mode → select "2. Heat"
3. Option 2: Set Target Temperature → enter "22"
4. System automatically starts heating when temp drops below 21°C
```

### Use Case 2: Creating Profiles for Different Times (Homeowner)
```
Scenario: Create "Sleep Mode" profile for nighttime
1. Login as homeowner
2. Option 7: Manage Profiles
3. Option 3: Create Profile
4. Profile name: "Sleep Mode"
5. Target temperature: 18°C
6. HVAC mode: "heat"
7. Guest accessible: "yes" (so family can use it)
```

### Use Case 3: Granting Service Access (Homeowner)
```
Scenario: HVAC technician needs 4-hour access for maintenance
1. Login as homeowner
2. Option 8: Manage Users
3. Option 2: Create Technician Account
4. Username: "service_tech"
5. Password: "TechPass123!"
6. Option 3: Grant/Extend Technician Access
7. Username: "service_tech"
8. Duration: 4 hours
9. Share credentials with technician
10. After 4 hours, access automatically expires
```

### Use Case 4: Guest Applies a Profile (Guest)
```
Scenario: Guest feels cold and wants to activate heating
1. Login as guest with PIN
2. Option 6: Manage Profiles
3. Option 1: List Profiles (see guest-accessible profiles)
4. Option 2: Apply Profile
5. Enter: "Warm Mode" or "Sleep Mode"
6. HVAC automatically adjusts to profile settings
```

### Use Case 5: Running Diagnostics (Technician)
```
Scenario: Technician troubleshoots system issues
1. Login as technician
2. Option 9: Run Diagnostics
3. Review system health report:
   - Sensor status
   - HVAC functionality
   - Database integrity
   - Recent error logs
4. Document findings
```

### Use Case 6: Monitoring Energy Usage (Homeowner)
```
Scenario: Review last week's energy consumption
1. Login as homeowner
2. Option 6: View Energy Usage
3. Enter: 7 (for 7 days)
4. Review report:
   - Total kWh consumed
   - Heating/cooling/fan breakdown
   - Estimated cost
   - Daily averages
```

### Use Case 7: Revoking Access (Homeowner)
```
Scenario: Guest's visit ended, remove their access
1. Login as homeowner
2. Option 8: Manage Users
3. Option 4: Revoke User Access
4. Enter username: "admin_guest_john"
5. Guest can no longer login
```

### Use Case 8: Security Audit (Homeowner)
```
Scenario: Review system access logs
1. Login as homeowner
2. Option 10: View Audit Logs
3. Review recent entries:
   - Login attempts (successful and failed)
   - Temperature changes
   - Profile applications
   - User creation/deletion
   - HVAC mode changes
```

---

## Architecture

### File Structure
```
smart-thermostat-security/
├── main.go              # Main application & CLI with role-based menus
├── database.go          # SQLite database initialization & schema
├── auth.go              # Authentication & session management (Kailash)
├── user.go              # User & access management with RBAC (Kailash)
├── logging.go           # Audit logging system (Kailash)
├── sensor.go            # Sensor data collection (Krishita)
├── weather.go           # Weather data integration (Krishita)
├── diagnostics.go       # System diagnostics (Krishita)
├── hvac.go              # HVAC control logic (Dahyun)
├── profile.go           # Profile & schedule management with RBAC (Dahyun)
├── energy.go            # Energy tracking & reporting (Dahyun)
├── security.go          # Security utilities & validation (Nina)
├── notifications.go     # Alert & notification system (Nina)
├── go.mod               # Go module dependencies
├── go.sum               # Dependency checksums
├── thermostat.db        # SQLite database (auto-created)
└── README.md            # This comprehensive documentation
```

### Database Schema

The system uses SQLite with the following tables:

**users** - User accounts with role-based access
```sql
id                      INTEGER PRIMARY KEY
username                TEXT UNIQUE NOT NULL
password_hash           TEXT NOT NULL (bcrypt)
role                    TEXT NOT NULL (homeowner/technician/guest)
created_at              TIMESTAMP DEFAULT CURRENT_TIMESTAMP
last_login              TIMESTAMP
session_token           TEXT (secure random token)
session_expires_at      TIMESTAMP
is_active               INTEGER DEFAULT 1 (1=active, 0=disabled)
failed_login_attempts   INTEGER DEFAULT 0
locked_until            TIMESTAMP (null if not locked)
```

**logs** - Comprehensive audit trail
```sql
id          INTEGER PRIMARY KEY
timestamp   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
event_type  TEXT NOT NULL
details     TEXT
username    TEXT
severity    TEXT (info/warning/critical)
```

**profiles** - Temperature/mode profiles with access control
```sql
id                INTEGER PRIMARY KEY
profile_name      TEXT UNIQUE NOT NULL
target_temp       REAL NOT NULL CHECK(target_temp >= 10 AND target_temp <= 35)
hvac_mode         TEXT NOT NULL
owner             TEXT NOT NULL
created_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP
guest_accessible  INTEGER DEFAULT 0 (0=private, 1=guest-accessible)
```

**schedules** - Time-based temperature schedules
```sql
id          INTEGER PRIMARY KEY
profile_id  INTEGER FOREIGN KEY → profiles(id)
day_of_week INTEGER (0=Sunday, 6=Saturday)
start_time  TEXT (HH:MM format)
end_time    TEXT (HH:MM format)
target_temp REAL CHECK(target_temp >= 10 AND target_temp <= 35)
```

**energy_logs** - HVAC runtime and energy usage
```sql
id                INTEGER PRIMARY KEY
timestamp         TIMESTAMP DEFAULT CURRENT_TIMESTAMP
hvac_mode         TEXT
runtime_minutes   INTEGER
estimated_kwh     REAL
```

**guest_access** - Guest and technician access grants
```sql
id             INTEGER PRIMARY KEY
guest_username TEXT NOT NULL
granted_by     TEXT NOT NULL
granted_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP
expires_at     TIMESTAMP (for time-limited technician access)
is_active      INTEGER DEFAULT 1
```

**sensor_readings** - Historical sensor data
```sql
id          INTEGER PRIMARY KEY
timestamp   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
temperature REAL
humidity    REAL
co_level    REAL
```

**hvac_state** - HVAC operational history
```sql
id           INTEGER PRIMARY KEY
timestamp    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
mode         TEXT
target_temp  REAL
current_temp REAL
is_running   INTEGER
```

### Access Control Flow

```
User Login
    ↓
Authentication (auth.go)
    ↓
Role Verification (user.go)
    ↓
Session Token Generation
    ↓
Menu Display (role-based filtering in main.go)
    ↓
Action Request
    ↓
Permission Check (user.go, profile.go, hvac.go)
    ↓
Input Validation (security.go)
    ↓
Execute Action
    ↓
Audit Logging (logging.go)
```

### Security Architecture

**Authentication Layer** (auth.go)
- bcrypt password hashing (cost factor 10)
- Secure session token generation (32-byte random)
- Account lockout mechanism (5 attempts → 15 min lock)
- Session expiration (24 hours)

**Authorization Layer** (user.go, profile.go)
- Role-based access control checks
- Owner verification for profile operations
- Time-limited technician access validation
- Guest accessibility flags

**Validation Layer** (security.go)
- Input sanitization (SQL injection prevention)
- Temperature range validation (10-35°C)
- Username format validation (alphanumeric + underscore)
- Password complexity enforcement
- PIN format validation (numeric, 4+ digits)

**Audit Layer** (logging.go)
- All authentication events
- All authorization failures
- All data modifications
- Security-relevant actions
- Severity classification (info/warning/critical)

---

## Security Features

### OWASP Top 10 Compliance

| OWASP Category | Implementation | Verification |
|----------------|----------------|--------------|
| **A01: Broken Access Control** | Role-based permissions, session validation, owner checks | Test with different roles attempting restricted actions |
| **A02: Cryptographic Failures** | bcrypt password hashing, secure tokens, no plaintext storage | Inspect database - passwords are hashed |
| **A03: Injection** | Parameterized SQL queries, input sanitization | Attempt SQL injection in login/inputs |
| **A04: Insecure Design** | Secure-by-default, least privilege principle | Review code for security assumptions |
| **A05: Security Misconfiguration** | Secure defaults, minimal attack surface, constraints | Check default settings and database constraints |
| **A06: Vulnerable Components** | Updated dependencies, minimal external libraries | Run `go list -m all` |
| **A07: Identification/Authentication** | Strong password policy, account lockout, session management | Test login failures and session expiration |
| **A08: Software/Data Integrity** | Database constraints, validation, foreign keys | Test with invalid data inputs |
| **A09: Security Logging** | Comprehensive audit trail with severity levels | Check logs table for event coverage |
| **A10: Server-Side Request Forgery** | Input validation for external requests (weather API) | Test weather function with malicious input |

### Testing Security Controls

#### Test 1: Authentication Security
```bash
# Test failed login lockout
1. Attempt login with wrong password 5 times
2. Verify account is locked for 15 minutes
3. Wait 15 minutes
4. Verify account is automatically unlocked
```

#### Test 2: Password Complexity
```bash
# Test weak password rejection
1. Login as homeowner
2. Option 11: Change Password
3. Try: "test123" → Should reject (no uppercase)
4. Try: "TEST123" → Should reject (no lowercase)
5. Try: "TestTest" → Should reject (no digit)
6. Try: "Test123" → Should reject (less than 8 chars)
7. Try: "TestPass123" → Should accept ✓
```

#### Test 3: Access Control - Guest Restrictions
```bash
# Test guest cannot set temperature
1. Create guest account: admin_guest_test
2. Login as guest
3. Attempt Option 2: Set Temperature → Should not appear in menu
4. Attempt Option 6: View Energy → Should not appear in menu
5. Verify guest can only see allowed options
```

#### Test 4: Access Control - Technician Restrictions
```bash
# Test technician cannot create other technicians
1. Login as technician
2. Option 8: Manage Users
3. Verify Option 2 (Create Technician) does NOT appear
4. Verify Option 5 (List All Users) does NOT appear
5. Verify only guest management options appear
```

#### Test 5: Profile Access Control
```bash
# Test guest can only access guest-accessible profiles
1. Login as homeowner
2. Create profile "Private" with guest_accessible=no
3. Create profile "Shared" with guest_accessible=yes
4. Logout
5. Login as guest
6. Option 6 → 1: List Profiles
7. Verify only "Shared" appears, not "Private"
8. Attempt to apply "Private" → Should fail
```

#### Test 6: SQL Injection Prevention
```bash
# Test SQL injection in username field
1. Username: admin' OR '1'='1
2. Password: anything
3. Should fail safely without SQL error
4. Check logs for security event
```

#### Test 7: Temperature Validation
```bash
# Test temperature range limits
1. Login as homeowner
2. Option 2: Set Temperature
3. Try: 50 → Should reject (> 35°C)
4. Try: 5 → Should reject (< 10°C)
5. Try: 22 → Should accept ✓
```

#### Test 8: Session Expiration
```bash
# Test session timeout
1. Login as any user
2. Note the session token (in database)
3. Wait 24 hours (or manually update session_expires_at in DB)
4. Attempt to perform action
5. Should be logged out automatically
```

#### Test 9: Technician Time-Limited Access
```bash
# Test technician access expiration
1. Login as homeowner
2. Create technician account
3. Grant access for 1 hour
4. Wait 1 hour (or manually update expires_at in DB)
5. Attempt technician login
6. Should fail with "access expired" message
```

#### Test 10: Audit Trail Verification
```bash
# Test comprehensive logging
1. Login as homeowner
2. Perform various actions:
   - Set temperature
   - Create profile
   - Create guest
   - Revoke access
3. Option 10: View Audit Logs
4. Verify all actions are logged with:
   - Timestamp
   - Event type
   - Username
   - Details
   - Severity level
```

---

## Troubleshooting

### Problem: Cannot Login - "Account Temporarily Locked"

**Cause**: Too many failed login attempts (5 attempts)

**Solution**:
- Wait 15 minutes for automatic unlock, OR
- If you're a homeowner and need immediate access:
  ```bash
  # Manually unlock in database
  sqlite3 thermostat.db
  UPDATE users SET failed_login_attempts = 0, locked_until = NULL WHERE username = 'admin';
  .quit
  ```

### Problem: Technician Cannot Login - "Access Expired"

**Cause**: Time-limited access period has ended

**Solution**:
1. Contact the homeowner
2. Homeowner must extend access:
   ```
   Option 8 → Option 3: Grant/Extend Technician Access
   Enter technician username
   Enter new duration in hours
   ```

### Problem: Guest Cannot Set Temperature

**Cause**: This is **intentional** security behavior

**Solution**:
- Guests **cannot** set specific temperatures for safety
- Guests can only:
  - Change HVAC mode (Off/Heat/Cool/Fan)
  - Apply pre-approved guest-accessible profiles
- If guest needs specific temperature:
  1. Homeowner creates profile with that temperature
  2. Mark profile as "guest_accessible"
  3. Guest applies the profile

### Problem: Cannot Create Technician Account - "Only Homeowners Can..."

**Cause**: Logged in as technician or guest

**Solution**:
- Only **homeowners** can create technician accounts
- Technicians and guests cannot create other technicians
- This is a security feature to maintain control hierarchy

### Problem: Cannot See a Profile That Should Exist

**Cause**: Profile visibility is role-based

**Solution**:
- **Guests**: Can only see profiles marked "guest_accessible"
- **Technicians**: Can see their own profiles + guest-accessible profiles
- **Homeowners**: Can see all profiles
- Check profile ownership and guest_accessible flag:
  ```bash
  sqlite3 thermostat.db
  SELECT profile_name, owner, guest_accessible FROM profiles;
  .quit
  ```

### Problem: "Database is Locked" Error

**Cause**: Multiple instances or crashed process

**Solution**:
1. Close all instances of the application
2. Check for running processes:
   ```bash
   ps aux | grep thermostat
   kill <pid>  # if any found
   ```
3. If problem persists, restart:
   ```bash
   rm thermostat.db
   ./thermostat  # Fresh database with admin account
   ```
   **Warning**: This deletes all data!

### Problem: "Cannot Find Package github.com/mattn/go-sqlite3"

**Cause**: Dependencies not installed

**Solution**:
```bash
go get github.com/mattn/go-sqlite3
go get golang.org/x/crypto/bcrypt
go mod tidy
go build -o thermostat
```

### Problem: "Permission Denied" When Running

**Cause**: Binary not executable

**Solution**:
```bash
chmod +x thermostat
./thermostat
```

### Problem: Forgot Homeowner Password

**Solution**:
```bash
# Reset admin password to default
sqlite3 thermostat.db
# Generate new bcrypt hash for "Admin123!" or your new password
# You'll need to hash it externally then update:
UPDATE users SET password_hash = '<new_bcrypt_hash>' WHERE username = 'admin';
.quit
```

**Better approach**: Delete database and start fresh:
```bash
rm thermostat.db
./thermostat
# Login with admin/Admin123!
# Immediately change password
```

### Problem: Want to See Raw Data in Database

**Solution**:
```bash
sqlite3 thermostat.db
.tables                              # List all tables
.schema users                        # Show table structure
SELECT * FROM users;                 # View all users
SELECT * FROM logs ORDER BY timestamp DESC LIMIT 20;  # Recent logs
SELECT * FROM profiles;              # All profiles
.quit
```

### Problem: Energy Reports Show Zero Usage

**Cause**: HVAC hasn't run long enough or not running

**Solution**:
1. Verify HVAC is running: Option 1 (View Status)
2. Check if mode is "off"
3. Energy only accumulates when HVAC actively runs
4. Background loop logs runtime every 30 seconds
5. Try running system for a few minutes first

### Problem: Sensor Readings Look Random

**Cause**: This is **intentional** - simulated sensors

**Solution**:
- Current implementation uses simulated sensor data
- For production deployment, integrate real hardware sensors
- Modify `sensor.go` functions to read from actual hardware
- The architecture supports drop-in sensor replacement

### Problem: Weather Function Not Working

**Cause**: Mock weather API or network issues

**Solution**:
- Current implementation uses simulated weather data
- For production, integrate real weather API (OpenWeatherMap, etc.)
- Modify `weather.go` to call actual weather service
- Ensure network connectivity and API keys

---

## API / Function Reference

### Authentication Functions (auth.go)

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

### Authentication Functions (auth.go)

```go
RegisterUser(username, password, role string) error
RegisterGuestUser(username, pin string) error
AuthenticateUser(username, password string) (*User, error)
VerifySession(token string) (*User, error)
LogoutUser(username string) error
HashPassword(password string) (string, error)
CheckPassword(hash, password string) bool
ValidatePassword(password string) error
ValidatePin(pin string) error
```

### User Management Functions (user.go)

```go
CreateGuestAccount(creator, guestName, pin string, creatorRole string) error
CreateTechnicianAccount(homeowner, techName, password string, creatorRole string) error
GrantTechnicianAccess(homeowner, technician string, duration time.Duration, granterRole string) error
RevokeAccess(username string, revokerUsername string, revokerRole string) error
ListAllUsers(requesterRole string) ([]User, error)
DeleteUser(requester, usernameToDelete, requesterRole string) error
ChangePassword(username, oldPassword, newPassword string) error
ChangePIN(username, oldPIN, newPIN string) error
GetUserByUsername(username string) (*User, error)
```

### Sensor Functions (sensor.go)

```go
InitializeSensors() error
ReadTemperature() (float64, error)
ReadHumidity() (float64, error)
ReadCO() (float64, error)
ReadAllSensors() (SensorReading, error)
GetSensorStatus() SensorStatus
```

### HVAC Functions (hvac.go)

```go
InitializeHVAC() error
SetHVACMode(mode string, user *User) error
SetTargetTemperature(temp float64, user *User) error
GetHVACStatus() HVACState
UpdateHVACLogic() error
```

### Profile Functions (profile.go)

```go
CreateProfile(profileName string, targetTemp float64, hvacMode, owner string, user *User, guestAccessible int) error
GetProfile(profileName string) (*Profile, error)
ListProfiles(owner string, user *User) ([]Profile, error)
ApplyProfile(profileName string, user *User) error
DeleteProfile(profileName, username, role string) error
AddSchedule(profileID, dayOfWeek int, startTime, endTime string, targetTemp float64, user *User) error
GetSchedules(profileID int, user *User) ([]Schedule, error)
```

### Energy Functions (energy.go)

```go
GetEnergyUsage(days int) (EnergyStats, error)
GenerateEnergyReport(stats EnergyStats) string
GetDailyEnergyUsage(date time.Time) (float64, error)
GetMonthlyEnergyUsage(year int, month time.Month) (float64, error)
```

### Security Functions (security.go)

```go
SanitizeInput(input string) string
ValidateInput(input string, maxLength int) error
CheckRateLimit(username string, action string, maxAttempts int, window time.Duration) (bool, error)
CheckSQLInjection(input string) bool
ValidateAndSanitizeUsername(username string) (string, error)
ValidateTemperatureInput(temp float64) error
AuditSecurityEvent(eventType, details, username string)
```

### Notification Functions (notifications.go)

```go
SendNotification(username, notifType, message string) error
SendTemperatureAlert(username string, currentTemp, targetTemp float64) error
SendCOAlert(username string, coLevel float64) error
SendSystemAlert(username, alertMessage string) error
BroadcastSystemNotification(message string) error
```

### Logging Functions (logging.go)

```go
LogEvent(eventType, details, username, severity string) error
ViewAuditTrail(limit int) ([]LogEntry, error)
CleanExpiredSessions() error
```

---

## Known Limitations

1. **Simulated Sensors**: Current implementation uses random data generation for demonstration
   - For production: Integrate real hardware sensors (temperature probes, humidity sensors, CO detectors)
   
2. **Mock Weather API**: Uses simulated weather data
   - For production: Integrate real weather service (OpenWeatherMap, Weather.gov, etc.)
   
3. **In-Memory HVAC State**: HVAC control is simulated
   - For production: Integrate with actual HVAC hardware controllers
   
4. **CLI Only**: Command-line interface only
   - For production: Consider web/mobile interface with TLS encryption
   
5. **Single Instance**: No multi-device synchronization
   - For production: Consider cloud backend with synchronization
   
6. **No Email/SMS Notifications**: Alerts are database-only
   - For production: Integrate email/SMS gateway for alerts

---

## Future Enhancements

### Phase 1: Production Readiness
- [ ] Real sensor hardware integration (I2C/SPI interfaces)
- [ ] Actual HVAC system integration (relay controls, protocols)
- [ ] Real weather API integration with API key management
- [ ] Email/SMS notification system
- [ ] Data backup and restore functionality

### Phase 2: Advanced Features
- [ ] Web interface with responsive design
- [ ] Mobile application (iOS/Android)
- [ ] RESTful API with authentication
- [ ] Machine learning for temperature prediction
- [ ] Geofencing for automatic away mode
- [ ] Voice assistant integration (Alexa, Google Home)

### Phase 3: Enterprise Features
- [ ] Cloud synchronization and remote access
- [ ] Multi-thermostat support
- [ ] Advanced analytics and reporting
- [ ] Energy optimization algorithms
- [ ] Integration with smart home platforms
- [ ] Two-factor authentication
- [ ] API rate limiting and DDoS protection
- [ ] Field-level database encryption

---

## Development

### Adding New Features

**Adding a New User Role:**
1. Update `auth.go` to recognize new role
2. Update `user.go` permission checks
3. Update `main.go` menu display logic
4. Add role-specific checks in feature modules
5. Update database schema if needed
6. Add comprehensive tests

**Adding a New Sensor Type:**
1. Extend `sensor.go` with new reading function
2. Update database schema with new sensor table/fields
3. Add validation in `security.go`
4. Update diagnostics in `diagnostics.go`
5. Add alerting logic in `notifications.go`

**Adding a New Security Control:**
1. Implement in `security.go`
2. Apply checks across relevant modules
3. Add audit logging
4. Update tests
5. Document in README

### Running Tests

```bash
go test ./...
```

### Code Review Checklist

Before committing any changes, verify:

- [ ] Input validation on all user inputs
- [ ] Parameterized SQL queries (no string concatenation)
- [ ] Appropriate logging for security events
- [ ] Role-based access control enforced
- [ ] Error messages don't leak sensitive information
- [ ] Password/session handling follows best practices
- [ ] Constants used instead of magic numbers
- [ ] Functions are single-purpose and testable
- [ ] Comments explain "why" not "what"
- [ ] No hardcoded credentials or secrets

### Coding Standards

- **Error Handling**: Always check and handle errors appropriately
- **SQL**: Use parameterized queries exclusively
- **Logging**: Log all security-relevant events
- **Validation**: Validate all external inputs
- **Constants**: Use named constants for configuration values
- **Documentation**: Comment complex logic and security decisions

---

## License

**Academic Project** - EN601.643 Security and Privacy in Computing

This project is developed for educational purposes as part of the Johns Hopkins University course EN601.643. 

**Not for commercial use.**

---

## Support

### Contact Information

**Team Logan Members:**
- **Kailash Parshad**: kailashparshad7724@gmail.com
- **Krishita Choksi**: kchoksi1@jh.edu
- **Dahyun Hong**: dhong15@jh.edu
- **Nina Gao**: ngao6@jh.edu

**Course Information:**
- **Course**: EN601.643 Security and Privacy in Computing
- **Institution**: Johns Hopkins University
- **Semester**: Fall 2025
- **GitHub Repository**: https://github.com/at0m-b0mb/smart-thermostat-security

### Reporting Issues

If you encounter bugs or security issues:

1. **Security vulnerabilities**: Email team members directly (do not open public issues)
2. **Feature requests**: Open GitHub issue with "Feature Request" label
3. **Bug reports**: Open GitHub issue with detailed steps to reproduce

---

## Acknowledgments

- **Course Instructor and TAs** - For guidance on secure software development
- **OWASP Foundation** - For comprehensive security guidelines and Top 10 framework
- **Go Community** - For excellent libraries (sqlite3, bcrypt)
- **Team Logan** - For collaborative development and security-first mindset

---

## Security Disclosure

This system implements multiple security controls and follows OWASP Top 10 guidelines. However, as an academic project, it should be thoroughly audited before any production deployment.

**Security Features Implemented:**
- ✅ Strong authentication with bcrypt
- ✅ Role-based access control
- ✅ SQL injection prevention
- ✅ Input validation and sanitization
- ✅ Session management with expiration
- ✅ Account lockout mechanism
- ✅ Comprehensive audit logging
- ✅ Password complexity requirements
- ✅ Secure session tokens

**Known Security Considerations:**
- Session tokens are stored in database (consider Redis for production)
- No rate limiting on API level (only authentication)
- No field-level database encryption
- No two-factor authentication
- CLI-only interface (no TLS/HTTPS)

---

## Version History

### v1.0 (Current)
- Initial release with full RBAC implementation
- Three user roles: Homeowner, Technician, Guest
- Complete OWASP Top 10 security controls
- SQLite database with comprehensive schema
- Audit logging for all actions
- Profile management with guest accessibility
- Energy usage tracking
- System diagnostics

---

**Built with security in mind. Every line of code reviewed for vulnerabilities.**

---

## Quick Reference Card

### Default Login
```
Username: admin
Password: Admin123!
⚠️ Change immediately after first login!
```

### Guest Account Format
```
Username: {creator}_guest_{name}
Authentication: PIN (4+ numeric digits)
```

### Temperature Limits
```
Minimum: 10°C
Maximum: 35°C
```

### Account Lockout
```
Failed Attempts: 5
Lock Duration: 15 minutes
```

### Session Duration
```
Timeout: 24 hours
```

### Technician Access
```
Type: Time-limited
Granted By: Homeowner only
Duration: Set by homeowner (hours)
```

### Emergency Commands
```bash
# Build
go build -o thermostat

# Run
./thermostat

# View Database
sqlite3 thermostat.db

# Reset (deletes all data)
rm thermostat.db && ./thermostat
```

---

**END OF DOCUMENTATION**
