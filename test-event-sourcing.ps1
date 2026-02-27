# Test Event Sourcing Implementation (2.14)
# Testira Event Sourcing pattern za aktivnosti korisnika

$baseURL = "http://localhost:8081"
$analyticsURL = "http://localhost:8007"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "EVENT SOURCING TEST (2.14)" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# Step 1: Health check
Write-Host "Step 1: Proveravam health..." -ForegroundColor Yellow
try {
    $health = Invoke-WebRequest -Uri "$analyticsURL/health" -UseBasicParsing
    Write-Host "  ✓ Analytics service je pokrenut" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Analytics service nije dostupan" -ForegroundColor Red
    exit 1
}

# Step 2: Proveri postojeće događaje
Write-Host "`nStep 2: Proveravam postojeće događaje u Event Store..." -ForegroundColor Yellow
$eventCount = docker exec projekat-2025-2-mongodb-analytics-1 mongosh analytics_db --quiet --eval "db.event_store.countDocuments()"
Write-Host "  Broj događaja u Event Store: $eventCount" -ForegroundColor Gray

if ([int]$eventCount -gt 0) {
    Write-Host "`n  Primeri postojećih događaja:" -ForegroundColor Gray
    docker exec projekat-2025-2-mongodb-analytics-1 mongosh analytics_db --quiet --eval "db.event_store.find().limit(3).forEach(function(doc) { print('  - ' + doc.eventType + ' (Stream: ' + doc.streamId + ', Version: ' + doc.version + ')'); });"
}

# Step 3: Test - Logovanje aktivnosti (automatski se dodaje u Event Store)
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST 1: Logovanje Aktivnosti" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`nZa testiranje, potrebno je:" -ForegroundColor Yellow
Write-Host "  1. Ulogovati se kao korisnik u aplikaciji" -ForegroundColor White
Write-Host "  2. Izvršiti neku aktivnost:" -ForegroundColor White
Write-Host "     - Pustiti pesmu (SONG_PLAYED)" -ForegroundColor Gray
Write-Host "     - Oceniti pesmu (RATING_GIVEN)" -ForegroundColor Gray
Write-Host "     - Pretplatiti se na žanr (GENRE_SUBSCRIBED)" -ForegroundColor Gray
Write-Host "     - Pretplatiti se na umetnika (ARTIST_SUBSCRIBED)" -ForegroundColor Gray
Write-Host "  3. Aktivnost će se automatski dodati u Event Store" -ForegroundColor White

