# 📋 Vodič za Proveru Prijave, Odjave i Promene Lozinke

## ✅ Status Implementacije

**Sve funkcionalnosti su POTPUNO IMPLEMENTIRANE:**

### 1. ✅ Kombinovana Autentifikacija (Lozinka + OTP)
- Korisnik unosi username i lozinku
- Sistem šalje OTP kod na email adresu
- Korisnik unosi OTP kod za završetak prijave
- OTP ističe nakon 5 minuta

### 2. ✅ Odjava sa Sistema
- Implementirana na frontendu i backendu
- Briše token i korisničke podatke iz localStorage

### 3. ✅ Promena Lozinke
- Korisnik mora uneti staru lozinku
- **Lozinka mora biti bar 1 dan stara** pre promene
- Validacija jake lozinke
- Postavlja novi `PasswordExpiresAt` (60 dana)

### 4. ✅ Email-bazirani Reset Lozinke
- Korisnik unosi email adresu
- Sistem šalje kratkotrajni link (ističe za 1 sat)
- Link vodi na formu za unos nove lozinke
- Validacija jake lozinke

### 5. ✅ Auditabilnost - Onemogućavanje Prijave
- **Provera isteka lozinke**: Ako je `PasswordExpiresAt < now()`, prijava je onemogućena
- **Maksimalni period važenja**: 60 dana (konfigurabilno)
- **Poruka**: "password expired" (HTTP 403)

---

## 🧪 Kako Proveriti Funkcionalnosti

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

## 📝 Test 1: Kombinovana Autentifikacija (Lozinka + OTP)

### Korak 1: Otvorite Prijavu
1. Otvorite `http://localhost:3000/login`
2. Unesite:
   - **Username**: `admin` (ili bilo koji registrovan korisnik)
   - **Lozinka**: `admin123` (ili odgovarajuća lozinka)

### Korak 2: Zatražite OTP
1. Kliknite "Zatraži OTP"
2. **Očekivano**: Forma se menja i traži OTP kod
3. **Proverite konzolu servera** - OTP kod se ispisuje u logovima:
   ```powershell
   docker-compose logs users-service | Select-String "OTP"
   ```

### Korak 3: Unesite OTP
1. Unesite OTP kod iz logova
2. Kliknite "Verifikuj OTP"
3. **Očekivano**: Uspešna prijava, preusmeravanje na početnu stranicu

### Test Scenariji:
- ❌ **Pogrešna lozinka** → Greška "invalid credentials"
- ❌ **Nevažeći OTP** → Greška "invalid OTP"
- ❌ **Istekao OTP** (nakon 5 minuta) → Greška "invalid OTP"
- ❌ **Ne-verifikovan email** → Greška "email not verified"
- ❌ **Istekla lozinka** → Greška "password expired"

---

## 📝 Test 2: Odjava sa Sistema

### Korak 1: Prijavite se
- Prijavite se koristeći Test 1

### Korak 2: Odjavite se
1. Kliknite na "Odjavi se" u navigaciji (gornji desni ugao)
2. **Očekivano**: 
   - Preusmeravanje na `/login`
   - Token i korisnički podaci obrisani
   - Navigacija više ne prikazuje korisničke opcije

### Provera:
- Otvorite Developer Tools (F12) → Application → Local Storage
- **Očekivano**: `token` i `user` su obrisani

---

## 📝 Test 3: Promena Lozinke (Mora biti bar 1 dan stara)

### Korak 1: Prijavite se
- Prijavite se sa postojećim korisnikom

### Korak 2: Otvorite Promenu Lozinke
1. Kliknite na "Promena lozinke" u navigaciji
2. Ili otvorite direktno: `http://localhost:3000/change-password`

### Korak 3: Pokušajte Promenu (Ako je lozinka < 1 dan stara)
1. Unesite:
   - Stara lozinka: `admin123`
   - Nova lozinka: `NewPass123`
   - Potvrdi novu lozinku: `NewPass123`
2. Kliknite "Promeni lozinku"
3. **Očekivano**: 
   - Ako je lozinka promenjena pre manje od 24 sata → Greška "password too new"
   - Ako je lozinka stara → Uspešna promena

### Test Scenariji:
- ❌ **Lozinka < 1 dan stara** → Greška "password too new" (HTTP 403)
- ❌ **Pogrešna stara lozinka** → Greška "wrong password"
- ❌ **Slaba nova lozinka** → Greška o validaciji lozinke
- ✅ **Uspešna promena** → Poruka "Lozinka je uspešno promenjena!"

### Simulacija za Testiranje:
Da biste testirali proveru "1 dan stara", možete:
1. Promeniti lozinku jednom
2. Pokušati ponovo odmah → trebalo bi da dobijete grešku
3. Ili promeniti `PasswordChangedAt` u bazi podataka na stariji datum

