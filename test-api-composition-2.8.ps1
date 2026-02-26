# Test skripta za 2.8 - API Composition
# Testira da li se prikazuju broj ocena i prosečna ocena uz svaku pesmu

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Test API Composition (2.8)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$successColor = "Green"
$errorColor = "Red"
$infoColor = "Yellow"
$testColor = "Cyan"

# Funkcija za testiranje endpoint-a
function Test-Endpoint {
    param($url, $expectedStatus = 200, $description = "")
    try {
        $response = Invoke-WebRequest -Uri $url -UseBasicParsing -TimeoutSec 10 -ErrorAction Stop
        if ($response.StatusCode -eq $expectedStatus) {
            Write-Host "  [OK] $description - Status: $($response.StatusCode)" -ForegroundColor $successColor
            return $response
        } else {
            Write-Host "  [FAIL] $description - Očekivano: $expectedStatus, Dobijeno: $($response.StatusCode)" -ForegroundColor $errorColor
            return $null
        }
    } catch {
        Write-Host "  [FAIL] $description - Greška: $($_.Exception.Message)" -ForegroundColor $errorColor
        return $null
    }
}

Write-Host "1. TEST API COMPOSITION - GET ALL SONGS" -ForegroundColor $testColor
Write-Host "----------------------------------------" -ForegroundColor $testColor
Write-Host ""

Write-Host "1.1. Testiranje GET /api/content/songs..." -ForegroundColor $infoColor
$response = Test-Endpoint "http://localhost:8081/api/content/songs" 200 "Get all songs with ratings"

if ($response) {
    try {
        $songs = $response.Content | ConvertFrom-Json
        
        if ($songs.Count -gt 0) {
            Write-Host "  Pronađeno $($songs.Count) pesama" -ForegroundColor $infoColor
            
            # Proveri da li svaka pesma ima averageRating i ratingCount
            $allHaveRatings = $true
            foreach ($song in $songs) {
                if (-not $song.PSObject.Properties.Name -contains "averageRating") {
                    Write-Host "  [FAIL] Pesma $($song.id) nema averageRating" -ForegroundColor $errorColor
                    $allHaveRatings = $false
                }
                if (-not $song.PSObject.Properties.Name -contains "ratingCount") {
                    Write-Host "  [FAIL] Pesma $($song.id) nema ratingCount" -ForegroundColor $errorColor
                    $allHaveRatings = $false
                }
            }
            
            if ($allHaveRatings) {
                Write-Host "  [OK] Sve pesme imaju averageRating i ratingCount" -ForegroundColor $successColor
                
                # Prikaži primer
                $firstSong = $songs[0]
                Write-Host ""
                Write-Host "  Primer pesme:" -ForegroundColor $infoColor
                Write-Host "    ID: $($firstSong.id)" -ForegroundColor White
                Write-Host "    Name: $($firstSong.name)" -ForegroundColor White
                Write-Host "    Average Rating: $($firstSong.averageRating)" -ForegroundColor White
                Write-Host "    Rating Count: $($firstSong.ratingCount)" -ForegroundColor White
            }
        } else {
            Write-Host "  [WARN] Nema pesama u bazi" -ForegroundColor $infoColor
        }
    } catch {
        Write-Host "  [FAIL] Greška pri parsiranju JSON: $($_.Exception.Message)" -ForegroundColor $errorColor
    }
}

Write-Host ""
Write-Host "2. TEST API COMPOSITION - GET SONGS BY ALBUM" -ForegroundColor $testColor
Write-Host "---------------------------------------------" -ForegroundColor $testColor
Write-Host ""

Write-Host "2.1. Testiranje GET /api/content/songs/by-album..." -ForegroundColor $infoColor

# Prvo dobij album da imamo albumId
$albumsResponse = Test-Endpoint "http://localhost:8081/api/content/albums" 200 "Get albums"
if ($albumsResponse) {
    try {
        $albums = $albumsResponse.Content | ConvertFrom-Json
        if ($albums.Count -gt 0) {
            $albumId = $albums[0].id
            Write-Host "  Koristim albumId: $albumId" -ForegroundColor $infoColor
            
            $songsResponse = Test-Endpoint "http://localhost:8081/api/content/songs/by-album?albumId=$albumId" 200 "Get songs by album with ratings"
            
            if ($songsResponse) {
                try {
                    $songs = $songsResponse.Content | ConvertFrom-Json
                    if ($songs.Count -gt 0) {
                        Write-Host "  Pronađeno $($songs.Count) pesama u albumu" -ForegroundColor $infoColor
                        
                        $allHaveRatings = $true
                        foreach ($song in $songs) {
                            if (-not $song.PSObject.Properties.Name -contains "averageRating") {
                                $allHaveRatings = $false
                            }
                            if (-not $song.PSObject.Properties.Name -contains "ratingCount") {
                                $allHaveRatings = $false
                            }
                        }
                        
                        if ($allHaveRatings) {
                            Write-Host "  [OK] Sve pesme u albumu imaju averageRating i ratingCount" -ForegroundColor $successColor
                        } else {
                            Write-Host "  [FAIL] Neke pesme nemaju averageRating ili ratingCount" -ForegroundColor $errorColor
                        }
                    } else {
                        Write-Host "  [WARN] Album nema pesama" -ForegroundColor $infoColor
                    }
                } catch {
                    Write-Host "  [FAIL] Greška pri parsiranju JSON: $($_.Exception.Message)" -ForegroundColor $errorColor
                }
            }
        } else {
            Write-Host "  [WARN] Nema albuma u bazi" -ForegroundColor $infoColor
        }
    } catch {
        Write-Host "  [FAIL] Greška pri parsiranju albuma: $($_.Exception.Message)" -ForegroundColor $errorColor
    }
}

Write-Host ""
Write-Host "3. PROVERA RATINGS SERVICE ENDPOINT" -ForegroundColor $testColor
Write-Host "------------------------------------" -ForegroundColor $testColor
Write-Host ""

Write-Host "3.1. Testiranje GET /api/ratings/average-rating..." -ForegroundColor $infoColor

# Prvo dobij pesmu da imamo songId
$songsResponse = Test-Endpoint "http://localhost:8081/api/content/songs" 200 "Get songs"
if ($songsResponse) {
    try {
        $songs = $songsResponse.Content | ConvertFrom-Json
        if ($songs.Count -gt 0) {
            $songId = $songs[0].id
            Write-Host "  Koristim songId: $songId" -ForegroundColor $infoColor
            
            $ratingResponse = Test-Endpoint "http://localhost:8081/api/ratings/average-rating?songId=$songId" 200 "Get average rating"
            
            if ($ratingResponse) {
                try {
                    $rating = $ratingResponse.Content | ConvertFrom-Json
                    Write-Host "  [OK] Average rating endpoint radi" -ForegroundColor $successColor
                    Write-Host "    Average Rating: $($rating.averageRating)" -ForegroundColor White
                    Write-Host "    Rating Count: $($rating.ratingCount)" -ForegroundColor White
                } catch {
                    Write-Host "  [FAIL] Greška pri parsiranju JSON: $($_.Exception.Message)" -ForegroundColor $errorColor
                }
            }
        }
    } catch {
        Write-Host "  [FAIL] Greška: $($_.Exception.Message)" -ForegroundColor $errorColor
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  TESTIRANJE ZAVRŠENO" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Za testiranje u browseru:" -ForegroundColor $infoColor
Write-Host "  http://localhost:8081/api/content/songs" -ForegroundColor White
Write-Host "  http://localhost:8081/api/content/songs/by-album?albumId=album1" -ForegroundColor White
Write-Host ""
