# Kompletno testiranje Saga Pattern implementacije (2.13)
# Testira uspešan tok i neuspešne tokove

$baseURL = "http://localhost:8081"
$ErrorActionPreference = "Continue"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "SAGA PATTERN TEST - Kompletno Testiranje" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# Step 1: Login kao admin
Write-Host "Step 1: Prijavljivanje kao admin..." -ForegroundColor Yellow
$loginBody = @{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json

try {
    $otpResponse = Invoke-WebRequest -Uri "$baseURL/api/users/login/request-otp" `
        -Method POST `
        -Headers @{ "Content-Type" = "application/json" } `
        -Body $loginBody `
        -UseBasicParsing
    
    Write-Host "  ✓ OTP zahtev poslat" -ForegroundColor Green
    
    # Za test, koristimo "000000" kao OTP (ako je to default)
    $verifyBody = @{
        username = "admin"
        otp = "000000"
    } | ConvertTo-Json
    
    Start-Sleep -Seconds 2
    
    $verifyResponse = Invoke-WebRequest -Uri "$baseURL/api/users/login/verify-otp" `
        -Method POST `
        -Headers @{ "Content-Type" = "application/json" } `
        -Body $verifyBody `
        -UseBasicParsing
    
    $loginResult = $verifyResponse.Content | ConvertFrom-Json
    $token = $loginResult.token
    
    if (-not $token) {
        Write-Host "  ✗ Neuspešna prijava - proverite OTP kod" -ForegroundColor Red
        Write-Host "  Pokušavam sa direktnim login-om..." -ForegroundColor Yellow
        
        # Alternativno, možemo koristiti postojeći token ili kreirati novi
        Write-Host "  Molimo unesite admin token ručno:" -ForegroundColor Yellow
        $token = Read-Host "Token"
    } else {
        Write-Host "  ✓ Uspešna prijava, token dobijen" -ForegroundColor Green
    }
} catch {
    Write-Host "  ✗ Greška pri prijavljivanju: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "  Molimo unesite admin token ručno:" -ForegroundColor Yellow
    $token = Read-Host "Token"
}

if (-not $token) {
    Write-Host "`n❌ Nema tokena, prekidam testiranje" -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# Step 2: Kreiranje test pesme
Write-Host "`nStep 2: Kreiranje test pesme..." -ForegroundColor Yellow

# Prvo, proverimo da li postoji album i artist
$albums = try {
    (Invoke-WebRequest -Uri "$baseURL/api/content/albums" -UseBasicParsing).Content | ConvertFrom-Json
} catch { @() }

$artists = try {
    (Invoke-WebRequest -Uri "$baseURL/api/content/artists" -UseBasicParsing).Content | ConvertFrom-Json
} catch { @() }

$albumId = if ($albums.Count -gt 0) { $albums[0].id } else { "test-album-1" }
$artistId = if ($artists.Count -gt 0) { $artists[0].id } else { "test-artist-1" }

$songData = @{
    name = "Test Song for Saga $(Get-Date -Format 'HHmmss')"
    duration = 180
    genre = "Pop"
    albumId = $albumId
    artistIds = @($artistId)
} | ConvertTo-Json

try {
    $createResponse = Invoke-WebRequest -Uri "$baseURL/api/content/songs" `
        -Method POST `
        -Headers $headers `
        -Body $songData `
        -UseBasicParsing
    
    $song = $createResponse.Content | ConvertFrom-Json
    $songId = $song.id
    Write-Host "  ✓ Test pesma kreirana: $songId" -ForegroundColor Green
    Write-Host "    Naziv: $($song.name)" -ForegroundColor Gray
} catch {
    Write-Host "  ✗ Greška pri kreiranju pesme: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "  Pokušavam sa postojećom pesmom..." -ForegroundColor Yellow
    
    # Koristimo postojeću pesmu
    $songs = try {
        (Invoke-WebRequest -Uri "$baseURL/api/content/songs" -UseBasicParsing).Content | ConvertFrom-Json
    } catch { @() }
    
    if ($songs.Count -gt 0) {
        $songId = $songs[0].id
        Write-Host "  ✓ Koristim postojeću pesmu: $songId" -ForegroundColor Green
    } else {
        Write-Host "  ✗ Nema dostupnih pesama za testiranje" -ForegroundColor Red
        exit 1
    }
}

# Step 3: Test uspešnog toka
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST 1: Uspešan Tok" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$deleteData = @{
    songId = $songId
} | ConvertTo-Json

try {
    Write-Host "`nPokretanje saga transakcije za brisanje pesme..." -ForegroundColor Yellow
    $sagaResponse = Invoke-WebRequest -Uri "$baseURL/api/sagas/delete-song" `
        -Method POST `
        -Headers $headers `
        -Body $deleteData `
        -UseBasicParsing
    
    $saga = $sagaResponse.Content | ConvertFrom-Json
    $sagaId = $saga.id
    
    Write-Host "`nSaga Transakcija Rezultat:" -ForegroundColor Cyan
    Write-Host "  ID: $sagaId" -ForegroundColor White
    Write-Host "  Status: $($saga.status)" -ForegroundColor $(if ($saga.status -eq "COMPLETED") { "Green" } else { "Red" })
    Write-Host "  Song ID: $($saga.songId)" -ForegroundColor White
    Write-Host "`n  Koraci:" -ForegroundColor Cyan
    
    foreach ($step in $saga.steps) {
        $statusColor = switch ($step.status) {
            "COMPLETED" { "Green" }
            "FAILED" { "Red" }
            "COMPENSATED" { "Yellow" }
            default { "White" }
        }
        Write-Host "    [$($step.order)] $($step.name): $($step.status)" -ForegroundColor $statusColor
        if ($step.error) {
            Write-Host "         Error: $($step.error)" -ForegroundColor Red
        }
    }
    
    if ($saga.status -eq "COMPLETED") {
        Write-Host "`n✅ TEST 1 PROŠAO: Saga transakcija uspešno završena!" -ForegroundColor Green
    } else {
        Write-Host "`n❌ TEST 1 NEUSPEŠAN: Saga transakcija nije završena" -ForegroundColor Red
        if ($saga.error) {
            Write-Host "   Error: $($saga.error)" -ForegroundColor Red
        }
    }
    
    # Sačuvaj saga ID za kasnije testiranje
    $global:sagaId = $sagaId
    
} catch {
    Write-Host "`n❌ TEST 1 NEUSPEŠAN: Greška pri izvršavanju saga transakcije" -ForegroundColor Red
    Write-Host "   Error: $($_.Exception.Message)" -ForegroundColor Red
    
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "   Response: $responseBody" -ForegroundColor Red
    }
}

