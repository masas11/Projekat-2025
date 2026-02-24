# Test skripta za 2.7.5, 2.7.6, 2.7.7 - Resilience mehanizmi
# PokreÄ‡e sve testove odjednom

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Test Resilience (2.7.5, 2.7.6, 2.7.7)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Boje za output
$successColor = "Green"
$errorColor = "Red"
$infoColor = "Yellow"
$testColor = "Cyan"

# Funkcija za proveru statusa servisa
function Test-ServiceStatus {
    param($serviceName)
    $status = docker-compose ps $serviceName --format "{{.Status}}"
    if ($status -match "Up") {
        Write-Host "[OK] $serviceName je pokrenut" -ForegroundColor $successColor
        return $true
    } else {
        Write-Host "[FAIL] $serviceName NIJE pokrenut: $status" -ForegroundColor $errorColor
        return $false
    }
}

# Funkcija za proveru endpoint-a
function Test-Endpoint {
    param($url, $expectedStatus = 200, $description = "")
    try {
        $response = Invoke-WebRequest -Uri $url -UseBasicParsing -TimeoutSec 10 -ErrorAction Stop
        if ($response.StatusCode -eq $expectedStatus) {
            Write-Host "  [OK] $description - Status: $($response.StatusCode)" -ForegroundColor $successColor
            return $true
        } else {
            Write-Host "  [FAIL] $description - OÄekivano: $expectedStatus, Dobijeno: $($response.StatusCode)" -ForegroundColor $errorColor
            return $false
        }
    } catch {
        if ($_.Exception.Response) {
            $statusCode = $_.Exception.Response.StatusCode.value__
            if ($statusCode -eq $expectedStatus) {
                Write-Host "  [OK] $description - Status: $statusCode (oÄekivano)" -ForegroundColor $successColor
                return $true
            }
        }
        Write-Host "  [FAIL] $description - Greska: $($_.Exception.Message)" -ForegroundColor $errorColor
        return $false
    }
}

# Funkcija za proveru logova
function Test-Logs {
    param($serviceName, $pattern, $description)
    $logs = docker-compose logs $serviceName --tail 50 2>&1 | Select-String -Pattern $pattern -CaseSensitive:$false
    if ($logs) {
        Write-Host "  [OK] $description - PronaÄ‘eno u logovima" -ForegroundColor $successColor
        return $true
    } else {
        Write-Host "  [FAIL] $description - Nije pronaÄ‘eno u logovima" -ForegroundColor $errorColor
        return $false
    }
}

# PoÄetak testova
Write-Host "1. PROVERA STATUSA SERVISA" -ForegroundColor $testColor
Write-Host "---------------------------" -ForegroundColor $testColor

$services = @("api-gateway", "content-service", "ratings-service", "subscriptions-service")
$allServicesUp = $true
foreach ($service in $services) {
    if (-not (Test-ServiceStatus $service)) {
        $allServicesUp = $false
    }
}

if (-not $allServicesUp) {
    Write-Host ""
    Write-Host "[WARN] Neki servisi nisu pokrenuti. PokuÅ¡avam da ih pokrenem..." -ForegroundColor $infoColor
    docker-compose up -d api-gateway content-service ratings-service subscriptions-service
    Start-Sleep -Seconds 5
}

Write-Host ""
Write-Host "2. TEST 2.7.6 - API GATEWAY TIMEOUT" -ForegroundColor $testColor
Write-Host "-----------------------------------" -ForegroundColor $testColor

# Test health endpoint
Write-Host ""
Write-Host "2.1. Testiranje health endpoint-a..." -ForegroundColor $infoColor
Test-Endpoint "http://localhost:8081/health" 200 "API Gateway Health"

# Test timeout - zaustavimo ratings-service i pokuÅ¡ajmo zahtev
Write-Host ""
Write-Host "2.2. Testiranje timeout-a (zaustavljamo ratings-service)..." -ForegroundColor $infoColor
docker-compose stop ratings-service | Out-Null
Start-Sleep -Seconds 2

