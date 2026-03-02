# 📋 Plan Testiranja - Sve Ocene (6-10)

**Projekat:** Music Streaming Platform  
**Format:** Konkretno i jasno, bez previše teksta

---

## 🔧 Priprema

```powershell
# 1. Pokreni Docker
docker-compose up -d
Start-Sleep -Seconds 60

# 2. Proveri status
docker-compose ps

# 3. Pokreni frontend (novi terminal)
cd frontend
npm start
```

**Frontend:** http://localhost:3000  
**MailHog:** http://localhost:8025  
**Jaeger:** http://localhost:16686

---

## ✅ OCENA 6

### Test 1.1: Registracija
**Gde:** `/register`  
**Akcija:** Popuni formu (ime, prezime, email, username, password) → Register  
**Očekivano:** Email u MailHog-u sa verifikacionim linkom → Klik na link → Email verified

### Test 1.2: Prijava
**Gde:** `/login`  
**Akcija:** Unesi username/password → Login  
**Očekivano:** Email sa OTP kodom → Unesi OTP → Uspešna prijava

### Test 1.3: Kreiranje Umetnika (Admin)
**Gde:** `/artists` (kao admin)  
**Akcija:** Add Artist → Popuni (name, biography, genres) → Create  
**Očekivano:** Umetnik se pojavljuje u listi

### Test 1.4: Kreiranje Albuma (Admin)
**Gde:** `/artists/{id}` → Add Album  
**Akcija:** Popuni (name, date, genre) → Create  
**Očekivano:** Album se pojavljuje

### Test 1.5: Kreiranje Pesme (Admin)
**Gde:** `/albums/{id}` → Add Song  
**Akcija:** Popuni (name, duration, genre) → Create  
**Očekivano:** Pesma se pojavljuje

### Test 1.6: Pregled Sadržaja
**Gde:** `/artists`, `/albums/{id}`, `/songs/{id}`  
**Akcija:** Klik na umetnika → album → pesmu  
**Očekivano:** Sve se prikazuje korektno

### Test 1.7: Pregled Notifikacija
**Gde:** `/notifications`  
**Akcija:** Otvori stranicu  
**Očekivano:** Lista notifikacija (ručno popunjena baza)

**Provera:**
```powershell
docker-compose ps  # Svi servisi healthy
docker logs projekat-2025-2-users-service-1 --tail=10
```

---

## ✅ OCENA 7

### Sve iz Ocene 6 + sledeće:

### Test 2.1: Reprodukcija Pesme
**Gde:** `/songs/{id}`  
**Akcija:** Klik Play  
**Očekivano:** Audio player se pojavljuje, pesma se reprodukuje

### Test 2.2: Pretraga
**Gde:** `/songs` ili `/artists`  
**Akcija:** Unesi tekst u search box → Search  
**Očekivano:** Filtrirani rezultati

### Test 2.3: Filtriranje po Žanru
**Gde:** `/artists`  
**Akcija:** Izaberi žanr iz dropdown → Filter  
**Očekivano:** Lista filtrirana po žanru

### Test 2.4: Ocenjivanje Pesme
**Gde:** `/songs/{id}` (kao korisnik)  
**Akcija:** Klik na zvezdu (1-5)  
**Očekivano:** Ocena se prikazuje, prosek se ažurira

### Test 2.5: Pretplata na Umetnika
**Gde:** `/artists/{id}`  
**Akcija:** Klik Subscribe  
**Očekivano:** Button se menja u Unsubscribe

### Test 2.6: Pretplata na Žanr
**Gde:** `/profile` → Subscriptions  
**Akcija:** Izaberi žanr → Subscribe to Genre  
**Očekivano:** Žanr se pojavljuje u listi

**Provera sinhronih poziva:**
```powershell
docker logs projekat-2025-2-api-gateway-1 --tail=20 | Select-String "content|ratings"
```

---

## ✅ OCENA 8

### Sve iz Ocene 7 + sledeće:

### Test 3.1: Generisanje Notifikacije
**Akcija:** 
1. Pretplati se na umetnika kao korisnik
2. Kao admin, kreiraj novu pesmu za tog umetnika
3. Sačekaj 5-10 sekundi

**Očekivano:** Notifikacija se pojavljuje u `/notifications`

**Provera asinhrone komunikacije:**
```powershell
docker logs projekat-2025-2-subscriptions-service-1 --tail=20 | Select-String "new_song|notification"
docker logs projekat-2025-2-recommendation-service-1 --tail=20 | Select-String "rating_created|subscription_created"
```

### Test 3.2: Preporuke
**Gde:** `/home` (kao korisnik)  
**Akcija:** Otvori početnu stranicu  
**Očekivano:** Prikazuju se preporuke umesto svih umetnika
- Pesme iz žanrova na koje je pretplaćen (ocena >= 4 ili bez ocene)
- Pesma sa najviše ocena 5 (iz žanra na koji nije pretplaćen)

**Provera Neo4j:**
```powershell
docker exec -it projekat-2025-2-neo4j-1 cypher-shell -u neo4j -p password "MATCH (u:User)-[:RATED]->(s:Song) RETURN u.id, s.name LIMIT 5"
```

### Test 3.3: Otpornost - Retry/Timeout
**Akcija:** 
1. Zaustavi servis: `docker stop projekat-2025-2-ratings-service-1`
2. Pokušaj da oceniš pesmu
3. Restartuj: `docker start projekat-2025-2-ratings-service-1`
4. Pokušaj ponovo

**Očekivano:** 
- Circuit breaker se aktivira
- Retry mehanizam pokušava ponovo
- Nakon restart-a, operacija uspeva

**Provera:**
```powershell
docker logs projekat-2025-2-api-gateway-1 --tail=30 | Select-String "circuit|retry|timeout"
```

---

## ✅ OCENA 9

### Sve iz Ocene 8 + sledeće:

### Test 4.1: API Composition
**Gde:** `/songs/{id}` (stranica pesme)  
**Akcija:** Otvori stranicu pesme  
**Očekivano:** Prikazuje se:
- Broj ocena (npr. "Prosečna ocena: 4.5 (5 ocena)")
- Prosečna ocena sa ikonom ⭐
- Podaci iz Content + Ratings servisa (API Gateway kompozuje podatke)
- **Napomena:** Informacije se takođe prikazuju i na listi pesama (`/songs`), ali za 4.1 je obavezno na stranici pesme

**Provera:**
```powershell
docker logs projekat-2025-2-api-gateway-1 --tail=20 | Select-String "composition|songs"
```

### Test 4.2: CQRS
**Gde:** `/profile` → Subscriptions  
**Akcija:** Otvori Subscriptions tab  
**Očekivano:** 
- Lista pretplata prikazuje imena umetnika
- Nema ponovnih poziva ka Content servisu (keširano)

**Provera (Developer Tools → Network):**
- Refresh stranicu
- Proveri da nema poziva ka `/api/content/artists/{id}` za svaku pretplatu

