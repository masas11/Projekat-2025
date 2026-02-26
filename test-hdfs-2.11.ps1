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
Write-Host "5. Provera HDFS Web UI..." -ForegroundColor Yellow
Write-Host "   Otvori: http://localhost:9870" -ForegroundColor White
Write-Host "   Proveri da li postoje fajlovi u /audio/songs/" -ForegroundColor Gray

Write-Host ""
Write-Host "=== KRAJ TESTIRANJA ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Napomena: HDFS moze da traje 1-2 minuta da se potpuno pokrene." -ForegroundColor Yellow
Write-Host "Ako vidis greske, sacekaj i pokusaj ponovo." -ForegroundColor Yellow
