# Test skripta za SQL Injection napad demonstraciju
. .\https-helper.ps1

Write-Host "=== SQL INJECTION NAPAD DEMONSTRACIJA ===" -ForegroundColor Cyan
Write-Host ""

# Test 1: Osnovni SQL Injection Pattern
Write-Host "Test 1: Osnovni SQL Injection (' OR '1'='1)" -ForegroundColor Yellow
$body = @{
    firstName = "Test' OR '1'='1"
    lastName = "User"
    email = "sqli1@test.com"
    username = "sqliuser1"
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"
Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 400) { "Green" } else { "Red" })
Write-Host "Response: $($result.Content)" -ForegroundColor Gray
if ($result.StatusCode -eq 400) {
    Write-Host "[OK] SQL Injection napad je BLOKIRAN!" -ForegroundColor Green
} else {
    Write-Host "[FAIL] SQL Injection napad je PROSAO!" -ForegroundColor Red
}
Start-Sleep -Seconds 2

# Test 2: SQL Injection sa UNION
Write-Host "`nTest 2: SQL Injection sa UNION SELECT" -ForegroundColor Yellow
$body = @{
    firstName = "Test"
    lastName = "User' UNION SELECT * FROM users--"
    email = "sqli2@test.com"
    username = "sqliuser2"
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"
Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 400) { "Green" } else { "Red" })
if ($result.StatusCode -eq 400) {
    Write-Host "[OK] SQL Injection napad je BLOKIRAN!" -ForegroundColor Green
} else {
    Write-Host "[FAIL] SQL Injection napad je PROSAO!" -ForegroundColor Red
}
Start-Sleep -Seconds 2

# Test 3: SQL Injection sa DROP TABLE
Write-Host "`nTest 3: SQL Injection sa DROP TABLE" -ForegroundColor Yellow
$firstNameSQL = "Test'; DROP TABLE users--"
$body = @{
    firstName = $firstNameSQL
    lastName = "User"
    email = "sqli3@test.com"
    username = "sqliuser3"
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"
Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 400) { "Green" } else { "Red" })
if ($result.StatusCode -eq 400) {
    Write-Host "[OK] SQL Injection napad je BLOKIRAN!" -ForegroundColor Green
} else {
    Write-Host "[FAIL] SQL Injection napad je PROSAO!" -ForegroundColor Red
}

Write-Host "`n=== REZIME ===" -ForegroundColor Cyan
Write-Host "Svi SQL Injection napadi su testirani." -ForegroundColor White
Write-Host "Proverite logove: docker logs projekat-2025-2-users-service-1 | Select-String VALIDATION_FAILURE" -ForegroundColor Gray
