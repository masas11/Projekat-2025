# 📝 Ručno Testiranje - Korak po Korak

**Vodič za ručno testiranje svih zahteva iz Informacione bezbednosti**

---

## 🔧 PRIprema

### Korak 1: Pokretanje Sistema

1. Otvorite PowerShell terminal
2. Navigirajte do projekta:
   ```
   cd C:\Users\ivana\Desktop\Projekat-2025-1
   ```
3. Pokrenite Docker kontejnere:
   ```
   docker-compose up -d
   ```
4. Sačekajte 30 sekundi da se servisi pokrenu
5. Proverite da li su svi kontejneri pokrenuti:
   ```
   docker-compose ps
   ```
   **Očekivano:** Svi servisi treba da budu "Up"

### Korak 2: Pokretanje Frontend-a

1. Otvorite novi PowerShell terminal
2. Navigirajte do frontend direktorijuma:
   ```
   cd C:\Users\ivana\Desktop\Projekat-2025-1\frontend
   ```
3. Pokrenite frontend:
   ```
   npm start
   ```
4. Otvorite browser na `http://localhost:3000`

### Korak 3: Učitavanje Helper Funkcija

1. U prvom PowerShell terminalu (gde ste pokrenuli Docker):
   ```
   . .\https-helper.ps1
   ```
   **Očekivano:** Nema greške, funkcija je učitana

---

## ✅ TEST 1: Registracija Naloga (1.1)

### Test 1.1: Uspešna Registracija

**Šta testirate:** Da li registracija radi i validira podatke

**Koraci:**

1. Otvorite browser na `http://localhost:3000/register`
2. Popunite formu:
   - **Ime:** Test
   - **Prezime:** User
   - **Email:** testuser@example.com
   - **Username:** testuser
   - **Lozinka:** Test1234!
   - **Ponovljena lozinka:** Test1234!
3. Kliknite "Registruj se"
4. **Očekivano:** Poruka "Uspešna registracija! Email za verifikaciju je poslat..."

**Provera u kodu:**
- Otvorite: `services/users-service/internal/handler/register.go`
- Linija 32-218: `Register()` funkcija
- Proverite da se pozivaju:
  - `validation.ValidateEmail()` (linija 51)
  - `validation.ValidateUsername()` (linija 60)
  - `validation.IsStrongPassword()` (linija 147)
  - `bcrypt.GenerateFromPassword()` (linija 156)

**Šta proveriti:**
- ✅ Email je validan format
- ✅ Username je jedinstven
- ✅ Lozinka je jaka (8+ karaktera, veliko/malo/broj/specijalni)
- ✅ Lozinka je heširana u bazi (bcrypt format)

---

### Test 1.2: Validacija - Jedinstven Username

**Šta testirate:** Da li sistem sprečava duplikate username-a

**Koraci:**

1. Pokušajte ponovo da se registrujete sa istim username-om (`testuser`)
2. **Očekivano:** Poruka greške "Korisnik sa ovim korisničkim imenom ili email adresom već postoji"

**Provera u kodu:**
- `services/users-service/internal/handler/register.go:174-180`
- Provera: `err == store.ErrUserExists`
- Status: HTTP 409 Conflict

---

### Test 1.3: Validacija - Jaka Lozinka

**Šta testirate:** Da li sistem zahteva jaku lozinku

**Koraci:**

1. Pokušajte registraciju sa slabom lozinkom:
   - **Lozinka:** 123
   - **Očekivano:** Poruka greške "Lozinka mora imati najmanje 8 karaktera"
2. Pokušajte sa lozinkom bez velikog slova:
   - **Lozinka:** test1234!
   - **Očekivano:** Poruka greške "Lozinka mora sadržati najmanje jedno veliko slovo"
3. Pokušajte sa lozinkom bez broja:
   - **Lozinka:** TestPassword!
   - **Očekivano:** Poruka greške "Lozinka mora sadržati najmanje jedan broj"

**Provera u kodu:**
- `services/users-service/internal/validation/password.go`
- `IsStrongPassword()` funkcija proverava:
  - Minimum 8 karaktera
  - Najmanje jedno veliko slovo
  - Najmanje jedno malo slovo
  - Najmanje jedan broj
  - Najmanje jedan specijalni karakter

---

### Test 1.4: Email Verifikacija

**Šta testirate:** Da li se šalje email za verifikaciju

**Koraci:**

1. Nakon registracije, proverite logove servera:
   ```
   docker logs projekat-2025-2-users-service-1 | Select-String "verification"
   ```
2. **Očekivano:** Vidite verification link u logovima
3. Ili proverite response u browser Developer Tools (F12 → Network)
4. **Očekivano:** Response sadrži `verificationLink` (ako SMTP nije konfigurisan)

