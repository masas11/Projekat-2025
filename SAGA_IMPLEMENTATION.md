# Saga Pattern Implementation (2.13)

## Pregled

Saga pattern je implementiran za brisanje pesama kako bi se osigurala konzistentnost podataka kroz više mikroservisa. Saga orchestrator koordinira sve korake i izvršava kompenzaciju u slučaju greške.

## Arhitektura

### Saga Service
- **Port**: 8008
- **Baza**: MongoDB (`mongodb-saga:27017`, database: `saga_db`)
- **Pattern**: Orchestrator pattern

### Koraci za brisanje pesme

1. **BACKUP_SONG** - Snima podatke pesme pre brisanja
2. **DELETE_RATINGS** - Briše sve ocene za pesmu iz `ratings-service`
3. **DELETE_FROM_NEO4J** - Briše pesmu iz Neo4j grafa (`recommendation-service`)
4. **DELETE_FROM_HDFS** - Briše audio fajl iz HDFS (ako postoji)
5. **DELETE_FROM_MONGO** - Briše pesmu iz MongoDB (`content-service`)

### Kompenzacione akcije

Ako bilo koji korak ne uspe, kompenzacija se izvršava u **obrnutom redosledu**:

- **DELETE_FROM_MONGO** → `RESTORE_TO_MONGO` (vraća pesmu u MongoDB)
- **DELETE_FROM_HDFS** → `RESTORE_TO_HDFS` (vraća audio fajl)
- **DELETE_FROM_NEO4J** → `RESTORE_TO_NEO4J` (vraća pesmu u Neo4j)
- **DELETE_RATINGS** → Nema kompenzacije (ocene su već obrisane, ali to je OK)
- **BACKUP_SONG** → Nema kompenzacije (nije ništa obrisano)

## API Endpoints

### 1. Pokretanje Saga Transakcije
```http
POST /api/sagas/delete-song
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "songId": "song-id-here"
}
```

**Response (Success - 200 OK):**
```json
{
  "id": "saga_song-id_timestamp",
  "type": "DELETE_SONG",
  "status": "COMPLETED",
  "songId": "song-id-here",
  "songData": { ... },
  "steps": [
    { "name": "BACKUP_SONG", "status": "COMPLETED", "order": 1 },
    { "name": "DELETE_RATINGS", "status": "COMPLETED", "order": 2 },
    { "name": "DELETE_FROM_NEO4J", "status": "COMPLETED", "order": 3 },
    { "name": "DELETE_FROM_HDFS", "status": "COMPLETED", "order": 4 },
    { "name": "DELETE_FROM_MONGO", "status": "COMPLETED", "order": 5 }
  ]
}
```

**Response (Failure - 500 Internal Server Error):**
```json
{
  "error": "Step DELETE_FROM_NEO4J failed: connection refused",
  "saga": {
    "id": "saga_song-id_timestamp",
    "status": "COMPENSATED",
    "steps": [
      { "name": "BACKUP_SONG", "status": "COMPLETED" },
      { "name": "DELETE_RATINGS", "status": "COMPLETED" },
      { "name": "DELETE_FROM_NEO4J", "status": "FAILED", "error": "..." },
      { "name": "DELETE_FROM_HDFS", "status": "PENDING" },
      { "name": "DELETE_FROM_MONGO", "status": "PENDING" }
    ]
  }
}
```

### 2. Provera Statusa Saga Transakcije
```http
GET /api/sagas/{sagaId}
Authorization: Bearer <admin-token>
```

## Test Scenariji

### Uspešan Tok
1. Kreiraj test pesmu
2. Pozovi `POST /api/sagas/delete-song` sa `songId`
3. Očekivani rezultat: `status: "COMPLETED"`, svi koraci `COMPLETED`

### Neuspešni Tokovi

#### Scenario 1: Pesma ne postoji
- **Korak koji ne uspe**: `BACKUP_SONG`
- **Kompenzacija**: Nema (nije ništa obrisano)
- **Očekivani status**: `COMPENSATED` (ili `FAILED` ako nema koraka za kompenzaciju)

#### Scenario 2: Ratings Service nedostupan
- **Korak koji ne uspe**: `DELETE_RATINGS`
- **Kompenzacija**: Nema (nije ništa obrisano)
- **Test**: Zaustavi `ratings-service` container pre brisanja

#### Scenario 3: Neo4j Service nedostupan
- **Korak koji ne uspe**: `DELETE_FROM_NEO4J`
- **Kompenzacija**: Nema (ocene su već obrisane, ali to je OK)
- **Test**: Zaustavi `recommendation-service` container pre brisanja

#### Scenario 4: MongoDB Service nedostupan
- **Korak koji ne uspe**: `DELETE_FROM_MONGO`
- **Kompenzacija**: 
  - `RESTORE_TO_NEO4J` (vraća pesmu u Neo4j)
  - `RESTORE_TO_HDFS` (vraća audio fajl)
- **Test**: Zaustavi `content-service` container pre brisanja

## Integracija

### Content Service
- `DeleteSong` endpoint sada poziva `saga-service` umesto direktnog brisanja
- Fallback na staru implementaciju ako `saga-service` nije dostupan
- Interni endpoint `/songs/internal/delete` za direktno brisanje iz MongoDB (koristi ga saga-service)

### API Gateway
- `POST /api/sagas/delete-song` - pokretanje saga transakcije (admin only)
- `GET /api/sagas/{id}` - provera statusa saga transakcije

## Testiranje

### PowerShell Skripte

1. **test-saga-success.ps1** - Testira uspešan tok
   ```powershell
   .\test-saga-success.ps1
   ```

2. **test-saga-failure.ps1** - Testira neuspešne tokove
   ```powershell
   .\test-saga-failure.ps1
   ```

### Manualno Testiranje

1. **Uspešan tok**:
   ```bash
   curl -X POST http://localhost:8081/api/sagas/delete-song \
     -H "Authorization: Bearer <token>" \
     -H "Content-Type: application/json" \
     -d '{"songId": "song-id"}'
   ```

2. **Provera statusa**:
   ```bash
   curl http://localhost:8081/api/sagas/{saga-id} \
     -H "Authorization: Bearer <token>"
   ```

3. **Simulacija greške**:
   - Zaustavi `ratings-service`: `docker compose stop ratings-service`
   - Pokušaj da obrišeš pesmu
   - Proveri status saga transakcije
   - Restartuj servis: `docker compose start ratings-service`

## Logovi

Saga transakcije se čuvaju u MongoDB (`saga_db.saga_transactions`) sa sledećim informacijama:
- ID transakcije
- Status (PENDING, IN_PROGRESS, COMPLETED, FAILED, COMPENSATING, COMPENSATED)
- Detalji svakog koraka (status, vreme izvršavanja, greške)
- Backup podataka pesme (za kompenzaciju)

## Napomene

- HDFS brisanje je trenutno samo logovano (nije implementirano)
- Kompenzacija za Neo4j vraćanje je trenutno samo logovana (nije potpuno implementirana)
- Saga transakcije se čuvaju u MongoDB za audit i debugging