Write-Host "  PokuÅ¡avam zahtev ka ratings-service (trebalo bi timeout nakon ~5s)..." -ForegroundColor $infoColor
$startTime = Get-Date
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8081/api/ratings/health" -UseBasicParsing -TimeoutSec 10 -ErrorAction Stop
    Write-Host "  [FAIL] Timeout nije aktiviran - Status: $($response.StatusCode)" -ForegroundColor $errorColor
} catch {
    $elapsed = ((Get-Date) - $startTime).TotalSeconds
    if ($_.Exception.Message -match "timeout" -or $elapsed -ge 5) {
        Write-Host "  [OK] Timeout aktiviran nakon ~$([math]::Round($elapsed, 1))s" -ForegroundColor $successColor
    } else {
        Write-Host "  [WARN] Greska: $($_.Exception.Message)" -ForegroundColor $infoColor
    }
}

# Restartuj ratings-service
docker-compose start ratings-service | Out-Null
Start-Sleep -Seconds 3

Write-Host ""
Write-Host "3. TEST 2.7.5 - RETRY MEHANIZAM" -ForegroundColor $testColor
Write-Host "--------------------------------" -ForegroundColor $testColor

# Zaustavimo content-service da testiramo retry
Write-Host ""
Write-Host "3.1. Testiranje retry mehanizma (zaustavljamo content-service)..." -ForegroundColor $infoColor
docker-compose stop content-service | Out-Null
Start-Sleep -Seconds 2

Write-Host "  PokuÅ¡avam da ocenim pesmu (trebalo bi da vidiÅ¡ retry pokuÅ¡aje u logovima)..." -ForegroundColor $infoColor

# Napravi stvarni zahtev koji će aktivirati retry mehanizam
# Prvo pročitaj logove pre zahteva
$logsBefore = docker-compose logs ratings-service --tail 10 2>&1

# Pokušaj da oceniš pesmu (ovo će aktivirati retry jer content-service nije dostupan)
try {
    # Koristimo bilo koji songId - retry će se aktivirati kada ratings-service pokuša da proveri da li pesma postoji
    $testUrl = "http://localhost:8081/api/ratings/rate-song?songId=test123&userId=testuser&rating=5"
    # Ne koristimo -ErrorAction Stop jer želimo da vidimo šta se dešava
    $response = Invoke-WebRequest -Uri $testUrl -Method POST -UseBasicParsing -TimeoutSec 10 -ErrorAction SilentlyContinue
} catch {
    # Očekivano - zahtev će verovatno da ne uspe jer content-service nije dostupan
    Write-Host "  Zahtev završen (očekivano da ne uspe)" -ForegroundColor $infoColor
}

# SaÄekaj malo da se retry pokuÅ¡aji izvrÅ¡e
Start-Sleep -Seconds 4

# Proveri logove za retry
Write-Host "  Proveravam logove za retry pokuÅ¡aje..." -ForegroundColor $infoColor
$retryFound = Test-Logs "ratings-service" "retry|Retry|attempt|Retrying" "Retry mehanizam"

# Restartuj content-service
docker-compose start content-service | Out-Null
Start-Sleep -Seconds 3

Write-Host ""
Write-Host "4. TEST 2.7.7 - UPSTREAM SERVIS ODUSTAJE OD OBRADE" -ForegroundColor $testColor
Write-Host "---------------------------------------------------" -ForegroundColor $testColor

Write-Host ""
Write-Host "4.1. Testiranje context cancellation..." -ForegroundColor $infoColor
Write-Host "  Proveravam da li servisi koriste request context..." -ForegroundColor $infoColor

# Proveri kod da li koristi r.Context()
$ratingsCode = Get-Content "services/ratings-service/cmd/main.go" -Raw
if ($ratingsCode -match "r\.Context\(\)") {
    Write-Host "  [OK] Ratings-service koristi request context" -ForegroundColor $successColor
} else {
    Write-Host "  [FAIL] Ratings-service NE koristi request context" -ForegroundColor $errorColor
}

$subscriptionsCode = Get-Content "services/subscriptions-service/cmd/main.go" -Raw
if ($subscriptionsCode -match "r\.Context\(\)") {
    Write-Host "  [OK] Subscriptions-service koristi request context" -ForegroundColor $successColor
} else {
    Write-Host "  [FAIL] Subscriptions-service NE koristi request context" -ForegroundColor $errorColor
}

