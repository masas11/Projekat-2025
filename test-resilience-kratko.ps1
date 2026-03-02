# KRATKA TEST SKRIPTA - 2.7 Otpornost na parcijalne otkaze sistema
# Pokriva svih 7 mehanizama: 2.7.1 - 2.7.7

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  TEST 2.7 - OTPORNOST NA OTKAZE" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$successColor = "Green"
$errorColor = "Red"
$infoColor = "Yellow"
$testColor = "Cyan"

# Funkcija za proveru logova
function Test-LogPattern {
    param($serviceName, $pattern, $description)
    $logs = docker-compose logs $serviceName --tail 100 2>&1 | Select-String -Pattern $pattern -CaseSensitive:$false
    if ($logs) {
        Write-Host "  [OK] $description" -ForegroundColor $successColor
        return $true
    } else {
        Write-Host "  [FAIL] $description" -ForegroundColor $errorColor
        return $false
    }
}

# Funkcija za proveru koda
function Test-CodePattern {
    param($filePath, $pattern, $description)
    if (Test-Path $filePath) {
        $content = Get-Content $filePath -Raw
        if ($content -match $pattern) {
            Write-Host "  [OK] $description" -ForegroundColor $successColor
            return $true
        } else {
            Write-Host "  [FAIL] $description" -ForegroundColor $errorColor
            return $false
        }
    } else {
        Write-Host "  [FAIL] Fajl ne postoji: $filePath" -ForegroundColor $errorColor
        return $false
    }
}

# Provera statusa servisa
Write-Host "1. PROVERA STATUSA SERVISA" -ForegroundColor $testColor
Write-Host "---------------------------" -ForegroundColor $testColor
$services = @("api-gateway", "content-service", "ratings-service", "subscriptions-service")
foreach ($service in $services) {
    $status = docker-compose ps $service --format "{{.Status}}" 2>&1
    if ($status -match "Up") {
        Write-Host "  [OK] $service je pokrenut" -ForegroundColor $successColor
    } else {
        Write-Host "  [FAIL] $service NIJE pokrenut" -ForegroundColor $errorColor
        Write-Host "      Pokrecem servise..." -ForegroundColor $infoColor
        docker-compose up -d $service | Out-Null
        Start-Sleep -Seconds 3
    }
}
Write-Host ""

# TEST 2.7.1 - Konfiguracija HTTP klijenta
Write-Host "2. TEST 2.7.1 - KONFIGURACIJA HTTP KLIJENTA" -ForegroundColor $testColor
Write-Host "--------------------------------------------" -ForegroundColor $testColor
$null = Test-CodePattern "services/ratings-service/cmd/main.go" "http\.Transport|TLSClientConfig|MaxIdleConns|IdleConnTimeout" "Ratings-service: HTTP Transport konfiguracija"
$null = Test-CodePattern "services/subscriptions-service/cmd/main.go" "http\.Transport|TLSClientConfig|MaxIdleConns|IdleConnTimeout" "Subscriptions-service: HTTP Transport konfiguracija"
Write-Host ""

# TEST 2.7.2 - Timeout na nivou zahteva
Write-Host "3. TEST 2.7.2 - TIMEOUT NA NIVOU ZAHTEVA" -ForegroundColor $testColor
Write-Host "-----------------------------------------" -ForegroundColor $testColor
$null = Test-CodePattern "services/ratings-service/cmd/main.go" "clientHTTP.*Timeout.*2.*time\.Second|Timeout.*2.*time\.Second" "Ratings-service: Timeout na HTTP klijentu (2s)"
$null = Test-CodePattern "services/subscriptions-service/cmd/main.go" "client.*Timeout.*5.*time\.Second|Timeout.*5.*time\.Second" "Subscriptions-service: Timeout na HTTP klijentu"
Write-Host ""

# TEST 2.7.3 - Fallback logika
Write-Host "4. TEST 2.7.3 - FALLBACK LOGIKA" -ForegroundColor $testColor
Write-Host "-------------------------------" -ForegroundColor $testColor
$null = Test-CodePattern "services/ratings-service/cmd/main.go" "fallback activated|using fallback|Fallback" "Ratings-service: Fallback logika"
$null = Test-CodePattern "services/subscriptions-service/cmd/main.go" "fallback activated|using fallback|Fallback" "Subscriptions-service: Fallback logika"
Write-Host ""

