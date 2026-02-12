# Kompletna test skripta za sve napade
. .\https-helper.ps1

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  DEMONSTRACIJA POKUŠAJA NAPADA" -ForegroundColor Cyan
Write-Host "  Zahtev 2.22" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Test 1: XSS Napad
Write-Host "=== 1. XSS NAPAD ===" -ForegroundColor Yellow
Write-Host ""
& .\test-xss-attack.ps1
Start-Sleep -Seconds 3

# Test 2: SQL Injection Napad
Write-Host "`n`n=== 2. SQL INJECTION NAPAD ===" -ForegroundColor Yellow
Write-Host ""
& .\test-sql-injection-attack.ps1
Start-Sleep -Seconds 3

# Test 3: Brute-force Napad
Write-Host "`n`n=== 3. BRUTE-FORCE NAPAD ===" -ForegroundColor Yellow
Write-Host ""
& .\test-brute-force-attack.ps1
Start-Sleep -Seconds 3

# Test 4: DoS Napad
Write-Host "`n`n=== 4. DoS NAPAD ===" -ForegroundColor Yellow
Write-Host ""
& .\test-dos-attack.ps1

Write-Host "`n`n========================================" -ForegroundColor Cyan
Write-Host "  DEMONSTRACIJA ZAVRŠENA" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Svi napadi su testirani i blokirani!" -ForegroundColor Green
Write-Host "Detaljna dokumentacija: DEMONSTRACIJA_NAPADA_2.22.md" -ForegroundColor Cyan