---

## 📝 Test 4: Email-bazirani Reset Lozinke

### Korak 1: Otvorite Zaboravljenu Lozinku
1. Otvorite `http://localhost:3000/forgot-password`
2. Ili kliknite "Zaboravljena lozinka?" na login stranici

### Korak 2: Zatražite Reset Link
1. Unesite email adresu registrovanog korisnika
2. Kliknite "Pošalji link za reset"
3. **Očekivano**: Poruka "Ako email postoji, link za reset lozinke je poslat..."

### Korak 3: Proverite Email Link
1. **Proverite konzolu servera** za reset link:
   ```powershell
   docker-compose logs users-service | Select-String "reset"
   ```
2. Link bi trebao biti: `http://localhost:3000/reset-password?token=...`
3. Token ističe nakon **1 sata**

### Korak 4: Resetujte Lozinku
1. Otvorite reset link u browseru
2. Unesite novu lozinku koja ispunjava kriterijume (npr. `NewPass123`)
3. Potvrdite lozinku
4. Kliknite "Resetuj lozinku"
5. **Očekivano**: 
   - Poruka "Lozinka je uspešno promenjena!"
   - Preusmeravanje na login stranicu

### Test Scenariji:
- ❌ **Istekao token** (nakon 1 sata) → Greška "invalid or expired reset token"
- ❌ **Nevažeći token** → Greška "invalid or expired reset token"
- ❌ **Slaba lozinka** → Greška o validaciji lozinke
- ✅ **Uspešan reset** → Poruka o uspehu

---

## 📝 Test 5: Auditabilnost - Onemogućavanje Prijave nakon Isteka Lozinke

### Simulacija Isteka Lozinke

#### Opcija A: Promenite Konfiguraciju (Za Testiranje)
```powershell
# U docker-compose.yml ili .env fajlu, postavite:
PASSWORD_EXPIRATION_DAYS=0  # Lozinka ističe odmah
```

#### Opcija B: Promenite u Bazi Podataka
```powershell
# Povežite se na MongoDB
docker exec -it mongodb-users mongosh

# U MongoDB shell-u:
use users_db
db.users.updateOne(
  {username: "admin"},
  {$set: {passwordExpiresAt: new Date(Date.now() - 86400000)}}  # -1 dan
)
```

### Korak 1: Pokušajte Prijavu
1. Otvorite `http://localhost:3000/login`
2. Unesite username i lozinku
3. Kliknite "Zatraži OTP"
4. **Očekivano**: 
   - Greška "password expired" (HTTP 403)
   - Prijava je **onemogućena**

### Provera u Kodu:
- `services/users-service/internal/handler/login_handler.go` linija 56-58:
  ```go
  if time.Now().After(user.PasswordExpiresAt) {
      http.Error(w, "password expired", http.StatusForbidden)
      return
  }
  ```

### Maksimalni Period Važenja:
- **Podrazumevano**: 60 dana
- **Konfigurabilno**: Preko `PASSWORD_EXPIRATION_DAYS` u `config/config.go`
- **Postavlja se**: Pri registraciji i promeni lozinke

---

## 🔍 Provera Preko API-ja (curl/Postman)

### Test 1: Request OTP
```powershell
$body = @{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8081/api/users/login/request-otp" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

**Očekivani odgovor:** HTTP 200 (OTP poslat na email)

### Test 2: Verify OTP
```powershell
# Prvo proverite OTP iz logova
$otp = "123456"  # Zamenite sa stvarnim OTP kodom

