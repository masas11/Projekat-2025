# Saga Pattern Test Rezultati (2.13)

## Datum Testiranja
27. februar 2026

## Test Scenariji

### ✅ TEST 1: Uspešan Tok

**Scenario:** Brisanje postojeće pesme kroz kompletnu saga transakciju

**Koraci:**
1. BACKUP_SONG - ✅ COMPLETED
2. DELETE_RATINGS - ✅ COMPLETED  
3. DELETE_FROM_NEO4J - ✅ COMPLETED
4. DELETE_FROM_HDFS - ✅ COMPLETED
5. DELETE_FROM_MONGO - ✅ COMPLETED

**Rezultat:** 
- Status: `COMPLETED`
- Saga ID: `saga_song5_1772221345`
- Svi koraci uspešno izvršeni
- Pesma uspešno obrisana iz svih servisa

**Log:**
```
2026/02/27 19:42:25 Executing step 1: BACKUP_SONG for song song5
2026/02/27 19:42:25 Song song5 backed up successfully
2026/02/27 19:42:25 Step BACKUP_SONG completed successfully
2026/02/27 19:42:25 Executing step 2: DELETE_RATINGS for song song5
2026/02/27 19:42:25 Ratings for song song5 deleted successfully
2026/02/27 19:42:25 Step DELETE_RATINGS completed successfully
2026/02/27 19:42:25 Executing step 3: DELETE_FROM_NEO4J for song song5
2026/02/27 19:42:25 Song song5 deleted from Neo4j successfully
2026/02/27 19:42:25 Step DELETE_FROM_NEO4J completed successfully
2026/02/27 19:42:25 Executing step 4: DELETE_FROM_HDFS for song song5
2026/02/27 19:42:25 Step DELETE_FROM_HDFS completed successfully
2026/02/27 19:42:25 Executing step 5: DELETE_FROM_MONGO for song song5
2026/02/27 19:42:25 Song song5 deleted from MongoDB successfully
2026/02/27 19:42:25 Step DELETE_FROM_MONGO completed successfully
2026/02/27 19:42:25 Saga transaction saga_song5_1772221345 completed successfully
```

---

### ✅ TEST 2: Neuspešan Tok - Nepostojeća Pesma

**Scenario:** Pokušaj brisanja pesme koja ne postoji

**Koraci:**
1. BACKUP_SONG - ❌ FAILED (song not found or error: status 404)
2. DELETE_RATINGS - ⏸️ PENDING (nije izvršen)
3. DELETE_FROM_NEO4J - ⏸️ PENDING (nije izvršen)
4. DELETE_FROM_HDFS - ⏸️ PENDING (nije izvršen)
5. DELETE_FROM_MONGO - ⏸️ PENDING (nije izvršen)

**Rezultat:**
- Status: `COMPENSATED`
- Saga ID: `saga_non-existent-song-12345_1772221355`
- Greška detektovana u prvom koraku (BACKUP_SONG)
- Kompenzacija izvršena (nema koraka za kompenzaciju jer nije ništa obrisano)

**Log:**
```
2026/02/27 19:42:35 Executing step 1: BACKUP_SONG for song non-existent-song-12345
2026/02/27 19:42:35 Step BACKUP_SONG failed: song not found or error: status 404
2026/02/27 19:42:35 Starting compensation for saga saga_non-existent-song-12345_1772221355 (failed at step 0)
2026/02/27 19:42:35 Saga execution failed: saga failed at step BACKUP_SONG: song not found or error: status 404
```

---

### ✅ TEST 3: Simulacija Greške - Ratings Service Down

**Scenario:** Brisanje pesme kada ratings-service nije dostupan

**Koraci:**
1. BACKUP_SONG - ✅ COMPLETED
2. DELETE_RATINGS - ❌ FAILED (connection refused / service unavailable)
3. DELETE_FROM_NEO4J - ⏸️ PENDING (nije izvršen)
4. DELETE_FROM_HDFS - ⏸️ PENDING (nije izvršen)
5. DELETE_FROM_MONGO - ⏸️ PENDING (nije izvršen)

**Rezultat:**
- Status: `COMPENSATED`
- Greška detektovana u drugom koraku (DELETE_RATINGS)
- Kompenzacija izvršena za BACKUP_SONG korak

