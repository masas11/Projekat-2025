# API Gateway Test Script
Write-Host "=== Testing API Gateway Endpoints ===" -ForegroundColor Green
Write-Host ""

# Test 1: Users Health
Write-Host "1. GET /api/users/health" -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri http://localhost:8080/api/users/health -UseBasicParsing
    Write-Host "   Status: $($response.StatusCode) - $($response.Content)" -ForegroundColor Green
} catch {
    Write-Host "   Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# Test 2: Content Health
Write-Host "2. GET /api/content/health" -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri http://localhost:8080/api/content/health -UseBasicParsing
    Write-Host "   Status: $($response.StatusCode) - $($response.Content)" -ForegroundColor Green
} catch {
    Write-Host "   Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# Test 3: Register
Write-Host "3. POST /api/users/register" -ForegroundColor Yellow
$registerBody = @{
    firstName = "Test"
    lastName = "User"
    email = "test@example.com"
    username = "testuser123"
    password = "StrongP@ss123"
    confirmPassword = "StrongP@ss123"
} | ConvertTo-Json

try {
    $response = Invoke-WebRequest -Uri http://localhost:8080/api/users/register -Method POST -Body $registerBody -ContentType "application/json" -UseBasicParsing
    Write-Host "   Status: $($response.StatusCode) - $($response.Content)" -ForegroundColor Green
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    $stream = $_.Exception.Response.GetResponseStream()
    $reader = New-Object System.IO.StreamReader($stream)
    $responseBody = $reader.ReadToEnd()
    Write-Host "   Status: $statusCode - $responseBody" -ForegroundColor Cyan
}

Write-Host ""

# Test 4: Login Request OTP
Write-Host "4. POST /api/users/login/request-otp" -ForegroundColor Yellow
$loginBody = @{
    username = "testuser123"
    password = "StrongP@ss123"
} | ConvertTo-Json

try {
    $response = Invoke-WebRequest -Uri http://localhost:8080/api/users/login/request-otp -Method POST -Body $loginBody -ContentType "application/json" -UseBasicParsing
    Write-Host "   Status: $($response.StatusCode)" -ForegroundColor Green
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    Write-Host "   Status: $statusCode (Expected if user doesn't exist)" -ForegroundColor Cyan
}

Write-Host ""
Write-Host "=== All Tests Complete ===" -ForegroundColor Green
Write-Host ""
Write-Host "Available endpoints:" -ForegroundColor White
Write-Host "  GET  http://localhost:8080/api/users/health" -ForegroundColor Gray
Write-Host "  GET  http://localhost:8080/api/content/health" -ForegroundColor Gray
Write-Host "  POST http://localhost:8080/api/users/register" -ForegroundColor Gray
Write-Host "  POST http://localhost:8080/api/users/login/request-otp" -ForegroundColor Gray
Write-Host "  POST http://localhost:8080/api/users/login/verify-otp" -ForegroundColor Gray
Write-Host "  POST http://localhost:8080/api/users/password/change" -ForegroundColor Gray
Write-Host "  POST http://localhost:8080/api/users/password/reset" -ForegroundColor Gray
