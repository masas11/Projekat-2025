# 📚 MongoDB Povezivanje - Korak po Korak Uputstvo

## 🎯 Šta je MongoDB i zašto ga koristimo?

MongoDB je **NoSQL baza podataka** (dokument-orijentisana). Za razliku od običnih baza (kao što je MySQL), MongoDB čuva podatke u formatu sličnom JSON-u, što je lakše za rad sa podacima.

U vašem projektu, MongoDB čuva:
- **Korisnike** (users-service) → baza: `users_db`
- **Muzički sadržaj** (content-service) → baza: `music_streaming` (artists, albums, songs)
- **Notifikacije** (notifications-service) → baza: `notifications_db`

---

## 🐳 METODA 1: MongoDB preko Docker-a (PREPORUČENO - NAJLAKŠE)

Ovo je **najlakši način** jer Docker automatski instalira i pokreće MongoDB za vas!

### Korak 1: Proverite da li imate Docker Desktop

1. Otvorite PowerShell ili Command Prompt
2. Ukucajte:
```bash
docker --version
```

Ako vidite verziju (npr. `Docker version 24.0.0`), imate Docker! ✅

Ako ne, **preuzmite Docker Desktop** sa: https://www.docker.com/products/docker-desktop/

### Korak 2: Pokrenite samo MongoDB

Idite u folder gde je vaš `docker-compose.yml` fajl i pokrenite:

```bash
cd Projekat-2025
docker-compose up mongodb
```

**Šta se dešava:**
- Docker preuzima MongoDB sliku (prvi put može potrajati)
- MongoDB se pokreće na portu 27017
- Podaci se čuvaju u Docker volumenu (ne gube se kada zatvorite)

**Kako znati da radi:**
Videćete poruku sličnu:
```
mongodb_1  | {"t":{"$date":"2025-01-XX..."},"s":"I",  "c":"NETWORK",  "id":23015,   "ctx":"listener","msg":"Listening on","attr":{"address":"0.0.0.0:27017"}}
```

### Korak 3: Povežite servise sa MongoDB

Vaš `docker-compose.yml` je **već konfigurisan**! Servisi se automatski povezuju kada pokrenete:

```bash
docker-compose up
```

**Kako to radi:**
- `mongodb` servis se pokreće prvi
- `users-service`, `content-service`, i `notifications-service` čekaju da MongoDB bude spreman
- Svaki servis dobija adresu: `mongodb://mongodb:27017`

---

## 💻 METODA 2: Lokalna instalacija MongoDB (Naprednije)

Ako ne želite da koristite Docker, možete instalirati MongoDB direktno na Windows.

### Korak 1: Preuzmite MongoDB

1. Idite na: https://www.mongodb.com/try/download/community
2. Izaberite:
   - Version: Latest (npr. 7.0)
   - Platform: Windows
   - Package: MSI
3. Preuzmite i instalirajte

### Korak 2: Pokrenite MongoDB kao servis

MongoDB se automatski pokreće kao Windows servis nakon instalacije.

Proverite da li radi:
```bash
# U PowerShell-u
Get-Service MongoDB
```

Trebalo bi da vidite status "Running".

### Korak 3: Povežite servise

Kada pokrenete servise **bez Docker-a**, oni će koristiti:
```
mongodb://localhost:27017
```

Ovo je već postavljeno kao podrazumevana vrednost u `config.go` fajlovima!

---

## 🔍 Kako proveriti da li je MongoDB povezan?

### Metoda 1: Provera kroz Docker

```bash
# Proverite da li MongoDB kontejner radi
docker ps

# Trebalo bi da vidite nešto kao:
# CONTAINER ID   IMAGE       COMMAND                  STATUS
# abc123def456   mongo:7.0   "docker-entrypoint..."   Up 5 minutes
```

### Metoda 2: Provera kroz servis logove

Kada pokrenete servise, proverite logove:

```bash
docker-compose logs users-service
```

Ako vidite:
```
users-service_1  | Connected to MongoDB
```

**MongoDB je uspešno povezan!** ✅

