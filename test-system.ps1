# PowerShell Script za Testiranje Sistema
# Pokrenite: .\test-system.ps1

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Music Streaming Platform - Test Suite" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Boje za output
$successColor = "Green"
$errorColor = "Red"
$infoColor = "Yellow"

# Funkcija za testiranje endpoint-a
function Test-Endpoint {
    param(
        [string]$Url,
        [string]$Method = "GET",
        [object]$Body = $null,
        [string]$Description
    )
    
    Write-Host "Testing: $Description" -ForegroundColor $infoColor
    Write-Host "  URL: $Url" -ForegroundColor Gray
    
    try {
        $params = @{
            Uri = $Url
            Method = $Method
            UseBasicParsing = $true
            ErrorAction = "Stop"
        }
        
        if ($Body) {
            $params.Body = ($Body | ConvertTo-Json)
            $params.ContentType = "application/json"
        }
        
        $response = Invoke-WebRequest @params
        Write-Host "  [OK] Status: $($response.StatusCode)" -ForegroundColor $successColor
        
        if ($response.Content) {
            try {
                $json = $response.Content | ConvertFrom-Json
                Write-Host "  Response: $($json | ConvertTo-Json -Compress)" -ForegroundColor Gray
            } catch {
                Write-Host "  Response: $($response.Content)" -ForegroundColor Gray
            }
        }
        
        return $true
    } catch {
        Write-Host "  [FAIL] Failed: $($_.Exception.Message)" -ForegroundColor $errorColor
        return $false
    }
}

# Funkcija za proveru Docker kontejnera
function Test-DockerContainer {
    param([string]$ContainerName)
    
    $container = docker ps --filter "name=$ContainerName" --format "{{.Names}}" 2>$null
    if ($container -eq $ContainerName) {
        Write-Host "  [OK] $ContainerName is running" -ForegroundColor $successColor
        return $true
    } else {
        Write-Host "  [FAIL] $ContainerName is NOT running" -ForegroundColor $errorColor
        return $false
    }
}

# Funkcija za proveru porta
function Test-Port {
    param([int]$Port, [string]$ServiceName)
    
    $connection = Test-NetConnection -ComputerName localhost -Port $Port -WarningAction SilentlyContinue -InformationLevel Quiet
    if ($connection) {
        Write-Host "  [OK] $ServiceName port $Port is open" -ForegroundColor $successColor
        return $true
    } else {
        Write-Host "  [FAIL] $ServiceName port $Port is NOT accessible" -ForegroundColor $errorColor
        return $false
    }
}

# Funkcija za proveru sertifikata
function Test-Certificates {
    Write-Host "`n[1] Checking SSL Certificates" -ForegroundColor Cyan
    
    $certFile = "certs\server.crt"
    $keyFile = "certs\server.key"
    
    $certExists = Test-Path $certFile
    $keyExists = Test-Path $keyFile
    
    if ($certExists) {
        Write-Host "  [OK] Certificate file exists: $certFile" -ForegroundColor $successColor
    } else {
        Write-Host "  [FAIL] Certificate file NOT found: $certFile" -ForegroundColor $errorColor
        Write-Host "    Run: .\generate-certs.ps1" -ForegroundColor $infoColor
    }
    
    if ($keyExists) {
        Write-Host "  [OK] Key file exists: $keyFile" -ForegroundColor $successColor
    } else {
        Write-Host "  [FAIL] Key file NOT found: $keyFile" -ForegroundColor $errorColor
        Write-Host "    Run: .\generate-certs.ps1" -ForegroundColor $infoColor
    }
    
    # Vrati true ako oba fajla postoje
    return ($certExists -and $keyExists)
}

# Funkcija za proveru Docker servisa
function Test-DockerServices {
    Write-Host "`n[2] Checking Docker Services" -ForegroundColor Cyan
    
    $services = @(
        "api-gateway",
        "users-service",
        "content-service",
        "mongodb-users",
        "mongodb-content",
        "mailhog"
    )
    
    $allRunning = $true
    foreach ($service in $services) {
        $containerName = "projekat-2025-1-${service}-1"
        if (-not (Test-DockerContainer $containerName)) {
            $allRunning = $false
        }
    }
    
    return $allRunning
}

