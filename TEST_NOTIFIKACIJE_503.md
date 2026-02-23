# 🔧 Rešavanje problema sa notifikacijama (503 Service Unavailable)

## 🔍 Problem

Frontend prikazuje grešku "Service unavailable" (503) kada pokušava da učita notifikacije.

## ✅ Rešenja

### 1. Proveri da li servisi rade

```bash
# Proveri status svih servisa
docker-compose ps

# Proveri logove notifications-service
docker-compose logs notifications-service --tail 50

# Proveri logove api-gateway
docker-compose logs api-gateway --tail 50 | grep -i "notifications\|503\|unavailable"
```

### 2. Proveri Cassandra konekciju

```bash
# Proveri da li Cassandra radi
docker-compose logs cassandra --tail 30

# Proveri health notifications-service
curl http://localhost:8005/health
```

### 3. Restart servisa ako je potrebno

```bash
# Restart notifications-service
docker-compose restart notifications-service

# Restart api-gateway
docker-compose restart api-gateway

# Ako i dalje ne radi, restartuj sve
docker-compose restart
```

### 4. Proveri autentifikaciju (401 greške)

**Problem:** JWT token je istekao ili nevažeći.

**Rešenje:**
1. Odjavi se iz aplikacije
2. Prijavi se ponovo
3. Proveri da li token postoji u localStorage:
   ```javascript
   // U browser konzoli
   localStorage.getItem('token')
   ```

### 5. Proveri da li je Cassandra spremna

Cassandra može da traje 30-60 sekundi da se inicijalizuje. Notifications-service ima retry mehanizam (30 pokušaja, svaki 2 sekunde = 60 sekundi).

```bash
# Proveri da li Cassandra prihvata konekcije
docker-compose exec cassandra cqlsh -e "DESCRIBE KEYSPACES;"
```

### 6. Ručno testiranje API-ja

```bash
# 1. Prijavi se i uzmi JWT token
# 2. Testiraj health endpoint
curl http://localhost:8005/health

# 3. Testiraj notifications endpoint (zameni TOKEN sa pravim tokenom)
curl -H "Authorization: Bearer TOKEN" http://localhost:8081/api/notifications

# 4. Testiraj direktno notifications-service (zameni USER_ID)
curl "http://localhost:8005/notifications?userId=USER_ID"
```

## 🐛 Debugging koraci

### Korak 1: Proveri logove
```bash
# Svi logovi vezani za notifikacije
docker-compose logs | grep -i "notification\|cassandra" | tail -50
```

### Korak 2: Proveri network
```bash
# Proveri da li servisi mogu da komuniciraju
docker-compose exec api-gateway ping notifications-service
docker-compose exec notifications-service ping cassandra
```

### Korak 3: Proveri environment varijable
```bash
# Proveri konfiguraciju api-gateway
docker-compose exec api-gateway env | grep NOTIFICATIONS

# Proveri konfiguraciju notifications-service
docker-compose exec notifications-service env | grep CASSANDRA
```

## ✅ Checklist

- [ ] Cassandra radi (`docker-compose ps cassandra`)
- [ ] Notifications-service radi (`docker-compose ps notifications-service`)
- [ ] API Gateway radi (`docker-compose ps api-gateway`)
- [ ] Cassandra je spremna (`docker-compose logs cassandra | grep "Startup complete"`)
- [ ] Notifications-service je povezan na Cassandra (`docker-compose logs notifications-service | grep "Connected to Cassandra"`)
- [ ] JWT token je važeći (korisnik je prijavljen)
- [ ] Health endpoint radi (`curl http://localhost:8005/health`)

## 🔄 Ako ništa ne pomaže

1. **Restartuj sve servise:**
   ```bash
   docker-compose down
   docker-compose up -d
   ```

2. **Proveri da li postoje greške u logovima:**
   ```bash
   docker-compose logs --tail 100 | grep -i "error\|fatal\|panic"
   ```

3. **Proveri da li postoje problemi sa mrežom:**
   ```bash
   docker network ls
   docker network inspect projekat-2025-1_music-streaming-network
   ```

## 📝 Napomene

- **Timeout je povećan na 15 sekundi** za notifications-service zbog Cassandra inicijalizacije
- **401 greške** znače da JWT token nije važeći - korisnik mora da se ponovo prijavi
- **503 greške** znače da notifications-service nije dostupan - proveri logove