### Test 4.3: Tracing
**Gde:** Jaeger UI (http://localhost:16686)  
**Akcija:** 
1. Izvrši akciju u frontendu (npr. oceni pesmu)
2. Otvori Jaeger → Find Traces

**Očekivano:** 
- Trace se pojavljuje
- Prikazuje flow: API Gateway → Content → Ratings → Recommendation
- Svaki span ima trajanje

### Test 4.4: HDFS Upload
**Gde:** `/songs` → Edit pesmu (kao admin)  
**Akcija:** 
1. Klikni "Edit" na postojećoj pesmi
2. Izaberi novi .mp3 fajl u polju "Audio File"
3. Klikni "Save"
4. Sačekaj da se upload završi (može potrajati 10-30 sekundi za velike fajlove)

**Očekivano:** 
- Poruka "Audio uploaded successfully" ili "Pesma je ažurirana, ali upload audio fajla nije uspeo" (ako upload ne uspe)
- Ako upload uspe, pesma se može pustiti sa novim audio fajlom
- Ako upload ne uspe, proveri backend logove

**Provera HDFS:**
```powershell
# Proveri da li fajl postoji u HDFS
docker exec hdfs-namenode curl -s "http://localhost:9870/webhdfs/v1/audio/songs/?op=LISTSTATUS" | ConvertFrom-Json | Select-Object -ExpandProperty FileStatuses | Select-Object -ExpandProperty FileStatus | Select-Object pathSuffix, length | Format-Table
```

**Provera backend logova (ako upload ne uspe):**
```powershell
docker logs projekat-2025-2-content-service-1 --tail=50 | Select-String "upload|Upload|HDFS|error|Error"
```

**Napomena:** 
- Upload može potrajati 10-30 sekundi za velike fajlove
- Ako dobiješ "connection reset by peer" grešku, HDFS možda ima problema sa velikim fajlovima
- Proveri da li je HDFS namenode i datanode pokrenut: `docker ps | Select-String "hdfs"`

---

## ✅ OCENA 10

### Sve iz Ocene 9 + sledeće:

### Test 5.1: Brisanje Pesme - Saga (Uspešan Tok)
**Gde:** `/songs/{id}` (kao admin)  
**Akcija:** Delete Song → Potvrdi  
**Očekivano:** Poruka "Song deleted successfully"

**Provera Saga:**
```powershell
docker logs projekat-2025-2-saga-service-1 --tail=30 | Select-String "delete-song|saga"
```

**Očekivano u logovima:**
- Saga transaction pokrenuta
- Delete from Content → Delete ratings → Delete from Neo4j
- Sve uspešno

**Provera baza:**
```powershell
# Content
docker exec -it projekat-2025-2-mongodb-content-1 mongosh --eval "use music_streaming; db.songs.find({_id: 'song_id'})"

# Ratings
docker exec -it projekat-2025-2-mongodb-ratings-1 mongosh --eval "use ratings_db; db.ratings.find({songId: 'song_id'})"

# Neo4j
docker exec -it projekat-2025-2-neo4j-1 cypher-shell -u neo4j -p password "MATCH (s:Song {id: 'song_id'}) RETURN s"
```

**Očekivano:** Pesma obrisana iz svih baza

### Test 5.2: Saga Rollback
**Akcija:** 
1. Zaustavi ratings-service: `docker stop projekat-2025-2-ratings-service-1`
2. Pokušaj da obrišeš pesmu
3. Restartuj: `docker start projekat-2025-2-ratings-service-1`

**Očekivano:** 
- Saga se rollback-uje
- Poruka o grešci
- Pesma nije obrisana

**Provera:**
```powershell
docker logs projekat-2025-2-saga-service-1 --tail=30 | Select-String "rollback|compensate"
```

### Test 5.3: Istorija Aktivnosti
**Gde:** `/profile` → Activity History  
**Akcija:** 
1. Izvrši akcije: slušaj pesmu, oceni, pretplati se
2. Otvori Activity History

**Očekivano:** 
- Lista aktivnosti sa tipom, opisom, datumom
- Sortirano po datumu (najnovije prvo)

**Tipovi aktivnosti:**
- `song_played`, `rating_created`, `artist_subscribed`, `genre_subscribed`

**Provera Event Sourcing:**
```powershell
docker logs projekat-2025-2-analytics-service-1 --tail=20 | Select-String "event|activity"
```

### Test 5.4: Event Stream
**Gde:** Developer Tools → Network  
**Akcija:** Otvori `/profile` → Activity History  
**Očekivano:** Zahtev ka `/api/analytics/events/stream` vraća stream događaja

**Kako dobiti token i user ID:**
1. Otvori browser console (F12)
2. Ukucaj: `localStorage.getItem('token')` - kopiraj token
3. Za user ID, izvuci iz JWT tokena:
   ```javascript
   // U browser console:
   const token = localStorage.getItem('token');
   const payload = JSON.parse(atob(token.split('.')[1]));
   console.log('User ID:', payload.userId);
   ```

**Provera direktno:**
```powershell
# Zameni {token} i {user_id} sa stvarnim vrednostima
$token = "{token}"  # Iz browser console: localStorage.getItem('token')
$userId = "{user_id}"  # Iz JWT tokena (payload.userId)
$headers = @{ "Authorization" = "Bearer $token" }
Invoke-RestMethod -Uri "http://localhost:8081/api/analytics/events/stream?userId=$userId" -Headers $headers
```

**Napomena:** Token može isteći. Ako dobiješ "invalid or expired token", prijavi se ponovo i uzmi novi token.

**Alternativa - proveri u Network tab-u:**
- Otvori Activity History stranicu
- U Network tab-u, filtriraj po "stream"
- Ako ne vidiš zahtev, endpoint postoji ali frontend koristi `/activities` umesto `/events/stream`

### Test 5.5: Analitike
**Gde:** `/profile` → Analytics  
**Akcija:** Otvori Analytics tab  
**Očekivano:** Prikazuje se:
- Broj odslušanih pesama
- Prosek ocena
- Broj odslušanih pesama po žanru
- Top 5 umetnika
- Broj pretplata

**Provera Event Sourcing + CQRS:**
- Analitike izračunate iz event stream-a
- Podaci keširani (CQRS read model)

### Test 5.6: Keširanje
**Akcija:** 
1. Slušaj pesmu nekoliko puta (npr. 4 puta)
2. Proveri Redis

**Provera:**
```powershell
# Proveri play count za pesmu
docker exec projekat-2025-2-redis-1 redis-cli KEYS "song_play_count:*"
docker exec projekat-2025-2-redis-1 redis-cli GET "song_play_count:{song_id}"

# Proveri most played songs cache
docker exec projekat-2025-2-redis-1 redis-cli GET "most_played_songs"
```

**Očekivano:** 
- `song_play_count:{song_id}` = broj reprodukcija (npr. 4)
- `most_played_songs` = JSON sa najslušanijim pesmama (može biti prazan ako je invalidiran)

**Provera performansi:**
- Prvi put → zahtev ka HDFS
- Drugi put → iz Redis keša (brže)

---

## 📊 Checklist Finalna Provera

### Pre odbrane proveri:

- [ ] Svi Docker kontejneri healthy (`docker-compose ps`)
- [ ] Frontend radi (http://localhost:3000)
- [ ] MailHog radi (http://localhost:8025)
- [ ] Jaeger radi (http://localhost:16686)
- [ ] Svi testovi prošli
- [ ] Logovi bez grešaka
- [ ] Eventi se šalju asinhrono
- [ ] Saga transaction-i rade
- [ ] Event sourcing čuva aktivnosti
- [ ] Analitike se računaju iz event stream-a
- [ ] Keširanje radi (Redis)

---

## 🔍 Brze Debug Komande

```powershell
# Status servisa
docker-compose ps

# Logovi
docker-compose logs --tail=20 [service-name]

# MongoDB
docker exec -it projekat-2025-2-mongodb-[service]-1 mongosh --eval "use [db]; db.[collection].find().pretty()"

# Neo4j
docker exec -it projekat-2025-2-neo4j-1 cypher-shell -u neo4j -p password "MATCH (n) RETURN n LIMIT 10"

# Redis
docker exec -it projekat-2025-2-redis-1 redis-cli KEYS "*"

# Cassandra
docker exec -it projekat-2025-2-cassandra-1 cqlsh -e "SELECT * FROM notifications_keyspace.notifications LIMIT 10;"
```

---

## 📝 Napomene

1. **Sve testove radite kroz frontend** - ne Postman/cURL
2. **Proveri logove** nakon svake akcije
3. **Koristi Developer Tools** za network zahteve
4. **Proveri Jaeger** za tracing
5. **MailHog** za email notifikacije

---

**Srećno! 🎉**
