# Test skripta za Jaeger Tracing (2.10)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  TESTIRANJE JAEGER TRACING (2.10)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Proveri da li je Jaeger pokrenut
Write-Host "1. Proveravam da li je Jaeger pokrenut..." -ForegroundColor Yellow
$jaegerContainer = docker ps --filter "name=jaeger" --format "{{.Names}}" | Select-Object -First 1
if ([string]::IsNullOrEmpty($jaegerContainer)) {
    Write-Host "   [ERROR] Jaeger kontejner nije pokrenut!" -ForegroundColor Red
    Write-Host "   Pokrenite: docker-compose up -d jaeger" -ForegroundColor Yellow
    exit 1
}
Write-Host "   [OK] Jaeger je pokrenut: $jaegerContainer" -ForegroundColor Green

# Proveri da li su servisi pokrenuti
Write-Host ""
Write-Host "2. Proveravam da li su servisi pokrenuti..." -ForegroundColor Yellow
$services = @("api-gateway", "users-service", "content-service", "ratings-service", "subscriptions-service")
$allRunning = $true
foreach ($service in $services) {
    $container = docker ps --filter "name=$service" --format "{{.Names}}" | Select-Object -First 1
    if ([string]::IsNullOrEmpty($container)) {
        Write-Host "   [WARN] $service nije pokrenut" -ForegroundColor Yellow
        $allRunning = $false
    } else {
        Write-Host "   [OK] $service je pokrenut" -ForegroundColor Green
    }
}

if (-not $allRunning) {
    Write-Host ""
    Write-Host "   Pokrenite servise: docker-compose up -d" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "3. Testiram API pozive da generišem trace-ove..." -ForegroundColor Yellow
Write-Host ""

# Test 1: Health check
Write-Host "   Test 1: Health check..." -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8081/health" -UseBasicParsing -ErrorAction Stop
    Write-Host "      [OK] Health check: $($response.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "      [ERROR] Health check failed: $_" -ForegroundColor Red
}

Start-Sleep -Seconds 1

# Test 2: Get songs (synchronous operation)
Write-Host "   Test 2: Get songs (synchronous)..." -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8081/api/content/songs" -UseBasicParsing -ErrorAction Stop
    Write-Host "      [OK] Get songs: $($response.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "      [ERROR] Get songs failed: $_" -ForegroundColor Red
}

Start-Sleep -Seconds 1

# Test 3: Get subscriptions (requires auth - skip if fails)
Write-Host "   Test 3: Get subscriptions (synchronous)..." -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8081/api/subscriptions?userId=test" -UseBasicParsing -ErrorAction Stop
    Write-Host "      [OK] Get subscriptions: $($response.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "      [INFO] Get subscriptions requires auth (expected): $($_.Exception.Response.StatusCode)" -ForegroundColor Gray
}

Start-Sleep -Seconds 2

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  TESTIRANJE ZAVRSENO!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Sledeci koraci:" -ForegroundColor Yellow
Write-Host "  1. Otvori Jaeger UI: http://localhost:16686" -ForegroundColor White
Write-Host "  2. Izaberi servis iz dropdown-a (npr. 'api-gateway')" -ForegroundColor White
Write-Host "  3. Klikni 'Find Traces'" -ForegroundColor White
Write-Host "  4. Trebalo bi da vidis trace-ove za sve pozive" -ForegroundColor White
Write-Host ""
Write-Host "Proveri:" -ForegroundColor Yellow
Write-Host "  - Trace-ovi za sinhronne operacije (HTTP pozivi)" -ForegroundColor White
Write-Host "  - Trace-ovi za asinhrone operacije (event emisije)" -ForegroundColor White
Write-Host "  - Span hijerarhija između servisa" -ForegroundColor White
Write-Host "  - Trace context propagacija" -ForegroundColor White
Write-Host ""
