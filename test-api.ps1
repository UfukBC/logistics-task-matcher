# Quick Test Script for Task Assignment Engine
# This script demonstrates the complete workflow

Write-Host "`n=== Task Assignment Engine - Quick Test ===" -ForegroundColor Green
Write-Host "Starting test workflow...`n" -ForegroundColor Yellow

$baseUrl = "http://localhost:8080"

# Function to display JSON nicely
function Show-Result {
    param($response, $title)
    Write-Host "`n--- $title ---" -ForegroundColor Cyan
    $response | ConvertTo-Json -Depth 10
}

try {
    # 1. Health Check
    Write-Host "1. Checking API health..." -ForegroundColor Yellow
    $health = Invoke-RestMethod -Uri "$baseUrl/health"
    Show-Result $health "Health Check"
    Start-Sleep -Seconds 1

    # 2. Create Employees
    Write-Host "`n2. Creating employees..." -ForegroundColor Yellow

    $alice = @{
        name = "Alice Johnson"
        location = @{ lat = 60.1699; lon = 24.9384 }
        skills = @("delivery", "driving")
    } | ConvertTo-Json

    $aliceResult = Invoke-RestMethod -Uri "$baseUrl/employees" -Method Post -Body $alice -ContentType "application/json"
    Show-Result $aliceResult "Employee: Alice (Helsinki Center)"
    Start-Sleep -Seconds 1

    $bob = @{
        name = "Bob Smith"
        location = @{ lat = 60.2055; lon = 24.6559 }
        skills = @("delivery")
    } | ConvertTo-Json

    $bobResult = Invoke-RestMethod -Uri "$baseUrl/employees" -Method Post -Body $bob -ContentType "application/json"
    Show-Result $bobResult "Employee: Bob (Espoo - Far)"
    Start-Sleep -Seconds 1

    $charlie = @{
        name = "Charlie Brown"
        location = @{ lat = 60.1741; lon = 24.9416 }
        skills = @("delivery")
    } | ConvertTo-Json

    $charlieResult = Invoke-RestMethod -Uri "$baseUrl/employees" -Method Post -Body $charlie -ContentType "application/json"
    Show-Result $charlieResult "Employee: Charlie (Near Helsinki)"
    Start-Sleep -Seconds 1

    # 3. List all employees
    Write-Host "`n3. Listing all employees..." -ForegroundColor Yellow
    $employees = Invoke-RestMethod -Uri "$baseUrl/employees"
    Show-Result $employees "All Employees (all available: true)"
    Start-Sleep -Seconds 1

    # 4. Create a task
    Write-Host "`n4. Creating task near Helsinki center..." -ForegroundColor Yellow
    $task = @{
        location = @{ lat = 60.1700; lon = 24.9400 }
        required_skill = "delivery"
    } | ConvertTo-Json

    $taskResult = Invoke-RestMethod -Uri "$baseUrl/tasks" -Method Post -Body $task -ContentType "application/json"
    Show-Result $taskResult "Task Created (pending)"

    Write-Host "`nWaiting for assignment (2 seconds)..." -ForegroundColor Yellow
    Start-Sleep -Seconds 2

    # 5. Check task status
    Write-Host "`n5. Checking task assignment status..." -ForegroundColor Yellow
    $tasks = Invoke-RestMethod -Uri "$baseUrl/tasks"
    Show-Result $tasks "All Tasks"

    # Extract assigned employee
    $assignedTask = $tasks.data[0]
    if ($assignedTask.status -eq "assigned") {
        Write-Host "`nSUCCESS: Task assigned to employee: $($assignedTask.assigned_employee_id)" -ForegroundColor Green
    } else {
        Write-Host "`nWARNING: Task not yet assigned. Status: $($assignedTask.status)" -ForegroundColor Yellow
    }

    # 6. Check employee availability
    Write-Host "`n6. Checking employee availability after assignment..." -ForegroundColor Yellow
    $updatedEmployees = Invoke-RestMethod -Uri "$baseUrl/employees"
    Show-Result $updatedEmployees "Updated Employees"

    $assignedEmployee = $updatedEmployees.data | Where-Object { $_.id -eq $assignedTask.assigned_employee_id }
    if ($assignedEmployee) {
        Write-Host "`nAssigned employee '$($assignedEmployee.name)' availability: $($assignedEmployee.is_available)" -ForegroundColor $(if ($assignedEmployee.is_available) { "Yellow" } else { "Green" })
    }

    # 7. Create another task
    Write-Host "`n7. Creating second task..." -ForegroundColor Yellow
    $task2 = @{
        location = @{ lat = 60.2000; lon = 24.7000 }
        required_skill = "delivery"
    } | ConvertTo-Json

    $task2Result = Invoke-RestMethod -Uri "$baseUrl/tasks" -Method Post -Body $task2 -ContentType "application/json"
    Show-Result $task2Result "Second Task Created"

    Write-Host "`nWaiting for assignment (2 seconds)..." -ForegroundColor Yellow
    Start-Sleep -Seconds 2

    # 8. Final status
    Write-Host "`n8. Final status check..." -ForegroundColor Yellow
    $finalTasks = Invoke-RestMethod -Uri "$baseUrl/tasks"
    Show-Result $finalTasks "All Tasks (Final)"

    Write-Host "`n=== Test Completed Successfully ===" -ForegroundColor Green
    Write-Host "`nSummary:" -ForegroundColor Cyan
    Write-Host "- Created 3 employees" -ForegroundColor White
    Write-Host "- Created 2 tasks" -ForegroundColor White
    Write-Host "- Both tasks should be assigned to the nearest available employees" -ForegroundColor White
    Write-Host "- The first task should be assigned to Alice (closest to task location)" -ForegroundColor White
    Write-Host "- The second task should be assigned to Charlie or Bob (Alice is now unavailable)" -ForegroundColor White

} catch {
    Write-Host "`n=== ERROR ===" -ForegroundColor Red
    Write-Host "Failed to complete test. Make sure the server is running on $baseUrl" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "`nTo start the server, run:" -ForegroundColor Yellow
    Write-Host "  go run ." -ForegroundColor White
    Write-Host "or" -ForegroundColor Yellow
    Write-Host "  docker-compose up" -ForegroundColor White
}

Write-Host ""
