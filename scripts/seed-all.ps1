# PowerShell script to seed all databases with initial data
# Make sure Docker containers are running first!

Write-Host "Seeding databases with initial data..." -ForegroundColor Green

# Wait for MongoDB containers to be ready
Write-Host "Waiting for MongoDB containers to be ready..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Get container names
$usersContainer = docker ps --filter "name=mongodb-users" --format "{{.Names}}" | Select-Object -First 1
$contentContainer = docker ps --filter "name=mongodb-content" --format "{{.Names}}" | Select-Object -First 1
$ratingsContainer = docker ps --filter "name=mongodb-ratings" --format "{{.Names}}" | Select-Object -First 1

if ([string]::IsNullOrEmpty($usersContainer) -or [string]::IsNullOrEmpty($contentContainer) -or [string]::IsNullOrEmpty($ratingsContainer)) {
    Write-Host "Error: MongoDB containers not found. Make sure they are running!" -ForegroundColor Red
    Write-Host "   Run: docker-compose up -d" -ForegroundColor Yellow
    exit 1
}

Write-Host "Seeding content database..." -ForegroundColor Cyan
Get-Content scripts/seed-content.js | docker exec -i $contentContainer mongosh music_streaming | Out-Null
if ($LASTEXITCODE -eq 0) {
    Write-Host "Content database seeded successfully!" -ForegroundColor Green
} else {
    Write-Host "Warning: Content seeding may have failed." -ForegroundColor Yellow
}

Write-Host "Seeding ratings database..." -ForegroundColor Cyan
Get-Content scripts/seed-ratings.js | docker exec -i $ratingsContainer mongosh ratings_db | Out-Null
if ($LASTEXITCODE -eq 0) {
    Write-Host "Ratings database seeded successfully!" -ForegroundColor Green
} else {
    Write-Host "Warning: Ratings seeding may have failed." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "All databases seeded successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Note: Users database is automatically seeded by users-service on startup" -ForegroundColor Yellow
Write-Host "   (Admin user: username='admin', password='admin123')" -ForegroundColor Yellow
