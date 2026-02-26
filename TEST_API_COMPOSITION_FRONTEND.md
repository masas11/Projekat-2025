# 🧪 Test API Composition u Frontendu (2.8)

## 🎯 Kako Testirati

### 1. Pokreni Servise
```powershell
docker-compose up -d
```

### 2. Pokreni Frontend
```powershell
cd frontend
npm start
```

Frontend će biti dostupan na: `http://localhost:3000`

---

## ✅ Šta Proveriti

### 1. **Stranica sa Pesmama** (`/songs`)
- Otvori: `http://localhost:3000/songs`
- **Proveri:** Svaka pesma treba da prikazuje:
  - ⭐ **Prosečna ocena:** (npr. 4.5)
  - **Broj ocena:** (npr. 10 ocena)
  - Ili "Još nema ocena" ako nema ocena

### 2. **Stranica sa Albumom** (`/albums/:id`)
- Otvori: `http://localhost:3000/albums/album1` (ili bilo koji album)
- **Proveri:** Svaka pesma u albumu treba da prikazuje:
  - ⭐ **Prosečna ocena** i **broj ocena**

### 3. **Browser Developer Tools**
- Otvori **F12** → **Network** tab
- Učitaj stranicu sa pesmama
- Proveri zahtev: `GET http://localhost:8081/api/content/songs`
- **Proveri response:** Treba da sadrži `averageRating` i `ratingCount` za svaku pesmu

---

## 🔍 Primer Odgovora iz API-ja

```json
[
  {
    "id": "song1",
    "name": "Billie Jean",
    "duration": 294,
    "genre": "Pop",
    "averageRating": 4.5,
    "ratingCount": 10,
    ...
  }
]
```

---

## ⚠️ Ako Ne Radi

### Problem: Ne prikazuje se ocena
**Rešenje:**
1. Proveri da li API vraća `averageRating` i `ratingCount`:
   ```powershell
   Invoke-WebRequest -Uri "http://localhost:8081/api/content/songs" -UseBasicParsing | Select-Object -ExpandProperty Content
   ```

2. Proveri da li su servisi pokrenuti:
   ```powershell
   docker-compose ps
   ```

3. Proveri logove:
   ```powershell
   docker-compose logs api-gateway --tail 50
   docker-compose logs ratings-service --tail 50
   ```

### Problem: Prikazuje "Još nema ocena"
**Rešenje:** To je OK ako pesma stvarno nema ocena. Dodaj ocenu:
1. Prijavi se kao korisnik (ne admin)
2. Otvori pesmu
3. Oceni pesmu (1-5)
4. Osveži stranicu - trebalo bi da se prikaže ocena

---

## 📋 Checklist

- [ ] Frontend se pokreće bez grešaka
- [ ] Stranica `/songs` prikazuje pesme
- [ ] Svaka pesma prikazuje `averageRating` i `ratingCount`
- [ ] Stranica sa albumom prikazuje ocene za pesme
- [ ] API vraća `averageRating` i `ratingCount` u response-u
- [ ] Ako nema ocena, prikazuje se "Još nema ocena"

---

**Napomena:** API Composition automatski kombinuje podatke iz content-service i ratings-service. Nema potrebe za dodatnim pozivima iz frontenda!
