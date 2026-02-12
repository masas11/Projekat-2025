# Test skripta za logovanje
# Učitaj helper funkciju
. .\https-helper.ps1

Write-Host "=== TESTIRANJE LOGOVANJA ===" -ForegroundColor Cyan
Write-Host ""

# TEST 1: Health Check
Write-Host "1. Health Check..." -ForegroundColor Yellow
$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/health"
Write-Host "   Status: $($result.StatusCode) - $(if ($result.StatusCode -eq 200) { 'OK' } else { 'FAIL' })" -ForegroundColor $(if ($result.StatusCode -eq 200) { "Green" } else { "Red" })
Start-Sleep -Seconds 1

# TEST 2: Neuspeh kontrole pristupa (bez tokena)
Write-Host "`n2. Neuspeh kontrole pristupa (bez tokena)..." -ForegroundColor Yellow
$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/logout" -Method "GET"
if ($result.StatusCode -eq 401) {
    Write-Host "   Status: 401 - OK (ocekivano)" -ForegroundColor Green
} else {
    Write-Host "   Status: $($result.StatusCode) - $(if ($result.Error) { $result.Error } else { 'Nesto nije u redu' })" -ForegroundColor Yellow
    if ($result.Content) {
        Write-Host "   Response: $($result.Content)" -ForegroundColor Gray
    }
}
Start-Sleep -Seconds 2

# TEST 3: Nevalidni token
Write-Host "`n3. Nevalidni token..." -ForegroundColor Yellow
$headers = @{ "Authorization" = "Bearer invalid_token_12345" }
$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/logout" -Method "GET" -Headers $headers
Write-Host "   Status: $($result.StatusCode) - $(if ($result.StatusCode -eq 401) { 'OK (ocekivano)' } else { 'Nesto nije u redu' })" -ForegroundColor $(if ($result.StatusCode -eq 401) { "Green" } else { "Yellow" })
Start-Sleep -Seconds 2

# TEST 4: Neuspeh kontrole pristupa - RequireRole
Write-Host "`n4. Neuspeh kontrole pristupa - RequireRole..." -ForegroundColor Yellow
$headers = @{ 
    "Authorization" = "Bearer invalid_token"
    "Content-Type" = "application/json"
}
$body = '{"name":"Test","biography":"Test","genres":["Rock"]}'
$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/content/artists" -Method "POST" -Headers $headers -Body $body
Write-Host "   Status: $($result.StatusCode) - $(if ($result.StatusCode -eq 401 -or $result.StatusCode -eq 403) { 'OK (ocekivano)' } else { 'Nesto nije u redu' })" -ForegroundColor $(if ($result.StatusCode -eq 401 -or $result.StatusCode -eq 403) { "Green" } else { "Yellow" })
Start-Sleep -Seconds 2

# Proveri logove
Write-Host "`n=== PROVERAVANJE LOGOVA ===" -ForegroundColor Cyan
Write-Host "`nAPI Gateway logovi (poslednjih 50):" -ForegroundColor Yellow
$logs = docker logs projekat-2025-2-api-gateway-1 --tail 50 2>&1 | Select-String -Pattern "ACCESS_CONTROL|INVALID_TOKEN|ERROR|WARN|AUDIT"
if ($logs) {
    $logs | Select-Object -Last 10 | ForEach-Object { Write-Host "   $_" -ForegroundColor Gray }
} else {
    Write-Host "   Nema logova ovog tipa u poslednjih 50 linija" -ForegroundColor Yellow
    Write-Host "   Proverite sve logove: docker logs projekat-2025-2-api-gateway-1 --tail 100" -ForegroundColor Gray
}

Write-Host "`n=== REZIME ===" -ForegroundColor Cyan
Write-Host "`n✅ Testovi zavrseni!" -ForegroundColor Green
Write-Host "`nZa detaljne uputstva, otvori: KORAK_PO_KORAK_TESTIRANJE.md" -ForegroundColor Yellow