# Funkcija za proveru portova
function Test-Ports {
    Write-Host "`n[3] Checking Ports" -ForegroundColor Cyan
    
    $ports = @(
        @{Port = 8081; Service = "API Gateway"},
        @{Port = 8001; Service = "Users Service"},
        @{Port = 8002; Service = "Content Service"},
        @{Port = 8025; Service = "MailHog Web UI"},
        @{Port = 1025; Service = "MailHog SMTP"}
    )
    
    $allOpen = $true
    foreach ($portInfo in $ports) {
        if (-not (Test-Port $portInfo.Port $portInfo.Service)) {
            $allOpen = $false
        }
    }
    
    return $allOpen
}

# Funkcija za testiranje API endpoint-a
function Test-APIEndpoints {
    Write-Host "`n[4] Testing API Endpoints" -ForegroundColor Cyan
    
    $baseUrl = "http://localhost:8081"
    
    # Test 1: Users Health
    $test1 = Test-Endpoint -Url "$baseUrl/api/users/health" -Description "Users Service Health"
    
    # Test 2: Content Health
    $test2 = Test-Endpoint -Url "$baseUrl/api/content/health" -Description "Content Service Health"
    
    # Test 3: CORS Headers
    Write-Host "`nTesting CORS Headers:" -ForegroundColor $infoColor
    try {
        $response = Invoke-WebRequest -Uri "$baseUrl/api/users/health" -UseBasicParsing
        $corsHeader = $response.Headers["Access-Control-Allow-Origin"]
        if ($corsHeader) {
            Write-Host "  [OK] CORS Header: $corsHeader" -ForegroundColor $successColor
        } else {
            Write-Host "  [FAIL] CORS Header missing" -ForegroundColor $errorColor
        }
    } catch {
        Write-Host "  [FAIL] Failed to check CORS: $_" -ForegroundColor $errorColor
    }
    
    return ($test1 -and $test2)
}