**Provera u kodu:**
- `services/users-service/internal/handler/register.go:184-200`
- Generiše se `verificationToken`
- Šalje se email sa linkom
- URL encoding za sigurno slanje tokena

---

## ✅ TEST 2: Prijava na Sistem (1.2)

### Test 2.1: Kombinovana Autentifikacija (Lozinka + OTP)

**Šta testirate:** Da li kombinovana autentifikacija radi

**Koraci:**

1. Otvorite browser na `http://localhost:3000/login`
2. Unesite:
   - **Username:** admin
   - **Lozinka:** admin123
3. Kliknite "Zatraži OTP"
4. **Očekivano:** Forma se menja i traži OTP kod
5. Proverite logove za OTP kod:
   ```
   docker logs projekat-2025-2-users-service-1 | Select-String "OTP"
   ```
6. Unesite OTP kod iz logova
7. Kliknite "Potvrdi OTP"
8. **Očekivano:** Uspešna prijava, preusmeravanje na početnu stranicu

**Provera u kodu:**
- `services/users-service/internal/handler/login_handler.go:33-108` - RequestOTP
- `services/users-service/internal/handler/login_handler.go:110-174` - VerifyOTP
- Proverite:
  - Provera lozinke (bcrypt)
  - Generisanje OTP koda
  - Slanje email-a sa OTP kodom
  - Generisanje JWT tokena nakon uspešne verifikacije

---

### Test 2.2: Neuspešna Prijava - Pogrešna Lozinka

**Šta testirate:** Da li sistem loguje neuspešne pokušaje

**Koraci:**

1. Pokušajte prijavu sa pogrešnom lozinkom:
   - **Username:** admin
   - **Lozinka:** wrongpassword
2. **Očekivano:** Poruka greške "Nevažeći podaci"
3. Proverite logove:
   ```
   docker logs projekat-2025-2-users-service-1 | Select-String "LOGIN_FAILURE"
   ```
4. **Očekivano:** Log entry sa `EventType=LOGIN_FAILURE` i `reason=invalid password`

**Provera u kodu:**
- `services/users-service/internal/handler/login_handler.go:78-89`
- `h.Logger.LogLoginFailure()` poziva se za neuspešne pokušaje

---

### Test 2.3: Account Locking - 5 Neuspešnih Pokušaja

**Šta testirate:** Da li se nalog zaključava nakon 5 neuspešnih pokušaja

**Koraci:**

1. Pokušajte prijavu sa pogrešnom lozinkom **5 puta** za isti username
2. Nakon 5. pokušaja, pokušajte ponovo
3. **Očekivano:** Poruka greške "Nalog je zaključan" (HTTP 403)
4. Proverite logove:
   ```
   docker logs projekat-2025-2-users-service-1 | Select-String "account locked|LOGIN_FAILURE"
   ```
5. **Očekivano:** Log entry sa `reason=account locked`

**Napomena:** Logovi se mogu videti preko `docker logs` komande - to je potpuno validno za Docker okruženje.

**Provera u kodu:**
- `services/users-service/internal/handler/login_handler.go:78-84`
- `user.FailedLoginAttempts++`
- `if user.FailedLoginAttempts >= 5` → `user.LockedUntil = time.Now().Add(15 * time.Minute)`

---

### Test 2.4: Promena Lozinke (Mora biti bar 1 dan stara)

**Šta testirate:** Da li sistem sprečava promenu lozinke ako nije bar 1 dan stara

**Koraci:**

1. Prijavite se kao korisnik
2. Idite na stranicu za promenu lozinke
3. Pokušajte promeniti lozinku odmah nakon registracije
4. **Očekivano:** Poruka greške "Lozinka mora biti bar 1 dan stara"

**Provera u kodu:**
- `services/users-service/internal/handler/password_handler.go`
- Provera: `time.Since(user.PasswordChangedAt) < 24*time.Hour`

---

### Test 2.5: Email-bazirani Reset Lozinke

**Šta testirate:** Da li reset lozinke radi

**Koraci:**

1. Idite na stranicu za reset lozinke
2. Unesite email adresu registrovanog korisnika
3. Kliknite "Pošalji reset link"
4. **Očekivano:** Poruka "Reset link je poslat na vašu email adresu"
5. Proverite logove za reset link:
   ```
   docker logs projekat-2025-2-users-service-1 | Select-String "reset"
   ```
6. Kliknite na reset link (ili kopirajte iz logova)
7. **Očekivano:** Forma za unos nove lozinke
8. Unesite novu lozinku i potvrdite
9. **Očekivano:** Lozinka je promenjena, možete se prijaviti sa novom lozinkom

**Provera u kodu:**
- `services/users-service/internal/handler/password_handler.go`
- Generiše se reset token (ističe za 1 sat)
- Šalje se email sa linkom

