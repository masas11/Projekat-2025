# Test skripta za ocenu 7 - SOA & NoSQL
# Pokreni sa: .\test-grade7.ps1

Write-Host "🚀 Pokretanje testiranja za ocenu 7..." -ForegroundColor Green

# Proveri da li Docker radi
try {
    docker version | Out-Null
    Write-Host "✅ Docker je dostupan" -ForegroundColor Green
} catch {
    Write-Host "❌ Docker nije dostupan. Instaliraj Docker prvo." -ForegroundColor Red
    exit 1
}

# Pokreni sve servise
Write-Host "📦 Pokretanje svih servisa..." -ForegroundColor Yellow
docker-compose up -d

# Sačekaj da servisi startuju
Write-Host "⏳ Čekam da servisi startuju (30 sekundi)..." -ForegroundColor Yellow
Start-Sleep -Seconds 30

# Proveri health endpoint-e
$services = @(
    @{name="API Gateway"; url="http://localhost:8080/health"},
    @{name="Content Service"; url="http://localhost:8081/health"},
    @{name="Users Service"; url="http://localhost:8082/health"},
    @{name="Ratings Service"; url="http://localhost:8083/health"},
    @{name="Subscriptions Service"; url="http://localhost:8084/health"},
    @{name="Notifications Service"; url="http://localhost:8085/health"}
)

Write-Host "🏥 Provera health endpoint-a..." -ForegroundColor Yellow
$allHealthy = $true

foreach ($service in $services) {
    try {
        $response = Invoke-WebRequest -Uri $service.url -TimeoutSec 5
        if ($response.StatusCode -eq 200) {
            Write-Host "✅ $($service.name) je zdrav" -ForegroundColor Green
        } else {
            Write-Host "❌ $($service.name) nije zdrav (Status: $($response.StatusCode))" -ForegroundColor Red
            $allHealthy = $false
        }
    } catch {
        Write-Host "❌ $($service.name) nije dostupan" -ForegroundColor Red
        $allHealthy = $false
    }
}

if (-not $allHealthy) {
    Write-Host "⚠️ Neki servisi nisu zdravi. Nastavljam testiranje..." -ForegroundColor Yellow
}