# TEST 2.7.4 - Circuit Breaker
Write-Host "5. TEST 2.7.4 - CIRCUIT BREAKER" -ForegroundColor $testColor
Write-Host "------------------------------" -ForegroundColor $testColor
$null = Test-CodePattern "services/ratings-service/cmd/main.go" "CircuitBreaker|circuit breaker|Circuit breaker" "Ratings-service: Circuit Breaker implementacija"
$null = Test-CodePattern "services/subscriptions-service/cmd/main.go" "CircuitBreaker|circuit breaker|Circuit breaker" "Subscriptions-service: Circuit Breaker implementacija"

# Test circuit breaker u akciji - zaustavimo content-service
Write-Host ""
Write-Host "  Testiranje Circuit Breaker u akciji..." -ForegroundColor $infoColor
docker-compose stop content-service 2>&1 | Out-Null
Start-Sleep -Seconds 2

# Pokusaj zahteva koji ce aktivirati circuit breaker
$testUrl = "http://localhost:8081/api/ratings/rate-song?songId=test123&userId=testuser&rating=5"
try {
    $null = Invoke-WebRequest -Uri $testUrl -Method POST -UseBasicParsing -TimeoutSec 10 -ErrorAction SilentlyContinue
} catch {
    # Ocekivano
}

Start-Sleep -Seconds 3
$null = Test-LogPattern "ratings-service" "Circuit breaker|circuit breaker opened" "Circuit Breaker se aktivirao"

docker-compose start content-service 2>&1 | Out-Null
Start-Sleep -Seconds 3
Write-Host ""

# TEST 2.7.5 - Retry mehanizam
Write-Host "6. TEST 2.7.5 - RETRY MEHANIZAM" -ForegroundColor $testColor
Write-Host "------------------------------" -ForegroundColor $testColor
$null = Test-CodePattern "services/ratings-service/cmd/main.go" "RetryWithExponentialBackoff|ExponentialBackoff|retry attempt" "Ratings-service: Retry sa exponential backoff"
$null = Test-CodePattern "services/subscriptions-service/cmd/main.go" "RetryWithExponentialBackoff|ExponentialBackoff|retry attempt" "Subscriptions-service: Retry sa exponential backoff"

# Test retry u akciji
Write-Host ""
Write-Host "  Testiranje Retry mehanizma..." -ForegroundColor $infoColor
docker-compose stop content-service 2>&1 | Out-Null
Start-Sleep -Seconds 2

$testUrl2 = "http://localhost:8081/api/ratings/rate-song?songId=test123&userId=testuser&rating=5"
try {
    $null = Invoke-WebRequest -Uri $testUrl2 -Method POST -UseBasicParsing -TimeoutSec 10 -ErrorAction SilentlyContinue
} catch {
    # Ocekivano
}

Start-Sleep -Seconds 4
$null = Test-LogPattern "ratings-service" "Retry attempt|retrying in|Retry attempt|retry attempt" "Retry pokusaji sa exponential backoff"

docker-compose start content-service 2>&1 | Out-Null
Start-Sleep -Seconds 3
Write-Host ""

# TEST 2.7.6 - Eksplicitno postavljen timeout za korisnika
Write-Host "7. TEST 2.7.6 - TIMEOUT ZA KORISNIKA (API GATEWAY)" -ForegroundColor $testColor
Write-Host "--------------------------------------------------" -ForegroundColor $testColor
$null = Test-CodePattern "services/api-gateway/cmd/main.go" "context\.WithTimeout.*r\.Context|StatusRequestTimeout|408" "API Gateway: Timeout sa request context i 408 status"

# Test timeout u akciji
Write-Host ""
Write-Host "  Testiranje API Gateway timeout-a..." -ForegroundColor $infoColor
docker-compose stop ratings-service 2>&1 | Out-Null
Start-Sleep -Seconds 2

