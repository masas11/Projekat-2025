# 🧪 Brzi test - 2.6 i Notifikacije

## 📋 Test 1: Asinhrona komunikacija (2.6)

### Test kreiranja pesme → Notifikacije
```bash
# 1. Prijavi se kao korisnik (npr. "ana")
# 2. Pretplati se na žanr (npr. "Pop")
# 3. Odjavi se, prijavi se kao admin
# 4. Kreiraj novu Pop pesmu
# 5. Odjavi se, prijavi se kao korisnik
# 6. Proveri notifikacije - trebalo bi da vidiš notifikaciju
```

**Provera logova (PowerShell):**
```powershell
# Proveri da li content-service emituje event-e
docker-compose logs content-service --tail 20 | Select-String -Pattern "Event emitted" -CaseSensitive:$false

# Proveri da li subscriptions-service prima event-e
docker-compose logs subscriptions-service --tail 20 | Select-String -Pattern "Received event|event" -CaseSensitive:$false

# Proveri da li recommendation-service prima event-e
docker-compose logs recommendation-service --tail 20 | Select-String -Pattern "Received event|event" -CaseSensitive:$false
```

### Test ocenjivanja → Neo4j
```powershell
# 1. Oceni pesmu (npr. 5 zvezdica)
# 2. Proveri logove
docker-compose logs recommendation-service --tail 20 | Select-String -Pattern "rating" -CaseSensitive:$false
```

### Test brisanja pesme
```powershell
# 1. Obriši pesmu kao admin
# 2. Proveri logove
docker-compose logs recommendation-service --tail 20 | Select-String -Pattern "deleted" -CaseSensitive:$false
```

---

## 📋 Test 2: Notifikacije (503/401 greške)

### Brza provera (PowerShell)
```powershell
# 1. Proveri status servisa
docker-compose ps | Select-String -Pattern "notifications|api-gateway|cassandra"

# 2. Proveri health
Invoke-WebRequest -Uri http://localhost:8005/health -UseBasicParsing | Select-Object -ExpandProperty Content

# 3. Proveri logove
docker-compose logs notifications-service --tail 30
docker-compose logs api-gateway --tail 30 | Select-String -Pattern "notifications" -CaseSensitive:$false
```

### Ako ne radi
```powershell
# Restart servisa
docker-compose restart notifications-service api-gateway

# Sačekaj 30-60 sekundi (Cassandra inicijalizacija)
# Proveri ponovo
Invoke-WebRequest -Uri http://localhost:8005/health -UseBasicParsing | Select-Object -ExpandProperty Content
```

### Test u browseru
1. **401 greška** → Odjavi se i prijavi se ponovo (JWT token istekao)
2. **503 greška** → Proveri logove i restartuj servise
3. **Proveri u browser konzoli:**
   ```javascript
   // Proveri da li postoji token
   localStorage.getItem('token')
   
   // Proveri network tab - vidi status kod
   ```

---

## ✅ Checklist

### 2.6 Asinhrona komunikacija
- [ ] Kreiranje pesme emituje događaj → subscriptions-service
- [ ] Kreiranje pesme emituje događaj → recommendation-service
- [ ] Ocenjivanje pesme emituje događaj → recommendation-service
- [ ] Brisanje pesme emituje događaj → recommendation-service
- [ ] Neo4j graf se ažurira na osnovu događaja

### Notifikacije
- [ ] Cassandra radi (`docker-compose ps | Select-String cassandra`)
- [ ] Notifications-service radi (`docker-compose ps | Select-String notifications`)
- [ ] Health endpoint radi (`Invoke-WebRequest -Uri http://localhost:8005/health -UseBasicParsing`)
- [ ] JWT token je važeći (korisnik je prijavljen)
- [ ] Notifikacije se prikazuju u frontendu

---

## 🔍 Provera svih logova odjednom (PowerShell)

```powershell
# Svi event-i
docker-compose logs --tail 50 | Select-String -Pattern "event|received|processed" -CaseSensitive:$false

# Svi problemi
docker-compose logs --tail 100 | Select-String -Pattern "error|fatal|503|401" -CaseSensitive:$false
```

---

## ⚡ Najbrži test (sve odjednom) - PowerShell

```powershell
# 1. Proveri servise
docker-compose ps

# 2. Testiraj health
Invoke-WebRequest -Uri http://localhost:8005/health -UseBasicParsing | Select-Object -ExpandProperty Content

# 3. Proveri logove za event-e
docker-compose logs --tail 30 | Select-String -Pattern "event|notification" -CaseSensitive:$false

# 4. Ako nešto ne radi, restartuj
docker-compose restart notifications-service api-gateway recommendation-service
```

---

## 🎯 Ključni indikatori uspeha

**2.6:**
- ✅ `Event emitted` u logovima
- ✅ `Event received` u logovima
- ✅ Neo4j graf se ažurira

**Notifikacije:**
- ✅ Health endpoint vraća 200
- ✅ Notifikacije se prikazuju u frontendu
- ✅ Nema 503/401 grešaka u browser konzoli