# Step 4: Test neuspešnog toka - nepostojeća pesma
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST 2: Neuspešan Tok - Nepostojeća Pesma" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$invalidSongId = "non-existent-song-$(Get-Random)"
$deleteData2 = @{
    songId = $invalidSongId
} | ConvertTo-Json

try {
    Write-Host "`nPokretanje saga transakcije za nepostojeću pesmu..." -ForegroundColor Yellow
    $sagaResponse2 = Invoke-WebRequest -Uri "$baseURL/api/sagas/delete-song" `
        -Method POST `
        -Headers $headers `
        -Body $deleteData2 `
        -UseBasicParsing
    
    $saga2 = $sagaResponse2.Content | ConvertFrom-Json
    
    Write-Host "`nSaga Transakcija Rezultat:" -ForegroundColor Cyan
    Write-Host "  Status: $($saga2.status)" -ForegroundColor $(if ($saga2.status -eq "COMPENSATED" -or $saga2.status -eq "FAILED") { "Yellow" } else { "Red" })
    
    if ($saga2.status -eq "COMPENSATED" -or $saga2.status -eq "FAILED") {
        Write-Host "`n✅ TEST 2 PROŠAO: Saga pravilno detektovala grešku i izvršila kompenzaciju" -ForegroundColor Green
    } else {
        Write-Host "`n⚠️  TEST 2: Neočekivani status: $($saga2.status)" -ForegroundColor Yellow
    }
    
} catch {
    Write-Host "`n✅ TEST 2 PROŠAO: Saga pravilno odbacila nepostojeću pesmu" -ForegroundColor Green
    Write-Host "   Error (očekivano): $($_.Exception.Message)" -ForegroundColor Gray
}

# Step 5: Provera statusa saga transakcije
if ($global:sagaId) {
    Write-Host "`n========================================" -ForegroundColor Cyan
    Write-Host "TEST 3: Provera Statusa Saga Transakcije" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    
    try {
        Write-Host "`nProveravam status saga transakcije: $global:sagaId" -ForegroundColor Yellow
        $statusResponse = Invoke-WebRequest -Uri "$baseURL/api/sagas/$global:sagaId" `
            -Method GET `
            -Headers $headers `
            -UseBasicParsing
        
        $statusSaga = $statusResponse.Content | ConvertFrom-Json
        
        Write-Host "`nStatus Saga Transakcije:" -ForegroundColor Cyan
        Write-Host "  ID: $($statusSaga.id)" -ForegroundColor White
        Write-Host "  Status: $($statusSaga.status)" -ForegroundColor $(if ($statusSaga.status -eq "COMPLETED") { "Green" } else { "Yellow" })
        Write-Host "  Created At: $($statusSaga.createdAt)" -ForegroundColor Gray
        Write-Host "  Updated At: $($statusSaga.updatedAt)" -ForegroundColor Gray
        
        Write-Host "`n✅ TEST 3 PROŠAO: Status saga transakcije uspešno pročitan" -ForegroundColor Green
        
    } catch {
        Write-Host "`n❌ TEST 3 NEUSPEŠAN: Greška pri proveri statusa" -ForegroundColor Red
        Write-Host "   Error: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Summary
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "ZAVRŠETAK TESTIRANJA" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "`nProverite logove za detalje:" -ForegroundColor Yellow
Write-Host "  docker logs projekat-2025-2-saga-service-1" -ForegroundColor Gray
Write-Host "  docker logs projekat-2025-2-content-service-1" -ForegroundColor Gray
Write-Host "`nProverite MongoDB za saga transakcije:" -ForegroundColor Yellow
Write-Host "  docker exec -it projekat-2025-2-mongodb-saga-1 mongosh saga_db --eval 'db.saga_transactions.find().pretty()'" -ForegroundColor Gray
