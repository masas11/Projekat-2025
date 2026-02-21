# Test skripta za zahteve 2.5 i 2.7
# Testira sinhronu komunikaciju i otpornost na parcijalne otkaze sistema

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "TESTIRANJE ZAHTEVA 2.5 I 2.7" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Provera statusa servisa
Write-Host "[1/6] Provera statusa servisa..." -ForegroundColor Yellow
docker-compose ps | Select-String "content-service|ratings-service|subscriptions-service"
Write-Host ""

# Test 1: Dobijanje liste umetnika i pesama
Write-Host "[2/6] Dobijanje liste umetnika i pesama..." -ForegroundColor Yellow
try {
    $artistsResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/content/artists" -Method GET -UseBasicParsing
    $artists = $artistsResponse.Content | ConvertFrom-Json
    $artistId = if ($artists.Count -gt 0) { $artists[0].id } else { $null }
    Write-Host "  Pronadeno umetnika: $($artists.Count)" -ForegroundColor Green
    if ($artistId) {
        Write-Host "  Koristicemo umetnika ID: $artistId" -ForegroundColor Green
    }
} catch {
    Write-Host "  Greska pri dobijanju umetnika: $_" -ForegroundColor Red
    $artistId = $null
}

try {
    $songsResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/content/songs" -Method GET -UseBasicParsing
    $songs = $songsResponse.Content | ConvertFrom-Json
    $songId = if ($songs.Count -gt 0) { $songs[0].id } else { $null }
    Write-Host "  Pronadeno pesama: $($songs.Count)" -ForegroundColor Green
    if ($songId) {
        Write-Host "  Koristicemo pesmu ID: $songId" -ForegroundColor Green
    }
} catch {
    Write-Host "  Greska pri dobijanju pesama: $_" -ForegroundColor Red
    $songId = $null
}
Write-Host ""

# Test 2: Zaustavljanje content-service
Write-Host "[3/6] Zaustavljanje content-service za test otpornosti..." -ForegroundColor Yellow
docker-compose stop content-service
Start-Sleep -Seconds 2
Write-Host "  content-service zaustavljen" -ForegroundColor Green
Write-Host ""

# Test 3: Test pretplate na umetnika (content-service DOWN)
Write-Host "[4/6] Test 2.7: Pretplata na umetnika (content-service DOWN)..." -ForegroundColor Yellow
Write-Host "  Ocekivano: Retry logika (2.7.2), Fallback (2.7.3), Circuit Breaker (2.7.4)" -ForegroundColor Cyan
Write-Host "  Napomena: Potrebno je biti ulogovan kao korisnik (RK role)" -ForegroundColor Yellow
Write-Host "  Koristite frontend ili curl za testiranje pretplate" -ForegroundColor Yellow
Write-Host "  Proverite logove: docker-compose logs subscriptions-service --tail=50" -ForegroundColor Cyan
Write-Host ""

# Test 4: Test ocenjivanja pesme (content-service DOWN)
Write-Host "[5/6] Test 2.7: Ocenjivanje pesme (content-service DOWN)..." -ForegroundColor Yellow
Write-Host "  Ocekivano: Retry logika (2.7.2), Fallback (2.7.3), Circuit Breaker (2.7.4)" -ForegroundColor Cyan
Write-Host "  Napomena: Potrebno je biti ulogovan kao korisnik (RK role)" -ForegroundColor Yellow
Write-Host "  Koristite frontend ili curl za testiranje ocenjivanja" -ForegroundColor Yellow
Write-Host "  Proverite logove: docker-compose logs ratings-service --tail=50" -ForegroundColor Cyan
Write-Host ""

# Test 5: Analiza logova
Write-Host "[6/6] Analiza logova..." -ForegroundColor Yellow
Write-Host ""
Write-Host "  LOGOVI ZA SUBSCRIPTIONS-SERVICE:" -ForegroundColor Cyan
Write-Host "  docker-compose logs subscriptions-service --tail=50" -ForegroundColor White
Write-Host ""
Write-Host "  LOGOVI ZA RATINGS-SERVICE:" -ForegroundColor Cyan
Write-Host "  docker-compose logs ratings-service --tail=50" -ForegroundColor White
Write-Host ""
Write-Host "  Trazite sledece u logovima:" -ForegroundColor Yellow
Write-Host "    - retry ili Retry (2.7.2 - Retry logika)" -ForegroundColor White
Write-Host "    - fallback ili Fallback (2.7.3 - Fallback logika)" -ForegroundColor White
Write-Host "    - circuit breaker ili Circuit breaker (2.7.4 - Circuit Breaker)" -ForegroundColor White
Write-Host "    - checkArtistExists ili checkSpecificSongExists (2.5 - Sinhrona komunikacija)" -ForegroundColor White
Write-Host "    - timeout ili Timeout (2.7.2 - Timeout)" -ForegroundColor White
Write-Host ""

# Ponovno pokretanje content-service
Write-Host "Ponovno pokretanje content-service..." -ForegroundColor Yellow
docker-compose start content-service
Start-Sleep -Seconds 2
Write-Host "  content-service ponovo pokrenut" -ForegroundColor Green
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "TESTIRANJE ZAVRSENO" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "REZIME:" -ForegroundColor Yellow
Write-Host "  Test 2.5: Sinhrona komunikacija - proverite logove za checkArtistExists i checkSpecificSongExists" -ForegroundColor White
Write-Host "  Test 2.7.1: HTTP Client konfiguracija - proverite kod (TLSClientConfig, MaxIdleConns, IdleConnTimeout)" -ForegroundColor White
Write-Host "  Test 2.7.2: Timeout - proverite logove za timeout ili kod (2 sekunde)" -ForegroundColor White
Write-Host "  Test 2.7.3: Fallback - proverite logove za fallback ili kod (vraca false kada servis nije dostupan)" -ForegroundColor White
Write-Host "  Test 2.7.4: Circuit Breaker - proverite logove za circuit breaker i stanja (closed, open, half-open)" -ForegroundColor White
Write-Host ""