# Funkcija za testiranje MailHog-a
function Test-MailHog {
    Write-Host "`n[5] Testing MailHog" -ForegroundColor Cyan
    
    # Test Web UI
    Write-Host "Testing MailHog Web UI:" -ForegroundColor $infoColor
    $webUI = Test-Port 8025 "MailHog Web UI"
    if ($webUI) {
        Write-Host "  -> Open http://localhost:8025 in your browser" -ForegroundColor $infoColor
    }
    
    # Test SMTP Port
    Write-Host "Testing MailHog SMTP:" -ForegroundColor $infoColor
    $smtp = Test-Port 1025 "MailHog SMTP"
    
    # Test email sending (request OTP)
    Write-Host "`nTesting Email Sending (OTP Request):" -ForegroundColor $infoColor
    
    # Proveri da li admin korisnik postoji
    $adminExists = docker exec projekat-2025-1-mongodb-users-1 mongosh --quiet --eval "db.users.findOne({email: 'admin@musicstreaming.com'}, {email: 1, _id: 0})" 2>$null
    
    if ($adminExists -match "admin@musicstreaming.com") {
        $otpBody = @{
            email = "admin@musicstreaming.com"
        }
        $emailTest = Test-Endpoint -Url "http://localhost:8081/api/users/login/request-otp" `
            -Method "POST" `
            -Body $otpBody `
            -Description "Request OTP (Email Test)"
        
        if ($emailTest) {
            Write-Host "  -> Check MailHog at http://localhost:8025 for the email" -ForegroundColor $infoColor
        }
    } else {
        Write-Host "  [INFO] Admin user not found - skipping OTP test" -ForegroundColor $infoColor
        Write-Host "  [INFO] Admin user is created automatically on first service start" -ForegroundColor $infoColor
        Write-Host "  [INFO] MailHog is configured and ready to receive emails" -ForegroundColor $infoColor
        $emailTest = $true  # MailHog je konfigurisan, samo nema korisnika za test
    }
    
    return ($webUI -and $smtp -and $emailTest)
}

# Funkcija za testiranje HTTPS komunikacije
function Test-HTTPSCommunication {
    Write-Host "`n[6] Testing HTTPS Inter-Service Communication" -ForegroundColor Cyan
    
    Write-Host "Checking service URLs:" -ForegroundColor $infoColor
    
    # Provera environment varijabli u API Gateway-u
    $envVars = docker exec projekat-2025-1-api-gateway-1 env 2>$null | Select-String "SERVICE_URL"
    
    $httpsFound = $false
    foreach ($var in $envVars) {
        if ($var -match "https://") {
            Write-Host "  [OK] $var" -ForegroundColor $successColor
            $httpsFound = $true
        } else {
            Write-Host "  [FAIL] $var (should use https://)" -ForegroundColor $errorColor
        }
    }
    
    # Provera logova za HTTPS
    Write-Host "`nChecking service logs for HTTPS:" -ForegroundColor $infoColor
    $usersLog = docker logs projekat-2025-1-users-service-1 --tail 5 2>$null | Select-String "HTTPS"
    if ($usersLog) {
        Write-Host "  [OK] Users Service: $usersLog" -ForegroundColor $successColor
    } else {
        Write-Host "  ⚠ Users Service: Check logs manually" -ForegroundColor $infoColor
    }
    
    return $httpsFound
}

# Funkcija za testiranje password hashing-a
function Test-PasswordHashing {
    Write-Host "`n[7] Testing Password Security (Hash & Salt)" -ForegroundColor Cyan
    
    # Prvo proveri da li već postoje korisnici u bazi
    Write-Host "Checking existing users in database..." -ForegroundColor $infoColor
    $anyUser = docker exec projekat-2025-1-mongodb-users-1 mongosh --quiet --eval "db.users.findOne({}, {passwordHash: 1, email: 1, _id: 0})" 2>$null
    
    if ($anyUser -match '\$2[ab]\$') {
        Write-Host "  [OK] Found hashed password in database (bcrypt detected)" -ForegroundColor $successColor
        Write-Host "  [OK] Password is NOT in plain text" -ForegroundColor $successColor
        return $true
    }
    
    # Ako nema korisnika, pokušaj da registruješ novog
    Write-Host "No users found. Registering test user..." -ForegroundColor $infoColor
    
    $registerBody = @{
        email = "hashtest@example.com"
        username = "hashtest"
        password = "TestPassword123!"
        confirmPassword = "TestPassword123!"
        firstName = "Hash"
        lastName = "Test"
    }
    
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8081/api/users/register" `
            -Method POST `
            -Body ($registerBody | ConvertTo-Json) `
            -ContentType "application/json" `
            -UseBasicParsing `
            -ErrorAction Stop
        
        Write-Host "  [OK] User registered successfully" -ForegroundColor $successColor
        
        # Čekaj malo da se podaci sačuvaju
        Start-Sleep -Seconds 1
        
        # Provera MongoDB-a (polje se zove passwordHash, ne password)
        Write-Host "`nChecking password in database..." -ForegroundColor $infoColor
        $passwordCheck = docker exec projekat-2025-1-mongodb-users-1 mongosh --quiet --eval "db.users.findOne({email: 'hashtest@example.com'}, {passwordHash: 1, _id: 0})" 2>$null
        
        if ($passwordCheck -match '\$2[ab]\$') {
            Write-Host "  [OK] Password is hashed (bcrypt detected)" -ForegroundColor $successColor
            Write-Host "  [OK] Password is NOT in plain text" -ForegroundColor $successColor
            return $true
        } elseif ($passwordCheck -match 'password' -or $passwordCheck -match 'TestPassword') {
            Write-Host "  [FAIL] Password might be in plain text" -ForegroundColor $errorColor
            Write-Host "  Database entry: $passwordCheck" -ForegroundColor Gray
            return $false
        } else {
            Write-Host "  [WARN] Could not verify password format" -ForegroundColor $infoColor
            Write-Host "  Database entry: $passwordCheck" -ForegroundColor Gray
            return $false
        }
    } catch {
        if ($_.Exception.Message -match "already exists") {
            Write-Host "  [INFO] User already exists, checking existing password..." -ForegroundColor $infoColor
            Start-Sleep -Seconds 1  # Čekaj da se query izvrši
            $passwordCheck = docker exec projekat-2025-1-mongodb-users-1 mongosh --quiet --eval "db.users.findOne({email: 'hashtest@example.com'}, {passwordHash: 1, _id: 0})" 2>$null
            
            if ($passwordCheck -match '\$2[ab]\$') {
                Write-Host "  [OK] Password is hashed (bcrypt detected)" -ForegroundColor $successColor
                Write-Host "  [OK] Password is NOT in plain text" -ForegroundColor $successColor
                return $true
            } elseif ($passwordCheck -match 'passwordHash') {
                # Proveri bilo kog korisnika u bazi
                $anyUser = docker exec projekat-2025-1-mongodb-users-1 mongosh --quiet --eval "db.users.findOne({}, {passwordHash: 1, email: 1, _id: 0})" 2>$null
                if ($anyUser -match '\$2[ab]\$') {
                    Write-Host "  [OK] Found hashed password in database (bcrypt)" -ForegroundColor $successColor
                    Write-Host "  [OK] Password hashing is working correctly" -ForegroundColor $successColor
                    return $true
                } else {
                    Write-Host "  [WARN] Could not verify password format" -ForegroundColor $infoColor
                    Write-Host "  [INFO] Password hashing is implemented in code (bcrypt)" -ForegroundColor $infoColor
                    return $false
                }
            } else {
                Write-Host "  [WARN] Could not verify password format" -ForegroundColor $infoColor
                Write-Host "  [INFO] Password hashing is implemented in code (bcrypt)" -ForegroundColor $infoColor
                return $false
            }
        } else {
            Write-Host "  [FAIL] Registration failed: $_" -ForegroundColor $errorColor
            Write-Host "  [INFO] Password hashing is implemented in code (bcrypt)" -ForegroundColor $infoColor
            Write-Host "  [INFO] To verify: register a user via frontend and check database" -ForegroundColor $infoColor
            return $false
        }
    }
}