Ako vidite grešku:
```
users-service_1  | Failed to connect to MongoDB: connection refused
```

**Problem:** MongoDB nije pokrenut ili servisi ne mogu da ga pronađu.

---

## 🛠️ Rešavanje problema

### Problem 1: "Cannot connect to MongoDB"

**Rešenje:**
1. Proverite da li MongoDB radi:
   ```bash
   docker ps | findstr mongo
   ```

2. Ako ne radi, pokrenite ga:
   ```bash
   docker-compose up mongodb -d
   ```

3. Sačekajte 10-15 sekundi da se MongoDB potpuno pokrene

### Problem 2: "Port 27017 is already in use"

**Rešenje:**
Neko već koristi port 27017. Možete:
- Zatvoriti drugu MongoDB instancu
- Ili promeniti port u `docker-compose.yml`:
  ```yaml
  mongodb:
    ports:
      - "27018:27017"  # Promenite 27017 u 27018
  ```

### Problem 3: Servisi se pokreću pre MongoDB-a

**Rešenje:**
Vaš `docker-compose.yml` već ima `depends_on`, ali možete dodati health check:

```yaml
mongodb:
  image: mongo:7.0
  healthcheck:
    test: echo 'db.runCommand("ping").ok' | mongosh localhost:27017/test --quiet
    interval: 10s
    timeout: 5s
    retries: 5
```

---

## 📊 Struktura baza podataka

Kada se servisi povežu, MongoDB automatski kreira baze:

### 1. `users_db` (users-service)
- **Collection:** `users`
- **Sadrži:** Korisnike sa username, email, password hash, itd.

### 2. `music_streaming` (content-service)
- **Collections:** 
  - `artists` - Izvođači
  - `albums` - Albumi
  - `songs` - Pesme

### 3. `notifications_db` (notifications-service)
- **Collection:** `notifications`
- **Sadrži:** Notifikacije za korisnike

---

## 🧪 Testiranje konekcije

### Test 1: Povezivanje preko MongoDB Shell

```bash
# Ako koristite Docker:
docker exec -it projekat-2025-mongodb-1 mongosh

# U MongoDB shell-u:
show dbs
# Trebalo bi da vidite: users_db, music_streaming, notifications_db
```

### Test 2: Test kroz API

1. Pokrenite servise:
   ```bash
   docker-compose up
   ```

2. Registrujte novog korisnika:
   ```bash
   curl -X POST http://localhost:8081/api/users/register \
     -H "Content-Type: application/json" \
     -d '{"firstName":"Test","lastName":"User","email":"test@test.com","username":"testuser","password":"Test123!","confirmPassword":"Test123!"}'
   ```

3. Ako dobijete uspešan odgovor, **MongoDB radi!** ✅

---

## 📝 Rezime - Brzi start

**Za početak, samo pokrenite:**

```bash
cd Projekat-2025
docker-compose up
```

Docker će:
1. ✅ Preuzeti MongoDB (ako nije već preuzet)
2. ✅ Pokrenuti MongoDB
3. ✅ Pokrenuti sve servise
4. ✅ Automatski povezati servise sa MongoDB-om

**To je sve!** 🎉

---

## ❓ Često postavljana pitanja

**P: Gde se čuvaju podaci?**
A: U Docker volumenu `mongodb_data`. Podaci se ne gube kada zatvorite Docker, ali se gube ako obrišete volumen.

**P: Kako da vidim podatke u MongoDB-u?**
A: Koristite MongoDB Compass (GUI alat) ili `mongosh` (command line).

**P: Kako da obrišem sve podatke?**
A: 
```bash
docker-compose down -v  # -v briše volumene
```

**P: Mogu li da koristim MongoDB Atlas (cloud)?**
A: Da! Promenite `MONGODB_URI` u environment varijablama na Atlas connection string.

---

## 🎓 Dodatni resursi

- MongoDB dokumentacija: https://docs.mongodb.com/
- Docker dokumentacija: https://docs.docker.com/
- MongoDB Compass (GUI): https://www.mongodb.com/products/compass