$body = @{
    username = "admin"
    otp = $otp
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8081/api/users/login/verify-otp" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

**Očekivani odgovor:** JSON sa tokenom i korisničkim podacima

### Test 3: Logout
```powershell
$token = "your-jwt-token"  # Zamenite sa stvarnim tokenom

Invoke-RestMethod -Uri "http://localhost:8081/api/users/logout" `
    -Method POST `
    -Headers @{Authorization = "Bearer $token"} `
    -ContentType "application/json"
```

**Očekivani odgovor:** `{"message": "logged out successfully"}`

### Test 4: Change Password (Mora biti bar 1 dan stara)
```powershell
$token = "your-jwt-token"

$body = @{
    username = "admin"
    oldPassword = "admin123"
    newPassword = "NewPass123"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8081/api/users/password/change" `
    -Method POST `
    -Headers @{Authorization = "Bearer $token"} `
    -ContentType "application/json" `
    -Body $body
```

**Očekivani odgovor:**
- Ako lozinka < 1 dan stara: HTTP 403 "password too new"
- Ako uspešno: `{"message": "password changed successfully"}`

### Test 5: Request Password Reset
```powershell
$body = @{
    email = "admin@example.com"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8081/api/users/password/reset/request" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

**Očekivani odgovor:** `{"message": "if email exists, password reset link has been sent"}`

### Test 6: Reset Password
```powershell
# Prvo proverite token iz logova
$token = "reset-token-from-email"

$body = @{
    token = $token
    newPassword = "NewPass123"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8081/api/users/password/reset" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

**Očekivani odgovor:** `{"message": "password reset successfully"}`

---

## 📁 Relevantni Fajlovi

### Frontend:
- `frontend/src/components/Login.js` - Prijava sa OTP
- `frontend/src/components/ChangePassword.js` - Promena lozinke
- `frontend/src/components/ForgotPassword.js` - Zatraži reset link
- `frontend/src/components/ResetPassword.js` - Reset lozinke
- `frontend/src/components/Navbar.js` - Logout dugme
- `frontend/src/context/AuthContext.js` - Logout funkcija

### Backend:
- `services/users-service/internal/handler/login_handler.go` - OTP autentifikacija i logout
- `services/users-service/internal/handler/password_handler.go` - Promena i reset lozinke
- `services/users-service/internal/security/otp.go` - Generisanje OTP koda
- `services/users-service/config/config.go` - Konfiguracija (password expiration days)

---

## ✅ Checklist za Proveru

### Kombinovana Autentifikacija:
- [ ] Korisnik može uneti username i lozinku
- [ ] Sistem šalje OTP na email
- [ ] Korisnik može uneti OTP kod
- [ ] Uspešna prijava sa ispravnim OTP-om
- [ ] Greška sa pogrešnim OTP-om
- [ ] Greška sa isteklim OTP-om (nakon 5 minuta)
- [ ] Greška ako lozinka nije ispravna
- [ ] Greška ako email nije verifikovan
- [ ] Greška ako je lozinka istekla

### Odjava:
- [ ] Logout dugme u navigaciji
- [ ] Logout briše token iz localStorage
- [ ] Logout briše korisničke podatke
- [ ] Preusmeravanje na login stranicu

### Promena Lozinke:
- [ ] Forma za promenu lozinke
- [ ] Provera da lozinka mora biti bar 1 dan stara
- [ ] Validacija jake lozinke
- [ ] Provera stare lozinke
- [ ] Uspešna promena lozinke
- [ ] Postavljanje novog `PasswordExpiresAt`

### Reset Lozinke:
- [ ] Forma za zatraživanje reset linka
- [ ] Email sa reset linkom se šalje
- [ ] Reset link ističe nakon 1 sata
- [ ] Forma za unos nove lozinke
- [ ] Validacija jake lozinke
- [ ] Uspešan reset lozinke

### Auditabilnost:
- [ ] Provera `PasswordExpiresAt` pri prijavi
- [ ] Onemogućavanje prijave ako je lozinka istekla
- [ ] Maksimalni period važenja: 60 dana
- [ ] Konfigurabilno preko environment varijable

---

## 🐛 Troubleshooting

### Problem: OTP se ne šalje
- Proverite logove: `docker-compose logs users-service`
- Email funkcionalnost možda koristi mock implementaciju
- Proverite `services/users-service/internal/mail/mailer.go`

### Problem: "password too new" greška
- Lozinka mora biti promenjena pre najmanje 24 sata
- Proverite `PasswordChangedAt` u bazi podataka
- Za testiranje, možete promeniti datum u bazi

### Problem: Reset link ne radi
- Proverite da li je token ispravno URL-encoded
- Proverite da li je token istekao (1 sat)
- Proverite logove za detalje

### Problem: "password expired" greška
- Proverite `PasswordExpiresAt` u bazi podataka
- Podrazumevano je 60 dana od poslednje promene
- Možete promeniti preko `PASSWORD_EXPIRATION_DAYS`

---

## 📝 Napomene

- **OTP ističe**: Nakon 5 minuta
- **Reset token ističe**: Nakon 1 sata
- **Lozinka mora biti stara**: Najmanje 24 sata pre promene
- **Maksimalni period važenja lozinke**: 60 dana (konfigurabilno)
- **Auditabilnost**: Prijava je onemogućena ako je `PasswordExpiresAt < now()`
- **Sve lozinke**: Čuvaju se kao bcrypt hash

---

## 🎯 Simulacija za Demonstraciju

Za demonstraciju na odbrani, možete simulirati kraće periode:

### Simulacija Isteka Lozinke (1 dan):
```powershell
# U docker-compose.yml ili .env:
PASSWORD_EXPIRATION_DAYS=1
```

### Simulacija "Lozinka mora biti stara" (1 sat):
U `password_handler.go` linija 57, promenite:
```go
// Umesto 24*time.Hour, koristite:
if time.Since(user.PasswordChangedAt) < 1*time.Hour {
```

**Napomena**: Vratite na 24 sata pre produkcije!