---

### Test 2.6: Auditabilnost - Istek Lozinke

**Šta testirate:** Da li sistem blokira prijavu ako je lozinka istekla

**Koraci:**

**Napomena:** Ovo zahteva direktnu manipulaciju baze podataka za simulaciju.

1. Otvorite MongoDB shell:
   ```
   docker exec -it projekat-2025-2-mongodb-users-1 mongosh
   ```
2. Postavite `PasswordExpiresAt` u prošlosti:
   ```javascript
   use music_streaming_users
   db.users.updateOne(
     {username: "testuser"},
     {$set: {passwordExpiresAt: new Date("2024-01-01")}}
   )
   ```
3. Pokušajte prijavu sa tim korisnikom
4. **Očekivano:** Poruka greške "Lozinka je istekla" (HTTP 403)
5. Proverite logove:
   ```
   docker logs projekat-2025-2-users-service-1 | Select-String "password expired|LOGIN_FAILURE"
   ```

**Napomena:** Logovi se čuvaju unutar Docker kontejnera. Koristite `docker logs` komandu za pristup.

**Provera u kodu:**
- `services/users-service/internal/handler/login_handler.go:70-76`
- Provera: `time.Now().After(user.PasswordExpiresAt)`
- Maksimalni period: 60 dana (konfigurabilno)

---

## ✅ TEST 3: Povraćaj Naloga - Magic Link (1.3)

### Test 3.1: Request Magic Link

**Šta testirate:** Da li magic link autentifikacija radi

**Koraci:**

1. Otvorite browser na `http://localhost:3000/recover-account`
2. Unesite email adresu registrovanog korisnika
3. Kliknite "Pošalji magic link"
4. **Očekivano:** Poruka "Ako email postoji, magic link je poslat..."
5. Proverite logove za magic link:
   ```
   docker logs projekat-2025-2-users-service-1 | Select-String "magic"
   ```
6. **Očekivano:** Magic link u logovima (format: `https://localhost:3000/verify-magic-link?token=...`)

**Napomena:** Logovi se mogu videti preko `docker logs` komande - to je standardni način pristupa logovima u Docker okruženju.

**Provera u kodu:**
- `services/users-service/internal/handler/magic_link_handler.go:33-81`
- Generiše se siguran token (32 bajta)
- Token se čuva u bazi (ističe za 15 minuta)
- Šalje se email sa magic link-om

---

### Test 3.2: Verify Magic Link

**Šta testirate:** Da li magic link automatski prijavljuje korisnika

**Koraci:**

1. Kopirajte magic link iz logova
2. Otvorite link u browseru
3. **Očekivano:** Automatska prijava, preusmeravanje na početnu stranicu
4. Proverite da li ste prijavljeni (treba videti korisničke podatke)

**Provera u kodu:**
- `services/users-service/internal/handler/magic_link_handler.go:84-143`
- Proverava token (postoji, nije istekao)
- Proverava status naloga (zaključan, istekla lozinka)
- Generiše JWT token
- Briše magic link token (jednokratna upotreba)

---

## ✅ TEST 4: Kontrola Pristupa (2.17)

### Test 4.1: Autorizacija - Zaštićeni Endpoint bez Tokena

**Šta testirate:** Da li sistem zahteva autentifikaciju za zaštićene endpoint-e

**Koraci:**

1. Otvorite browser Developer Tools (F12)
2. Idite na Network tab
3. Pokušajte pristup zaštićenom endpoint-u (npr. logout):
   - U browser konzoli:
     ```javascript
     fetch('https://localhost:8081/api/users/logout', {
       method: 'POST',
       headers: {'Content-Type': 'application/json'}
     })
     ```
4. **Očekivano:** HTTP 401 "authorization header required"
5. Proverite logove:
   ```
   docker logs projekat-2025-2-api-gateway-1 | Select-String "ACCESS_CONTROL_FAILURE"
   ```

**Provera u kodu:**
- `services/api-gateway/internal/middleware/auth.go:70-76`
- Provera Authorization header-a
- `log.LogAccessControlFailure()` za neuspehe

---

### Test 4.2: Autorizacija - Nevažeći Token

**Šta testirate:** Da li sistem detektuje nevažeće tokene

**Koraci:**

1. U browser konzoli:
   ```javascript
   fetch('https://localhost:8081/api/users/logout', {
     method: 'POST',
     headers: {
       'Content-Type': 'application/json',
       'Authorization': 'Bearer invalid-token'
     }
   })
   ```
2. **Očekivano:** HTTP 401 "invalid or expired token"
3. Proverite logove:
   ```
   docker logs projekat-2025-2-api-gateway-1 | Select-String "INVALID_TOKEN"
   ```

