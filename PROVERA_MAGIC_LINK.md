# 📋 Vodič za Proveru Povraćaja Naloga - Magic Link

## ✅ Status Implementacije

**Povraćaj naloga pomoću magičnog linka je POTPUNO IMPLEMENTIRAN:**

### Implementirane Funkcionalnosti:

1. ✅ **Zahtev za Magic Link** - Korisnik unosi email adresu
2. ✅ **Generisanje sigurnog tokena** - 32 bajta, base64 encoded
3. ✅ **Slanje email-a sa magic link-om** - Link vodi na frontend
4. ✅ **Verifikacija magic link-a** - Automatska prijava korisnika
5. ✅ **Kratkotrajni link** - Magic link ističe nakon **15 minuta**
6. ✅ **Jednokratna upotreba** - Token se briše nakon korišćenja
7. ✅ **Provera statusa naloga** - Proverava da li je nalog zaključan ili lozinka istekla
8. ✅ **Automatska prijava** - Korisnik se automatski prijavljuje nakon klika na link

---

## 🧪 Kako Proveriti Funkcionalnost

### Metoda 1: Preko Frontend Aplikacije (Preporučeno)

#### Korak 1: Pokrenite Sistem
```powershell
docker-compose up -d
Start-Sleep -Seconds 20
```

#### Korak 2: Pokrenite Frontend
```powershell
cd frontend
npm start
```

---

## 📝 Test: Povraćaj Naloga sa Magic Link-om

### Korak 1: Otvorite Povraćaj Naloga
1. Otvorite `http://localhost:3000/recover-account`
2. Ili kliknite na "Povraćaj naloga (Magic Link)" na login stranici

### Korak 2: Zatražite Magic Link
1. Unesite email adresu registrovanog korisnika (npr. `admin@example.com`)
2. Kliknite "Pošalji magic link"
3. **Očekivano**: 
   - Poruka "Ako email postoji, magic link je poslat na vašu email adresu..."
   - Poruka se prikazuje i ako email ne postoji (security best practice)

### Korak 3: Proverite Magic Link u Logovima
1. **Proverite konzolu servera** za magic link:
   ```powershell
   docker-compose logs users-service | Select-String "magic"
   ```
2. Link bi trebao biti: `http://localhost:3000/verify-magic-link?token=...`
3. Token je **base64 encoded** i siguran (32 bajta)

### Korak 4: Kliknite na Magic Link
1. Kopirajte link iz logova ili otvorite direktno u browseru
2. **Očekivano**: 
   - Automatska verifikacija tokena
   - Poruka "Uspešno ste se prijavili pomoću magic link-a!"
   - Automatska prijava korisnika
   - Preusmeravanje na početnu stranicu

### Test Scenariji:

#### ✅ Uspešan Povraćaj:
- Email postoji → Magic link se šalje
- Klik na link → Automatska prijava
- Token se briše nakon korišćenja

#### ❌ Nevažeći Email:
- Email ne postoji → Ista poruka (security best practice)
- Ne otkriva da li email postoji ili ne

#### ❌ Istekao Magic Link:
- Čekajte više od 15 minuta
- Kliknite na link → Greška "invalid or expired magic link"

#### ❌ Nevažeći Token:
- Promenite token u URL-u → Greška "invalid or expired magic link"

#### ❌ Zaključan Nalog:
- Ako je nalog zaključan → Greška "account locked"

#### ❌ Istekla Lozinka:
- Ako je lozinka istekla → Greška "password expired"

---

## 🔍 Provera Preko API-ja (curl/Postman)

### Test 1: Request Magic Link
```powershell
$body = @{
    email = "admin@example.com"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8081/api/users/recover/request" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

**Očekivani odgovor:**
```json
{
  "message": "if email exists, magic link has been sent"
}
```

### Test 2: Verify Magic Link
```powershell
# Prvo proverite token iz logova
$token = "magic-link-token-from-email"

# Token se prosleđuje kao query parameter
Invoke-RestMethod -Uri "http://localhost:8081/api/users/recover/verify?token=$token" `
    -Method GET
```

**Očekivani odgovor:**
```json
{
  "token": "jwt-token-here",
  "id": "user-id",
  "username": "admin",
  "email": "admin@example.com",
  "firstName": "Admin",
  "lastName": "User",
  "role": "ADMIN"
}
```

**Napomena**: Token se automatski briše nakon korišćenja, tako da drugi poziv sa istim tokenom neće raditi.

---

## 📁 Relevantni Fajlovi

### Frontend:
- `frontend/src/components/RecoverAccount.js` - Forma za zahtev magic link-a
- `frontend/src/components/VerifyMagicLink.js` - Verifikacija i automatska prijava
- `frontend/src/services/api.js` - API pozivi (`requestMagicLink`, `verifyMagicLink`)

### Backend:
- `services/users-service/internal/handler/magic_link_handler.go` - Handler za magic link
- `services/users-service/internal/security/magic_link.go` - Generisanje i provera tokena
- `services/users-service/internal/store/user_repository.go` - Čuvanje magic link-a u bazi

---

## 🔐 Sigurnosne Karakteristike

### 1. Siguran Token
- **Dužina**: 32 bajta (256 bita)
- **Kodiranje**: Base64 URL encoding
- **Generisanje**: Kriptografski siguran random (`crypto/rand`)

