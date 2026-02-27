# Kako Testirati Event Sourcing (2.14)

## Metoda 1: Kroz Aplikaciju (Najlakše)

### Koraci:

1. **Pokreni aplikaciju**
   ```bash
   docker compose up -d
   ```

2. **Uloguj se kao korisnik**
   - Otvori browser: `http://localhost:3000`
   - Uloguj se sa svojim nalogom

3. **Izvrši aktivnosti:**
   - **Pusti pesmu** → automatski se loguje `SONG_PLAYED` događaj
   - **Oceni pesmu** → automatski se loguje `RATING_GIVEN` događaj
   - **Pretplati se na žanr** → automatski se loguje `GENRE_SUBSCRIBED` događaj
   - **Pretplati se na umetnika** → automatski se loguje `ARTIST_SUBSCRIBED` događaj

4. **Proveri Istoriju Aktivnosti**
   - Idi na "Istorija Aktivnosti" stranicu
   - Sve aktivnosti su automatski sačuvane u Event Store

---

## Metoda 2: Direktno API Pozivi

### 1. Proveri Health

```powershell
Invoke-WebRequest -Uri "http://localhost:8007/health" -UseBasicParsing
```

### 2. Proveri Event Store u MongoDB

```powershell
# Broj događaja
docker exec projekat-2025-2-mongodb-analytics-1 mongosh analytics_db --quiet --eval "db.event_store.countDocuments()"

# Svi događaji
docker exec projekat-2025-2-mongodb-analytics-1 mongosh analytics_db --quiet --eval "db.event_store.find().pretty()"

# Događaji po tipu
docker exec projekat-2025-2-mongodb-analytics-1 mongosh analytics_db --quiet --eval "db.event_store.aggregate([{$group: {_id: `$eventType, count: {$sum: 1}}}]).forEach(function(doc) { print(doc._id + ': ' + doc.count); });"
```

### 3. Čitaj Event Stream za Korisnika

**Preko API Gateway (zahteva autentifikaciju):**
```powershell
# Treba JWT token
$token = "your-jwt-token"
$userId = "user-id-here"

Invoke-WebRequest -Uri "http://localhost:8081/api/analytics/events/stream?userId=$userId" `
    -Headers @{"Authorization" = "Bearer $token"} `
    -UseBasicParsing
```

**Direktno analytics-service (za testiranje):**
```powershell
$userId = "user-id-here"

Invoke-WebRequest -Uri "http://localhost:8007/events/stream?userId=$userId" `
    -UseBasicParsing
```

### 4. Replay Događaja (Rekonstrukcija Stanja)

**Preko API Gateway:**
```powershell
$token = "your-jwt-token"
$userId = "user-id-here"

Invoke-WebRequest -Uri "http://localhost:8081/api/analytics/events/replay?userId=$userId" `
    -Headers @{"Authorization" = "Bearer $token"} `
    -UseBasicParsing
```

**Direktno analytics-service:**
```powershell
$userId = "user-id-here"

$response = Invoke-WebRequest -Uri "http://localhost:8007/events/replay?userId=$userId" -UseBasicParsing
$state = $response.Content | ConvertFrom-Json

Write-Host "Total Songs Played: $($state.totalSongsPlayed)"
Write-Host "Total Ratings: $($state.totalRatingsGiven)"
Write-Host "Subscribed Genres: $($state.subscribedGenres -join ', ')"
Write-Host "Subscribed Artists: $($state.subscribedArtists -join ', ')"
```

---

## Metoda 3: cURL Komande

### 1. Event Stream

```bash
curl "http://localhost:8007/events/stream?userId=user123"
```

### 2. Replay

```bash
curl "http://localhost:8007/events/replay?userId=user123"
```

---

## Metoda 4: MongoDB Direktno

### Pristup MongoDB Shell-u

```bash
docker exec -it projekat-2025-2-mongodb-analytics-1 mongosh analytics_db
```

### Komande u MongoDB:

```javascript
// Broj događaja
db.event_store.countDocuments()

// Svi događaji
db.event_store.find().pretty()

// Događaji za određenog korisnika (sortirano po version-u)
db.event_store.find({streamId: "user-id-here"}).sort({version: 1}).pretty()

// Događaji po tipu
db.event_store.aggregate([
  {$group: {_id: "$eventType", count: {$sum: 1}}},
  {$sort: {count: -1}}
])

// Poslednji događaj za korisnika
db.event_store.find({streamId: "user-id-here"}).sort({version: -1}).limit(1).pretty()

// Događaji u određenom vremenskom periodu
db.event_store.find({
  streamId: "user-id-here",
  timestamp: {
    $gte: ISODate("2026-02-27T00:00:00Z"),
    $lte: ISODate("2026-02-27T23:59:59Z")
  }
}).sort({version: 1}).pretty()
```

---

## Šta Proveriti

### ✅ Event Store je Append-Only
- Događaji se samo dodaju, nikada ne menjaju ili brišu
- Svaki događaj ima jedinstveni version u stream-u

### ✅ Event Stream
- Svaki korisnik ima svoj event stream
- Događaji su sekvencijalno numerisani (version: 1, 2, 3, ...)

### ✅ State Reconstruction
- Replay vraća agregirano stanje
- Stanje se rekonstruiše primenom svih događaja u redosledu

### ✅ Backward Compatibility
- Postojeći `/activities` endpoint i dalje radi
- Aktivnosti se čuvaju i u ActivityStore i u EventStore

---

## Primer Test Scenarija

1. **Kreiraj korisnika** (ako ne postoji)
2. **Uloguj se** kao taj korisnik
3. **Pusti pesmu** → proveri Event Store
4. **Oceni pesmu** → proveri Event Store
5. **Pretplati se na žanr** → proveri Event Store
6. **Pretplati se na umetnika** → proveri Event Store
7. **Pozovi `/events/stream`** → vidi sve događaje
8. **Pozovi `/events/replay`** → vidi rekonstruisano stanje

---

## Očekivani Rezultati

### Event Stream Response:
```json
[
  {
    "id": "...",
    "eventId": "...",
    "eventType": "SONG_PLAYED",
    "streamId": "user123",
    "version": 1,
    "timestamp": "2026-02-27T19:00:00Z",
    "payload": {
      "songId": "song1",
      "songName": "Song Name"
    }
  },
  {
    "eventType": "RATING_GIVEN",
    "version": 2,
    ...
  }
]
```

### Replay Response:
```json
{
  "userId": "user123",
  "totalSongsPlayed": 5,
  "totalRatingsGiven": 3,
  "subscribedGenres": ["Pop", "Rock"],
  "subscribedArtists": ["artist1"],
  "activityBreakdown": {
    "SONG_PLAYED": 5,
    "RATING_GIVEN": 3,
    "GENRE_SUBSCRIBED": 2,
    "ARTIST_SUBSCRIBED": 1
  }
}
```

---

## Troubleshooting

### Problem: Nema događaja u Event Store
**Rešenje:** 
- Proveri da li su aktivnosti logovane kroz aplikaciju
- Proveri logove: `docker logs projekat-2025-2-analytics-service-1`

### Problem: Event Store nije inicijalizovan
**Rešenje:**
- Proveri da li je analytics-service pokrenut: `docker compose ps analytics-service`
- Proveri logove za greške

### Problem: Replay vraća prazno stanje
**Rešenje:**
- Proveri da li postoje događaji za tog korisnika u Event Store
- Proveri da li je `userId` tačan
