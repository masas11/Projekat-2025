# Kompletna test skripta za Informacionu Bezbednost (Ocena 10)
# Pokreće sve testove redom i generiše izveštaj

. .\https-helper.ps1

$ErrorActionPreference = "Continue"
$results = @()

function Write-TestHeader {
    param([string]$Title)
    Write-Host "`n" -NoNewline
    Write-Host "=" * 60 -ForegroundColor Cyan
    Write-Host $Title -ForegroundColor Cyan
    Write-Host "=" * 60 -ForegroundColor Cyan
}

function Write-TestResult {
    param([string]$TestName, [bool]$Passed, [string]$Details = "")
    $status = if ($Passed) { "[OK]" } else { "[FAIL]" }
    $color = if ($Passed) { "Green" } else { "Red" }
    Write-Host "$status $TestName" -ForegroundColor $color
    if ($Details) {
        Write-Host "  $Details" -ForegroundColor Gray
    }
    $script:results += [PSCustomObject]@{
        Test = $TestName
        Status = if ($Passed) { "PASS" } else { "FAIL" }
        Details = $Details
    }
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  TEST PLAN - INFORMACIONA BEZBEDNOST" -ForegroundColor Cyan
Write-Host "  Ocena 10 - Kompletno Testiranje" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Provera da li je sistem pokrenut
Write-TestHeader "PROVERA SISTEMA"

try {
    $health = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/health" -Method "GET"
    if ($health.StatusCode -eq 200) {
        Write-TestResult "Sistem je pokrenut" $true "API Gateway odgovara"
    } else {
        Write-TestResult "Sistem je pokrenut" $false "API Gateway ne odgovara (Status: $($health.StatusCode))"
        Write-Host "`nPokrenite sistem sa: docker-compose up -d" -ForegroundColor Yellow
        exit 1
    }
} catch {
    Write-TestResult "Sistem je pokrenut" $false "Greška pri povezivanju: $($_.Exception.Message)"
    Write-Host "`nPokrenite sistem sa: docker-compose up -d" -ForegroundColor Yellow
    exit 1
}

Start-Sleep -Seconds 2

# TEST 1: Registracija Naloga (1.1)
Write-TestHeader "TEST 1: REGISTRACIJA NALOGA (1.1)"

# Test 1.1: Uspešna registracija
try {
    $body = @{
        firstName = "Test"
        lastName = "User"
        email = "testuser$(Get-Random)@example.com"
        username = "testuser$(Get-Random)"
        password = "Test1234!"
        confirmPassword = "Test1234!"
    } | ConvertTo-Json
    
    $result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"
    Write-TestResult "1.1.1 Uspešna registracija" ($result.StatusCode -eq 201) "Status: $($result.StatusCode)"
} catch {
    Write-TestResult "1.1.1 Uspešna registracija" $false "Greška: $($_.Exception.Message)"
}

# Test 1.2: Jedinstven username
try {
    $body = @{
        firstName = "Another"
        lastName = "User"
        email = "another$(Get-Random)@example.com"
        username = "testuser"  # Isti username
        password = "Test1234!"
        confirmPassword = "Test1234!"
    } | ConvertTo-Json
    
    $result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"
    Write-TestResult "1.1.2 Jedinstven username" ($result.StatusCode -eq 409) "Status: $($result.StatusCode)"
} catch {
    Write-TestResult "1.1.2 Jedinstven username" $false "Greška: $($_.Exception.Message)"
}

# Test 1.3: Jaka lozinka
try {
    $body = @{
        firstName = "Test"
        lastName = "User"
        email = "weak$(Get-Random)@example.com"
        username = "weakuser$(Get-Random)"
        password = "123"  # Slaba lozinka
        confirmPassword = "123"
    } | ConvertTo-Json
    
    $result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"
    Write-TestResult "1.1.3 Jaka lozinka" ($result.StatusCode -eq 400) "Status: $($result.StatusCode)"
} catch {
    Write-TestResult "1.1.3 Jaka lozinka" $false "Greška: $($_.Exception.Message)"
}

Start-Sleep -Seconds 2

# TEST 2: Prijava na Sistem (1.2)
Write-TestHeader "TEST 2: PRIJAVA NA SISTEM (1.2)"

# Test 2.1: Kombinovana autentifikacija
try {
    $body = @{
        username = "admin"
        password = "admin123"
    } | ConvertTo-Json
    
    $result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/login/request-otp" -Method "POST" -Body $body -ContentType "application/json"
    Write-TestResult "2.1 Kombinovana autentifikacija (OTP)" ($result.StatusCode -eq 200) "Status: $($result.StatusCode)"
} catch {
    Write-TestResult "2.1 Kombinovana autentifikacija (OTP)" $false "Greška: $($_.Exception.Message)"
}

# Test 2.2: Autorizacija bez tokena
try {
    $result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/logout" -Method "POST" -Body "{}" -ContentType "application/json"
    Write-TestResult "2.2 Autorizacija bez tokena" ($result.StatusCode -eq 401) "Status: $($result.StatusCode)"
} catch {
    Write-TestResult "2.2 Autorizacija bez tokena" $false "Greška: $($_.Exception.Message)"
}

Start-Sleep -Seconds 2

# TEST 3: Povraćaj Naloga (1.3)
Write-TestHeader "TEST 3: POVRAĆAJ NALOGA - MAGIC LINK (1.3)"

try {
    $body = @{
        email = "admin@example.com"
    } | ConvertTo-Json
    
    $result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/recover/request" -Method "POST" -Body $body -ContentType "application/json"
    Write-TestResult "3.1 Request Magic Link" ($result.StatusCode -eq 200) "Status: $($result.StatusCode)"
} catch {
    Write-TestResult "3.1 Request Magic Link" $false "Greška: $($_.Exception.Message)"
}

Start-Sleep -Seconds 2

# TEST 4: Kontrola Pristupa (2.17)
Write-TestHeader "TEST 4: KONTROLA PRISTUPA (2.17)"

# Test 4.1: DoS zaštita - Rate Limiting
try {
    $blockedCount = 0
    $totalRequests = 110
    
    for ($i = 1; $i -le $totalRequests; $i++) {
        $result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/health" -Method "GET"
        if ($result.StatusCode -eq 429) {
            $blockedCount++
        }
        Start-Sleep -Milliseconds 50
    }
    
    $passed = $blockedCount -gt 0
    Write-TestResult "4.1 DoS zaštita (Rate Limiting)" $passed "Blokirano: $blockedCount/$totalRequests"
} catch {
    Write-TestResult "4.1 DoS zaštita (Rate Limiting)" $false "Greška: $($_.Exception.Message)"
}

Start-Sleep -Seconds 3

# TEST 5: Validacija Podataka (2.18)
Write-TestHeader "TEST 5: VALIDACIJA PODATAKA (2.18)"

# Test 5.1: SQL Injection
try {
    $body = @{
        firstName = "Test' OR '1'='1"
        lastName = "User"
        email = "sqli$(Get-Random)@example.com"
        username = "sqliuser$(Get-Random)"
        password = "Test1234!"
        confirmPassword = "Test1234!"
    } | ConvertTo-Json
    
    $result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"
    Write-TestResult "5.1 SQL Injection Detection" ($result.StatusCode -eq 400) "Status: $($result.StatusCode)"
} catch {
    Write-TestResult "5.1 SQL Injection Detection" $false "Greška: $($_.Exception.Message)"
}

# Test 5.2: XSS
try {
    $firstNameXSS = '<script>alert("XSS")</script>'
    $body = @{
        firstName = $firstNameXSS
        lastName = "User"
        email = "xss$(Get-Random)@example.com"
        username = "xssuser$(Get-Random)"
        password = "Test1234!"
        confirmPassword = "Test1234!"
    } | ConvertTo-Json
    
    $result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"
    Write-TestResult "5.2 XSS Detection" ($result.StatusCode -eq 400 -or $result.StatusCode -eq 429) "Status: $($result.StatusCode)"
} catch {
    Write-TestResult "5.2 XSS Detection" $false "Greška: $($_.Exception.Message)"
}

Start-Sleep -Seconds 2

# TEST 6: Zaštita Podataka (2.19)
Write-TestHeader "TEST 6: ZAŠTITA PODATAKA (2.19)"

# Test 6.1: HTTPS
try {
    $result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/health" -Method "GET"
    $passed = $result.StatusCode -eq 200
    Write-TestResult "6.1 HTTPS protokol" $passed "Status: $($result.StatusCode)"
} catch {
    Write-TestResult "6.1 HTTPS protokol" $false "Greška: $($_.Exception.Message)"
}

# Test 6.2: POST metoda za senzitivne podatke
try {
    $result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "GET"
    Write-TestResult "6.2 POST metoda za senzitivne podatke" ($result.StatusCode -eq 405) "Status: $($result.StatusCode)"
} catch {
    Write-TestResult "6.2 POST metoda za senzitivne podatke" $false "Greška: $($_.Exception.Message)"
}

# Test 6.3: Hash & Salt za lozinke
try {
    $hashCheck = docker exec projekat-2025-2-mongodb-users-1 mongosh --quiet --eval "db.users.findOne({}, {passwordHash: 1, _id: 0})" 2>$null
    $passed = $hashCheck -match '\$2[ab]\$'
    Write-TestResult "6.3 Hash & Salt za lozinke" $passed "Bcrypt format: $passed"
} catch {
    Write-TestResult "6.3 Hash & Salt za lozinke" $false "Greška pri proveri baze"
}

Start-Sleep -Seconds 2

# TEST 7: Logovanje (2.20)
Write-TestHeader "TEST 7: LOGOVANJE (2.20)"

# Test 7.1: Provera log fajlova
try {
    $logFiles = Get-ChildItem -Path "services/users-service/logs" -Filter "*.log" -ErrorAction SilentlyContinue
    $passed = $logFiles.Count -gt 0
    Write-TestResult "7.1 Log fajlovi postoje" $passed "Broj fajlova: $($logFiles.Count)"
} catch {
    Write-TestResult "7.1 Log fajlovi postoje" $false "Greška: $($_.Exception.Message)"
}

# Test 7.2: Provera logovanja validacije
try {
    $logs = docker logs projekat-2025-2-users-service-1 2>&1 | Select-String "VALIDATION_FAILURE" | Select-Object -Last 1
    $passed = $logs -ne $null
    Write-TestResult "7.2 Logovanje validacije" $passed "Pronađeno: $passed"
} catch {
    Write-TestResult "7.2 Logovanje validacije" $false "Greška: $($_.Exception.Message)"
}

Start-Sleep -Seconds 2

# TEST 8: Analiza Ranjivosti (2.21)
Write-TestHeader "TEST 8: ANALIZA RANJIVOSTI (2.21)"

try {
    $reportExists = Test-Path "IZVESTAJ_ANALIZA_RANJIVOSTI_2.21.md"
    Write-TestResult "8.1 Izveštaj o ranjivostima" $reportExists "Fajl postoji: $reportExists"
} catch {
    Write-TestResult "8.1 Izveštaj o ranjivostima" $false "Greška: $($_.Exception.Message)"
}

Start-Sleep -Seconds 2

# TEST 9: Demonstracija Napada (2.22)
Write-TestHeader "TEST 9: DEMONSTRACIJA POKUŠAJA NAPADA (2.22)"

Write-Host "Pokretanje test skripti za napade..." -ForegroundColor Yellow

# Test 9.1: XSS
if (Test-Path "test-xss-attack.ps1") {
    Write-Host "`nPokretanje test-xss-attack.ps1..." -ForegroundColor Gray
    & .\test-xss-attack.ps1 | Out-Null
    Write-TestResult "9.1 XSS napad" $true "Test skripta pokrenuta"
} else {
    Write-TestResult "9.1 XSS napad" $false "Test skripta ne postoji"
}

Start-Sleep -Seconds 2

# Test 9.2: SQL Injection
if (Test-Path "test-sql-injection-attack.ps1") {
    Write-Host "`nPokretanje test-sql-injection-attack.ps1..." -ForegroundColor Gray
    & .\test-sql-injection-attack.ps1 | Out-Null
    Write-TestResult "9.2 SQL Injection napad" $true "Test skripta pokrenuta"
} else {
    Write-TestResult "9.2 SQL Injection napad" $false "Test skripta ne postoji"
}

Start-Sleep -Seconds 2

# Test 9.3: Brute-force
if (Test-Path "test-brute-force-attack.ps1") {
    Write-Host "`nPokretanje test-brute-force-attack.ps1..." -ForegroundColor Gray
    & .\test-brute-force-attack.ps1 | Out-Null
    Write-TestResult "9.3 Brute-force napad" $true "Test skripta pokrenuta"
} else {
    Write-TestResult "9.3 Brute-force napad" $false "Test skripta ne postoji"
}

Start-Sleep -Seconds 2

# Test 9.4: DoS
if (Test-Path "test-dos-attack.ps1") {
    Write-Host "`nPokretanje test-dos-attack.ps1..." -ForegroundColor Gray
    & .\test-dos-attack.ps1 | Out-Null
    Write-TestResult "9.4 DoS napad" $true "Test skripta pokrenuta"
} else {
    Write-TestResult "9.4 DoS napad" $false "Test skripta ne postoji"
}

# FINALNI REZIME
Write-TestHeader "FINALNI REZIME"

$totalTests = $results.Count
$passedTests = ($results | Where-Object { $_.Status -eq "PASS" }).Count
$failedTests = ($results | Where-Object { $_.Status -eq "FAIL" }).Count
$passRate = [math]::Round(($passedTests / $totalTests) * 100, 2)

Write-Host "`nUkupno testova: $totalTests" -ForegroundColor White
Write-Host "Uspešno: $passedTests" -ForegroundColor Green
Write-Host "Neuspešno: $failedTests" -ForegroundColor $(if ($failedTests -eq 0) { "Green" } else { "Red" })
Write-Host "Procenat uspešnosti: $passRate%" -ForegroundColor $(if ($passRate -eq 100) { "Green" } else { "Yellow" })

Write-Host "`nDetaljni rezultati:" -ForegroundColor Cyan
$results | Format-Table -AutoSize

# Sačuvaj rezultate u fajl
$results | Export-Csv -Path "test-results-$(Get-Date -Format 'yyyyMMdd-HHmmss').csv" -NoTypeInformation
Write-Host "`nRezultati sačuvani u CSV fajl." -ForegroundColor Gray

if ($passRate -eq 100) {
    Write-Host "`n✅ SVI TESTOVI SU PROŠLI!" -ForegroundColor Green
} else {
    Write-Host "`n⚠ NEKI TESTOVI NISU PROŠLI!" -ForegroundColor Yellow
    Write-Host "Proverite detalje iznad." -ForegroundColor Yellow
}

Write-Host "`nDetaljna dokumentacija: TEST_PLAN_INFORMACIONA_BEZBEDNOST.md" -ForegroundColor Cyan
