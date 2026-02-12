# Test skripta za XSS napad demonstraciju
. .\https-helper.ps1

Write-Host "=== XSS NAPAD DEMONSTRACIJA ===" -ForegroundColor Cyan
Write-Host ""

# Test 1: Osnovni XSS Pattern
Write-Host "Test 1: Osnovni XSS Pattern (<script>)" -ForegroundColor Yellow
$body = @{
    firstName = "<script>alert('XSS')</script>"
    lastName = "User"
    email = "xss1@test.com"
    username = "xssuser1"
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"
Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 400) { "Green" } else { "Red" })
Write-Host "Response: $($result.Content)" -ForegroundColor Gray
if ($result.StatusCode -eq 400) {
    Write-Host "✓ XSS napad je BLOKIRAN!" -ForegroundColor Green
} else {
    Write-Host "✗ XSS napad je PROŠAO!" -ForegroundColor Red
}
Start-Sleep -Seconds 2

# Test 2: XSS sa Event Handler-om
Write-Host "`nTest 2: XSS sa Event Handler-om (onerror)" -ForegroundColor Yellow
$body = @{
    firstName = "<img src=x onerror=alert('XSS')>"
    lastName = "User"
    email = "xss2@test.com"
    username = "xssuser2"
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"
Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 400) { "Green" } else { "Red" })
if ($result.StatusCode -eq 400) {
    Write-Host "✓ XSS napad je BLOKIRAN!" -ForegroundColor Green
} else {
    Write-Host "✗ XSS napad je PROŠAO!" -ForegroundColor Red
}
Start-Sleep -Seconds 2

# Test 3: XSS sa JavaScript Protocol
Write-Host "`nTest 3: XSS sa JavaScript Protocol" -ForegroundColor Yellow
$body = @{
    firstName = "Test"
    lastName = "User"
    email = "javascript:alert('XSS')@test.com"
    username = "xssuser3"
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"
Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 400) { "Green" } else { "Red" })
if ($result.StatusCode -eq 400) {
    Write-Host "✓ XSS napad je BLOKIRAN!" -ForegroundColor Green
} else {
    Write-Host "✗ XSS napad je PROŠAO!" -ForegroundColor Red
}

Write-Host "`n=== REZIME ===" -ForegroundColor Cyan
Write-Host "Svi XSS napadi su testirani." -ForegroundColor White
Write-Host "Proverite logove: docker logs projekat-2025-2-users-service-1 | grep VALIDATION_FAILURE" -ForegroundColor Gray
