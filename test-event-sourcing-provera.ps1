# KOMPLETNA PROVERA EVENT SOURCING (2.14)
# Proverava da li sve radi kako treba

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  PROVERA EVENT SOURCING (2.14)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$baseURL = "http://localhost:8081"
$analyticsURL = "http://localhost:8007"
$successColor = "Green"
$errorColor = "Red"
$infoColor = "Yellow"
$testColor = "Cyan"

# 1. Provera analytics-service
Write-Host "1. PROVERA ANALYTICS-SERVICE" -ForegroundColor $testColor
Write-Host "----------------------------" -ForegroundColor $testColor
try {
    $health = Invoke-WebRequest -Uri "$analyticsURL/health" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "   [OK] Analytics-service je pokrenut" -ForegroundColor $successColor
} catch {
    Write-Host "   [FAIL] Analytics-service nije dostupan: $($_.Exception.Message)" -ForegroundColor $errorColor
    Write-Host "   Pokreni: docker-compose up -d analytics-service" -ForegroundColor $infoColor
    exit 1
}

# 2. Provera Event Store u MongoDB
Write-Host ""
Write-Host "2. PROVERA EVENT STORE (MongoDB)" -ForegroundColor $testColor
Write-Host "--------------------------------" -ForegroundColor $testColor

Write-Host ""
Write-Host "2.1. Broj dogadjaja u Event Store..." -ForegroundColor $infoColor
try {
    $eventCount = docker exec projekat-2025-2-mongodb-analytics-1 mongosh analytics_db --quiet --eval "db.event_store.countDocuments()" 2>&1
    if ($eventCount -match "^\d+$") {
        $count = [int]$eventCount
        Write-Host "   [OK] Ukupno dogadjaja: $count" -ForegroundColor $successColor
        if ($count -eq 0) {
            Write-Host "   [INFO] Event Store je prazan - izvrsi neku aktivnost u frontendu" -ForegroundColor $infoColor
        }
    } else {
        Write-Host "   [FAIL] Nije moguce proveriti Event Store" -ForegroundColor $errorColor
    }
} catch {
    Write-Host "   [FAIL] Greska pri proveri Event Store: $($_.Exception.Message)" -ForegroundColor $errorColor
}

Write-Host ""
Write-Host "2.2. Tipovi dogadjaja..." -ForegroundColor $infoColor
try {
    $mongoQuery = 'db.event_store.aggregate([{$group: {_id: "$eventType", count: {$sum: 1}}}, {$sort: {count: -1}}]).forEach(function(doc) { print(doc._id + ": " + doc.count); });'
    $eventTypes = docker exec projekat-2025-2-mongodb-analytics-1 mongosh analytics_db --quiet --eval $mongoQuery 2>&1
    if ($eventTypes -and $eventTypes -notmatch "null" -and $eventTypes.Trim() -ne "") {
        Write-Host "   [OK] Događaji po tipu:" -ForegroundColor $successColor
        $eventTypes -split "`n" | Where-Object { $_ -ne "" -and $_ -notmatch "null" } | ForEach-Object {
            Write-Host "      - $_" -ForegroundColor Gray
        }
    } else {
        Write-Host "   [INFO] Nema dogadjaja u Event Store" -ForegroundColor $infoColor
    }
} catch {
    Write-Host "   [INFO] Nije moguce proveriti tipove dogadjaja" -ForegroundColor $infoColor
}

# 3. Provera da li se eventi loguju kada se oceni pesma
Write-Host ""
Write-Host "3. PROVERA AUTOMATSKOG LOGOVANJA" -ForegroundColor $testColor
Write-Host "--------------------------------" -ForegroundColor $testColor

Write-Host ""
Write-Host "3.1. Provera logova ratings-service..." -ForegroundColor $infoColor
try {
    $logs = docker logs projekat-2025-2-ratings-service-1 --tail 20 2>&1 | Select-String -Pattern "LogActivity|analytics" -CaseSensitive:$false
    if ($logs) {
        Write-Host "   [OK] Ratings-service poziva LogActivity" -ForegroundColor $successColor
        $logs | Select-Object -First 2 | ForEach-Object {
            Write-Host "      $_" -ForegroundColor Gray
        }
    } else {
        Write-Host "   [INFO] Nema nedavnih poziva ka analytics (ocekivano ako nije bilo ocenjivanja)" -ForegroundColor $infoColor
    }
} catch {
    Write-Host "   [INFO] Nije moguce proveriti logove" -ForegroundColor $infoColor
}

