# Test skripta za DoS napad demonstraciju
. .\https-helper.ps1

Write-Host "=== DoS NAPAD DEMONSTRACIJA ===" -ForegroundColor Cyan
Write-Host ""

$totalRequests = 150  # Više od rate limit-a (100/min)
Write-Host "Slanje $totalRequests zahteva (limit: 100/min)..." -ForegroundColor Yellow
Write-Host "Očekivano: Prvih ~100 prolazi, preko 100 se blokira" -ForegroundColor Gray
Write-Host ""

$successCount = 0
$blockedCount = 0
$errorCount = 0
$startTime = Get-Date

for ($i = 1; $i -le $totalRequests; $i++) {
    $result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/health" -Method "GET"
    
    if ($result.StatusCode -eq 200) {
        $successCount++
    } elseif ($result.StatusCode -eq 429) {
        $blockedCount++
        if ($blockedCount -eq 1) {
            Write-Host "`nPrvi blokirani zahtev na poziciji: $i" -ForegroundColor Red
        }
    } else {
        $errorCount++
    }
    
    # Prikaži progres svakih 25 zahteva
    if ($i % 25 -eq 0) {
        $elapsed = (Get-Date) - $startTime
        $rate = [math]::Round($i / $elapsed.TotalSeconds, 2)
        Write-Host "Progres: $i/$totalRequests (Uspešno: $successCount, Blokirano: $blockedCount, Greške: $errorCount) - $rate req/s" -ForegroundColor Gray
    }
    
    Start-Sleep -Milliseconds 100
}

$endTime = Get-Date
$totalTime = ($endTime - $startTime).TotalSeconds

Write-Host "`n=== REZULTATI ===" -ForegroundColor Cyan
Write-Host "Ukupno zahteva: $totalRequests" -ForegroundColor White
Write-Host "Uspešno (200): $successCount" -ForegroundColor Green
Write-Host "Blokirano (429): $blockedCount" -ForegroundColor Red
Write-Host "Greške: $errorCount" -ForegroundColor Yellow
Write-Host "Vreme: $([math]::Round($totalTime, 2)) sekundi" -ForegroundColor White
Write-Host "Prosečna brzina: $([math]::Round($totalRequests / $totalTime, 2)) req/s" -ForegroundColor White

$blockedPercentage = [math]::Round(($blockedCount / $totalRequests) * 100, 2)
Write-Host "Procenat blokiranih: $blockedPercentage%" -ForegroundColor $(if ($blockedPercentage -gt 0) { "Green" } else { "Yellow" })

if ($blockedCount -gt 0) {
    Write-Host "`n✓ DoS NAPAD JE BLOKIRAN!" -ForegroundColor Green
    Write-Host "  Rate limiting je zaštitio server od preopterećenja." -ForegroundColor White
    Write-Host "  $blockedCount zahteva je blokirano (HTTP 429)." -ForegroundColor White
} else {
    Write-Host "`n⚠ DoS NAPAD NIJE BLOKIRAN!" -ForegroundColor Yellow
    Write-Host "  Rate limiting možda ne radi pravilno." -ForegroundColor White
}

Write-Host "`nProverite logove: docker logs projekat-2025-2-api-gateway-1 | grep 'too many requests'" -ForegroundColor Gray
