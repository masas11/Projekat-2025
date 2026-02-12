# Test skripta za Brute-force napad demonstraciju
. .\https-helper.ps1

Write-Host "=== BRUTE-FORCE NAPAD DEMONSTRACIJA ===" -ForegroundColor Cyan
Write-Host ""

$username = "testuser"
$passwords = @("password", "123456", "admin", "test", "qwerty", "password123", "admin123", "root")

Write-Host "Napad na korisnika: $username" -ForegroundColor Yellow
Write-Host "Pokušaj probijanja lozinke kroz $($passwords.Count) pokušaja..." -ForegroundColor Yellow
Write-Host ""

$successCount = 0
$failureCount = 0
$locked = $false

for ($i = 0; $i -lt $passwords.Count; $i++) {
    $password = $passwords[$i]
    $attempt = $i + 1
    
    Write-Host "Pokušaj $attempt/$($passwords.Count): '$password'" -ForegroundColor Gray -NoNewline
    
    $body = @{
        username = $username
        password = $password
    } | ConvertTo-Json
    
    $result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/login/request-otp" -Method "POST" -Body $body -ContentType "application/json"
    
    if ($result.StatusCode -eq 200) {
        $successCount++
        Write-Host " → ✓ USPEH! Lozinka probijena: $password" -ForegroundColor Green
        break
    } elseif ($result.StatusCode -eq 403) {
        $locked = $true
        Write-Host " → ✗ NALOG ZAKLJUČAN (Status: 403)" -ForegroundColor Red
        Write-Host "   Account je zaključan nakon 5 neuspešnih pokušaja!" -ForegroundColor Yellow
        break
    } else {
        $failureCount++
        Write-Host " → ✗ Neuspešno (Status: $($result.StatusCode))" -ForegroundColor Red
    }
    
    Start-Sleep -Seconds 1
}

Write-Host "`n=== REZULTATI ===" -ForegroundColor Cyan
Write-Host "Ukupno pokušaja: $($successCount + $failureCount)" -ForegroundColor White
Write-Host "Uspešno: $successCount" -ForegroundColor $(if ($successCount -gt 0) { "Red" } else { "Green" })
Write-Host "Neuspešno: $failureCount" -ForegroundColor White
Write-Host "Nalog zaključan: $locked" -ForegroundColor $(if ($locked) { "Green" } else { "Yellow" })

if ($locked) {
    Write-Host "`n✓ BRUTE-FORCE NAPAD JE BLOKIRAN!" -ForegroundColor Green
    Write-Host "  Account locking mehanizam je zaštitio nalog." -ForegroundColor White
} elseif ($successCount -eq 0) {
    Write-Host "`n✓ BRUTE-FORCE NAPAD JE NEUSPEŠAN!" -ForegroundColor Green
    Write-Host "  Niti jedna lozinka nije probijena." -ForegroundColor White
} else {
    Write-Host "`n✗ BRUTE-FORCE NAPAD JE USPEŠAN!" -ForegroundColor Red
    Write-Host "  Lozinka je probijena: $($passwords[$successCount-1])" -ForegroundColor Red
}

Write-Host "`nProverite logove: docker logs projekat-2025-2-users-service-1 | grep LOGIN_FAILURE" -ForegroundColor Gray
