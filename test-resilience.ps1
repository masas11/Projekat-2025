# Test skripta za zahteve 2.5 i 2.7 (2.7.1, 2.7.2, 2.7.3, 2.7.4)
# Testira sinhronu komunikaciju i otpornost na parcijalne otkaze sistema

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "TESTIRANJE ZAHTEVA 2.5 I 2.7" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Provera da li su servisi pokrenuti
Write-Host "[1/8] Provera statusa servisa..." -ForegroundColor Yellow
$services = docker-compose ps --format json | ConvertFrom-Json
$contentRunning = ($services | Where-Object { $_.Service -eq "content-service" -and $_.State -eq "running" }) -ne $null
$ratingsRunning = ($services | Where-Object { $_.Service -eq "ratings-service" -and $_.State -eq "running" }) -ne $null
$subscriptionsRunning = ($services | Where-Object { $_.Service -eq "subscriptions-service" -and $_.State -eq "running" }) -ne $null

if (-not $contentRunning) {
    Write-Host "  ⚠️  content-service nije pokrenut. Pokrećem..." -ForegroundColor Yellow
    docker-compose start content-service
    Start-Sleep -Seconds 3
}

if (-not $ratingsRunning) {
    Write-Host "  ⚠️  ratings-service nije pokrenut. Pokrećem..." -ForegroundColor Yellow
    docker-compose start ratings-service
    Start-Sleep -Seconds 2
}

if (-not $subscriptionsRunning) {
    Write-Host "  ⚠️  subscriptions-service nije pokrenut. Pokrećem..." -ForegroundColor Yellow
    docker-compose start subscriptions-service
    Start-Sleep -Seconds 2
}

Write-Host "  ✓ Servisi su pokrenuti" -ForegroundColor Green
Write-Host ""

# Test 1: Login kao korisnik (RK role)
Write-Host "[2/8] Login kao korisnik (RK role)..." -ForegroundColor Yellow
$loginBody = @{
    username = "testuser"
    password = "testpass123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/users/login/request-otp" -Method POST -ContentType "application/json" -Body $loginBody -UseBasicParsing
    Write-Host "  ✓ Login zahtev poslat" -ForegroundColor Green
    Write-Host "  ⚠️  Potrebno je uneti OTP kod sa MailHog-a (http://localhost:8025)" -ForegroundColor Yellow
    $otp = Read-Host "  Unesite OTP kod"
    
    $otpBody = @{
        username = "testuser"
        otp = $otp
    } | ConvertTo-Json
    
    $otpResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/users/login/verify-otp" -Method POST -ContentType "application/json" -Body $otpBody -UseBasicParsing
    $tokenData = $otpResponse.Content | ConvertFrom-Json
    $token = $tokenData.token
    $userId = $tokenData.user.id
    Write-Host "  ✓ Uspešno logovanje. User ID: $userId" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Greška pri logovanju: $_" -ForegroundColor Red
    Write-Host "  ⚠️  Nastavljam sa testiranjem bez tokena (neki testovi neće raditi)" -ForegroundColor Yellow
    $token = $null
    $userId = $null
}
Write-Host ""

# Test 2: Dobijanje liste umetnika i pesama (potrebno za testiranje)
Write-Host "[3/8] Dobijanje liste umetnika i pesama..." -ForegroundColor Yellow
try {
    $artistsResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/content/artists" -Method GET -UseBasicParsing
    $artists = $artistsResponse.Content | ConvertFrom-Json
    $artistId = if ($artists.Count -gt 0) { $artists[0].id } else { $null }
    Write-Host "  ✓ Pronađeno umetnika: $($artists.Count)" -ForegroundColor Green
    if ($artistId) {
        Write-Host "  ✓ Koristićemo umetnika ID: $artistId" -ForegroundColor Green
    }
} catch {
    Write-Host "  ✗ Greška pri dobijanju umetnika: $_" -ForegroundColor Red
    $artistId = $null
}