Write-Host ""
Write-Host "3.2. Provera logova analytics-service..." -ForegroundColor $infoColor
try {
    $logs = docker logs projekat-2025-2-analytics-service-1 --tail 20 2>&1 | Select-String -Pattern "Event|event|AppendEvent|RATING_GIVEN" -CaseSensitive:$false
    if ($logs) {
        Write-Host "   [OK] Analytics-service prima i cuva dogadjaje" -ForegroundColor $successColor
        $logs | Select-Object -First 2 | ForEach-Object {
            Write-Host "      $_" -ForegroundColor Gray
        }
    } else {
        Write-Host "   [INFO] Nema nedavnih dogadjaja u logovima" -ForegroundColor $infoColor
    }
} catch {
    Write-Host "   [INFO] Nije moguce proveriti logove" -ForegroundColor $infoColor
}

# 4. Provera Event Stream endpoint
Write-Host ""
Write-Host "4. PROVERA EVENT STREAM ENDPOINT" -ForegroundColor $testColor
Write-Host "--------------------------------" -ForegroundColor $testColor

Write-Host ""
Write-Host "4.1. Test event stream endpoint..." -ForegroundColor $infoColor
Write-Host "   Endpoint: GET $baseURL/api/analytics/events/stream?userId=<user-id>" -ForegroundColor Gray
Write-Host "   Napomena: Za testiranje, unesite user ID iz frontenda" -ForegroundColor Gray

# 5. Provera strukture Event Store
Write-Host ""
Write-Host "5. PROVERA STRUKTURE DOGADJAJA" -ForegroundColor $testColor
Write-Host "-----------------------------" -ForegroundColor $testColor

Write-Host ""
Write-Host "5.1. Primer dogadjaja (RATING_GIVEN)..." -ForegroundColor $infoColor
try {
    $mongoQuery2 = 'var event = db.event_store.findOne({eventType: "RATING_GIVEN"}); if (event) { print("Event Type: " + event.eventType); print("Stream ID (User ID): " + event.streamId); print("Version: " + event.version); print("Timestamp: " + event.timestamp); if (event.payload) { print("Payload: rating=" + event.payload.rating + ", songId=" + event.payload.songId); } } else { print("Nema RATING_GIVEN dogadjaja"); }'
    $sampleEvent = docker exec projekat-2025-2-mongodb-analytics-1 mongosh analytics_db --quiet --eval $mongoQuery2 2>&1
    if ($sampleEvent -and $sampleEvent -notmatch "Nema") {
        Write-Host "   [OK] Primer RATING_GIVEN dogadjaja:" -ForegroundColor $successColor
        $sampleEvent -split "`n" | Where-Object { $_ -ne "" } | ForEach-Object {
            Write-Host "      $_" -ForegroundColor Gray
        }
    } else {
        Write-Host "   [INFO] Nema RATING_GIVEN dogadjaja - oceni pesmu u frontendu pa pokreni ponovo" -ForegroundColor $infoColor
    }
} catch {
    Write-Host "   [INFO] Nije moguce prikazati primer dogadjaja" -ForegroundColor $infoColor
}

# 6. Provera da li frontend koristi activities endpoint
Write-Host ""
Write-Host "6. PROVERA CQRS (2.15)" -ForegroundColor $testColor
Write-Host "----------------------" -ForegroundColor $testColor

Write-Host ""
Write-Host "6.1. Provera Projection Store (Read Model)..." -ForegroundColor $infoColor
try {
    $projectionCount = docker exec projekat-2025-2-mongodb-analytics-1 mongosh analytics_db --quiet --eval "db.analytics_projections.countDocuments()" 2>&1
    if ($projectionCount -match "^\d+$") {
        $count = [int]$projectionCount
        Write-Host "   [OK] Ukupno projekcija (read model): $count" -ForegroundColor $successColor
        if ($count -gt 0) {
            Write-Host "   [OK] CQRS Read Model postoji i sadrzi podatke" -ForegroundColor $successColor
        } else {
            Write-Host "   [INFO] Nema projekcija (ocekivano ako nema aktivnosti)" -ForegroundColor $infoColor
        }
    }
} catch {
    Write-Host "   [INFO] Nije moguce proveriti Projection Store" -ForegroundColor $infoColor
}

