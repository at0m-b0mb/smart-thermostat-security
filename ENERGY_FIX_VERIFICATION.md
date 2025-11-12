# Energy Usage Tracking Fix - Verification Guide

## Problem
Energy usage was showing 0.0 for all HVAC modes (fan, heat, cool) even when the system was running for a long time.

## Root Causes
1. **Fan mode never logged energy**: When fan mode was active, energy was never tracked
2. **Energy only logged on stop**: Heat and cool modes only logged energy when stopping, not during continuous operation  
3. **Insufficient runtime**: The update loop runs every 30 seconds, but energy wasn't logged until at least 1 minute had passed
4. **No periodic tracking**: Long-running operations weren't tracked until they stopped

## Solution
Modified `hvac.go` to add periodic energy logging:

1. **Added `lastEnergyLog` variable**: Tracks when energy was last logged to prevent too-frequent logging
2. **Added periodic logging**: When HVAC is running, energy is logged every 2 minutes
3. **Fixed fan mode**: Fan mode now properly starts tracking time and logs energy periodically
4. **Fixed mode OFF**: When switching to OFF, any running energy is logged before stopping

## How It Works Now

### Fan Mode
- First call: Sets IsRunning=true, startTime=now, logs "Fan started"
- Every 30 seconds after: Calls logPeriodicRuntime()
- After 2+ minutes: Logs accumulated energy, resets startTime
- Continues logging every 2 minutes while running

### Heat/Cool Mode  
- When starts running: Sets startTime=now
- While continuously running (target not reached): Calls logPeriodicRuntime() every 30 seconds
- After 2+ minutes of running: Logs accumulated energy, resets startTime
- When stops (target reached): Logs final accumulated energy, resets startTime

## Verification Steps

### Quick Test (2-3 minutes)
1. Start the thermostat: `./thermostat`
2. Login with admin/Admin123!
3. Set mode to fan: Option 3 → 4 (Fan)
4. Wait 2-3 minutes
5. Check energy: Option 6 → Enter 1 (for 1 day)
6. Should see:
   - Total Energy Used: > 0.0 kWh
   - Fan: > 0.0 kWh

### Full Test (5 minutes)
1. Login and set to cool mode: Option 3 → 3 (Cool)
2. Set target to 11°C: Option 2 → 11
3. Wait 2-3 minutes (system should run continuously)
4. Check energy: Option 6
5. Should see cooling energy > 0.0 kWh

### Database Verification
```bash
sqlite3 thermostat.db
SELECT hvac_mode, runtime_minutes, estimated_kwh, datetime(timestamp, 'localtime') 
FROM energy_logs 
ORDER BY timestamp DESC 
LIMIT 10;
```

Expected output: Multiple entries showing energy logs every 2 minutes while system is running

## Expected Energy Values

### Fan Mode (0.5 kWh per hour)
- 2 minutes = 0.0167 kWh
- 4 minutes = 0.0333 kWh  
- 10 minutes = 0.0833 kWh

### Cooling Mode (3.0 kWh per hour)
- 2 minutes = 0.1 kWh
- 4 minutes = 0.2 kWh
- 10 minutes = 0.5 kWh

### Heating Mode (2.5 kWh per hour)
- 2 minutes = 0.0833 kWh
- 4 minutes = 0.1667 kWh
- 10 minutes = 0.4167 kWh

## Timing Notes
- HVAC control loop runs every 30 seconds
- Energy is logged every 2+ minutes during continuous operation
- Minimum runtime of 1 minute required before any energy is logged (int truncation)
- So expect first energy log after ~2 minutes of continuous operation

## Files Changed
- `hvac.go`: Added periodic energy logging functionality
