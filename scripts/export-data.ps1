# PowerShell skripta za eksport podataka iz baza u JSON format
# Ovi fajlovi se mogu commit-ovati u git i deliti sa timom

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  EKSPORT PODATAKA IZ BAZA" -ForegroundColor Cyan
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
    exit 1
}

# Kreiraj folder za eksportovane podatke
$exportDir = "scripts\seed-data"
if (-not (Test-Path $exportDir)) {
    New-Item -ItemType Directory -Path $exportDir | Out-Null
    Write-Host "Kreiran folder: $exportDir" -ForegroundColor Green
}

Write-Host ""
Write-Host "Eksportujem podatke..." -ForegroundColor Yellow
Write-Host ""

# Eksport Content baze - Artists
Write-Host "1. Eksport Artists..." -ForegroundColor Cyan
docker exec $contentContainer mongoexport --db=music_streaming --collection=artists --jsonArray --pretty | Out-File -FilePath "$exportDir\artists.json" -Encoding UTF8
if ($LASTEXITCODE -eq 0) {
    Write-Host "   [OK] Artists eksportovani" -ForegroundColor Green
}

# Eksport Content baze - Albums
Write-Host "2. Eksport Albums..." -ForegroundColor Cyan
docker exec $contentContainer mongoexport --db=music_streaming --collection=albums --jsonArray --pretty | Out-File -FilePath "$exportDir\albums.json" -Encoding UTF8
if ($LASTEXITCODE -eq 0) {
    Write-Host "   [OK] Albums eksportovani" -ForegroundColor Green
}

# Eksport Content baze - Songs
Write-Host "3. Eksport Songs..." -ForegroundColor Cyan
docker exec $contentContainer mongoexport --db=music_streaming --collection=songs --jsonArray --pretty | Out-File -FilePath "$exportDir\songs.json" -Encoding UTF8
if ($LASTEXITCODE -eq 0) {
    Write-Host "   [OK] Songs eksportovani" -ForegroundColor Green
}

# Eksport Ratings baze
Write-Host "4. Eksport Ratings..." -ForegroundColor Cyan
docker exec $ratingsContainer mongoexport --db=ratings_db --collection=ratings --jsonArray --pretty | Out-File -FilePath "$exportDir\ratings.json" -Encoding UTF8
if ($LASTEXITCODE -eq 0) {
    Write-Host "   [OK] Ratings eksportovani" -ForegroundColor Green
}

# Eksport Subscriptions baze
Write-Host "5. Eksport Subscriptions..." -ForegroundColor Cyan
docker exec $subscriptionsContainer mongoexport --db=subscriptions_db --collection=subscriptions --jsonArray --pretty | Out-File -FilePath "$exportDir\subscriptions.json" -Encoding UTF8
if ($LASTEXITCODE -eq 0) {
    Write-Host "   [OK] Subscriptions eksportovani" -ForegroundColor Green
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  EKSPORT ZAVRSEN!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Eksportovani fajlovi:" -ForegroundColor Yellow
Write-Host "  - $exportDir\artists.json" -ForegroundColor White
Write-Host "  - $exportDir\albums.json" -ForegroundColor White
Write-Host "  - $exportDir\songs.json" -ForegroundColor White
Write-Host "  - $exportDir\ratings.json" -ForegroundColor White
Write-Host "  - $exportDir\subscriptions.json" -ForegroundColor White
Write-Host ""
Write-Host "Sledeci koraci:" -ForegroundColor Yellow
Write-Host "  1. Proverite eksportovane fajlove" -ForegroundColor White
Write-Host "  2. Commit-ujte ih u git:" -ForegroundColor White
Write-Host "     git add scripts/seed-data/*.json" -ForegroundColor Cyan
Write-Host "     git commit -m 'Update seed data'" -ForegroundColor Cyan
Write-Host "     git push" -ForegroundColor Cyan
Write-Host ""
Write-Host "Drugi clanovi tima mogu sada:" -ForegroundColor Yellow
Write-Host "  1. git pull" -ForegroundColor White
Write-Host '  2. .\scripts\import-data.ps1' -ForegroundColor White
Write-Host ""
