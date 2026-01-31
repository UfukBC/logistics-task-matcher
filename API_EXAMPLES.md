# API Testing Examples

This file contains sample curl commands to test the Task Assignment Engine API.

## Prerequisites
- Server running on http://localhost:8080
- curl installed

## 1. Health Check

```bash
curl http://localhost:8080/health
```

## 2. Create Employees

### Employee 1 - Alice (Helsinki Center)
```bash
curl -X POST http://localhost:8080/employees \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Johnson",
    "location": {
      "lat": 60.1699,
      "lon": 24.9384
    },
    "skills": ["delivery", "driving", "navigation"]
  }'
```

### Employee 2 - Bob (Espoo)
```bash
curl -X POST http://localhost:8080/employees \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Bob Smith",
    "location": {
      "lat": 60.2055,
      "lon": 24.6559
    },
    "skills": ["delivery", "driving"]
  }'
```

### Employee 3 - Charlie (Near Helsinki)
```bash
curl -X POST http://localhost:8080/employees \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Charlie Brown",
    "location": {
      "lat": 60.1741,
      "lon": 24.9416
    },
    "skills": ["delivery", "package_handling"]
  }'
```

### Employee 4 - Diana (Vantaa)
```bash
curl -X POST http://localhost:8080/employees \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Diana Prince",
    "location": {
      "lat": 60.2934,
      "lon": 25.0378
    },
    "skills": ["delivery", "express_delivery"]
  }'
```

### Employee 5 - Eve (Cooking specialist)
```bash
curl -X POST http://localhost:8080/employees \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Eve Martinez",
    "location": {
      "lat": 60.1800,
      "lon": 24.9500
    },
    "skills": ["cooking", "food_preparation"]
  }'
```

## 3. Get All Employees

```bash
curl http://localhost:8080/employees
```

## 4. Create Tasks

### Task 1 - Delivery near Helsinki Center
```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "location": {
      "lat": 60.1700,
      "lon": 24.9400
    },
    "required_skill": "delivery"
  }'
```

### Task 2 - Delivery in Espoo
```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "location": {
      "lat": 60.2050,
      "lon": 24.6600
    },
    "required_skill": "delivery"
  }'
```

### Task 3 - Package handling
```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "location": {
      "lat": 60.1750,
      "lon": 24.9450
    },
    "required_skill": "package_handling"
  }'
```

### Task 4 - Express delivery
```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "location": {
      "lat": 60.2900,
      "lon": 25.0400
    },
    "required_skill": "express_delivery"
  }'
```

### Task 5 - Cooking (will test no available employee scenario)
```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "location": {
      "lat": 60.1800,
      "lon": 24.9500
    },
    "required_skill": "cooking"
  }'
```

## 5. Get All Tasks

```bash
curl http://localhost:8080/tasks
```

## 6. Get Specific Task (replace {task-id} with actual ID)

```bash
curl http://localhost:8080/tasks/{task-id}
```

## Complete Test Workflow

Run this script to test the complete workflow:

```bash
#!/bin/bash

echo "=== Testing Task Assignment Engine ==="
echo ""

echo "1. Health Check"
curl -s http://localhost:8080/health | json_pp
echo ""
sleep 1

echo "2. Creating Employees..."
ALICE=$(curl -s -X POST http://localhost:8080/employees \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","location":{"lat":60.1699,"lon":24.9384},"skills":["delivery"]}')
echo $ALICE | json_pp
sleep 1

BOB=$(curl -s -X POST http://localhost:8080/employees \
  -H "Content-Type: application/json" \
  -d '{"name":"Bob","location":{"lat":60.2055,"lon":24.6559},"skills":["delivery"]}')
echo $BOB | json_pp
sleep 1

echo ""
echo "3. Listing Employees..."
curl -s http://localhost:8080/employees | json_pp
sleep 1

echo ""
echo "4. Creating Task..."
TASK=$(curl -s -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{"location":{"lat":60.1700,"lon":24.9400},"required_skill":"delivery"}')
echo $TASK | json_pp

# Wait for assignment
sleep 2

echo ""
echo "5. Checking Task Status..."
curl -s http://localhost:8080/tasks | json_pp
sleep 1

echo ""
echo "6. Checking Updated Employee Availability..."
curl -s http://localhost:8080/employees | json_pp

echo ""
echo "=== Test Complete ==="
```

## PowerShell Version (for Windows)

```powershell
Write-Host "=== Testing Task Assignment Engine ===" -ForegroundColor Green

Write-Host "`n1. Health Check" -ForegroundColor Cyan
Invoke-RestMethod -Uri "http://localhost:8080/health" | ConvertTo-Json

Start-Sleep -Seconds 1

Write-Host "`n2. Creating Employees..." -ForegroundColor Cyan
$alice = @{
    name = "Alice"
    location = @{ lat = 60.1699; lon = 24.9384 }
    skills = @("delivery")
} | ConvertTo-Json

$bob = @{
    name = "Bob"
    location = @{ lat = 60.2055; lon = 24.6559 }
    skills = @("delivery")
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/employees" -Method Post -Body $alice -ContentType "application/json" | ConvertTo-Json
Start-Sleep -Seconds 1
Invoke-RestMethod -Uri "http://localhost:8080/employees" -Method Post -Body $bob -ContentType "application/json" | ConvertTo-Json

Start-Sleep -Seconds 1

Write-Host "`n3. Listing Employees..." -ForegroundColor Cyan
Invoke-RestMethod -Uri "http://localhost:8080/employees" | ConvertTo-Json

Start-Sleep -Seconds 1

Write-Host "`n4. Creating Task..." -ForegroundColor Cyan
$task = @{
    location = @{ lat = 60.1700; lon = 24.9400 }
    required_skill = "delivery"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/tasks" -Method Post -Body $task -ContentType "application/json" | ConvertTo-Json

Start-Sleep -Seconds 2

Write-Host "`n5. Checking Task Status..." -ForegroundColor Cyan
Invoke-RestMethod -Uri "http://localhost:8080/tasks" | ConvertTo-Json

Write-Host "`n=== Test Complete ===" -ForegroundColor Green
```

## Expected Behavior

1. **Employee Creation**: Each employee should receive a unique UUID and be marked as `available: true`
2. **Task Creation**: Task should be created with `status: pending` and immediately submitted for assignment
3. **Task Assignment**:
   - Within 1-2 seconds, the task should be assigned to the closest available employee
   - Task status should change to `assigned`
   - Assigned employee should be marked as `is_available: false`
4. **No Eligible Employee**: If no employee has the required skill, task status should be `failed`

## Testing Edge Cases

### Test: No Available Employees
1. Create a task requiring "delivery"
2. Create a task requiring "delivery" again
3. Keep creating until all employees are busy
4. Next task should fail with "no eligible employee" error

### Test: Skill Mismatch
1. Create employees with only "cooking" skill
2. Create a task requiring "delivery"
3. Task should fail immediately

### Test: Distance-Based Assignment
1. Create 3 employees at different distances from task location
2. Create a task
3. Verify the closest employee is assigned

## Performance Testing

### Load Test with curl
```bash
for i in {1..100}; do
  curl -X POST http://localhost:8080/tasks \
    -H "Content-Type: application/json" \
    -d "{\"location\":{\"lat\":60.17,\"lon\":24.94},\"required_skill\":\"delivery\"}" &
done
wait
```

### Check Results
```bash
curl http://localhost:8080/tasks | json_pp | grep -c "assigned"
```
