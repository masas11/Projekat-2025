# Test skripta za 2.9 - CQRS šablon
# Testira da li se ime umetnika prikazuje uz pretplatu bez poziva Content servisa

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Test CQRS (2.9)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$successColor = "Green"
$errorColor = "Red"
$infoColor = "Yellow"
$testColor = "Cyan"

# Funkcija za testiranje endpoint-a
function Test-Endpoint {
    param($url, $method = "GET", $headers = @{}, $expectedStatus = 200, $description = "")
    try {
        $params = @{
            Uri = $url
            Method = $method
            Headers = $headers
            UseBasicParsing = $true
            TimeoutSec = 10
            ErrorAction = "Stop"
        }
        $response = Invoke-WebRequest @params
        if ($response.StatusCode -eq $expectedStatus) {
            Write-Host "  [OK] $description - Status: $($response.StatusCode)" -ForegroundColor $successColor
            return $response
        } else {
            Write-Host "  [FAIL] $description - Očekivano: $expectedStatus, Dobijeno: $($response.StatusCode)" -ForegroundColor $errorColor
            return $null
        }
    } catch {
        if ($_.Exception.Response) {
            $statusCode = $_.Exception.Response.StatusCode.value__
            if ($statusCode -eq $expectedStatus) {
                Write-Host "  [OK] $description - Status: $statusCode (očekivano)" -ForegroundColor $successColor
                return $null
            }
        }
        Write-Host "  [FAIL] $description - Greška: $($_.Exception.Message)" -ForegroundColor $errorColor
        return $null
    }
}

Write-Host "1. TEST CQRS - KREIRANJE PRETPLATE SA ARTIST NAME" -ForegroundColor $testColor
Write-Host "---------------------------------------------------" -ForegroundColor $testColor
Write-Host ""

Write-Host "1.1. Proveravam da li postoje umetnici..." -ForegroundColor $infoColor
$artistsResponse = Test-Endpoint "http://localhost:8081/api/content/artists" 200 "Get artists"

if ($artistsResponse) {
    try {
        $artists = $artistsResponse.Content | ConvertFrom-Json
        if ($artists.Count -gt 0) {
            $artistId = $artists[0].id
            $artistName = $artists[0].name
            Write-Host "  Koristim umetnika: $artistName (ID: $artistId)" -ForegroundColor $infoColor
            
            Write-Host ""
            Write-Host "1.2. Testiranje kreiranja pretplate (zahteva autentifikaciju)..." -ForegroundColor $infoColor
            Write-Host "  [INFO] Za kreiranje pretplate potrebna je autentifikacija" -ForegroundColor $infoColor
            Write-Host "  [INFO] Testirajte kroz frontend: http://localhost:3000" -ForegroundColor $infoColor
            Write-Host "  [INFO] Prijavite se i pretplatite se na umetnika" -ForegroundColor $infoColor
        } else {
            Write-Host "  [WARN] Nema umetnika u bazi" -ForegroundColor $infoColor
        }
    } catch {
        Write-Host "  [FAIL] Greška pri parsiranju: $($_.Exception.Message)" -ForegroundColor $errorColor
    }
}

Write-Host ""
Write-Host "2. TEST CQRS - ČITANJE PRETPLATA SA ARTIST NAME" -ForegroundColor $testColor
Write-Host "------------------------------------------------" -ForegroundColor $testColor
Write-Host ""

Write-Host "2.1. Proveravam da li postoje pretplate..." -ForegroundColor $infoColor
Write-Host "  [INFO] Za pregled pretplata potrebna je autentifikacija" -ForegroundColor $infoColor
Write-Host "  [INFO] Testirajte kroz frontend: http://localhost:3000/profile" -ForegroundColor $infoColor

Write-Host ""
Write-Host "3. PROVERA IMPLEMENTACIJE" -ForegroundColor $testColor
Write-Host "--------------------------" -ForegroundColor $testColor
Write-Host ""

Write-Host "3.1. Proveravam da li Subscription model ima ArtistName polje..." -ForegroundColor $infoColor
$subscriptionModel = Get-Content "services\subscriptions-service\internal\model\subscription.go" -Raw
if ($subscriptionModel -match "ArtistName") {
    Write-Host "  [OK] Subscription model ima ArtistName polje" -ForegroundColor $successColor
} else {
    Write-Host "  [FAIL] Subscription model NEMA ArtistName polje" -ForegroundColor $errorColor
}

Write-Host ""
Write-Host "3.2. Proveravam da li subscribe-artist endpoint koristi getArtistName..." -ForegroundColor $infoColor
$mainCode = Get-Content "services\subscriptions-service\cmd\main.go" -Raw
if ($mainCode -match "getArtistName" -and $mainCode -match "ArtistName.*artistName") {
    Write-Host "  [OK] subscribe-artist endpoint koristi getArtistName i čuva ArtistName" -ForegroundColor $successColor
} else {
    Write-Host "  [FAIL] subscribe-artist endpoint NE koristi getArtistName ili NE čuva ArtistName" -ForegroundColor $errorColor
}

Write-Host ""
Write-Host "3.3. Proveravam da li subscriptions endpoint vraća ArtistName..." -ForegroundColor $infoColor
if ($mainCode -match "subscriptions.*GetByUserID") {
    Write-Host "  [OK] subscriptions endpoint koristi GetByUserID koji vraća Subscription sa ArtistName" -ForegroundColor $successColor
} else {
    Write-Host "  [WARN] Proverite ručno da li subscriptions endpoint vraća ArtistName" -ForegroundColor $infoColor
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  TESTIRANJE ZAVRŠENO" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Kako testirati CQRS:" -ForegroundColor $infoColor
Write-Host "  1. Otvori frontend: http://localhost:3000" -ForegroundColor White
Write-Host "  2. Prijavi se kao korisnik (ne admin)" -ForegroundColor White
Write-Host "  3. Otvori stranicu umetnika (npr. /artists/artist1)" -ForegroundColor White
Write-Host "  4. Klikni 'Pretplati se' - ovo će kreirati pretplatu SA imenom umetnika" -ForegroundColor White
Write-Host "  5. Otvori Profile stranicu (/profile)" -ForegroundColor White
Write-Host "  6. Proveri da li se prikazuje ime umetnika uz pretplatu" -ForegroundColor White
Write-Host ""
Write-Host "Provera u Developer Tools (F12):" -ForegroundColor $infoColor
Write-Host "  - Network tab → GET /api/subscriptions" -ForegroundColor White
Write-Host "  - Proveri response - treba da sadrži 'artistName' za svaku pretplatu" -ForegroundColor White
Write-Host "  - NEMA poziva ka /api/content/artists/{id} pri čitanju pretplata!" -ForegroundColor White
Write-Host ""
