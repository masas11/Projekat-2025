# 🧪 Test CQRS (2.9) - Kratko

## ✅ Šta je Implementirano

1. **Subscription model** - Dodato `ArtistName` polje (denormalizacija)
2. **subscribe-artist endpoint** - Uzima ime umetnika iz content-service i čuva ga
3. **subscriptions endpoint** - Vraća pretplate SA imenom umetnika (bez poziva content-service)

---

## 🧪 Kako Testirati

### 1. Pokreni test skriptu:
```powershell
.\test-cqrs-2.9.ps1
```

### 2. Testiraj kroz frontend:

**Korak 1:** Prijavi se kao korisnik (ne admin)
```
http://localhost:3000
```

**Korak 2:** Otvori stranicu umetnika
```
http://localhost:3000/artists/artist1
```

**Korak 3:** Klikni "Pretplati se"
- Ovo kreira pretplatu SA imenom umetnika (CQRS write)

**Korak 4:** Otvori Profile stranicu
```
http://localhost:3000/profile
```

**Korak 5:** Proveri da li se prikazuje ime umetnika uz pretplatu
- Treba da vidiš: "Pretplaćen na: [Ime Umetnika]"

---

## 🔍 Provera u Developer Tools (F12)

1. **Network tab** → Otvori Profile stranicu
2. Proveri zahtev: `GET /api/subscriptions?userId=...`
3. **Proveri response:**
   ```json
   [
     {
       "id": "...",
       "userId": "...",
       "type": "artist",
       "artistId": "artist1",
       "artistName": "Michael Jackson",  // ← OVO TREBA DA POSTOJI
       ...
     }
   ]
   ```
4. **Proveri da NEMA poziva ka:** `/api/content/artists/{id}` pri čitanju pretplata!

---

## ✅ Checklist

- [ ] Subscription model ima `ArtistName` polje
- [ ] subscribe-artist endpoint čuva `ArtistName`
- [ ] subscriptions endpoint vraća `ArtistName` bez poziva content-service
- [ ] Frontend prikazuje ime umetnika uz pretplatu

---

**Napomena:** CQRS znači da se ime umetnika čuva u subscriptions bazi prilikom kreiranja pretplate, tako da se ne mora pozivati content-service svaki put pri čitanju pretplata.
