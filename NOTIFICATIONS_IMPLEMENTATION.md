# ✅ Implementacija Generisanja Notifikacija (1.11)

## 📋 Pregled Implementacije

### **Zahtev 1.11: Generisanje notifikacija**

Korisnik treba da dobije notifikaciju:
- ✅ Kada se doda novi album umetnika na čiji sadržaj je pretplaćen
- ✅ Kada se doda nova pesma umetnika na čiji sadržaj je pretplaćen
- ✅ Kada se doda novi umetnik žanra na koji je pretplaćen

---

## 🏗️ Arhitektura

### **Asinhrona komunikacija između servisa (2.6)**

Implementacija koristi **event-driven arhitekturu**:

1. **Content-Service** → emituje event kada se kreira novi artist/album/song
2. **Subscriptions-Service** → prima evente i proverava pretplate
3. **Notifications-Service** → kreira notifikacije za pretplaćene korisnike

```
Content-Service (kreira artist/album/song)
    ↓ [HTTP POST event]
Subscriptions-Service (/events endpoint)
    ↓ [proverava pretplate]
    ↓ [HTTP POST notification]
Notifications-Service (/notifications endpoint)
    ↓ [čuva u Cassandra]
```

---

## 📝 Detalji Implementacije

### **1. Content-Service - Event Emitter**

**Lokacija:** `services/content-service/internal/events/emitter.go`

**Funkcionalnost:**
- Asinhrono šalje HTTP POST zahtev ka subscriptions-service
- Timeout: 2 sekunde
- Ne blokira glavni tok izvršavanja

**Event tipovi:**
- `new_artist` - kada se kreira novi umetnik
- `new_album` - kada se kreira novi album
- `new_song` - kada se kreira nova pesma

**Integracija:**
- `CreateArtist` handler emituje `new_artist` event
- `CreateAlbum` handler emituje `new_album` event
- `CreateSong` handler emituje `new_song` event

---

### **2. Subscriptions-Service - Event Handler**

**Lokacija:** `services/subscriptions-service/cmd/main.go`

**Endpoint:** `POST /events`

**Funkcionalnost:**
1. Prima event od content-service
2. Proverava tip eventa (`new_artist`, `new_album`, `new_song`)
3. Pronalazi sve pretplaćene korisnike:
   - Za `new_artist`: pretrage po žanrovima umetnika
   - Za `new_album`: pretrage po artist ID-ovima albuma
   - Za `new_song`: pretrage po artist ID-ovima pesme
4. Za svakog pretplaćenog korisnika poziva notifications-service

**Nove metode u SubscriptionRepository:**
- `GetByArtistID(ctx, artistID)` - vraća sve pretplate za određenog umetnika
- `GetByGenre(ctx, genre)` - vraća sve pretplate za određeni žanr

---

### **3. Notifications-Service - Notification Creation**

**Lokacija:** `services/notifications-service/internal/handler/notification_handler.go`

**Endpoint:** `POST /notifications`

**Funkcionalnost:**
- Prima zahtev za kreiranje notifikacije
- Validira podatke (userId, type, message, contentId)
- Kreira notifikaciju u Cassandra bazi
- Vraća kreiranu notifikaciju

**Tipovi notifikacija:**
- `new_artist` - "New artist 'X' in genre Y has been added"
- `new_album` - "New album 'X' by artist has been released"
- `new_song` - "New song 'X' by artist has been added"

---

## 🔧 Konfiguracija

### **Content-Service Environment Variables:**
```yaml
SUBSCRIPTIONS_SERVICE_URL=http://subscriptions-service:8004
```

### **Subscriptions-Service Environment Variables:**
```yaml
NOTIFICATIONS_SERVICE_URL=http://notifications-service:8005
```

### **Docker Compose:**
- `content-service` zavisi od `subscriptions-service`
- `subscriptions-service` zavisi od `notifications-service`

---

## 🧪 Kako Testirati

### **Test 1: Notifikacija za novi umetnik**

1. Pretplatite se na žanr (npr. "Pop")
2. Kreirajte novog umetnika sa žanrom "Pop" (kao admin)
3. Proverite notifikacije korisnika koji je pretplaćen na "Pop"

**Očekivani rezultat:**
- Notifikacija tipa `new_artist` sa porukom: "New artist 'X' in genre Pop has been added"

### **Test 2: Notifikacija za novi album**

1. Pretplatite se na umetnika (npr. "artist1")
2. Kreirajte novi album za tog umetnika (kao admin)
3. Proverite notifikacije korisnika koji je pretplaćen na umetnika

**Očekivani rezultat:**
- Notifikacija tipa `new_album` sa porukom: "New album 'X' by artist has been released"

### **Test 3: Notifikacija za novu pesmu**

1. Pretplatite se na umetnika (npr. "artist1")
2. Kreirajte novu pesmu za tog umetnika (kao admin)
3. Proverite notifikacije korisnika koji je pretplaćen na umetnika

**Očekivani rezultat:**
- Notifikacija tipa `new_song` sa porukom: "New song 'X' by artist has been added"

---

## 📊 Flow Diagram

```
┌─────────────────┐
│ Content-Service │
│  Create Artist  │
└────────┬────────┘
         │
         │ emit event (async)
         ↓
┌──────────────────────┐
│ Subscriptions-Service│
│   /events endpoint   │
└────────┬─────────────┘
         │
         │ GetByGenre(genre)
         ↓
┌──────────────────────┐
│  Subscription Repo   │
│  [pretplaćeni users] │
└────────┬─────────────┘
         │
         │ for each user
         ↓
┌──────────────────────┐
│ Notifications-Service│
│  POST /notifications │
└────────┬─────────────┘
         │
         │ Create notification
         ↓
┌──────────────────────┐
│     Cassandra        │
│  [notifications]     │
└──────────────────────┘
```

---

## ✅ Checklist

- [x] **Event emitter u content-service**
  - [x] `NewArtistEvent` struktura
  - [x] `NewAlbumEvent` struktura
  - [x] `NewSongEvent` struktura
  - [x] Asinhrono emitovanje eventa

- [x] **Event handler u subscriptions-service**
  - [x] `POST /events` endpoint
  - [x] `handleNewArtistEvent` funkcija
  - [x] `handleNewAlbumEvent` funkcija
  - [x] `handleNewSongEvent` funkcija

- [x] **Repository metode**
  - [x] `GetByArtistID` metoda
  - [x] `GetByGenre` metoda

- [x] **Notification creation**
  - [x] `POST /notifications` endpoint u notifications-service
  - [x] `CreateNotification` handler metoda
  - [x] Validacija podataka

- [x] **Konfiguracija**
  - [x] Environment varijable u docker-compose.yml
  - [x] Dependencies između servisa

---

## 🎯 Status: KOMPLETNO IMPLEMENTIRANO

Svi zahtevi za generisanje notifikacija (1.11) su implementirani:
- ✅ Notifikacija za novi album umetnika
- ✅ Notifikacija za novu pesmu umetnika
- ✅ Notifikacija za novog umetnika žanra
- ✅ Asinhrona komunikacija između servisa (2.6)

---

## 📝 Napomene

1. **Asinhronost**: Eventi se šalju asinhrono, tako da ne blokiraju glavni tok izvršavanja
2. **Resilience**: Ako subscriptions-service ili notifications-service nisu dostupni, event se jednostavno gubi (može se dodati retry mehanizam)
3. **Scalability**: Svaki servis može da se skalira nezavisno