Write-Host ""
Write-Host "6.2. Provera CQRS Command Handler..." -ForegroundColor $infoColor
try {
    $logs = docker logs projekat-2025-2-analytics-service-1 --tail 30 2>&1 | Select-String -Pattern "Command handled|CQRS" -CaseSensitive:$false
    if ($logs) {
        Write-Host "   [OK] CQRS Command Handler radi" -ForegroundColor $successColor
        $logs | Select-Object -First 1 | ForEach-Object {
            Write-Host "      $_" -ForegroundColor Gray
        }
    } else {
        Write-Host "   [INFO] Nema nedavnih CQRS komandi u logovima" -ForegroundColor $infoColor
    }
} catch {
    Write-Host "   [INFO] Nije moguce proveriti logove" -ForegroundColor $infoColor
}

Write-Host ""
Write-Host "6.3. Provera CQRS Query Handler (Analitike)..." -ForegroundColor $infoColor
Write-Host "   Endpoint: GET $baseURL/api/analytics/analytics?userId=<user-id>" -ForegroundColor Gray
Write-Host "   Query Handler cita iz Projection Store (read model)" -ForegroundColor Gray

Write-Host ""
Write-Host "7. PROVERA FRONTEND INTEGRACIJE" -ForegroundColor $testColor
Write-Host "------------------------------" -ForegroundColor $testColor

Write-Host ""
Write-Host "7.1. Provera activities endpoint..." -ForegroundColor $infoColor
try {
    # Test da li endpoint postoji
    $testResponse = Invoke-WebRequest -Uri "$baseURL/api/analytics/activities" -UseBasicParsing -TimeoutSec 5 -ErrorAction SilentlyContinue
    if ($testResponse.StatusCode -eq 200 -or $testResponse.StatusCode -eq 401) {
        Write-Host "   [OK] Activities endpoint je dostupan" -ForegroundColor $successColor
    }
} catch {
    if ($_.Exception.Response.StatusCode.value__ -eq 401) {
        Write-Host "   [OK] Activities endpoint zahteva autentifikaciju (ocekivano)" -ForegroundColor $successColor
    } else {
        Write-Host "   [INFO] Activities endpoint: $($_.Exception.Message)" -ForegroundColor $infoColor
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  ZAKLJUCAK" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Event Sourcing (2.14) + CQRS (2.15) radi na sledeci nacin:" -ForegroundColor $infoColor
Write-Host ""
Write-Host "1. Kada ocenis pesmu (CQRS Command Side):" -ForegroundColor White
Write-Host "   - Frontend salje POST ka /api/ratings/rate-song" -ForegroundColor Gray
Write-Host "   - Ratings-service poziva analytics.LogActivity()" -ForegroundColor Gray
Write-Host "   - CommandHandler kreira RateSongCommand i cuva dogadjaj u Event Store" -ForegroundColor Gray
Write-Host "   - EventHandler azurira Projection Store (read model)" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Kada trazis analitike (CQRS Query Side):" -ForegroundColor White
Write-Host "   - QueryHandler cita iz Projection Store (read model)" -ForegroundColor Gray
Write-Host "   - Vraca analitike bez citanja Event Store-a (brze)" -ForegroundColor Gray
Write-Host ""
Write-Host "3. Event Store (2.14):" -ForegroundColor White
Write-Host "   - Kolekcija: event_store" -ForegroundColor Gray
Write-Host "   - Polja: eventType, streamId, version, timestamp, payload" -ForegroundColor Gray
Write-Host ""
Write-Host "4. Projection Store (2.15 CQRS Read Model):" -ForegroundColor White
Write-Host "   - Kolekcija: analytics_projections" -ForegroundColor Gray
Write-Host "   - Cuva izracunate analitike (totalSongsPlayed, totalRatings, itd.)" -ForegroundColor Gray
Write-Host ""
Write-Host "5. Za proveru:" -ForegroundColor White
Write-Host "   - Profile -> Analytics (koristi CQRS Query Handler)" -ForegroundColor Gray
Write-Host "   - Profile -> Activity History (koristi Event Store)" -ForegroundColor Gray
Write-Host ""