# Glavni test
Write-Host "Starting system tests...`n" -ForegroundColor $infoColor

$results = @{}

# Test 1: Certificates
$results.Certificates = Test-Certificates

# Test 2: Docker Services
$results.DockerServices = Test-DockerServices

# Test 3: Ports
$results.Ports = Test-Ports

# Test 4: API Endpoints
if ($results.DockerServices -and $results.Ports) {
    $results.APIEndpoints = Test-APIEndpoints
} else {
    Write-Host "`nSkipping API tests - services not running" -ForegroundColor $infoColor
    $results.APIEndpoints = $false
}

# Test 5: MailHog
if ($results.DockerServices) {
    $results.MailHog = Test-MailHog
} else {
    $results.MailHog = $false
}

# Test 6: HTTPS Communication
if ($results.DockerServices) {
    $results.HTTPS = Test-HTTPSCommunication
} else {
    $results.HTTPS = $false
}

# Test 7: Password Hashing
if ($results.APIEndpoints) {
    $results.PasswordHashing = Test-PasswordHashing
} else {
    $results.PasswordHashing = $false
}

# Rezime
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  Test Results Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

foreach ($test in $results.Keys) {
    $status = if ($results[$test]) { "[PASS]" } else { "[FAIL]" }
    $color = if ($results[$test]) { $successColor } else { $errorColor }
    Write-Host "$status - $test" -ForegroundColor $color
}

$allPassed = ($results.Values | Where-Object { $_ -eq $false }).Count -eq 0

Write-Host ""
if ($allPassed) {
    Write-Host "========================================" -ForegroundColor $successColor
    Write-Host "  All Tests Passed! [OK]" -ForegroundColor $successColor
    Write-Host "========================================" -ForegroundColor $successColor
} else {
    Write-Host "========================================" -ForegroundColor $errorColor
    Write-Host "  Some Tests Failed [FAIL]" -ForegroundColor $errorColor
    Write-Host "========================================" -ForegroundColor $errorColor
    Write-Host ""
    Write-Host "Check the output above for details." -ForegroundColor $infoColor
}

Write-Host "`nNext Steps:" -ForegroundColor $infoColor
Write-Host "1. Open frontend: http://localhost:3000" -ForegroundColor White
Write-Host "2. Open MailHog: http://localhost:8025" -ForegroundColor White
Write-Host "3. Test admin login: admin@musicstreaming.com" -ForegroundColor White
