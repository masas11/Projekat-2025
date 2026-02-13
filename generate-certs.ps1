# PowerShell skripta za kreiranje SSL sertifikata
Write-Host "🔐 Kreiranje SSL sertifikata..." -ForegroundColor Cyan

# Kreiraj certs direktorijum ako ne postoji
if (-not (Test-Path ".\certs")) {
    New-Item -ItemType Directory -Path ".\certs" | Out-Null
}

# Kreiraj self-signed sertifikat sa privatnim ključem
$cert = New-SelfSignedCertificate `
    -DnsName "localhost" `
    -CertStoreLocation "Cert:\CurrentUser\My" `
    -KeyAlgorithm RSA `
    -KeyLength 2048 `
    -NotAfter (Get-Date).AddYears(1) `
    -KeyExportPolicy Exportable `
    -KeySpec Signature

# Exportuj sertifikat u PEM format
$certBytes = $cert.Export([System.Security.Cryptography.X509Certificates.X509ContentType]::Cert)
[System.IO.File]::WriteAllBytes(".\certs\server.crt", $certBytes)

# Exportuj privatni ključ
$rsa = [System.Security.Cryptography.X509Certificates.RSACertificateExtensions]::GetRSAPrivateKey($cert)
$keyBytes = $rsa.ExportRSAPrivateKey()
[System.IO.File]::WriteAllBytes(".\certs\server.key", $keyBytes)

# Obriši sertifikat iz Windows sertifikat store-a
Remove-Item "Cert:\CurrentUser\My\$($cert.Thumbprint)" -ErrorAction SilentlyContinue

Write-Host "✅ SSL sertifikati kreirani!" -ForegroundColor Green
Write-Host "📁 Fajlovi:" -ForegroundColor Yellow
Write-Host "   - certs/server.crt (sertifikat)" -ForegroundColor White
Write-Host "   - certs/server.key (privatni ključ)" -ForegroundColor White
Write-Host ""
Write-Host "🚀 Pokreni sa: docker-compose up" -ForegroundColor Cyan