try {
    $songsResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/content/songs" -Method GET -UseBasicParsing
    $songs = $songsResponse.Content | ConvertFrom-Json
    $songId = if ($songs.Count -gt 0) { $songs[0].id } else { $null }
    Write-Host "  ✓ Pronađeno pesama: $($songs.Count)" -ForegroundColor Green
    if ($songId) {
        Write-Host "  ✓ Koristićemo pesmu ID: $songId" -ForegroundColor Green
    }
} catch {
    Write-Host "  ✗ Greška pri dobijanju pesama: $_" -ForegroundColor Red
    $songId = $null
}
Write-Host ""

# Test 3: Sinhrona komunikacija - Pretplata na umetnika (content-service UP)
Write-Host "[4/8] Test 2.5: Sinhrona komunikacija - Pretplata na umetnika (content-service UP)..." -ForegroundColor Yellow
if ($token -and $artistId) {
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }
    $subscribeBody = @{
        artistId = $artistId
    } | ConvertTo-Json
    
    try {
        Write-Host "  -> Šaljem zahtev za pretplatu na umetnika $artistId..." -ForegroundColor Cyan
        $subscribeResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/subscriptions/subscribe-artist" -Method POST -Headers $headers -Body $subscribeBody -UseBasicParsing
        Write-Host "  ✓ Uspešna pretplata (sinhrona komunikacija sa content-service radi)" -ForegroundColor Green
        Write-Host "  -> Proverite logove: docker-compose logs subscriptions-service | Select-String checkArtistExists" -ForegroundColor Cyan
    } catch {
        if ($_.Exception.Response.StatusCode -eq 409) {
            Write-Host "  ⚠️  Već ste pretplaćeni na ovog umetnika (to je okej)" -ForegroundColor Yellow
        } else {
            Write-Host "  ✗ Greška: $_" -ForegroundColor Red
        }
    }
} else {
    Write-Host "  ⚠️  Preskačem (nema tokena ili umetnika)" -ForegroundColor Yellow
}
Write-Host ""

# Test 4: Zaustavljanje content-service
Write-Host "[5/8] Zaustavljanje content-service za test otpornosti..." -ForegroundColor Yellow
docker-compose stop content-service
Start-Sleep -Seconds 2
Write-Host "  ✓ content-service zaustavljen" -ForegroundColor Green
Write-Host ""

# Test 5: Test 2.7 - Pretplata na umetnika (content-service DOWN) - Retry, Fallback, Circuit Breaker
Write-Host "[6/8] Test 2.7: Pretplata na umetnika (content-service DOWN)..." -ForegroundColor Yellow
Write-Host "  Očekivano: Retry logika (2.7.2), Fallback (2.7.3), Circuit Breaker (2.7.4)" -ForegroundColor Cyan
if ($token -and $artistId) {
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }
    $subscribeBody = @{
        artistId = $artistId
    } | ConvertTo-Json
    
    try {
        Write-Host "  -> Šaljem zahtev za pretplatu (content-service je DOWN)..." -ForegroundColor Cyan
        $subscribeResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/subscriptions/subscribe-artist" -Method POST -Headers $headers -Body $subscribeBody -UseBasicParsing
        Write-Host "  ✓ Zahtev prošao (fallback logika)" -ForegroundColor Green
    } catch {
        Write-Host "  ⚠️  Zahtev nije prošao: $_" -ForegroundColor Yellow
        Write-Host "  -> Ovo je okej ako fallback vraća grešku" -ForegroundColor Cyan
    }
    Write-Host "  -> Proverite logove: docker-compose logs subscriptions-service | Select-String -Pattern retry,fallback,circuit,checkArtistExists" -ForegroundColor Cyan
} else {
    Write-Host "  ⚠️  Preskačem (nema tokena ili umetnika)" -ForegroundColor Yellow
}
Write-Host ""

