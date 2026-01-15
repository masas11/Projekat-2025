# 🚀 KAKO POKRENUTI PROJEKAT - Korak po Korak

## ✅ Šta već imate:
- ✅ Go 1.24.12 instaliran
- ✅ Docker 29.1.2 instaliran

---

## 📋 KORAK 1: Otvorite Command Prompt (CMD)

1. Pritisnite `Windows + R`
2. Ukucajte: `cmd`
3. Pritisnite Enter

---

## 📋 KORAK 2: Idite u folder projekta

U CMD-u ukucajte:

```cmd
cd C:\Users\boris\OneDrive\Desktop\projekat\Projekat-2025
```

Proverite da li ste na pravom mestu:

```cmd
dir docker-compose.yml
```

Ako vidite `docker-compose.yml`, dobro ste! ✅

---

## 📋 KORAK 3: Pokrenite MongoDB i sve servise

### Opcija A: Pokrenite SVE ODJEDNOM (preporučeno)

```cmd
docker-compose up
```

**Šta se dešava:**
1. Docker preuzima MongoDB sliku (prvi put može potrajati 2-3 minuta)
2. Build-uje sve Go servise (može potrajati prvi put)
3. Pokreće MongoDB
4. Pokreće sve servise (users-service, content-service, itd.)
5. Svi se automatski povezuju na MongoDB! 🎉

**Kako znati da radi:**
Videćete poruke kao:
```
mongodb_1              | {"t":{"$date":"..."},"s":"I",...}
users-service_1        | Connected to MongoDB
users-service_1        | Users service running on port 8001
content-service_1      | Connected to MongoDB
content-service_1      | Content service running on port 8002
api-gateway_1          | API Gateway running on port 8081
```

**✅ Ako vidite "Connected to MongoDB" - SVE RADI!**

---

### Opcija B: Pokrenite samo MongoDB prvo (za testiranje)

Ako želite prvo da testirate samo MongoDB:

```cmd
docker-compose up mongodb
```

Sada ćete videti samo MongoDB logove. Proverite da li piše:
```
Listening on 0.0.0.0:27017
```

**Za zaustavljanje:** Pritisnite `Ctrl + C`

---

## 📋 KORAK 4: Proverite da li MongoDB radi

### Test 1: Proverite Docker kontejnere

Otvorite NOVI CMD prozor (ostavite prvi da radi) i ukucajte:

```cmd
docker ps
```

Trebalo bi da vidite nešto kao:
```
CONTAINER ID   IMAGE          STATUS
abc123def456   mongo:7.0      Up 2 minutes
xyz789ghi012   api-gateway    Up 2 minutes
...
```

### Test 2: Proverite da li servisi rade

```cmd
curl http://localhost:8081/api/users/health
```

Ako vidite: `users-service is running` - **RADI!** ✅

---

## 📋 KORAK 5: Kako zaustaviti sve

Vratite se u prvi CMD prozor gde ste pokrenuli `docker-compose up` i pritisnite:

```
Ctrl + C
```

Zatim:

```cmd
docker-compose down
```

Ovo zaustavlja sve kontejnere, ali **podaci u MongoDB-u se ne brišu!**

---

## 🔍 KAKO VIDETI PODATKE U MONGODB-U

### Opcija 1: Preko MongoDB Shell (u Docker-u)

```cmd
docker exec -it projekat-2025-mongodb-1 mongosh
```

U MongoDB shell-u:
```javascript
show dbs
// Trebalo bi da vidite: users_db, music_streaming, notifications_db

use users_db
show collections
// Trebalo bi da vidite: users

db.users.find()
// Ovo će pokazati sve korisnike
```

Za izlaz iz MongoDB shell-a: ukucajte `exit`

---

### Opcija 2: Preko MongoDB Compass (GUI - LEPŠE)

1. Preuzmite MongoDB Compass: https://www.mongodb.com/products/compass
2. Installirajte
3. Otvorite Compass
4. U connection string unesite: `mongodb://localhost:27017`
5. Kliknite "Connect"
6. Videćete sve baze i kolekcije! 🎉

---

## ❓ ČESTI PROBLEMI I REŠENJA

### Problem 1: "docker-compose: command not found"

**Rešenje:** 
- Instalirajte Docker Desktop ponovo
- Ili pokušajte: `docker compose up` (bez crtice)

### Problem 2: "Cannot connect to MongoDB"

**Rešenje:**
1. Proverite da li MongoDB radi: `docker ps`
2. Ako ne radi, pokrenite ponovo:
   ```cmd
   docker-compose restart mongodb
   ```

### Problem 3: "Port 27017 is already in use"

**Rešenje:**
Neko drugi MongoDB već radi na portu 27017.

Proverite:
```cmd
netstat -ano | findstr :27017
```

Zaustavite proces ili promenite port u `docker-compose.yml`:
```yaml
mongodb:
  ports:
    - "27018:27017"  # Promenite u 27018
```

### Problem 4: "Build failed" za neki servis

**Rešenje:**
1. Proverite da li ste u pravom folderu
2. Pokušajte da build-ujete jedan po jedan:
   ```cmd
   docker-compose build users-service
   docker-compose build content-service
   ```

---

## 📊 KAKO SE SERVISI POVEZUJU NA MONGODB

### Automatski preko Docker Compose

Kada pokrenete `docker-compose up`, sve se dešava automatski:

1. **MongoDB se pokreće prvi** (jer drugi servisi imaju `depends_on: mongodb`)
2. **Servisi se pokreću** i čitaju `MONGODB_URI` iz environment varijabli
3. **URI je:** `mongodb://mongodb:27017`
   - `mongodb` = ime servisa u Docker Compose (Docker automatski razrešava IP adresu)
   - `27017` = MongoDB port

### Gde se konfiguriše?

U `docker-compose.yml`:
```yaml
users-service:
  environment:
    - MONGODB_URI=mongodb://mongodb:27017  # ← Ovo se prosleđuje u servis
    - MONGODB_DATABASE=users_db            # ← Ime baze
```

U Go kodu (`config/config.go`):
```go
mongoURI := os.Getenv("MONGODB_URI")  // Čita iz environment varijable
// Ako nije postavljena, koristi podrazumevanu: mongodb://localhost:27017
```

---

## 🎯 REZIME - BRZI START

**Samo 2 komande:**

```cmd
cd C:\Users\boris\OneDrive\Desktop\projekat\Projekat-2025
docker-compose up
```

**To je sve!** MongoDB i svi servisi će se pokrenuti i automatski povezati! 🎉

---

## 📝 DODATNE KOMANDE

```cmd
# Vidite logove samo za MongoDB
docker-compose logs mongodb

# Vidite logove za users-service
docker-compose logs users-service

# Restartujte samo jedan servis
docker-compose restart users-service

# Vidite status svih servisa
docker-compose ps

# Obrišite sve (uključujući podatke!)
docker-compose down -v
```

---

## ✅ PROVERA DA LI SVE RADI

Nakon što pokrenete `docker-compose up`, trebalo bi da vidite:

1. ✅ `mongodb_1 | Listening on 0.0.0.0:27017`
2. ✅ `users-service_1 | Connected to MongoDB`
3. ✅ `content-service_1 | Connected to MongoDB`
4. ✅ `notifications-service_1 | Connected to MongoDB`
5. ✅ `api-gateway_1 | API Gateway running on port 8081`

Ako vidite sve ovo - **SVE PERFEKTNO RADI!** 🎉🎉🎉