**Napomena:** Logovi se mogu videti preko `docker logs` komande.

**Provera u kodu:**
- `services/api-gateway/internal/middleware/auth.go:98-137`
- JWT validacija
- `log.LogInvalidToken()` za nevažeće tokene

---

### Test 4.3: Autorizacija - Validni Token

**Šta testirate:** Da li sistem dozvoljava pristup sa validnim tokenom

**Koraci:**

1. Prijavite se i dobijte token
2. U browser konzoli:
   ```javascript
   const token = localStorage.getItem('token');
   fetch('https://localhost:8081/api/users/logout', {
     method: 'POST',
     headers: {
       'Content-Type': 'application/json',
       'Authorization': `Bearer ${token}`
     }
   })
   ```
3. **Očekivano:** HTTP 200 "logged out successfully"

**Provera u kodu:**
- `services/api-gateway/internal/middleware/auth.go:139-152`
- JWT token validacija
- Dodavanje user claims u context

---

### Test 4.4: Role-Based Access Control (ADMIN)

**Šta testirate:** Da li sistem proverava uloge korisnika

**Koraci:**

1. Prijavite se kao **regular user** (ne admin)
2. Pokušajte kreirati artist-a (admin funkcija):
   - Idite na admin stranicu ili koristite API direktno
3. **Očekivano:** HTTP 403 "forbidden: ADMIN access required"
4. Prijavite se kao **admin** korisnik
5. Pokušajte ponovo
6. **Očekivano:** HTTP 200/201 (uspešno kreiranje)

**Provera u kodu:**
- `services/api-gateway/internal/middleware/auth.go:161-188`
- `RequireRole("ADMIN")` middleware
- Provera `claims.Role`

---

### Test 4.5: Šifrovanje i Integritet State Podataka

**Šta testirate:** Da li su podaci šifrovani i zaštićeni od manipulacije

**Koraci:**

1. Prijavite se u aplikaciju
2. Otvorite Developer Tools (F12)
3. Idite na **Application → Local Storage**
4. Proverite ključeve:
   - `user` - treba biti šifrovano (base64 encoded string)
   - `user_checksum` - treba postojati (checksum za integritet)
   - `token` - plain text (JWT je već encoded)
5. **Test manipulacije:**
   - Promenite vrednost `user` ključa (npr. dodajte karakter)
   - Osvežite stranicu (F5)
   - **Očekivano:** Podaci se brišu, korisnik se odjavljuje, preusmeravanje na login

**Provera u kodu:**
- `frontend/src/utils/encryption.js`
- `setEncryptedItem()` - šifruje podatke
- `getEncryptedItem()` - dešifruje i proverava integritet
- `calculateChecksum()` - generiše checksum

---

### Test 4.6: DoS Zaštita - Rate Limiting

**Šta testirate:** Da li rate limiting sprečava DoS napade

**Koraci:**

1. Otvorite browser Developer Tools (F12)
2. Idite na **Console** tab
3. Pokrenite sledeći kod (šalje 150 zahteva):
   ```javascript
   let success = 0;
   let blocked = 0;
   
   for (let i = 1; i <= 150; i++) {
     fetch('https://localhost:8081/api/users/health')
       .then(res => {
         if (res.status === 200) success++;
         if (res.status === 429) blocked++;
         console.log(`Request ${i}: ${res.status} (Success: ${success}, Blocked: ${blocked})`);
       });
     await new Promise(r => setTimeout(r, 100)); // 100ms delay
   }
   ```
4. **Očekivano:**
   - Prvih ~100 zahteva: HTTP 200
   - Preko 100 zahteva: HTTP 429 "too many requests"
5. Proverite logove:
   ```
   docker logs projekat-2025-2-api-gateway-1 | Select-String "too many requests"
   ```

**Provera u kodu:**
- `services/api-gateway/internal/middleware/rate_limit.go`
- Limit: 100 zahteva/min po IP adresi
- HTTP 429 za prekoračenje limita

---

## ✅ TEST 5: Validacija Podataka (2.18)

### Test 5.1: SQL Injection Detection

**Šta testirate:** Da li sistem detektuje SQL injection napade

**Koraci:**

1. Otvorite browser na `http://localhost:3000/register`
2. Pokušajte registraciju sa SQL injection payload-om:
   - **Ime:** Test' OR '1'='1
   - **Prezime:** User
   - **Email:** sqli@example.com
   - **Username:** sqliuser
   - **Lozinka:** Test1234!
   - **Ponovljena lozinka:** Test1234!
3. Kliknite "Registruj se"
4. **Očekivano:** Poruka greške "Nevažeći unos" (HTTP 400)
5. Proverite logove:
   ```
   docker logs projekat-2025-2-users-service-1 | Select-String "SQL injection|VALIDATION_FAILURE"
   ```

