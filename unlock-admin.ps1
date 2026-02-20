# PowerShell script to unlock admin account
# Resets failed login attempts and unlocks the account

Write-Host "Unlocking admin account..." -ForegroundColor Green

# Get MongoDB users container name
$usersContainer = docker ps --filter "name=mongodb-users" --format "{{.Names}}" | Select-Object -First 1

if ([string]::IsNullOrEmpty($usersContainer)) {
    Write-Host "Error: MongoDB users container not found. Make sure it's running!" -ForegroundColor Red
    Write-Host "   Run: docker-compose up -d mongodb-users" -ForegroundColor Yellow
    exit 1
}

Write-Host "Found MongoDB users container: $usersContainer" -ForegroundColor Cyan

# MongoDB commands to unlock admin account
$unlockCommands = @"
use music_streaming;

// Reset failed login attempts for admin user
db.users.updateOne(
    { username: "admin" },
    { 
        "`$set": { 
            failedLoginAttempts: 0,
            lockedUntil: null,
            accountLocked: false
        }
    }
);

// Verify the update
db.users.findOne({ username: "admin" }, { username: 1, email: 1, failedLoginAttempts: 1, lockedUntil: 1, accountLocked: 1 });

print("Admin account unlocked successfully!");
"@

Write-Host "Executing unlock commands..." -ForegroundColor Yellow
$unlockCommands | docker exec -i $usersContainer mongosh | Out-Null

if ($LASTEXITCODE -eq 0) {
    Write-Host "Admin account unlocked successfully!" -ForegroundColor Green
    Write-Host "You can now login with:" -ForegroundColor Cyan
    Write-Host "   Username: admin" -ForegroundColor White
    Write-Host "   Password: admin123" -ForegroundColor White
} else {
    Write-Host "Error: Failed to unlock admin account!" -ForegroundColor Red
    Write-Host "Check MongoDB container logs for details." -ForegroundColor Yellow
    exit 1
}