# Test 1: Registracija
Write-Host "👤 Test 1: Registracija korisnika..." -ForegroundColor Yellow
try {
    $registerResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/users/register" `
        -Method POST `
        -ContentType "application/json" `
        -Body @{
            username="testuser$(Get-Random)"
            password="StrongPass123!"
            email="test$(Get-Random)@example.com"
            firstName="Test"
            lastName="User"
        } | ConvertTo-Json
    
    Write-Host "✅ Registracija uspešna" -ForegroundColor Green
} catch {
    Write-Host "❌ Registracija neuspešna: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 2: Kreiranje umetnika
Write-Host "🎨 Test 2: Kreiranje umetnika..." -ForegroundColor Yellow
try {
    $artistResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/content/artists" `
        -Method POST `
        -ContentType "application/json" `
        -Body @{
            name="Test Artist $(Get-Random)"
            biography="Test biography for testing purposes"
            genres=@("Pop", "Rock")
        } | ConvertTo-Json
    
    $artistId = $artistResponse.id
    Write-Host "✅ Umetnik kreiran (ID: $artistId)" -ForegroundColor Green
} catch {
    Write-Host "❌ Kreiranje umetnika neuspešno: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 3: Kreiranje albuma
Write-Host "💿 Test 3: Kreiranje albuma..." -ForegroundColor Yellow
try {
    $albumResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/content/albums" `
        -Method POST `
        -ContentType "application/json" `
        -Body @{
            name="Test Album $(Get-Random)"
            releaseDate="2024-01-01"
            genre="Pop"
            artistIds=@($artistId)
        } | ConvertTo-Json
    
    $albumId = $albumResponse.id
    Write-Host "✅ Album kreiran (ID: $albumId)" -ForegroundColor Green
} catch {
    Write-Host "❌ Kreiranje albuma neuspešno: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 4: Kreiranje pesme
Write-Host "🎵 Test 4: Kreiranje pesme..." -ForegroundColor Yellow
try {
    $songResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/content/songs" `
        -Method POST `
        -ContentType "application/json" `
        -Body @{
            name="Test Song $(Get-Random)"
            duration=180
            genre="Pop"
            albumId=$albumId
            artistIds=@($artistId)
            audioFileUrl="https://www.soundhelix.com/examples/mp3/SoundHelix-Song-1.mp3"
        } | ConvertTo-Json
    
    $songId = $songResponse.id
    Write-Host "✅ Pesma kreirana (ID: $songId)" -ForegroundColor Green
} catch {
    Write-Host "❌ Kreiranje pesme neuspešno: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 5: Ocenjivanje pesme (sa sinhronom validacijom)
Write-Host "⭐ Test 5: Ocenjivanje pesme..." -ForegroundColor Yellow
try {
    $ratingResponse = Invoke-RestMethod -Uri "http://localhost:8083/rate-song?songId=$songId&rating=5&userId=testuser" `
        -Method POST
    
    Write-Host "✅ Pesma ocenjena uspešno" -ForegroundColor Green
} catch {
    Write-Host "❌ Ocenjivanje neuspešno: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 6: Pretplata na umetnika (sa sinhronom validacijom)
Write-Host "🔔 Test 6: Pretplata na umetnika..." -ForegroundColor Yellow
try {
    $subscriptionResponse = Invoke-RestMethod -Uri "http://localhost:8084/subscribe-artist?artistId=$artistId&userId=testuser" `
        -Method POST
    
    Write-Host "✅ Pretplata uspešna" -ForegroundColor Green
} catch {
    Write-Host "❌ Pretplata neuspešna: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 7: Circuit Breaker
Write-Host "⚡ Test 7: Circuit Breaker..." -ForegroundColor Yellow
try {
    # Napravi 3 neuspešna poziva
    for ($i = 1; $i -le 3; $i++) {
        try {
            Invoke-RestMethod -Uri "http://localhost:8083/rate-song?songId=invalid&rating=5&userId=test" -Method POST | Out-Null
        } catch {
            Write-Host "  Neuspešan poziv $i (očekivano)" -ForegroundColor Gray
        }
    }
    
    # Četvrti poziv treba da bude blokiran
    try {
        Invoke-RestMethod -Uri "http://localhost:8083/rate-song?songId=$songId&rating=5&userId=test" -Method POST | Out-Null
        Write-Host "❌ Circuit breaker se nije aktivirao" -ForegroundColor Red
    } catch {
        if ($_.Exception.Message -like "*circuit breaker*") {
            Write-Host "✅ Circuit breaker se aktivirao" -ForegroundColor Green
        } else {
            Write-Host "⚠️ Circuit breaker test nejasan: $($_.Exception.Message)" -ForegroundColor Yellow
        }
    }
} catch {
    Write-Host "❌ Circuit breaker test neuspešno: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 8: Audio Streaming
Write-Host "🎧 Test 8: Audio streaming..." -ForegroundColor Yellow
try {
    $streamResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/content/songs/$songId/stream" -Method Head
    if ($streamResponse.StatusCode -eq 200) {
        Write-Host "✅ Audio streaming endpoint radi" -ForegroundColor Green
    } else {
        Write-Host "❌ Audio streaming ne radi (Status: $($streamResponse.StatusCode))" -ForegroundColor Red
    }
} catch {
    Write-Host "❌ Audio streaming neuspešno: $($_.Exception.Message)" -ForegroundColor Red
}

# Prikaz logova
Write-Host "📋 Poslednje log poruke iz servisa..." -ForegroundColor Yellow
Write-Host "--- Ratings Service Logs ---" -ForegroundColor Cyan
docker-compose logs --tail=5 ratings-service

Write-Host "--- Subscriptions Service Logs ---" -ForegroundColor Cyan
docker-compose logs --tail=5 subscriptions-service

# Frontend test
Write-Host "🌐 Frontend test..." -ForegroundColor Yellow
try {
    $frontendResponse = Invoke-WebRequest -Uri "http://localhost:3000" -TimeoutSec 5
    if ($frontendResponse.StatusCode -eq 200) {
        Write-Host "✅ Frontend je dostupan na http://localhost:3000" -ForegroundColor Green
        Write-Host "🔗 Test rute:" -ForegroundColor Cyan
        Write-Host "  - http://localhost:3000/songs (lista pesama)" -ForegroundColor Gray
        Write-Host "  - http://localhost:3000/songs/$songId (AudioPlayer test)" -ForegroundColor Gray
        Write-Host "  - http://localhost:3000/url-tester (URL tester)" -ForegroundColor Gray
    } else {
        Write-Host "❌ Frontend nije dostupan" -ForegroundColor Red
    }
} catch {
    Write-Host "❌ Frontend test neuspešan: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "💡 Pokreni frontend sa: cd frontend && npm start" -ForegroundColor Yellow
}

Write-Host "🎉 Testiranje završeno!" -ForegroundColor Green
Write-Host "📊 Proveri detaljne rezultate u TESTING_GUIDE.md" -ForegroundColor Cyan
Write-Host "🔧 Za debugiranje koristi: docker-compose logs [service-name]" -ForegroundColor Cyan
