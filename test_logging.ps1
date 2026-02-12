# PowerShell skripta za testiranje logovanja

Write-Host "=== TESTIRANJE LOGOVANJA ===" -ForegroundColor Cyan
Write-Host ""

# 1. Test bez tokena
Write-Host "1. Test bez tokena (ACCESS_CONTROL_FAILURE)..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "https://localhost:8081/api/users/logout" -Method GET -SkipCertificateCheck -ErrorAction SilentlyContinue
} catch {
    # Očekivana greška
}
Start-Sleep -Seconds 1
Write-Host "   Proverite: Get-Content services\api-gateway\logs\app.log | Select-String 'ACCESS_CONTROL_FAILURE'" -ForegroundColor Gray
Write-Host ""

# 2. Test sa nevalidnim tokenom
Write-Host "2. Test sa nevalidnim tokenom (INVALID_TOKEN)..." -ForegroundColor Yellow
$headers = @{
    "Authorization" = "Bearer invalid_token_12345"
}
try {
    $response = Invoke-WebRequest -Uri "https://localhost:8081/api/users/logout" -Method GET -Headers $headers -SkipCertificateCheck -ErrorAction SilentlyContinue
} catch {
    # Očekivana greška
}
Start-Sleep -Seconds 1
Write-Host "   Proverite: Get-Content services\api-gateway\logs\app.log | Select-String 'INVALID_TOKEN'" -ForegroundColor Gray
Write-Host ""

# 3. Test neuspeha kontrole pristupa (RequireRole)
Write-Host "3. Test neuspeha kontrole pristupa - RequireRole..." -ForegroundColor Yellow
Write-Host "   (Pokušaj pristupa admin endpoint-u sa non-admin tokenom)" -ForegroundColor Gray
Write-Host "   Proverite: Get-Content services\api-gateway\logs\app.log | Select-String 'insufficient permissions'" -ForegroundColor Gray
Write-Host ""

# 4. Test administratorskih aktivnosti
Write-Host "4. Test administratorskih aktivnosti..." -ForegroundColor Yellow
Write-Host "   (Kreirajte artist/album/song kao admin)" -ForegroundColor Gray
Write-Host "   Proverite: Get-Content services\content-service\logs\app.log | Select-String 'ADMIN_ACTIVITY'" -ForegroundColor Gray
Write-Host ""

# 5. Test promene state podataka
Write-Host "5. Test promene state podataka..." -ForegroundColor Yellow
Write-Host "   (Ažurirajte artist/album/song kao admin)" -ForegroundColor Gray
Write-Host "   Proverite: Get-Content services\content-service\logs\app.log | Select-String 'STATE_CHANGE'" -ForegroundColor Gray
Write-Host ""

# 6. Pregled logova
Write-Host "=== PREGLED LOGOVA ===" -ForegroundColor Cyan
Write-Host ""

# Proverite da li postoje log fajlovi
$logFiles = Get-ChildItem -Path "services" -Recurse -Filter "app.log" -ErrorAction SilentlyContinue
if ($logFiles) {
    Write-Host "Pronađeni log fajlovi:" -ForegroundColor Green
    foreach ($file in $logFiles) {
        $size = (Get-Item $file.FullName).Length / 1KB
        Write-Host "  $($file.FullName) - $([math]::Round($size, 2)) KB" -ForegroundColor White
    }
    Write-Host ""
    
    # Prikaži poslednje logove
    Write-Host "Poslednje logove iz API Gateway:" -ForegroundColor Yellow
    $apiGatewayLog = "services\api-gateway\logs\app.log"
    if (Test-Path $apiGatewayLog) {
        Get-Content $apiGatewayLog -Tail 10 | ForEach-Object {
            Write-Host "  $_" -ForegroundColor Gray
        }
    }
    Write-Host ""
    
    Write-Host "Poslednje logove iz Content Service:" -ForegroundColor Yellow
    $contentServiceLog = "services\content-service\logs\app.log"
    if (Test-Path $contentServiceLog) {
        Get-Content $contentServiceLog -Tail 10 | ForEach-Object {
            Write-Host "  $_" -ForegroundColor Gray
        }
    }
} else {
    Write-Host "Nisu pronađeni log fajlovi. Proverite da li su servisi pokrenuti." -ForegroundColor Red
}

Write-Host ""
Write-Host "=== KORISNE KOMANDE ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "# Pregled svih ACCESS_CONTROL_FAILURE logova:" -ForegroundColor White
Write-Host "Get-Content services\api-gateway\logs\app.log | Select-String 'ACCESS_CONTROL_FAILURE'" -ForegroundColor Gray
Write-Host ""
Write-Host "# Pregled svih INVALID_TOKEN logova:" -ForegroundColor White
Write-Host "Get-Content services\api-gateway\logs\app.log | Select-String 'INVALID_TOKEN'" -ForegroundColor Gray
Write-Host ""
Write-Host "# Pregled svih ADMIN_ACTIVITY logova:" -ForegroundColor White
Write-Host "Get-Content services\content-service\logs\app.log | Select-String 'ADMIN_ACTIVITY'" -ForegroundColor Gray
Write-Host ""
Write-Host "# Pregled svih STATE_CHANGE logova:" -ForegroundColor White
Write-Host "Get-Content services\content-service\logs\app.log | Select-String 'STATE_CHANGE'" -ForegroundColor Gray
Write-Host ""
Write-Host "# Pregled svih TLS_FAILURE logova:" -ForegroundColor White
Write-Host "Get-ChildItem services\*\logs\app.log | ForEach-Object { Get-Content `$_.FullName | Select-String 'TLS_FAILURE' }" -ForegroundColor Gray
Write-Host ""