**Napomena:** Logovi se mogu videti preko `docker logs` komande.

**Provera u kodu:**
- `services/users-service/internal/validation/input.go:98-119`
- `CheckSQLInjection()` detektuje pattern-e:
  - `' OR '1'='1`
  - `'; DROP TABLE`
  - `UNION SELECT`
  - itd.

---

### Test 5.2: XSS Detection

**Šta testirate:** Da li sistem detektuje XSS napade

**Koraci:**

1. Pokušajte registraciju sa XSS payload-om:
   - **Ime:** <script>alert('XSS')</script>
   - **Prezime:** User
   - **Email:** xss@example.com
   - **Username:** xssuser
   - **Lozinka:** Test1234!
   - **Ponovljena lozinka:** Test1234!
2. **Očekivano:** Poruka greške "Nevažeći unos" (HTTP 400)
3. Proverite logove:
   ```
   docker logs projekat-2025-2-users-service-1 | Select-String "XSS attempt|VALIDATION_FAILURE"
   ```

**Napomena:** Logovi se mogu videti preko `docker logs` komande.

**Provera u kodu:**
- `services/users-service/internal/validation/input.go:121-144`
- `CheckXSS()` detektuje pattern-e:
  - `<script>`
  - `javascript:`
  - `onerror=`
  - `<iframe>`
  - itd.

---

### Test 5.3: Whitelisting (Username)

**Šta testirate:** Da li sistem dozvoljava samo dozvoljene karaktere

**Koraci:**

1. Pokušajte registraciju sa nevažećim karakterima u username-u:
   - **Username:** test@user!
   - **Očekivano:** Poruka greške "Username mora sadržati samo slova, brojeve i underscore"
2. Pokušajte sa validnim username-om:
   - **Username:** test_user123
   - **Očekivano:** Uspešna validacija

**Provera u kodu:**
- `services/users-service/internal/validation/input.go`
- `ValidateUsername()` - whitelist: `[a-zA-Z0-9_]`
- Boundary: 3-20 karaktera

---

### Test 5.4: Boundary Checking

**Šta testirate:** Da li sistem proverava dužinu input-a

**Koraci:**

1. Pokušajte registraciju sa predugačkim email-om (>254 karaktera):
   - **Email:** aaaaaaaaa... (255 karaktera)@example.com
   - **Očekivano:** Poruka greške "Input length exceeds maximum allowed"
2. Pokušajte sa prekratkim username-om (<3 karaktera):
   - **Username:** ab
   - **Očekivano:** Poruka greške "Username mora biti 3-20 karaktera"

**Provera u kodu:**
- `services/users-service/internal/validation/input.go`
- Email: max 254 karaktera
- Username: 3-20 karaktera
- Name: max 100 karaktera

---

### Test 5.5: File Upload Validation

**Šta testirate:** Da li sistem validira upload fajlova

**Napomena:** Ovo zahteva stvarni file upload endpoint. Ako nemate endpoint, pokažite kod.

**Koraci (ako imate endpoint):**

1. Pokušajte upload fajla sa nevažećim tipom (npr. .exe)
2. **Očekivano:** Poruka greške "File type not allowed"
3. Pokušajte upload prevelikog fajla (>10MB)
4. **Očekivano:** Poruka greške "File size exceeds maximum allowed"

**Provera u kodu:**
- `services/users-service/internal/validation/file.go`
- `ValidateFileType()` - MIME type whitelisting
- `ValidateFileSize()` - max 10MB
- `CalculateFileHash()` - MD5 hash za integritet
- `VerifyFileIntegrity()` - provera integriteta

---

### Test 5.6: Output Encoding

**Šta testirate:** Da li sistem escape-uje output

**Koraci:**

1. Registrujte korisnika sa imenom koje sadrži HTML karaktere (ako validacija dozvoljava)
2. Prijavite se i proverite da li se ime prikazuje ispravno
3. Otvorite Developer Tools (F12) → Network
4. Proverite API response za korisničke podatke
5. **Očekivano:** HTML karakteri su escape-ovani u JSON response-u

**Provera u kodu:**
- `services/users-service/internal/security/encoding.go`
- `EscapeHTML()` - HTML escaping
- JSON encoding automatski escape-uje

---

## ✅ TEST 6: Zaštita Podataka (2.19)

### Test 6.1: HTTPS Protokol

**Šta testirate:** Da li komunikacija koristi HTTPS

**Koraci:**

1. Otvorite browser Developer Tools (F12)
2. Idite na **Network** tab
3. Napravite bilo koji API zahtev (npr. login)
4. Kliknite na zahtev i proverite **Headers**
5. **Očekivano:**
   - **Request URL:** `https://localhost:8081/...` (HTTPS, ne HTTP)
   - **Protocol:** h2 (HTTP/2) ili h3