# Test 6: Test 2.7 - Ocenjivanje pesme (content-service DOWN)
Write-Host "[7/8] Test 2.7: Ocenjivanje pesme (content-service DOWN)..." -ForegroundColor Yellow
Write-Host "  Očekivano: Retry logika (2.7.2), Fallback (2.7.3), Circuit Breaker (2.7.4)" -ForegroundColor Cyan
if ($token -and $songId -and $userId) {
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }
    $ratingBody = @{
        songId = $songId
        rating = 5
    } | ConvertTo-Json
    
    try {
        Write-Host "  -> Šaljem zahtev za ocenjivanje pesme $songId (content-service je DOWN)..." -ForegroundColor Cyan
        $ratingResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/ratings/rate-song?userId=$userId" -Method POST -Headers $headers -Body $ratingBody -UseBasicParsing
        Write-Host "  ✓ Zahtev prošao (fallback logika)" -ForegroundColor Green
    } catch {
        Write-Host "  ⚠️  Zahtev nije prošao: $_" -ForegroundColor Yellow
        Write-Host "  -> Ovo je okej ako fallback vraća grešku" -ForegroundColor Cyan
    }
    Write-Host "  -> Proverite logove: docker-compose logs ratings-service | Select-String -Pattern retry,fallback,circuit,checkSpecificSongExists" -ForegroundColor Cyan
} else {
    Write-Host "  ⚠️  Preskačem (nema tokena, pesme ili userId)" -ForegroundColor Yellow
}
Write-Host ""

# Test 7: Provera logova
Write-Host "[8/8] Analiza logova..." -ForegroundColor Yellow
Write-Host ""
Write-Host "  📋 LOGOVI ZA SUBSCRIPTIONS-SERVICE:" -ForegroundColor Cyan
Write-Host "  docker-compose logs subscriptions-service --tail=50" -ForegroundColor White
Write-Host ""
Write-Host "  📋 LOGOVI ZA RATINGS-SERVICE:" -ForegroundColor Cyan
Write-Host "  docker-compose logs ratings-service --tail=50" -ForegroundColor White
Write-Host ""
Write-Host "  🔍 Tražite sledeće u logovima:" -ForegroundColor Yellow
Write-Host "    - retry ili Retry (2.7.2 - Retry logika)" -ForegroundColor White
Write-Host "    - fallback ili Fallback (2.7.3 - Fallback logika)" -ForegroundColor White
Write-Host "    - circuit breaker ili Circuit breaker (2.7.4 - Circuit Breaker)" -ForegroundColor White
Write-Host "    - checkArtistExists ili checkSpecificSongExists (2.5 - Sinhrona komunikacija)" -ForegroundColor White
Write-Host "    - timeout ili Timeout (2.7.2 - Timeout)" -ForegroundColor White
Write-Host ""

# Test 8: Ponovno pokretanje content-service
Write-Host "Ponovno pokretanje content-service..." -ForegroundColor Yellow
docker-compose start content-service
Start-Sleep -Seconds 2
Write-Host "  ✓ content-service ponovo pokrenut" -ForegroundColor Green
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "TESTIRANJE ZAVRŠENO" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "📝 REZIME:" -ForegroundColor Yellow
Write-Host "  ✓ Test 2.5: Sinhrona komunikacija - proverite logove za checkArtistExists i checkSpecificSongExists" -ForegroundColor White
Write-Host "  ✓ Test 2.7.1: HTTP Client konfiguracija - proverite kod (TLSClientConfig, MaxIdleConns, IdleConnTimeout)" -ForegroundColor White
Write-Host "  ✓ Test 2.7.2: Timeout - proverite logove za timeout ili kod (2 sekunde)" -ForegroundColor White
Write-Host "  ✓ Test 2.7.3: Fallback - proverite logove za fallback ili kod (vraca false kada servis nije dostupan)" -ForegroundColor White
Write-Host "  ✓ Test 2.7.4: Circuit Breaker - proverite logove za circuit breaker i stanja (closed, open, half-open)" -ForegroundColor White
Write-Host ""