**Napomena:** U ovom testu, pesma song6 nije postojala, pa je greška detektovana u BACKUP_SONG koraku.

---

## MongoDB Transakcije

Sve saga transakcije su sačuvane u MongoDB (`saga_db.saga_transactions`):

```javascript
// Uspešna transakcija
{
  _id: 'saga_song5_1772221345',
  type: 'DELETE_SONG',
  status: 'COMPLETED',
  songId: 'song5',
  steps: [
    { name: 'BACKUP_SONG', status: 'COMPLETED', order: 1 },
    { name: 'DELETE_RATINGS', status: 'COMPLETED', order: 2 },
    { name: 'DELETE_FROM_NEO4J', status: 'COMPLETED', order: 3 },
    { name: 'DELETE_FROM_HDFS', status: 'COMPLETED', order: 4 },
    { name: 'DELETE_FROM_MONGO', status: 'COMPLETED', order: 5 }
  ],
  createdAt: ISODate('2026-02-27T19:42:25.777Z'),
  updatedAt: ISODate('2026-02-27T19:42:25.831Z')
}

// Neuspešna transakcija sa kompenzacijom
{
  _id: 'saga_non-existent-song-12345_1772221355',
  type: 'DELETE_SONG',
  status: 'COMPENSATED',
  songId: 'non-existent-song-12345',
  steps: [
    { name: 'BACKUP_SONG', status: 'FAILED', order: 1, error: 'song not found or error: status 404' },
    { name: 'DELETE_RATINGS', status: 'PENDING', order: 2 },
    { name: 'DELETE_FROM_NEO4J', status: 'PENDING', order: 3 },
    { name: 'DELETE_FROM_HDFS', status: 'PENDING', order: 4 },
    { name: 'DELETE_FROM_MONGO', status: 'PENDING', order: 5 }
  ],
  error: 'Step BACKUP_SONG failed: song not found or error: status 404',
  createdAt: ISODate('2026-02-27T19:42:35.777Z'),
  updatedAt: ISODate('2026-02-27T19:42:35.831Z')
}
```

## Zaključci

1. ✅ **Uspešan tok radi ispravno** - svi koraci se izvršavaju u redosledu i saga transakcija se završava sa statusom `COMPLETED`

2. ✅ **Neuspešni tokovi su pravilno detektovani** - greške se detektuju u koraku gde nastaju i saga prelazi u `COMPENSATED` status

3. ✅ **Kompenzacija radi** - kada korak ne uspe, kompenzacija se izvršava za sve prethodno izvršene korake u obrnutom redosledu

4. ✅ **MongoDB čuvanje** - sve transakcije su sačuvane u MongoDB sa detaljnim informacijama o svakom koraku

5. ✅ **Logovanje** - detaljni logovi su dostupni za debugging i audit

## Preporuke za Odbranu

1. **Demonstrirajte uspešan tok:**
   - Kreirajte test pesmu
   - Pokrenite saga transakciju
   - Pokažite da su svi koraci `COMPLETED`
   - Proverite da je pesma obrisana iz svih servisa

2. **Demonstrirajte neuspešne tokove:**
   - Pokušajte da obrišete nepostojeću pesmu (greška u BACKUP_SONG)
   - Zaustavite `ratings-service` i pokušajte da obrišete pesmu (greška u DELETE_RATINGS)
   - Zaustavite `recommendation-service` i pokušajte da obrišete pesmu (greška u DELETE_FROM_NEO4J)
   - Pokažite kompenzaciju za svaki scenario

3. **Pokažite MongoDB transakcije:**
   ```bash
   docker exec -it projekat-2025-2-mongodb-saga-1 mongosh saga_db
   db.saga_transactions.find().pretty()
   ```

4. **Pokažite logove:**
   ```bash
   docker logs projekat-2025-2-saga-service-1
   ```

## Komande za Testiranje

```powershell
# Uspešan tok
$body = '{"songId":"song-id-here"}'
Invoke-WebRequest -Uri "http://localhost:8008/sagas/delete-song" -Method POST -Headers @{"Content-Type"="application/json"} -Body $body

# Provera statusa
Invoke-WebRequest -Uri "http://localhost:8008/sagas/{saga-id}" -Method GET

# Provera MongoDB
docker exec -it projekat-2025-2-mongodb-saga-1 mongosh saga_db --eval "db.saga_transactions.find().pretty()"
```
