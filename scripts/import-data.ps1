# PowerShell skripta za import podataka iz eksportovanih JSON fajlova
# Koristi se nakon git pull da se učitaju najnoviji podaci

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  IMPORT PODATAKA U BAZE" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Proveri da li su kontejneri pokrenuti
$contentContainer = docker ps --filter "name=mongodb-content" --format "{{.Names}}" | Select-Object -First 1
$ratingsContainer = docker ps --filter "name=mongodb-ratings" --format "{{.Names}}" | Select-Object -First 1
$subscriptionsContainer = docker ps --filter "name=mongodb-subscriptions" --format "{{.Names}}" | Select-Object -First 1
$usersContainer = docker ps --filter "name=mongodb-users" --format "{{.Names}}" | Select-Object -First 1

if ([string]::IsNullOrEmpty($contentContainer)) {
    Write-Host "ERROR: MongoDB kontejneri nisu pokrenuti!" -ForegroundColor Red
    Write-Host "Pokrenite: docker-compose up -d" -ForegroundColor Yellow
    Write-Host "Sačekajte 20 sekundi, pa pokrenite ovu skriptu ponovo." -ForegroundColor Yellow
    exit 1
}

# Proveri da li postoje eksportovani fajlovi
$exportDir = "scripts\seed-data"
if (-not (Test-Path $exportDir)) {
    Write-Host "WARNING: Folder $exportDir ne postoji!" -ForegroundColor Yellow
    Write-Host "Koristite osnovne seed skripte: .\scripts\seed-all.ps1" -ForegroundColor Yellow
    exit 0
}

Write-Host "Sačekajte da se baze potpuno pokrenu..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

Write-Host ""
Write-Host "Importujem podatke..." -ForegroundColor Yellow
Write-Host ""

# Import Artists
if (Test-Path "$exportDir\artists.json") {
    Write-Host "1. Import Artists..." -ForegroundColor Cyan
    Get-Content "$exportDir\artists.json" | docker exec -i $contentContainer mongoimport --db=music_streaming --collection=artists --jsonArray --drop
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   ✓ Artists importovani" -ForegroundColor Green
    }
} else {
    Write-Host "   ⚠ artists.json ne postoji, preskačem..." -ForegroundColor Yellow
}

# Import Albums
if (Test-Path "$exportDir\albums.json") {
    Write-Host "2. Import Albums..." -ForegroundColor Cyan
    Get-Content "$exportDir\albums.json" | docker exec -i $contentContainer mongoimport --db=music_streaming --collection=albums --jsonArray --drop
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   ✓ Albums importovani" -ForegroundColor Green
    }
} else {
    Write-Host "   ⚠ albums.json ne postoji, preskačem..." -ForegroundColor Yellow
}

# Import Songs
if (Test-Path "$exportDir\songs.json") {
    Write-Host "3. Import Songs..." -ForegroundColor Cyan
    Get-Content "$exportDir\songs.json" | docker exec -i $contentContainer mongoimport --db=music_streaming --collection=songs --jsonArray --drop
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   ✓ Songs importovani" -ForegroundColor Green
    }
} else {
    Write-Host "   ⚠ songs.json ne postoji, preskačem..." -ForegroundColor Yellow
}

# Import Ratings
if (Test-Path "$exportDir\ratings.json") {
    Write-Host "4. Import Ratings..." -ForegroundColor Cyan
    Get-Content "$exportDir\ratings.json" | docker exec -i $ratingsContainer mongoimport --db=ratings_db --collection=ratings --jsonArray --drop
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   ✓ Ratings importovani" -ForegroundColor Green
    }
} else {
    Write-Host "   ⚠ ratings.json ne postoji, preskačem..." -ForegroundColor Yellow
}

# Import Subscriptions
if (Test-Path "$exportDir\subscriptions.json") {
    Write-Host "5. Import Subscriptions..." -ForegroundColor Cyan
    Get-Content "$exportDir\subscriptions.json" | docker exec -i $subscriptionsContainer mongoimport --db=subscriptions_db --collection=subscriptions --jsonArray --drop
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   [OK] Subscriptions importovani" -ForegroundColor Green
    }
} else {
    Write-Host "   [WARN] subscriptions.json ne postoji, preskačem..." -ForegroundColor Yellow
}

# Import Users
if (Test-Path "$exportDir\users.json") {
    Write-Host "6. Import Users..." -ForegroundColor Cyan
    Get-Content "$exportDir\users.json" | docker exec -i $usersContainer mongoimport --db=users_db --collection=users --jsonArray --drop
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   [OK] Users importovani" -ForegroundColor Green
    }
} else {
    Write-Host "   [WARN] users.json ne postoji, preskačem..." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  IMPORT ZAVRSEN!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Napomena: Ako JSON fajlovi ne postoje, koristite:" -ForegroundColor Yellow
Write-Host "  .\scripts\seed-all.ps1" -ForegroundColor White
Write-Host ""