6. Proverite da li browser prikazuje "Secure" ikonicu pored URL-a

**Provera u kodu:**
- `services/api-gateway/cmd/main.go`
- `ListenAndServeTLS()` na portu 8081
- SSL sertifikati u `certs/` direktorijumu

---

### Test 6.2: POST Metoda za Senzitivne Parametre

**Šta testirate:** Da li senzitivni podaci se šalju preko POST metode

**Koraci:**

1. Otvorite Developer Tools (F12) → Network
2. Napravite registraciju ili login
3. Kliknite na zahtev i proverite:
   - **Request Method:** POST (ne GET)
   - **Request Payload:** Podaci su u body-u (ne u URL-u)
4. **Očekivano:** Svi senzitivni endpoint-i koriste POST metodu

**Provera u kodu:**
- `services/users-service/internal/handler/register.go:33-36`
- `if r.Method != http.MethodPost` provera
- HTTP 405 za nevažeće metode

---

### Test 6.3: Hash & Salt za Lozinke

**Šta testirate:** Da li su lozinke heširane u bazi

**Koraci:**

1. Otvorite MongoDB shell:
   ```
   docker exec -it projekat-2025-2-mongodb-users-1 mongosh
   ```
2. Proverite hash lozinke:
   ```javascript
   use music_streaming_users
   db.users.findOne({username: "testuser"}, {passwordHash: 1, email: 1, _id: 0})
   ```
3. **Očekivano:** `passwordHash` počinje sa `$2a$` ili `$2b$` (bcrypt format)
4. Proverite da li su različite lozinke imaju različite hash-eve:
   ```javascript
   db.users.find({}, {username: 1, passwordHash: 1, _id: 0})
   ```

**Provera u kodu:**
- `services/users-service/internal/security/password.go`
- `HashPassword()` koristi `bcrypt.GenerateFromPassword()`
- Format: `$2a$10$salt+hash`
- Automatski generiše salt za svaku lozinku

---

## ✅ TEST 7: Logovanje (2.20)

### Test 7.1: Logovanje Neuspeha Validacije

**Šta testirate:** Da li se loguju neuspehe validacije

**Koraci:**

1. Pokušajte registraciju sa nevažećim podacima (npr. XSS payload)
2. Proverite logove iz Docker kontejnera:
   ```
   docker logs projekat-2025-2-users-service-1 | Select-String "VALIDATION_FAILURE" | Select-Object -Last 5
   ```
3. **Alternativno:** Ako imate volume mount, proverite fajlove:
   ```
   ls logs/users-service/
   Get-Content logs/users-service/app-*.log | Select-String "VALIDATION_FAILURE"
   ```
4. **Očekivano:** Log entry sa `EventType=VALIDATION_FAILURE`

**Provera u kodu:**
- `services/users-service/internal/handler/register.go`
- `h.Logger.LogValidationFailure()` poziva se za svaku validaciju

---

### Test 7.2: Logovanje Pokušaja Prijave

**Šta testirate:** Da li se loguju pokušaji prijave

**Koraci:**

1. Pokušajte prijavu (uspešnu i neuspešnu)
2. Proverite logove iz Docker kontejnera:
   ```
   docker logs projekat-2025-2-users-service-1 | Select-String "LOGIN"
   ```
3. **Alternativno:** Ako imate volume mount:
   ```
   Get-Content logs/users-service/app-*.log | Select-String "LOGIN"
   ```
4. **Očekivano:**
   - `EventType=LOGIN_SUCCESS` za uspešne prijave
   - `EventType=LOGIN_FAILURE` za neuspešne prijave
   - Sa detaljima: username, IP adresa, razlog

**Provera u kodu:**
- `services/users-service/internal/handler/login_handler.go`
- `LogLoginSuccess()` i `LogLoginFailure()` pozivaju se za svaki pokušaj

---

### Test 7.3: Rotacija Logova

**Šta testirate:** Da li se logovi rotiraju kada dostignu maksimalnu veličinu

**Koraci:**

1. Proverite log fajlove u Docker kontejneru:
   ```
   docker exec projekat-2025-2-users-service-1 ls -la /app/logs/
   ```
2. **Alternativno:** Ako imate volume mount:
   ```
   ls logs/users-service/
   ```
3. **Očekivano:**
   - `app-YYYY-MM-DD.log` - trenutni fajl
   - `app-YYYY-MM-DD.log.YYYYMMDD-HHMMSS` - rotirani fajlovi
   - Maksimalno 5 rotiranih fajlova
   - Maksimalna veličina: 10MB po fajlu
