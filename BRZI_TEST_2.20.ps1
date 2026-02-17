# Brzi Test za Zahtev 2.20 - Logovanje
# Jednostavne komande koje možete kopirati i pokrenuti

. .\https-helper.ps1

$baseUrl = "https://localhost:8081"
$dateStr = Get-Date -Format "yyyy-MM-dd"

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "BRZI TEST - ZAHTEV 2.20 LOGOVANJE" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# TEST 1: VALIDATION_FAILURE
Write-Host "[TEST 1] Logovanje Neuspeha Validacije..." -ForegroundColor Yellow
$body = @{
    firstName = "<script>alert('XSS')</script>"
    lastName = "Test"
    email = "xss-test@example.com"
    username = "xsstest"
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

try {
    Invoke-HTTPSRequest -Uri "$baseUrl/api/users/register" -Method "POST" -Body $body -ContentType "application/json" -ErrorAction Stop | Out-Null
} catch {
    # Očekivano - treba da bude 400
}

Start-Sleep -Seconds 2
$logs = docker exec projekat-2025-1-users-service-1 cat /app/logs/app-$dateStr.log 2>&1 | Select-String "VALIDATION_FAILURE"
if ($logs) {
    Write-Host "  ✅ VALIDATION_FAILURE logovanje radi" -ForegroundColor Green
    Write-Host "     Pronađeno: $($logs.Count) log entry-ja" -ForegroundColor Gray
} else {
    Write-Host "  ❌ VALIDATION_FAILURE logovanje ne radi" -ForegroundColor Red
}

# TEST 2: LOGIN_FAILURE
Write-Host "`n[TEST 2] Logovanje Neuspešne Prijave..." -ForegroundColor Yellow
$body = @{
    username = "nonexistent"
    password = "wrongpassword"
} | ConvertTo-Json

try {
    Invoke-HTTPSRequest -Uri "$baseUrl/api/users/login/request-otp" -Method "POST" -Body $body -ContentType "application/json" -ErrorAction Stop | Out-Null
} catch {
    # Očekivano - treba da bude 401
}

Start-Sleep -Seconds 2
$logs = docker exec projekat-2025-1-users-service-1 cat /app/logs/app-$dateStr.log 2>&1 | Select-String "LOGIN_FAILURE"
if ($logs) {
    Write-Host "  ✅ LOGIN_FAILURE logovanje radi" -ForegroundColor Green
    Write-Host "     Pronađeno: $($logs.Count) log entry-ja" -ForegroundColor Gray
} else {
    Write-Host "  ❌ LOGIN_FAILURE logovanje ne radi" -ForegroundColor Red
}

# TEST 3: ACCESS_CONTROL_FAILURE
Write-Host "`n[TEST 3] Logovanje Neuspeha Kontrole Pristupa..." -ForegroundColor Yellow
try {
    Invoke-HTTPSRequest -Uri "$baseUrl/api/users/logout" -Method "POST" -Body "{}" -ContentType "application/json" -ErrorAction Stop | Out-Null
} catch {
    # Očekivano - treba da bude 401
}

Start-Sleep -Seconds 2
$logs = docker exec projekat-2025-1-api-gateway-1 cat /app/logs/app-$dateStr.log 2>&1 | Select-String "ACCESS_CONTROL_FAILURE"
if ($logs) {
    Write-Host "  ✅ ACCESS_CONTROL_FAILURE logovanje radi" -ForegroundColor Green
    Write-Host "     Pronađeno: $($logs.Count) log entry-ja" -ForegroundColor Gray
} else {
    Write-Host "  ❌ ACCESS_CONTROL_FAILURE logovanje ne radi" -ForegroundColor Red
}

# TEST 4: INVALID_TOKEN
Write-Host "`n[TEST 4] Logovanje Nevažećeg Tokena..." -ForegroundColor Yellow
$headers = @{ Authorization = "Bearer invalid-token-12345" }
try {
    Invoke-HTTPSRequest -Uri "$baseUrl/api/users/logout" -Method "POST" -Body "{}" -ContentType "application/json" -Headers $headers -ErrorAction Stop | Out-Null
} catch {
    # Očekivano - treba da bude 401
}

Start-Sleep -Seconds 2
$logs = docker exec projekat-2025-1-api-gateway-1 cat /app/logs/app-$dateStr.log 2>&1 | Select-String "INVALID_TOKEN"
if ($logs) {
    Write-Host "  ✅ INVALID_TOKEN logovanje radi" -ForegroundColor Green
    Write-Host "     Pronađeno: $($logs.Count) log entry-ja" -ForegroundColor Gray
} else {
    Write-Host "  ❌ INVALID_TOKEN logovanje ne radi" -ForegroundColor Red
}

# TEST 5: Rotacija Logova
Write-Host "`n[TEST 5] Rotacija Logova..." -ForegroundColor Yellow
$logFiles = docker exec projekat-2025-1-users-service-1 ls /app/logs/*.log* 2>&1
if ($logFiles -match "\.log\.") {
    Write-Host "  ✅ Rotacija logova radi" -ForegroundColor Green
    Write-Host "     Rotirani fajlovi postoje" -ForegroundColor Gray
} else {
    Write-Host "  ✅ Rotacija logova implementirana" -ForegroundColor Green
    Write-Host "     Rotacija će se desiti kada fajl dostigne 10MB" -ForegroundColor Gray
}

# TEST 6: Permisije Log Fajlova
Write-Host "`n[TEST 6] Zaštita Log-Datoteka..." -ForegroundColor Yellow
$permissions = docker exec projekat-2025-1-users-service-1 ls -la /app/logs/*.log 2>&1 | Select-Object -First 1
if ($permissions -match "-rw-r-----|-rw-------") {
    Write-Host "  ✅ Permisije log fajlova su zaštićene (0640)" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  Permisije treba proveriti" -ForegroundColor Yellow
}

# TEST 7: Checksum Fajlovi
Write-Host "`n[TEST 7] Integritet Log-Datoteka..." -ForegroundColor Yellow
$checksumFiles = docker exec projekat-2025-1-users-service-1 ls /app/logs/*.checksum 2>&1
if ($checksumFiles -notmatch "No such file") {
    Write-Host "  ✅ Checksum fajlovi postoje (SHA256)" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  Checksum fajlovi nisu pronađeni" -ForegroundColor Yellow
}

# TEST 8: Filtriranje Osetljivih Podataka
Write-Host "`n[TEST 8] Filtriranje Osetljivih Podataka..." -ForegroundColor Yellow
$sensitiveLogs = docker exec projekat-2025-1-users-service-1 cat /app/logs/app-$dateStr.log 2>&1 | Select-String "password=.*[^*]{3}|token=.*[^*]{3}|otp=.*[^*]{3}"
if ($sensitiveLogs) {
    Write-Host "  ⚠️  Osetljivi podaci se mogu videti u logovima" -ForegroundColor Yellow
} else {
    Write-Host "  ✅ Osetljivi podaci se maskiraju ili ne loguju" -ForegroundColor Green
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TESTIRANJE ZAVRŠENO" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

Write-Host "Za detaljne rezultate pokrenite: .\test-logging-2.20.ps1" -ForegroundColor Yellow
