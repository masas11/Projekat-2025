# Skripta za dodavanje sertifikata u Windows Certificate Store
Write-Host "🔐 Dodavanje sertifikata u Windows Certificate Store..." -ForegroundColor Cyan
Write-Host ""

# Provera da sertifikat postoji
if (-not (Test-Path ".\certs\server.crt")) {
    Write-Host "❌ Sertifikat ne postoji! Prvo generišite sertifikate." -ForegroundColor Red
    Write-Host "   Pokrenite: .\generate-certs.ps1" -ForegroundColor Yellow
    exit 1
}

# Eksportuj sertifikat u DER format (ako već ne postoji)
if (-not (Test-Path ".\certs\server.der")) {
    Write-Host "📦 Eksportovanje sertifikata u DER format..." -ForegroundColor Yellow
    
    # Koristi OpenSSL preko Docker-a
    docker run --rm -v "${PWD}/certs:/certs" alpine/openssl x509 -in /certs/server.crt -out /certs/server.der -outform DER 2>&1 | Out-Null
    
    if (-not (Test-Path ".\certs\server.der")) {
        Write-Host "❌ Neuspešno eksportovanje sertifikata." -ForegroundColor Red
        Write-Host "   Proverite da li Docker radi i da li sertifikat postoji." -ForegroundColor Yellow
        exit 1
    }
    
    Write-Host "✅ Sertifikat eksportovan u DER format." -ForegroundColor Green
}

# Dodaj sertifikat u Windows Certificate Store
Write-Host ""
Write-Host "➕ Dodavanje sertifikata u Trusted Root Certification Authorities..." -ForegroundColor Yellow
Write-Host "   (Može zahtevati administrator privilegije)" -ForegroundColor Gray
Write-Host ""

try {
    $result = certutil -addstore -f "Root" certs\server.der 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Sertifikat je dodat uspešno!" -ForegroundColor Green
        Write-Host ""
        Write-Host "🔄 Restartujte browser da bi sertifikat bio aktivan." -ForegroundColor Yellow
        Write-Host ""
        Write-Host "📝 Napomena: Ako browser i dalje prikazuje upozorenje," -ForegroundColor Gray
        Write-Host "   kliknite 'Advanced' → 'Proceed to localhost (unsafe)'" -ForegroundColor Gray
    } else {
        Write-Host "❌ Greška pri dodavanju sertifikata." -ForegroundColor Red
        Write-Host "   Pokušajte sa administrator privilegijama." -ForegroundColor Yellow
        Write-Host ""
        Write-Host "💡 Alternativno, prihvatite sertifikat direktno u browser-u:" -ForegroundColor Cyan
        Write-Host "   1. Otvorite https://localhost:8081/api/users/health" -ForegroundColor White
        Write-Host "   2. Kliknite 'Advanced' → 'Proceed to localhost (unsafe)'" -ForegroundColor White
    }
} catch {
    Write-Host "❌ Greška: $_" -ForegroundColor Red
    Write-Host ""
    Write-Host "💡 Alternativno, prihvatite sertifikat direktno u browser-u:" -ForegroundColor Cyan
    Write-Host "   1. Otvorite https://localhost:8081/api/users/health" -ForegroundColor White
    Write-Host "   2. Kliknite 'Advanced' → 'Proceed to localhost (unsafe)'" -ForegroundColor White
}
