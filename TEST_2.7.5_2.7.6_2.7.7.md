# 🧪 Test Vodič - 2.7.5, 2.7.6, 2.7.7

## 📋 Pregled Testova

### **2.7.5 Retry mehanizam sa exponential backoff**
### **2.7.6 Eksplicitno postavljen timeout za vraćanje odgovora korisniku**
### **2.7.7 Upstream servis odustaje od obrade zahteva ako je istekao timeout**

---

## 🧪 Test 1: Retry mehanizam (2.7.5)

**Cilj:** Proveriti da retry mehanizam radi sa exponential backoff-om.

### Koraci:

1. **Zaustavite content-service:**
   ```powershell
   docker-compose stop content-service
   ```

2. **Otvorite logove za ratings-service:**
   ```powershell
   docker-compose logs -f ratings-service
   ```

3. **Pokušajte da ocenite pesmu** preko frontenda ili API-ja:
   ```powershell
   # Primer zahtev
   Invoke-WebRequest -Uri "http://localhost:8000/api/ratings/rate-song?songId=test123&userId=user123&rating=5" -Method POST -Headers @{Authorization="Bearer YOUR_TOKEN"}
   ```

4. **Proverite logove** - trebalo bi da vidite:
   ```
   Retry attempt 1/3 failed: ..., retrying in 100ms...
   Retry attempt 2/3 failed: ..., retrying in 200ms...
   Retry attempt 3/3 failed: ..., retrying in 400ms...
   Content-service unavailable for song ..., fallback activated
   ```

5. **Ponovite za subscriptions-service:**
   ```powershell
   docker-compose logs -f subscriptions-service
   # Pokušajte da se pretplatite na umetnika
   ```

**Očekivani rezultat:**
- ✅ Retry pokušaji sa exponential backoff (100ms → 200ms → 400ms)
- ✅ Logovi pokazuju retry pokušaje
- ✅ Fallback se aktivira nakon svih retry pokušaja

---

## 🧪 Test 2: API Gateway timeout (2.7.6)

**Cilj:** Proveriti da API Gateway vraća timeout odgovor korisniku ako servis ne odgovori na vreme.

### Koraci:

1. **Zaustavite ratings-service:**
   ```powershell
   docker-compose stop ratings-service
   ```

2. **Otvorite logove za API Gateway:**
   ```powershell
   docker-compose logs -f api-gateway
   ```

3. **Pokušajte da ocenite pesmu** preko frontenda ili API-ja:
   ```powershell
   Invoke-WebRequest -Uri "http://localhost:8000/api/ratings/rate-song?songId=test123&userId=user123&rating=5" -Method POST -Headers @{Authorization="Bearer YOUR_TOKEN"}
   ```

4. **Proverite odgovor** - trebalo bi da dobijete:
   - **Status kod:** `408 Request Timeout`
   - **Body:** `"Request timeout - service did not respond in time"`

5. **Proverite logove** - trebalo bi da vidite:
   ```
   Request timeout for http://ratings-service:8083/rate-song: context deadline exceeded
   ```

**Očekivani rezultat:**
- ✅ API Gateway vraća HTTP 408 Request Timeout
- ✅ Odgovor stiže korisniku nakon ~5 sekundi
- ✅ Logovi pokazuju timeout

---

## 🧪 Test 3: Upstream servis odustaje od obrade (2.7.7)

**Cilj:** Proveriti da upstream servis prekida obradu ako je istekao timeout.

### Koraci:

1. **Pokrenite sve servise:**
   ```powershell
   docker-compose up -d
   ```

2. **Simulirajte spor odgovor** - možete koristiti delay u servisu ili:
   - Zaustavite content-service
   - Pokušajte da ocenite pesmu koja ne postoji

3. **Otvorite logove za ratings-service:**
   ```powershell
   docker-compose logs -f ratings-service
   ```

4. **Pokušajte da ocenite pesmu** sa kratkim timeout-om (možete modifikovati API Gateway timeout na 2 sekunde za test)

5. **Proverite logove** - trebalo bi da vidite:
   ```
   Request context cancelled before processing: context deadline exceeded
   Request context cancelled during circuit breaker call: context deadline exceeded
   Request context cancelled during database read: context deadline exceeded
   ```

6. **Proverite odgovor** - trebalo bi da dobijete:
   - **Status kod:** `408 Request Timeout`
   - **Body:** `"Request timeout - processing abandoned"`

