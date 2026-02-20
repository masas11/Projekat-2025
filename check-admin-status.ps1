# Check admin account status
Write-Host "Checking admin account status..." -ForegroundColor Green

$usersContainer = docker ps --filter "name=mongodb-users" --format "{{.Names}}" | Select-Object -First 1

if ([string]::IsNullOrEmpty($usersContainer)) {
    Write-Host "Error: MongoDB users container not found!" -ForegroundColor Red
    exit 1
}

Write-Host "Found container: $usersContainer" -ForegroundColor Cyan

$checkCommands = @"
use music_streaming;
db.users.findOne({username: "admin"}, {
    username: 1, 
    email: 1, 
    failedLoginAttempts: 1, 
    lockedUntil: 1, 
    accountLocked: 1
});
"@

Write-Host "Admin account status:" -ForegroundColor Yellow
$checkCommands | docker exec -i $usersContainer mongosh