4. Proverite veličinu fajlova (ako imate volume mount):
   ```
   ls logs/users-service/ | Get-Item | Select-Object Name, Length
   ```

**Provera u kodu:**
- `services/shared/logger/logger.go`
- `rotateLog()` - rotacija na 10MB
- `cleanupOldFiles()` - zadržava max 5 fajlova

---

### Test 7.4: Zaštita i Integritet Log-Datoteka

**Šta testirate:** Da li su log fajlovi zaštićeni i imaju integritet

**Koraci:**

1. Proverite permisije log fajlova u Docker kontejneru:
   ```
   docker exec projekat-2025-2-users-service-1 ls -la /app/logs/
   ```
2. **Očekivano:** Permisije 0640 (samo vlasnik i grupa mogu čitati)
3. Proverite checksum fajlove:
   ```
   docker exec projekat-2025-2-users-service-1 ls -la /app/logs/*.checksum
   ```
4. **Alternativno:** Ako imate volume mount:
   ```
   ls logs/users-service/*.checksum
   ```
5. **Očekivano:** `.checksum` fajlovi za svaki log fajl (SHA256)

**Provera u kodu:**
- `services/shared/logger/logger.go`
- `openLogFile()` - permissions 0640
- `updateChecksum()` - SHA256 checksum
- `VerifyIntegrity()` - provera integriteta

---

### Test 7.5: Filtriranje Osetljivih Podataka

**Šta testirate:** Da li se osetljivi podaci filtriraju iz logova

**Koraci:**

1. Napravite nekoliko zahteva sa različitim podacima
2. Proverite logove iz Docker kontejnera:
   ```
   docker logs projekat-2025-2-users-service-1 | Select-String "password|token|otp"
   ```
3. **Alternativno:** Ako imate volume mount:
   ```
   Get-Content logs/users-service/app-*.log | Select-String "password|token|otp"
   ```
4. **Očekivano:** Passwords, tokens, OTP se maskiraju kao `***` ili se ne loguju

**Provera u kodu:**
- `services/shared/logger/logger.go`
- `sanitizeMessage()` - filtriranje osetljivih podataka
- Automatsko maskiranje polja: password, token, otp, secret

---

## ✅ TEST 8: Analiza Ranjivosti (2.21)

### Test 8.1: Provera Izveštaja

**Šta testirate:** Da li postoji izveštaj o analizi ranjivosti

**Koraci:**

1. Proverite da li postoji izveštaj:
   ```
   ls IZVESTAJ_ANALIZA_RANJIVOSTI_2.21.md
   ```
2. Otvorite izveštaj i proverite da li sadrži:
   - ✅ Korišćene alate (Gosec, GolangCI-Lint, Semgrep, Snyk)
   - ✅ Identifikovane ranjivosti (Critical, High, Medium, Low)
   - ✅ Kako se mogu eksploatisati
   - ✅ Preporuke za prevazilaženje
   - ✅ Zaštita od eksploatacije

**Provera:**
- Lokacija: `IZVESTAJ_ANALIZA_RANJIVOSTI_2.21.md`
- Treba da sadrži sve navedene sekcije

---

## ✅ TEST 9: Demonstracija Pokušaja Napada (2.22)

### Test 9.1: XSS Napad

**Šta testirate:** Da li sistem blokira XSS napade

**Koraci:**

1. Otvorite browser na `http://localhost:3000/register`
2. Pokušajte registraciju sa XSS payload-om:
   - **Ime:** <script>alert('XSS')</script>
   - **Prezime:** User
   - **Email:** xss@example.com
   - **Username:** xssuser
   - **Lozinka:** Test1234!
   - **Ponovljena lozinka:** Test1234!
3. **Očekivano:** Poruka greške "Nevažeći unos" (HTTP 400)
4. Proverite logove:
   ```
   docker logs projekat-2025-2-users-service-1 | Select-String "XSS attempt"
   ```

**Alternativno:** Pokrenite test skriptu:
```
.\test-xss-attack.ps1
```

---

### Test 9.2: SQL Injection Napad

**Šta testirate:** Da li sistem blokira SQL injection napade

**Koraci:**

1. Pokušajte registraciju sa SQL injection payload-om:
   - **Ime:** Test' OR '1'='1
   - **Prezime:** User
   - **Email:** sqli@example.com
   - **Username:** sqliuser
   - **Lozinka:** Test1234!
   - **Ponovljena lozinka:** Test1234!
2. **Očekivano:** Poruka greške "Nevažeći unos" (HTTP 400)
3. Proverite logove:
   ```
   docker logs projekat-2025-2-users-service-1 | Select-String "SQL injection"
   ```

**Alternativno:** Pokrenite test skriptu:
```
.\test-sql-injection-attack.ps1
```