### 2. Kratkotrajni Link
- **Važenje**: 15 minuta
- **Provera**: `IsMagicLinkExpired()` funkcija

### 3. Jednokratna Upotreba
- Token se **briše** nakon uspešne verifikacije
- Ne može se koristiti više puta

### 4. Bezbednost Email-a
- Ne otkriva da li email postoji ili ne
- Ista poruka za sve zahteve (security best practice)

### 5. Provera Statusa Naloga
- Proverava da li je nalog zaključan
- Proverava da li je lozinka istekla
- Onemogućava prijavu ako je bilo koji uslov ispunjen

---

## ✅ Checklist za Proveru

### Zahtev Magic Link-a:
- [ ] Forma za unos email adrese
- [ ] Validacija email formata
- [ ] Slanje zahteva na backend
- [ ] Poruka o uspehu (ne otkriva da li email postoji)
- [ ] Magic link se generiše i šalje na email

### Verifikacija Magic Link-a:
- [ ] Automatska verifikacija tokena iz URL-a
- [ ] Provera da li token postoji u bazi
- [ ] Provera da li token nije istekao (15 minuta)
- [ ] Provera statusa naloga (zaključan, istekla lozinka)
- [ ] Generisanje JWT tokena
- [ ] Automatska prijava korisnika
- [ ] Brisanje korišćenog tokena
- [ ] Preusmeravanje na početnu stranicu

### Sigurnost:
- [ ] Token je siguran (32 bajta, base64)
- [ ] Token ističe nakon 15 minuta
- [ ] Token se briše nakon korišćenja
- [ ] Ne otkriva da li email postoji
- [ ] Provera zaključanog naloga
- [ ] Provera istekle lozinke

---

## 🐛 Troubleshooting

### Problem: Magic Link se ne šalje
- Proverite logove: `docker-compose logs users-service`
- Email funkcionalnost možda koristi mock implementaciju
- Proverite `services/users-service/internal/mail/mailer.go`

### Problem: Magic Link ne radi
- Proverite da li je token ispravno URL-encoded
- Proverite da li je token istekao (15 minuta)
- Proverite logove za detalje
- Proverite da li je token već korišćen (briše se nakon korišćenja)

### Problem: "invalid or expired magic link"
- Token možda nije ispravno dekodovan iz URL-a
- Token možda ističe (15 minuta)
- Token možda već korišćen (jednokratna upotreba)
- Proverite da li token postoji u bazi podataka

### Problem: "account locked" ili "password expired"
- Nalog je zaključan ili lozinka je istekla
- Magic link ne može se koristiti u ovim slučajevima
- Korisnik mora prvo rešiti problem sa nalogom

---

## 📝 Napomene

- **Magic link ističe**: Nakon 15 minuta
- **Jednokratna upotreba**: Token se briše nakon korišćenja
- **Automatska prijava**: Korisnik se automatski prijavljuje nakon klika
- **Bezbednost**: Ne otkriva da li email postoji ili ne
- **Provera statusa**: Proverava zaključan nalog i isteklu lozinku

---

## 🎯 Razlika između Magic Link i Reset Lozinke

| Karakteristika | Magic Link | Reset Lozinke |
|----------------|-----------|---------------|
| **Svrha** | Povraćaj naloga / Prijava | Promena lozinke |
| **Važenje** | 15 minuta | 1 sat |
| **Rezultat** | Automatska prijava | Promena lozinke |
| **Korak** | Jedan klik | Unos nove lozinke |
| **Endpoint** | `/recover/verify` | `/password/reset` |

---

## 🔄 Tok Radnje Magic Link-a

```
1. Korisnik unosi email
   ↓
2. Sistem generiše siguran token (32 bajta)
   ↓
3. Token se čuva u bazi sa vremenom isteka (15 min)
   ↓
4. Email sa magic link-om se šalje korisniku
   ↓
5. Korisnik klikne na link
   ↓
6. Frontend poziva API sa tokenom
   ↓
7. Backend proverava token (postoji, nije istekao)
   ↓
8. Backend proverava status naloga (zaključan, istekla lozinka)
   ↓
9. Backend generiše JWT token
   ↓
10. Backend briše magic link token (jednokratna upotreba)
    ↓
11. Frontend automatski prijavljuje korisnika
    ↓
12. Preusmeravanje na početnu stranicu
```

---

## 💡 Primer Korišćenja

### Scenario: Korisnik je zaboravio lozinku i želi da se prijavi

1. **Korisnik ide na**: `http://localhost:3000/recover-account`
2. **Unosi email**: `user@example.com`
3. **Klikne**: "Pošalji magic link"
4. **Dobija poruku**: "Ako email postoji, magic link je poslat..."
5. **Proverava email** (ili logove servera)
6. **Klikne na magic link**: `http://localhost:3000/verify-magic-link?token=...`
7. **Automatski se prijavljuje** bez unošenja lozinke
8. **Preusmeravanje** na početnu stranicu

**Prednosti**:
- ✅ Brzo i jednostavno
- ✅ Ne zahteva pamćenje lozinke
- ✅ Sigurno (kratkotrajni token)
- ✅ Jednokratna upotreba
