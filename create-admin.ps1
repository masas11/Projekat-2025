# Create admin account
Write-Host "Creating admin account..." -ForegroundColor Green

$usersContainer = docker ps --filter "name=mongodb-users" --format "{{.Names}}" | Select-Object -First 1

if ([string]::IsNullOrEmpty($usersContainer)) {
    Write-Host "Error: MongoDB users container not found!" -ForegroundColor Red
    exit 1
}

Write-Host "Found container: $usersContainer" -ForegroundColor Cyan

$createCommands = @"
use music_streaming;

// Check if admin already exists
var existingAdmin = db.users.findOne({username: "admin"});
if (existingAdmin) {
    print("Admin user already exists!");
    db.users.findOne({username: "admin"}, {username: 1, email: 1, failedLoginAttempts: 1, lockedUntil: 1, accountLocked: 1});
} else {
    // Create admin user
    db.users.insertOne({
        username: "admin",
        email: "admin@example.com",
        password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
        firstName: "Admin",
        lastName: "User",
        role: "ADMIN",
        failedLoginAttempts: 0,
        lockedUntil: null,
        accountLocked: false,
        createdAt: new Date(),
        updatedAt: new Date()
    });
    
    print("Admin user created successfully!");
    print("Username: admin");
    print("Password: password");
}
"@

Write-Host "Creating admin account..." -ForegroundColor Yellow
$createCommands | docker exec -i $usersContainer mongosh