---

### Test 9.3: Brute-force Napad

**Šta testirate:** Da li sistem blokira brute-force napade

**Koraci:**

1. Pokušajte prijavu sa pogrešnom lozinkom **5 puta** za isti username
2. Nakon 5. pokušaja, pokušajte ponovo
3. **Očekivano:** Poruka greške "Nalog je zaključan" (HTTP 403)
4. Proverite logove:
   ```
   docker logs projekat-2025-2-users-service-1 | Select-String "LOGIN_FAILURE|account locked"
   ```

**Napomena:** Logovi se mogu videti preko `docker logs` komande - to je standardni način pristupa logovima u Docker okruženju.

**Alternativno:** Pokrenite test skriptu:
```
.\test-brute-force-attack.ps1
```

---

### Test 9.4: DoS Napad

**Šta testirate:** Da li sistem blokira DoS napade

**Koraci:**

1. Otvorite browser Developer Tools (F12) → Console
2. Pokrenite kod koji šalje veliki broj zahteva:
   ```javascript
   let blocked = 0;
   for (let i = 1; i <= 150; i++) {
     fetch('https://localhost:8081/api/users/health')
       .then(res => {
         if (res.status === 429) blocked++;
         console.log(`Request ${i}: ${res.status} (Blocked: ${blocked})`);
       });
     await new Promise(r => setTimeout(r, 100));
   }
   ```
3. **Očekivano:**
   - Prvih ~100 zahteva: HTTP 200
   - Preko 100 zahteva: HTTP 429
4. Proverite logove:
   ```
   docker logs projekat-2025-2-api-gateway-1 | Select-String "too many requests|ACCESS_CONTROL_FAILURE"
   ```

**Napomena:** Logovi se čuvaju unutar Docker kontejnera. Koristite `docker logs` za pristup.

**Alternativno:** Pokrenite test skriptu:
```
.\test-dos-attack.ps1
```

---

## 📊 Finalni Rezime

### Checklist za Ručno Testiranje

- [ ] **1.1 Registracija naloga**
  - [ ] Uspešna registracija
  - [ ] Jedinstven username
  - [ ] Jaka lozinka
  - [ ] Email verifikacija

- [ ] **1.2 Prijava na sistem**
  - [ ] Kombinovana autentifikacija (OTP)
  - [ ] Neuspešna prijava
  - [ ] Account locking (5 pokušaja)
  - [ ] Promena lozinke (1 dan stara)
  - [ ] Reset lozinke
  - [ ] Istek lozinke (auditabilnost)

- [ ] **1.3 Povraćaj naloga**
  - [ ] Request magic link
  - [ ] Verify magic link

- [ ] **2.17 Kontrola pristupa**
  - [ ] Autorizacija bez tokena
  - [ ] Nevažeći token
  - [ ] Validni token
  - [ ] Role-based access (ADMIN)
  - [ ] Šifrovanje state podataka
  - [ ] DoS zaštita (rate limiting)

- [ ] **2.18 Validacija podataka**
  - [ ] SQL Injection detection
  - [ ] XSS detection
  - [ ] Whitelisting
  - [ ] Boundary checking
  - [ ] File upload validation
  - [ ] Output encoding

- [ ] **2.19 Zaštita podataka**
  - [ ] HTTPS protokol
  - [ ] POST metoda za senzitivne podatke
  - [ ] Hash & Salt za lozinke

- [ ] **2.20 Logovanje**
  - [ ] Logovanje validacije
  - [ ] Logovanje prijave
  - [ ] Rotacija logova
  - [ ] Zaštita log-datoteka
  - [ ] Filtriranje osetljivih podataka

- [ ] **2.21 Analiza ranjivosti**
  - [ ] Izveštaj postoji
  - [ ] Sadrži sve sekcije

- [ ] **2.22 Demonstracija napada**
  - [ ] XSS napad
  - [ ] SQL Injection napad
  - [ ] Brute-force napad
  - [ ] DoS napad

---

## 🎯 Za Odbranu

### Kako Demonstrirati:

1. **Pokrenite sistem** (`docker-compose up -d`)
2. **Pratite korake iz ovog vodiča** ručno
3. **Pokažite rezultate** u browser-u i logovima
4. **Pokažite kod** (otvorite relevantne fajlove u IDE-u)
5. **Objasnite mehanizme** (kako funkcioniše svaka zaštita)

### Ključne Tačke za Objašnjenje:

- **Gde je implementirano:** Navedite lokaciju fajla i liniju
- **Kako radi:** Objasnite algoritam/kod
- **Zašto je važno:** Objasnite sigurnosni aspekt
- **Primer:** Pokažite konkretan primer koda

---

**Srećno na testiranju i odbrani! 🎓**