**Očekivani rezultat:**
- ✅ Servis prekida obradu kada context istekne
- ✅ Vraća HTTP 408 Request Timeout
- ✅ Logovi pokazuju gde je obrada prekinuta

---

## 🧪 Test 4: Kompletan scenario (2.7.5 + 2.7.6 + 2.7.7)

**Cilj:** Proveriti da sve komponente rade zajedno.

### Koraci:

1. **Zaustavite content-service:**
   ```powershell
   docker-compose stop content-service
   ```

2. **Otvorite logove za sve servise:**
   ```powershell
   # Terminal 1
   docker-compose logs -f api-gateway
   
   # Terminal 2
   docker-compose logs -f ratings-service
   ```

3. **Pokušajte da ocenite pesmu:**
   ```powershell
   Invoke-WebRequest -Uri "http://localhost:8000/api/ratings/rate-song?songId=test123&userId=user123&rating=5" -Method POST -Headers @{Authorization="Bearer YOUR_TOKEN"}
   ```

4. **Posmatrajte tok:**
   - API Gateway šalje zahtev sa timeout-om (2.7.6)
   - Ratings-service pokušava retry sa exponential backoff (2.7.5)
   - Ako timeout istekne, ratings-service prekida obradu (2.7.7)
   - API Gateway vraća timeout odgovor korisniku (2.7.6)

**Očekivani rezultat:**
- ✅ Retry mehanizam radi
- ✅ API Gateway vraća timeout odgovor
- ✅ Upstream servis prekida obradu na vreme

---

## ⚡ Brzi Test (PowerShell)

```powershell
# 1. Zaustavite content-service
docker-compose stop content-service

# 2. Proverite logove za retry
docker-compose logs ratings-service --tail 50 | Select-String -Pattern "retry|Retry|timeout|Timeout" -CaseSensitive:$false

# 3. Pokušajte da ocenite pesmu
$response = Invoke-WebRequest -Uri "http://localhost:8000/api/ratings/rate-song?songId=test123&userId=user123&rating=5" -Method POST -ErrorAction SilentlyContinue
Write-Host "Status: $($response.StatusCode)"
Write-Host "Body: $($response.Content)"

# 4. Proverite API Gateway logove za timeout
docker-compose logs api-gateway --tail 20 | Select-String -Pattern "timeout|Timeout" -CaseSensitive:$false

# 5. Restartujte content-service
docker-compose start content-service
```

---

## ✅ Checklist

### 2.7.5 Retry mehanizam
- [ ] Retry pokušaji se izvršavaju (3 puta)
- [ ] Exponential backoff radi (delay se povećava)
- [ ] Logovi pokazuju retry pokušaje
- [ ] Fallback se aktivira nakon svih pokušaja

### 2.7.6 Eksplicitni timeout za korisnika
- [ ] API Gateway vraća HTTP 408 Request Timeout
- [ ] Timeout je 5 sekundi (15 za notifications-service)
- [ ] Odgovor stiže korisniku na vreme
- [ ] Logovi pokazuju timeout

### 2.7.7 Upstream servis odustaje od obrade
- [ ] Servis koristi request context
- [ ] Servis prekida obradu kada context istekne
- [ ] Vraća HTTP 408 Request Timeout
- [ ] Logovi pokazuju gde je obrada prekinuta

---

## 🎯 Ključni Indikatori Uspeha

**2.7.5:**
- ✅ `Retry attempt X/3 failed` u logovima
- ✅ Exponential backoff delay (100ms → 200ms → 400ms)

**2.7.6:**
- ✅ HTTP 408 Request Timeout odgovor
- ✅ `Request timeout for ...` u API Gateway logovima

**2.7.7:**
- ✅ `Request context cancelled` u servis logovima
- ✅ HTTP 408 Request Timeout odgovor
- ✅ Obrada se prekida na vreme

---

## 📝 Napomene

- **Timeout vrednosti:**
  - API Gateway: 5 sekundi (15 za notifications-service)
  - Ratings-service: 5 sekundi za database operacije
  - Subscriptions-service: 5 sekundi za database operacije

- **Retry konfiguracija:**
  - Max retries: 3
  - Initial delay: 100ms
  - Max delay: 2s
  - Backoff multiplier: 2.0

- **Za testiranje timeout-a:**
  - Možete privremeno smanjiti timeout vrednosti u kodu
  - Ili koristiti `sleep` u servisima za simulaciju spore obrade