# Proveri da li ima context cancellation provere
if ($ratingsCode -match "context cancelled|Request context cancelled") {
    Write-Host "  [OK] Ratings-service ima context cancellation provere" -ForegroundColor $successColor
} else {
    Write-Host "  [FAIL] Ratings-service NEMA context cancellation provere" -ForegroundColor $errorColor
}

Write-Host ""
Write-Host "5. PROVERA API GATEWAY TIMEOUT IMPLEMENTACIJE" -ForegroundColor $testColor
Write-Host "----------------------------------------------" -ForegroundColor $testColor

$apiGatewayCode = Get-Content "services/api-gateway/cmd/main.go" -Raw
if ($apiGatewayCode -match "context\.WithTimeout.*r\.Context") {
    Write-Host "  [OK] API Gateway koristi context.WithTimeout sa request context" -ForegroundColor $successColor
} else {
    Write-Host "  [FAIL] API Gateway NE koristi context.WithTimeout sa request context" -ForegroundColor $errorColor
}

if ($apiGatewayCode -match "StatusRequestTimeout|408") {
    Write-Host "  [OK] API Gateway vraÄ‡a 408 Request Timeout" -ForegroundColor $successColor
} else {
    Write-Host "  [FAIL] API Gateway NE vraÄ‡a 408 Request Timeout" -ForegroundColor $errorColor
}

Write-Host ""
Write-Host "6. PROVERA RETRY IMPLEMENTACIJE" -ForegroundColor $testColor
Write-Host "--------------------------------" -ForegroundColor $testColor

if ($ratingsCode -match "RetryWithExponentialBackoff|ExponentialBackoff") {
    Write-Host "  [OK] Ratings-service ima RetryWithExponentialBackoff" -ForegroundColor $successColor
} else {
    Write-Host "  [FAIL] Ratings-service NEMA RetryWithExponentialBackoff" -ForegroundColor $errorColor
}

if ($subscriptionsCode -match "RetryWithExponentialBackoff|ExponentialBackoff") {
    Write-Host "  [OK] Subscriptions-service ima RetryWithExponentialBackoff" -ForegroundColor $successColor
} else {
    Write-Host "  [FAIL] Subscriptions-service NEMA RetryWithExponentialBackoff" -ForegroundColor $errorColor
}

Write-Host ""
Write-Host "7. BRZA PROVERA ENDPOINT-A" -ForegroundColor $testColor
Write-Host "---------------------------" -ForegroundColor $testColor

Write-Host ""
Write-Host "7.1. Testiranje osnovnih endpoint-a..." -ForegroundColor $infoColor
Test-Endpoint "http://localhost:8081/api/content/health" 200 "Content Service Health"
Test-Endpoint "http://localhost:8081/api/ratings/health" 200 "Ratings Service Health"
Test-Endpoint "http://localhost:8081/api/subscriptions/health" 200 "Subscriptions Service Health"

Write-Host ""
Write-Host "7.2. Testiranje songs endpoint-a..." -ForegroundColor $infoColor
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8081/api/content/songs" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "  OK Songs endpoint radi - Status: $($response.StatusCode)" -ForegroundColor $successColor
} catch {
    Write-Host "  ERROR Songs endpoint greska: $($_.Exception.Message)" -ForegroundColor $errorColor
}

Write-Host ""
Write-Host '========================================' -ForegroundColor Cyan
Write-Host '  TESTIRANJE ZAVRÅ ENO' -ForegroundColor Cyan
Write-Host '========================================' -ForegroundColor Cyan
Write-Host ""
Write-Host 'Za detaljne logove, pokrenite:' -ForegroundColor $infoColor
Write-Host '  docker-compose logs ratings-service --tail 50' -ForegroundColor White
Write-Host '  docker-compose logs api-gateway --tail 50' -ForegroundColor White
Write-Host '  docker-compose logs subscriptions-service --tail 50' -ForegroundColor White
Write-Host '  docker-compose logs content-service --tail 50' -ForegroundColor White
Write-Host ""
Write-Host 'Za testiranje u browseru:' -ForegroundColor $infoColor
Write-Host '  http://localhost:3000 (frontend)' -ForegroundColor White
Write-Host '  http://localhost:8081/health (API Gateway health)' -ForegroundColor White
Write-Host ""