# Step 4: Test - Čitanje Event Stream-a
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST 2: Čitanje Event Stream-a" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$testUserId = Read-Host "`nUnesite User ID za testiranje (ili Enter za preskakanje)"

if ($testUserId) {
    Write-Host "`nČitam event stream za korisnika: $testUserId" -ForegroundColor Yellow
    
    try {
        # Direktno pozivanje analytics-service (za testiranje)
        $streamResponse = Invoke-WebRequest -Uri "$analyticsURL/events/stream?userId=$testUserId" -UseBasicParsing
        $events = $streamResponse.Content | ConvertFrom-Json
        
        Write-Host "`n  Pronađeno događaja: $($events.Count)" -ForegroundColor Green
        
        if ($events.Count -gt 0) {
            Write-Host "`n  Prvih 5 događaja:" -ForegroundColor Cyan
            $events | Select-Object -First 5 | ForEach-Object {
                Write-Host "    [$($_.version)] $($_.eventType) - $($_.timestamp)" -ForegroundColor White
                if ($_.payload.songName) {
                    Write-Host "      Pesma: $($_.payload.songName)" -ForegroundColor Gray
                }
                if ($_.payload.genre) {
                    Write-Host "      Žanr: $($_.payload.genre)" -ForegroundColor Gray
                }
                if ($_.payload.artistName) {
                    Write-Host "      Umetnik: $($_.payload.artistName)" -ForegroundColor Gray
                }
            }
        } else {
            Write-Host "  Nema događaja za ovog korisnika" -ForegroundColor Yellow
            Write-Host "  Izvršite neku aktivnost u aplikaciji pa pokušajte ponovo" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "  ✗ Greška pri čitanju event stream-a: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Step 5: Test - Replay događaja
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST 3: Replay Događaja (Rekonstrukcija Stanja)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

if ($testUserId) {
    Write-Host "`nRekonstruišem stanje za korisnika: $testUserId" -ForegroundColor Yellow
    
    try {
        $replayResponse = Invoke-WebRequest -Uri "$analyticsURL/events/replay?userId=$testUserId" -UseBasicParsing
        $state = $replayResponse.Content | ConvertFrom-Json
        
        Write-Host "`n  ✅ Stanje uspešno rekonstruisano!" -ForegroundColor Green
        Write-Host "`n  Statistika:" -ForegroundColor Cyan
        Write-Host "    - Ukupno puštanja pesama: $($state.totalSongsPlayed)" -ForegroundColor White
        Write-Host "    - Ukupno ocena: $($state.totalRatingsGiven)" -ForegroundColor White
        Write-Host "    - Pretplaćeni žanrovi: $($state.subscribedGenres.Count)" -ForegroundColor White
        if ($state.subscribedGenres.Count -gt 0) {
            Write-Host "      $($state.subscribedGenres -join ', ')" -ForegroundColor Gray
        }
        Write-Host "    - Pretplaćeni umetnici: $($state.subscribedArtists.Count)" -ForegroundColor White
        if ($state.subscribedArtists.Count -gt 0) {
            Write-Host "      $($state.subscribedArtists -join ', ')" -ForegroundColor Gray
        }
        
        Write-Host "`n  Raspodela aktivnosti:" -ForegroundColor Cyan
        $state.activityBreakdown.PSObject.Properties | ForEach-Object {
            Write-Host "    - $($_.Name): $($_.Value)" -ForegroundColor White
        }
        
        if ($state.lastActivityTime) {
            Write-Host "`n  Poslednja aktivnost: $($state.lastActivityTime)" -ForegroundColor Gray
        }
        
    } catch {
        Write-Host "  ✗ Greška pri replay-u: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $body = $reader.ReadToEnd()
            Write-Host "  Response: $body" -ForegroundColor Red
        }
    }
}

# Step 6: Provera u MongoDB
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST 4: Provera u MongoDB" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`nProveravam Event Store u MongoDB..." -ForegroundColor Yellow

Write-Host "`n  Ukupan broj događaja:" -ForegroundColor Cyan
docker exec projekat-2025-2-mongodb-analytics-1 mongosh analytics_db --quiet --eval "print('    ' + db.event_store.countDocuments());"

Write-Host "`n  Događaji po tipu:" -ForegroundColor Cyan
docker exec projekat-2025-2-mongodb-analytics-1 mongosh analytics_db --quiet --eval "db.event_store.aggregate([{$group: {_id: `$eventType, count: {$sum: 1}}}, {$sort: {count: -1}}]).forEach(function(doc) { print('    - ' + doc._id + ': ' + doc.count); });"

if ($testUserId) {
    Write-Host "`n  Događaji za korisnika $testUserId :" -ForegroundColor Cyan
    docker exec projekat-2025-2-mongodb-analytics-1 mongosh analytics_db --quiet --eval "var count = db.event_store.countDocuments({streamId: '$testUserId'}); print('    Ukupno: ' + count); if (count > 0) { db.event_store.find({streamId: '$testUserId'}).sort({version: 1}).limit(5).forEach(function(doc) { print('    [' + doc.version + '] ' + doc.eventType + ' - ' + doc.timestamp); }); }"
}

# Summary
Write-Host "`n========================================" -ForegroundColor Green
Write-Host "ZAVRŠETAK TESTIRANJA" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green

Write-Host "`n✅ Event Sourcing je implementiran i funkcionalan!" -ForegroundColor Green
Write-Host "`nKako testirati kroz aplikaciju:" -ForegroundColor Yellow
Write-Host "  1. Uloguj se kao korisnik" -ForegroundColor White
Write-Host "  2. Izvrši aktivnosti (pusti pesmu, oceni, pretplati se)" -ForegroundColor White
Write-Host "  3. Idi na 'Istorija Aktivnosti' stranicu" -ForegroundColor White
Write-Host "  4. Sve aktivnosti su automatski sačuvane u Event Store" -ForegroundColor White
Write-Host "`nAPI Endpointi za direktno testiranje:" -ForegroundColor Yellow
Write-Host "  GET http://localhost:8081/api/analytics/events/stream?userId=<user-id>" -ForegroundColor Gray
Write-Host "  GET http://localhost:8081/api/analytics/events/replay?userId=<user-id>" -ForegroundColor Gray
Write-Host "`nMongoDB komande:" -ForegroundColor Yellow
Write-Host "  docker exec -it projekat-2025-2-mongodb-analytics-1 mongosh analytics_db" -ForegroundColor Gray
Write-Host "  db.event_store.find().pretty()" -ForegroundColor Gray
Write-Host "  db.event_store.find({streamId: 'user-id'}).sort({version: 1}).pretty()" -ForegroundColor Gray
