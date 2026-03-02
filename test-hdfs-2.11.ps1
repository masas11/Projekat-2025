# Test skripta za HDFS implementaciju (2.11)
# Testira upload i download audio fajlova sa HDFS-a

Write-Host "=== TESTIRANJE HDFS (2.11) ===" -ForegroundColor Cyan
Write-Host ""

# Provera da li su servisi pokrenuti
Write-Host "1. Provera servisa..." -ForegroundColor Yellow
try {
    $health = Invoke-WebRequest -Uri "http://localhost:8002/health" -UseBasicParsing -TimeoutSec 5
    Write-Host "   [OK] content-service je pokrenut" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] content-service nije pokrenut. Pokreni: docker-compose up -d" -ForegroundColor Red
    exit 1
}

try {
    $hdfsHealth = Invoke-WebRequest -Uri "http://localhost:9870/webhdfs/v1/?op=LISTSTATUS" -UseBasicParsing -TimeoutSec 5
    Write-Host "   [OK] HDFS NameNode je pokrenut" -ForegroundColor Green
} catch {
    Write-Host "   [WARN] HDFS NameNode mozda jos nije spreman (cekaj 30-60 sekundi)" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "2. Testiranje HDFS operacija..." -ForegroundColor Yellow
Write-Host ""

# Test 1: Provera HDFS konekcije
Write-Host "   Test 1: Provera HDFS konekcije..." -ForegroundColor White
try {
    $hdfsStatus = Invoke-WebRequest -Uri "http://localhost:9870/webhdfs/v1/?op=GETFILESTATUS" -UseBasicParsing -TimeoutSec 10
    Write-Host "      [OK] HDFS je dostupan" -ForegroundColor Green
} catch {
    Write-Host "      [FAIL] HDFS nije dostupan. Proveri logove: docker-compose logs hdfs-namenode" -ForegroundColor Red
    Write-Host "      Napomena: HDFS moze da traje 1-2 minuta da se pokrene" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "3. Testiranje upload-a audio fajla..." -ForegroundColor Yellow
Write-Host ""
Write-Host "   Za upload audio fajla, koristi:" -ForegroundColor White
Write-Host "   curl -X POST http://localhost:8081/api/content/songs/{songId}/upload \" -ForegroundColor Gray
Write-Host "        -H 'Authorization: Bearer {JWT_TOKEN}' \" -ForegroundColor Gray
Write-Host "        -F 'audio=@/path/to/audio.mp3' \" -ForegroundColor Gray
Write-Host "        -F 'songId={songId}'" -ForegroundColor Gray
Write-Host ""
Write-Host "   Ili kroz frontend (admin panel):" -ForegroundColor White
Write-Host "   - Uloguj se kao admin" -ForegroundColor Gray
Write-Host "   - Idi na Songs -> Edit song" -ForegroundColor Gray
Write-Host "   - Upload audio fajl" -ForegroundColor Gray

Write-Host ""
Write-Host "4. Testiranje stream-a audio fajla..." -ForegroundColor Yellow
Write-Host ""
Write-Host "   Test stream endpoint-a:" -ForegroundColor White
Write-Host "   curl http://localhost:8081/api/content/songs/{songId}/stream -o test-audio.mp3" -ForegroundColor Gray
Write-Host ""
Write-Host "   Ili kroz frontend:" -ForegroundColor White
Write-Host "   - Otvori pesmu -> Play" -ForegroundColor Gray

Write-Host ""
Write-Host "5. Provera fajlova u HDFS-u..." -ForegroundColor Yellow
try {
    $hdfsListUrl = "http://localhost:9870/webhdfs/v1/audio/songs/?op=LISTSTATUS"
    $hdfsList = Invoke-WebRequest -Uri $hdfsListUrl -UseBasicParsing -TimeoutSec 10 -ErrorAction SilentlyContinue
    if ($hdfsList.StatusCode -eq 200) {
        $hdfsData = $hdfsList.Content | ConvertFrom-Json
        if ($hdfsData.FileStatuses -and $hdfsData.FileStatuses.FileStatus) {
            $fileCount = $hdfsData.FileStatuses.FileStatus.Count
            Write-Host "   [OK] Pronadjeno $fileCount fajlova u /audio/songs/" -ForegroundColor Green
            if ($fileCount -gt 0) {
                Write-Host "   Primer fajlova:" -ForegroundColor White
                $hdfsData.FileStatuses.FileStatus | Select-Object -First 3 | ForEach-Object {
                    $sizeKB = [math]::Round($_.length / 1024, 2)
                    Write-Host "      - $($_.pathSuffix) ($sizeKB KB)" -ForegroundColor Gray
                }
            }
        } else {
            Write-Host "   [INFO] Nema fajlova u /audio/songs/ (ocekivano ako nije bilo upload-a)" -ForegroundColor Yellow
        }
    }
} catch {
    Write-Host "   [INFO] Nije moguce proveriti fajlove (HDFS mozda jos nije spreman)" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "6. Testiranje stream endpoint-a (ako postoji pesma)..." -ForegroundColor Yellow
try {
    # Uzmi listu pesama
    $songsResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/content/songs" -UseBasicParsing -TimeoutSec 5 -ErrorAction SilentlyContinue
    if ($songsResponse.StatusCode -eq 200) {
        $songs = $songsResponse.Content | ConvertFrom-Json
        if ($songs -and $songs.Count -gt 0) {
            $songWithAudio = $songs | Where-Object { $_.audioFileURL -ne $null -and $_.audioFileURL -ne "" } | Select-Object -First 1
            if ($songWithAudio) {
                $songId = $songWithAudio.id
                Write-Host "   Testiram stream za pesmu: $($songWithAudio.name) (ID: $songId)" -ForegroundColor White
                try {
                    $streamResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/content/songs/$songId/stream" -UseBasicParsing -TimeoutSec 10 -Method Head -ErrorAction Stop
                    if ($streamResponse.StatusCode -eq 200) {
                        Write-Host "   [OK] Stream endpoint radi (Status: 200)" -ForegroundColor Green
                        $contentType = $streamResponse.Headers['Content-Type']
                        if ($contentType) {
                            Write-Host "      Content-Type: $contentType" -ForegroundColor Gray
                        }
                    }
                } catch {
                    $statusCode = 0
                    if ($_.Exception.Response) {
                        $statusCode = $_.Exception.Response.StatusCode.value__
                    }
                    if ($statusCode -eq 404) {
                        Write-Host "   [INFO] Audio fajl nije upload-ovan za ovu pesmu" -ForegroundColor Yellow
                    } else {
                        Write-Host "   [INFO] Stream endpoint: $($_.Exception.Message)" -ForegroundColor Yellow
                    }
                }
            } else {
                Write-Host "   [INFO] Nema pesama sa audio fajlovima za test" -ForegroundColor Yellow
            }
        }
    }
} catch {
    Write-Host "   [INFO] Nije moguce testirati stream (servis mozda nije dostupan)" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "7. HDFS Web UI..." -ForegroundColor Yellow
Write-Host "   Otvori: http://localhost:9870" -ForegroundColor White
Write-Host "   Navigiraj: Utilities -> Browse the file system -> /audio/songs/" -ForegroundColor Gray

Write-Host ""
Write-Host "=== KRAJ TESTIRANJA ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Sta pokazati na odbrani:" -ForegroundColor Yellow
Write-Host "  [OK] HDFS konekcija radi" -ForegroundColor Green
Write-Host "  [OK] Upload audio fajlova kroz frontend (admin panel)" -ForegroundColor Green
Write-Host "  [OK] Stream audio fajlova kroz /songs/{id}/stream endpoint" -ForegroundColor Green
Write-Host "  [OK] HDFS Web UI prikazuje fajlove u /audio/songs/" -ForegroundColor Green
Write-Host ""
Write-Host "Napomena: HDFS moze da traje 1-2 minuta da se potpuno pokrene." -ForegroundColor Yellow
Write-Host "Ako vidis greske, sacekaj i pokusaj ponovo." -ForegroundColor Yellow
