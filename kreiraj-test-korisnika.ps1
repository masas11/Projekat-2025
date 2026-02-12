# Skripta za kreiranje test korisnika za brute-force demonstraciju
. .\https-helper.ps1

Write-Host "=== KREIRANJE TEST KORISNIKA ===" -ForegroundColor Cyan
Write-Host ""

$body = @{
    firstName = "Test"
    lastName = "User"
    email = "testuser@example.com"
    username = "testuser"
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

Write-Host "Registracija korisnika: testuser" -ForegroundColor Yellow
$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"

if ($result.StatusCode -eq 201) {
    Write-Host "✓ Korisnik uspešno kreiran!" -ForegroundColor Green
    Write-Host "  Username: testuser" -ForegroundColor White
    Write-Host "  Password: Test1234!" -ForegroundColor White
    Write-Host "`nSada možete pokrenuti brute-force test:" -ForegroundColor Cyan
    Write-Host "  .\test-brute-force-attack.ps1" -ForegroundColor Gray
} elseif ($result.StatusCode -eq 409) {
    Write-Host "⚠ Korisnik već postoji!" -ForegroundColor Yellow
    Write-Host "  Možete direktno pokrenuti brute-force test:" -ForegroundColor Cyan
    Write-Host "  .\test-brute-force-attack.ps1" -ForegroundColor Gray
} else {
    Write-Host "✗ Greška pri kreiranju korisnika!" -ForegroundColor Red
    Write-Host "  Status: $($result.StatusCode)" -ForegroundColor Red
    Write-Host "  Response: $($result.Content)" -ForegroundColor Gray
}
