# 🔔 Frontend Pretplate - Implementacija

## ✅ Šta Je Dodato

### **1. API Metode (`frontend/src/services/api.js`)**

Dodate metode za pretplatu:
- `subscribeToArtist(artistId, userId)` - Pretplata na umetnika
- `subscribeToGenre(genre, userId)` - Pretplata na žanr

### **2. Pretplata na Umetnika (`frontend/src/components/ArtistDetail.js`)**

- ✅ Dugme "🔔 Pretplati se" na stranici detalja umetnika
- ✅ Vidljivo samo za prijavljene korisnike
- ✅ Prikazuje poruku o uspešnoj pretplati
- ✅ Error handling

### **3. Pretplata na Žanr (`frontend/src/components/Songs.js`)**

- ✅ Dugme "🔔" pored dropdown-a za filtriranje po žanru
- ✅ Vidljivo samo kada je žanr izabran i korisnik je prijavljen
- ✅ Prikazuje poruku o uspešnoj pretplati
- ✅ Error handling

---

## 🎯 Kako Koristiti

### **Pretplata na Umetnika:**

1. Idite na stranicu umetnika: `/artists/:id`
2. Kliknite na dugme "🔔 Pretplati se"
3. Poruka će se pojaviti: "Uspešno ste se pretplatili na ovog umetnika!"

### **Pretplata na Žanr:**

1. Idite na stranicu pesama: `/songs`
2. Izaberite žanr iz dropdown-a "Filtriranje po žanru"
3. Kliknite na dugme "🔔" pored dropdown-a
4. Poruka će se pojaviti: "Uspešno ste se pretplatili na žanr: [naziv žanra]!"

---

## 🔧 Tehnički Detalji

### **API Pozivi:**

```javascript
// Pretplata na umetnika
await api.subscribeToArtist(artistId, userId);

// Pretplata na žanr
await api.subscribeToGenre(genre, userId);
```

### **Backend Endpoint-i:**

- `POST /api/subscriptions/subscribe-artist?artistId={id}&userId={id}`
- `POST /api/subscriptions/subscribe-genre?genre={name}&userId={id}`

### **Autentifikacija:**

- Obavezna autentifikacija (JWT token)
- `userId` se automatski uzima iz JWT tokena (API Gateway)
- Frontend šalje `userId` u query parametrima (za kompatibilnost)

---

## 🎨 UI Elementi

### **ArtistDetail.js:**
- Dugme sa ikonom 🔔
- Pozicionirano pored naslova umetnika
- Disabled stanje tokom pretplaćivanja

### **Songs.js:**
- Mala ikona 🔔 pored dropdown-a
- Vidljiva samo kada je žanr izabran
- Tooltip sa opisom

---

## ✅ Testiranje

### **1. Test Pretplate na Umetnika:**

```powershell
# 1. Prijavite se kao korisnik
# 2. Idite na /artists/:id
# 3. Kliknite "Pretplati se"
# 4. Proverite poruku o uspehu
```

### **2. Test Pretplate na Žanr:**

```powershell
# 1. Prijavite se kao korisnik
# 2. Idite na /songs
# 3. Izaberite žanr (npr. "Pop")
# 4. Kliknite ikonu 🔔
# 5. Proverite poruku o uspehu
```

### **3. Provera Backend Logova:**

```powershell
# Proveri logove subscriptions-service
docker-compose logs subscriptions-service | Select-String -Pattern "subscribed"
```

---

## 🐛 Troubleshooting

### Problem: "Morate biti prijavljeni"
**Rešenje:** Prijavite se pre pokušaja pretplate

### Problem: "Greška pri pretplati"
**Rešenje:** 
- Proverite da li je subscriptions-service pokrenut
- Proverite logove: `docker-compose logs subscriptions-service`
- Proverite da li je API Gateway pokrenut

### Problem: Dugme se ne pojavljuje
**Rešenje:**
- Proverite da li ste prijavljeni
- Za žanr: Proverite da li ste izabrali žanr iz dropdown-a

---

## 📋 Checklist

- [x] Dodate API metode u `api.js`
- [x] Dodato dugme za pretplatu na umetnika u `ArtistDetail.js`
- [x] Dodato dugme za pretplatu na žanr u `Songs.js`
- [x] Error handling implementiran
- [x] Success poruke implementirane
- [x] Autentifikacija proverena
- [ ] Testirano u browser-u
- [ ] Backend logovi provereni

---

## 🚀 Sledeći Koraci (Opciono)

1. **Dodati pregled pretplata** - Stranica sa listom svih pretplata korisnika
2. **Dodati otkazivanje pretplate** - Dugme za otkazivanje pretplate
3. **Dodati notifikacije** - Automatske notifikacije kada umetnik/žanr dobije novi sadržaj
4. **Dodati bazu podataka** - Čuvanje pretplata u MongoDB umesto samo log-ovanja