$startTime = Get-Date
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8081/api/ratings/health" -UseBasicParsing -TimeoutSec 10 -ErrorAction Stop
    Write-Host "  [FAIL] Timeout nije aktiviran" -ForegroundColor $errorColor
} catch {
    $elapsed = ((Get-Date) - $startTime).TotalSeconds
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 408) {
            Write-Host "  [OK] API Gateway vraca 408 Request Timeout" -ForegroundColor $successColor
        } else {
            Write-Host "  [FAIL] Status kod: $statusCode (ocekivano 408)" -ForegroundColor $errorColor
        }
    } elseif (($elapsed -ge 4) -and ($elapsed -le 6)) {
        $rounded = [math]::Round($elapsed, 1)
        Write-Host "  [OK] Timeout aktiviran nakon ~${rounded}s" -ForegroundColor $successColor
    } else {
        $rounded = [math]::Round($elapsed, 1)
        Write-Host "  [~] Timeout aktiviran nakon ${rounded}s (ocekivano 4-6s)" -ForegroundColor $infoColor
    }
}

$null = Test-LogPattern "api-gateway" "Request timeout|timeout for" "API Gateway loguje timeout"

docker-compose start ratings-service 2>&1 | Out-Null
Start-Sleep -Seconds 3
Write-Host ""

# TEST 2.7.7 - Upstream servis odustaje od obrade
Write-Host "8. TEST 2.7.7 - UPSTREAM SERVIS ODUSTAJE OD OBRADE" -ForegroundColor $testColor
Write-Host "--------------------------------------------------" -ForegroundColor $testColor
$null = Test-CodePattern "services/ratings-service/cmd/main.go" "r\.Context\(\)|context cancelled|Context cancelled" "Ratings-service: Koristi request context i prekida obradu"
$null = Test-CodePattern "services/subscriptions-service/cmd/main.go" "r\.Context\(\)|context cancelled|Context cancelled" "Subscriptions-service: Koristi request context i prekida obradu"
$null = Test-CodePattern "services/ratings-service/cmd/main.go" "ctx\.Done\(\)|select.*ctx\.Done" "Ratings-service: Proverava context cancellation u retry"
$null = Test-CodePattern "services/subscriptions-service/cmd/main.go" "ctx\.Done\(\)|select.*ctx\.Done" "Subscriptions-service: Proverava context cancellation u retry"
Write-Host ""

# FINALNA PROVERA
Write-Host "9. FINALNA PROVERA - SVI SERVISI RADI" -ForegroundColor $testColor
Write-Host "--------------------------------------" -ForegroundColor $testColor
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8081/health" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "  [OK] API Gateway radi - Status: $($response.StatusCode)" -ForegroundColor $successColor
} catch {
    Write-Host "  [FAIL] API Gateway ne radi" -ForegroundColor $errorColor
}

try {
    $response = Invoke-WebRequest -Uri "http://localhost:8081/api/content/health" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "  [OK] Content Service radi - Status: $($response.StatusCode)" -ForegroundColor $successColor
} catch {
    Write-Host "  [FAIL] Content Service ne radi" -ForegroundColor $errorColor
}

try {
    $response = Invoke-WebRequest -Uri "http://localhost:8081/api/ratings/health" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "  [OK] Ratings Service radi - Status: $($response.StatusCode)" -ForegroundColor $successColor
} catch {
    Write-Host "  [FAIL] Ratings Service ne radi" -ForegroundColor $errorColor
}

try {
    $response = Invoke-WebRequest -Uri "http://localhost:8081/api/subscriptions/health" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "  [OK] Subscriptions Service radi - Status: $($response.StatusCode)" -ForegroundColor $successColor
} catch {
    Write-Host "  [FAIL] Subscriptions Service ne radi" -ForegroundColor $errorColor
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  TESTIRANJE ZAVRSENO" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Za detaljne logove:" -ForegroundColor $infoColor
$logCmd1 = 'docker-compose logs ratings-service --tail 50 | Select-String -Pattern "retry|circuit|timeout|fallback"'
Write-Host "  $logCmd1" -ForegroundColor White
$logCmd2 = 'docker-compose logs api-gateway --tail 50 | Select-String -Pattern "timeout|408"'
Write-Host "  $logCmd2" -ForegroundColor White
$logCmd3 = 'docker-compose logs subscriptions-service --tail 50 | Select-String -Pattern "retry|circuit|timeout|fallback"'
Write-Host "  $logCmd3" -ForegroundColor White
Write-Host ""
