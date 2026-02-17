# 📋 Test Plan - Informaciona Bezbednost (Ocena 10)

**Datum:** Februar 2025  
**Projekat:** Music Streaming Platform

---

## 🎯 Cilj Testiranja

Proveriti da li su svi zahtevi iz Informacione bezbednosti implementirani i funkcionalni.

---

## 📋 Checklist Pre Testiranja

- [ ] Docker kontejneri su pokrenuti (`docker-compose up -d`)
- [ ] Frontend je pokrenut (`cd frontend && npm start`)
- [ ] HTTPS sertifikati su generisani (`./generate-certs.ps1`)
- [ ] Admin korisnik postoji (username: `admin`, password: `admin123`)
- [ ] Test korisnik postoji ili će biti kreiran tokom testiranja

---

## 🔧 Priprema Okruženja

### Korak 1: Pokretanje Sistema

```powershell
# 1. Pokrenite Docker kontejnere
cd C:\Users\ivana\Desktop\Projekat-2025-1
docker-compose up -d

# 2. Sačekajte da se servisi pokrenu (20-30 sekundi)
Start-Sleep -Seconds 30

# 3. Proverite da li su svi servisi pokrenuti
docker-compose ps

# 4. Proverite logove za greške
docker-compose logs --tail=50
```

### Korak 2: Pokretanje Frontend-a

```powershell
# U novom terminalu
cd C:\Users\ivana\Desktop\Projekat-2025-1\frontend
npm start
```

### Korak 3: Učitavanje Helper Funkcija

```powershell
# U PowerShell terminalu gde ćete raditi testove
cd C:\Users\ivana\Desktop\Projekat-2025-1
. .\https-helper.ps1
```

---

## ✅ TEST 1: Registracija Naloga (1.1)

### Zahtevi:
- ✅ Jedinstven username
- ✅ Minimalni podaci: ime, prezime, email, lozinka, ponovljena lozinka
- ✅ Jaka lozinka
- ✅ Periodična promena lozinke (60 dana)
- ✅ Potvrda registracije (email verifikacija)

### Test 1.1: Uspešna Registracija

```powershell
Write-Host "=== TEST 1.1: Uspešna Registracija ===" -ForegroundColor Cyan

$body = @{
    firstName = "Test"
    lastName = "User"
    email = "testuser@example.com"
    username = "testuser"
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"

Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 201) { "Green" } else { "Red" })
Write-Host "Response: $($result.Content)" -ForegroundColor Gray

# Očekivano: HTTP 201 "registration successful"
```

### Test 1.2: Validacija - Jedinstven Username

```powershell
Write-Host "`n=== TEST 1.2: Jedinstven Username ===" -ForegroundColor Cyan

# Pokušaj registracije sa istim username-om
$body = @{
    firstName = "Another"
    lastName = "User"
    email = "another@example.com"
    username = "testuser"  # Isti username
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"

Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 409) { "Green" } else { "Red" })
# Očekivano: HTTP 409 "user already exists"
```

### Test 1.3: Validacija - Jaka Lozinka

```powershell
Write-Host "`n=== TEST 1.3: Jaka Lozinka ===" -ForegroundColor Cyan

# Pokušaj sa slabom lozinkom
$body = @{
    firstName = "Test"
    lastName = "User"
    email = "weak@example.com"
    username = "weakuser"
    password = "123"  # Slaba lozinka
    confirmPassword = "123"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"

Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 400) { "Green" } else { "Red" })
# Očekivano: HTTP 400 "password must be at least 8 characters"
```

### Test 1.4: Email Verifikacija

```powershell
Write-Host "`n=== TEST 1.4: Email Verifikacija ===" -ForegroundColor Cyan

# Proverite logove za verification link
docker logs projekat-2025-2-users-service-1 | Select-String "verification"

# Ili proverite response - ako SMTP nije konfigurisan, link će biti u response-u
# Očekivano: Verification link u email-u ili response-u
```

**Provera u kodu:**
- Lokacija: `services/users-service/internal/handler/register.go:184-200`
- Generiše se `verificationToken`
- Šalje se email sa linkom

---

## ✅ TEST 2: Prijava na Sistem (1.2)

### Zahtevi:
- ✅ Kombinovana autentifikacija (lozinka + OTP)
- ✅ Promena lozinke (mora biti bar 1 dan stara)
- ✅ Email-bazirani reset lozinke
- ✅ Auditabilnost (onemogućavanje prijave nakon isteka lozinke - 60 dana)

### Test 2.1: Kombinovana Autentifikacija (Lozinka + OTP)

```powershell
Write-Host "=== TEST 2.1: Kombinovana Autentifikacija ===" -ForegroundColor Cyan

# Korak 1: Request OTP
$body = @{
    username = "testuser"
    password = "Test1234!"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/login/request-otp" -Method "POST" -Body $body -ContentType "application/json"

Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 200) { "Green" } else { "Red" })
# Očekivano: HTTP 200

# Korak 2: Proverite logove za OTP kod
$otp = docker logs projekat-2025-2-users-service-1 | Select-String "OTP" | Select-Object -Last 1
Write-Host "OTP kod iz logova: $otp" -ForegroundColor Yellow

# Korak 3: Verify OTP
$body = @{
    username = "testuser"
    otp = "123456"  # Zamenite sa stvarnim OTP kodom iz logova
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/login/verify-otp" -Method "POST" -Body $body -ContentType "application/json"

Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 200) { "Green" } else { "Red" })
Write-Host "Response: $($result.Content)" -ForegroundColor Gray

# Očekivano: HTTP 200 sa JWT tokenom
$token = ($result.Content | ConvertFrom-Json).token
Write-Host "Token: $token" -ForegroundColor Green
```

**Provera u kodu:**
- Lokacija: `services/users-service/internal/handler/login_handler.go:33-108`
- RequestOTP proverava lozinku, generiše OTP, šalje email
- VerifyOTP proverava OTP, generiše JWT token

### Test 2.2: Promena Lozinke (Mora biti bar 1 dan stara)

```powershell
Write-Host "`n=== TEST 2.2: Promena Lozinke ===" -ForegroundColor Cyan

# Pokušaj promene lozinke odmah nakon registracije
$body = @{
    username = "testuser"
    oldPassword = "Test1234!"
    newPassword = "NewPassword123!"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/password/change" -Method "POST" -Body $body -ContentType "application/json" -Headers @{Authorization = "Bearer $token"}

Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 403) { "Green" } else { "Red" })
# Očekivano: HTTP 403 "password must be at least 1 day old"
```

**Provera u kodu:**
- Lokacija: `services/users-service/internal/handler/password_handler.go`
- Provera: `time.Since(user.PasswordChangedAt) < 24*time.Hour`

### Test 2.3: Email-bazirani Reset Lozinke

```powershell
Write-Host "`n=== TEST 2.3: Email-bazirani Reset Lozinke ===" -ForegroundColor Cyan

# Korak 1: Request reset token
$body = @{
    email = "testuser@example.com"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/password/reset/request" -Method "POST" -Body $body -ContentType "application/json"

Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 200) { "Green" } else { "Red" })
# Očekivano: HTTP 200

# Korak 2: Proverite logove za reset link
docker logs projekat-2025-2-users-service-1 | Select-String "reset"
```

**Provera u kodu:**
- Lokacija: `services/users-service/internal/handler/password_handler.go`
- Generiše se reset token (ističe za 1 sat)
- Šalje se email sa linkom

### Test 2.4: Auditabilnost - Istek Lozinke

```powershell
Write-Host "`n=== TEST 2.4: Auditabilnost - Istek Lozinke ===" -ForegroundColor Cyan

# Simulacija: Postavite PasswordExpiresAt u prošlosti u bazi
# Ovo zahteva direktan pristup MongoDB bazi

# Nakon što je lozinka istekla, pokušaj prijave
$body = @{
    username = "testuser"
    password = "Test1234!"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/login/request-otp" -Method "POST" -Body $body -ContentType "application/json"

Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 403) { "Green" } else { "Red" })
# Očekivano: HTTP 403 "password expired"
```

**Provera u kodu:**
- Lokacija: `services/users-service/internal/handler/login_handler.go:70-76`
- Provera: `time.Now().After(user.PasswordExpiresAt)`
- Maksimalni period: 60 dana (konfigurabilno)

---

## ✅ TEST 3: Povraćaj Naloga - Magic Link (1.3)

### Zahtevi:
- ✅ Autentifikacija upotrebom magičnog linka

### Test 3.1: Request Magic Link

```powershell
Write-Host "=== TEST 3.1: Request Magic Link ===" -ForegroundColor Cyan

$body = @{
    email = "testuser@example.com"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/recover/request" -Method "POST" -Body $body -ContentType "application/json"

Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 200) { "Green" } else { "Red" })
# Očekivano: HTTP 200

# Proverite logove za magic link
docker logs projekat-2025-2-users-service-1 | Select-String "magic"
```

### Test 3.2: Verify Magic Link

```powershell
Write-Host "`n=== TEST 3.2: Verify Magic Link ===" -ForegroundColor Cyan

# Iz logova uzmite magic link token
# Format: https://localhost:3000/verify-magic-link?token=...

# Ili direktno pozovite API
$token = "your-magic-link-token-from-logs"
$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/recover/verify?token=$token" -Method "GET"

Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 200) { "Green" } else { "Red" })
Write-Host "Response: $($result.Content)" -ForegroundColor Gray

# Očekivano: HTTP 200 sa JWT tokenom (automatska prijava)
```

**Provera u kodu:**
- Lokacija: `services/users-service/internal/handler/magic_link_handler.go`
- Token ističe za 15 minuta
- Jednokratna upotreba (briše se nakon korišćenja)

---

## ✅ TEST 4: Kontrola Pristupa (2.17)

### Zahtevi:
- ✅ Autorizacija za svaki zahtev
- ✅ Šifrovanje i provera integriteta state podataka
- ✅ DoS zaštita (rate limiting)

### Test 4.1: Autorizacija - Zaštićeni Endpoint bez Tokena

```powershell
Write-Host "=== TEST 4.1: Autorizacija - Bez Tokena ===" -ForegroundColor Cyan

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/logout" -Method "POST" -Body "{}" -ContentType "application/json"

Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 401) { "Green" } else { "Red" })
# Očekivano: HTTP 401 "authorization header required"
```

### Test 4.2: Autorizacija - Nevažeći Token

```powershell
Write-Host "`n=== TEST 4.2: Autorizacija - Nevažeći Token ===" -ForegroundColor Cyan

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/logout" -Method "POST" -Body "{}" -ContentType "application/json" -Headers @{Authorization = "Bearer invalid-token"}

Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 401) { "Green" } else { "Red" })
# Očekivano: HTTP 401 "invalid or expired token"
```

### Test 4.3: Autorizacija - Validni Token

```powershell
Write-Host "`n=== TEST 4.3: Autorizacija - Validni Token ===" -ForegroundColor Cyan

# Prvo se prijavite da dobijete token (koristite Test 2.1)
$token = "your-valid-token-from-login"

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/logout" -Method "POST" -Body "{}" -ContentType "application/json" -Headers @{Authorization = "Bearer $token"}

Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 200) { "Green" } else { "Red" })
# Očekivano: HTTP 200 "logged out successfully"
```

### Test 4.4: Role-Based Access Control (ADMIN)

```powershell
Write-Host "`n=== TEST 4.4: Role-Based Access Control ===" -ForegroundColor Cyan

# Test kao regular user (ne admin)
$regularUserToken = "token-for-regular-user"

$body = @{
    name = "Test Artist"
    bio = "Test bio"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/content/artists" -Method "POST" -Body $body -ContentType "application/json" -Headers @{Authorization = "Bearer $regularUserToken"}

Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 403) { "Green" } else { "Red" })
# Očekivano: HTTP 403 "forbidden: ADMIN access required"

# Test kao admin
$adminToken = "token-for-admin-user"
$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/content/artists" -Method "POST" -Body $body -ContentType "application/json" -Headers @{Authorization = "Bearer $adminToken"}

Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 200 -or $result.StatusCode -eq 201) { "Green" } else { "Red" })
# Očekivano: HTTP 200/201
```

**Provera u kodu:**
- Lokacija: `services/api-gateway/internal/middleware/auth.go`
- `RequireAuth` - zahteva validan token
- `RequireRole("ADMIN")` - zahteva ADMIN ulogu

### Test 4.5: Šifrovanje i Integritet State Podataka

```powershell
Write-Host "`n=== TEST 4.5: Šifrovanje State Podataka ===" -ForegroundColor Cyan

# Otvorite browser Developer Tools (F12)
# Idite na Application → Local Storage
# Proverite:
# - `user` - treba biti šifrovano (base64 encoded)
# - `user_checksum` - treba postojati za proveru integriteta
# - `token` - plain text (JWT je već encoded)

# Test manipulacije:
# 1. Promenite `user` vrednost u localStorage
# 2. Osvežite stranicu
# 3. Očekivano: Podaci se brišu, korisnik se odjavljuje
```

**Provera u kodu:**
- Lokacija: `frontend/src/utils/encryption.js`
- `setEncryptedItem()` - šifruje podatke
- `getEncryptedItem()` - dešifruje i proverava integritet

### Test 4.6: DoS Zaštita - Rate Limiting

```powershell
Write-Host "`n=== TEST 4.6: DoS Zaštita - Rate Limiting ===" -ForegroundColor Cyan

# Pokrenite test skriptu
.\test-dos-attack.ps1

# Ili ručno:
$totalRequests = 150
$blockedCount = 0

for ($i = 1; $i -le $totalRequests; $i++) {
    $result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/health" -Method "GET"
    
    if ($result.StatusCode -eq 429) {
        $blockedCount++
        if ($blockedCount -eq 1) {
            Write-Host "Prvi blokirani zahtev na poziciji: $i" -ForegroundColor Red
        }
    }
    
    Start-Sleep -Milliseconds 100
}

Write-Host "Ukupno blokirano: $blockedCount/$totalRequests" -ForegroundColor $(if ($blockedCount -gt 0) { "Green" } else { "Red" })
# Očekivano: ~64 zahteva blokirano (HTTP 429)
```

**Provera u kodu:**
- Lokacija: `services/api-gateway/internal/middleware/rate_limit.go`
- Limit: 100 zahteva/min po IP adresi

---

## ✅ TEST 5: Validacija Podataka (2.18)

### Zahtevi:
- ✅ Input validation (string provere, whitelisting, boundary checking, character escaping, numeric validation, specijalni karakteri)
- ✅ Bezbedno upravljanje datotekama (privilegije, tip datoteke, integritet)
- ✅ Validacija na client i server strani
- ✅ Output encoding

### Test 5.1: Input Validation - SQL Injection

```powershell
Write-Host "=== TEST 5.1: SQL Injection Detection ===" -ForegroundColor Cyan

.\test-sql-injection-attack.ps1

# Očekivano: Svi napadi blokirani (HTTP 400)
```

**Provera u kodu:**
- Lokacija: `services/users-service/internal/validation/input.go:98-119`
- `CheckSQLInjection()` detektuje SQL injection pattern-e

### Test 5.2: Input Validation - XSS

```powershell
Write-Host "`n=== TEST 5.2: XSS Detection ===" -ForegroundColor Cyan

.\test-xss-attack.ps1

# Očekivano: Svi napadi blokirani (HTTP 400)
```

**Provera u kodu:**
- Lokacija: `services/users-service/internal/validation/input.go:121-144`
- `CheckXSS()` detektuje XSS pattern-e

### Test 5.3: Input Validation - Whitelisting

```powershell
Write-Host "`n=== TEST 5.3: Whitelisting (Username) ===" -ForegroundColor Cyan

$body = @{
    firstName = "Test"
    lastName = "User"
    email = "test@example.com"
    username = "test@user!"  # Nevažeći karakteri
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"

Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 400) { "Green" } else { "Red" })
# Očekivano: HTTP 400 "username must contain only letters, numbers, and underscores"
```

**Provera u kodu:**
- Lokacija: `services/users-service/internal/validation/input.go`
- `ValidateUsername()` - whitelist: `[a-zA-Z0-9_]`

### Test 5.4: Input Validation - Boundary Checking

```powershell
Write-Host "`n=== TEST 5.4: Boundary Checking ===" -ForegroundColor Cyan

# Test predugačkog email-a (>254 karaktera)
$longEmail = "a" * 255 + "@example.com"
$body = @{
    firstName = "Test"
    lastName = "User"
    email = $longEmail
    username = "testuser"
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"

Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 400) { "Green" } else { "Red" })
# Očekivano: HTTP 400 "input length exceeds maximum allowed"
```

**Provera u kodu:**
- Lokacija: `services/users-service/internal/validation/input.go`
- Email: max 254 karaktera
- Username: 3-20 karaktera
- Name: max 100 karaktera

### Test 5.5: File Upload Validation

```powershell
Write-Host "`n=== TEST 5.5: File Upload Validation ===" -ForegroundColor Cyan

# Napomena: Ovo zahteva stvarni file upload endpoint
# Ako nemate endpoint, pokažite kod u file.go

Write-Host "Provera implementacije u kodu:" -ForegroundColor Yellow
Write-Host "- Lokacija: services/users-service/internal/validation/file.go" -ForegroundColor Gray
Write-Host "- ValidateFileType() - MIME type whitelisting" -ForegroundColor Gray
Write-Host "- ValidateFileSize() - max 10MB" -ForegroundColor Gray
Write-Host "- CalculateFileHash() - MD5 hash za integritet" -ForegroundColor Gray
Write-Host "- VerifyFileIntegrity() - provera integriteta" -ForegroundColor Gray
```

**Provera u kodu:**
- Lokacija: `services/users-service/internal/validation/file.go`
- `ValidateFileUpload()` - kompleksna validacija
- `CalculateFileHash()` - MD5 hash
- `VerifyFileIntegrity()` - provera integriteta

### Test 5.6: Output Encoding

```powershell
Write-Host "`n=== TEST 5.6: Output Encoding ===" -ForegroundColor Cyan

# Test sa HTML karakterima u odgovoru
# Registrujte korisnika sa imenom koje sadrži HTML karaktere
# Proverite da li su escape-ovani u JSON odgovoru

Write-Host "Provera implementacije u kodu:" -ForegroundColor Yellow
Write-Host "- Lokacija: services/users-service/internal/security/encoding.go" -ForegroundColor Gray
Write-Host "- EscapeHTML() - HTML escaping" -ForegroundColor Gray
Write-Host "- EscapeURL() - URL encoding" -ForegroundColor Gray
Write-Host "- JSON encoding automatski escape-uje" -ForegroundColor Gray
```

**Provera u kodu:**
- Lokacija: `services/users-service/internal/security/encoding.go`
- `EscapeHTML()` - HTML escaping
- `EscapeURL()` - URL encoding
- JSON encoding automatski escape-uje

---

## ✅ TEST 6: Zaštita Podataka (2.19)

### Zahtevi:
- ✅ HTTPS između servisa
- ✅ HTTPS između API Gateway-a i klijenta
- ✅ POST metoda za senzitivne parametre
- ✅ Hash & Salt za lozinke

### Test 6.1: HTTPS između Servisa

```powershell
Write-Host "=== TEST 6.1: HTTPS između Servisa ===" -ForegroundColor Cyan

# Proverite konfiguraciju u docker-compose.yml
docker-compose config | Select-String "TLS_CERT_FILE"

# Proverite da li servisi koriste HTTPS
docker exec projekat-2025-2-api-gateway-1 env | Select-String "SERVICE_URL"
# Očekivano: https:// za sve servise

Write-Host "Provera implementacije u kodu:" -ForegroundColor Yellow
Write-Host "- Lokacija: docker-compose.yml" -ForegroundColor Gray
Write-Host "- TLS_CERT_FILE i TLS_KEY_FILE su konfigurisani" -ForegroundColor Gray
Write-Host "- Svi servisi koriste HTTPS za inter-service komunikaciju" -ForegroundColor Gray
```

**Provera u kodu:**
- Lokacija: `docker-compose.yml`
- Svi servisi imaju `TLS_CERT_FILE` i `TLS_KEY_FILE`
- API Gateway koristi `https://` za pozive backend servisa

### Test 6.2: HTTPS API Gateway ↔ Klijent

```powershell
Write-Host "`n=== TEST 6.2: HTTPS API Gateway ↔ Klijent ===" -ForegroundColor Cyan

# Proverite da li API Gateway koristi HTTPS
$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/health" -Method "GET"

Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 200) { "Green" } else { "Red" })
# Očekivano: HTTP 200 (HTTPS radi)

# Proverite frontend konfiguraciju
Write-Host "`nProvera frontend konfiguracije:" -ForegroundColor Yellow
Write-Host "- Lokacija: frontend/src/services/api.js" -ForegroundColor Gray
Write-Host "- API_BASE_URL treba biti: https://localhost:8081" -ForegroundColor Gray
```

**Provera u kodu:**
- Lokacija: `services/api-gateway/cmd/main.go`
- `ListenAndServeTLS()` na portu 8081
- Frontend: `frontend/src/services/api.js` - `https://localhost:8081`

### Test 6.3: POST Metoda za Senzitivne Parametre

```powershell
Write-Host "`n=== TEST 6.3: POST Metoda za Senzitivne Parametre ===" -ForegroundColor Cyan

# Test GET zahtev na senzitivnom endpoint-u (ne bi trebalo raditi)
$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "GET"

Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 405) { "Green" } else { "Red" })
# Očekivano: HTTP 405 "method not allowed"

# Proverite da li svi senzitivni endpoint-i koriste POST
Write-Host "`nSenzitivni endpoint-i (mora biti POST):" -ForegroundColor Yellow
Write-Host "- /api/users/register" -ForegroundColor Gray
Write-Host "- /api/users/login/request-otp" -ForegroundColor Gray
Write-Host "- /api/users/login/verify-otp" -ForegroundColor Gray
Write-Host "- /api/users/password/change" -ForegroundColor Gray
Write-Host "- /api/users/password/reset/request" -ForegroundColor Gray
Write-Host "- /api/users/password/reset" -ForegroundColor Gray
```

**Provera u kodu:**
- Lokacija: `services/users-service/internal/handler/*.go`
- Svi handler-i proveravaju: `if r.Method != http.MethodPost`

### Test 6.4: Hash & Salt za Lozinke

```powershell
Write-Host "`n=== TEST 6.4: Hash & Salt za Lozinke ===" -ForegroundColor Cyan

# Proverite da li su lozinke heširane u bazi
docker exec projekat-2025-2-mongodb-users-1 mongosh --quiet --eval "db.users.findOne({}, {passwordHash: 1, email: 1, _id: 0})"

# Očekivano: passwordHash počinje sa $2a$ ili $2b$ (bcrypt format)

Write-Host "`nProvera implementacije u kodu:" -ForegroundColor Yellow
Write-Host "- Lokacija: services/users-service/internal/security/password.go" -ForegroundColor Gray
Write-Host "- HashPassword() koristi bcrypt.GenerateFromPassword()" -ForegroundColor Gray
Write-Host "- bcrypt.DefaultCost = 10 rounds" -ForegroundColor Gray
Write-Host "- Automatski generiše salt za svaku lozinku" -ForegroundColor Gray
```

**Provera u kodu:**
- Lokacija: `services/users-service/internal/security/password.go`
- `HashPassword()` koristi `bcrypt.GenerateFromPassword()`
- Format: `$2a$10$salt+hash`

---

## ✅ TEST 7: Logovanje (2.20)

### Zahtevi:
- ✅ Logovanje: neuspehe validacije, pokušaje prijave, neuspehe kontrole pristupa, neočekivane promene state, nevalidne/istekle tokene, administratorske aktivnosti, neuspešne TLS konekcije
- ✅ Rotacija logova
- ✅ Zaštita log-datoteka
- ✅ Integritet log-datoteka
- ✅ Filtriranje osetljivih podataka

### Test 7.1: Logovanje Neuspeha Validacije

```powershell
Write-Host "=== TEST 7.1: Logovanje Neuspeha Validacije ===" -ForegroundColor Cyan

# Pokušaj registracije sa nevažećim podacima
$body = @{
    firstName = "<script>alert('XSS')</script>"
    lastName = "User"
    email = "test@example.com"
    username = "testuser"
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"

# Proverite logove
docker logs projekat-2025-2-users-service-1 | Select-String "VALIDATION_FAILURE" | Select-Object -Last 5

# Očekivano: Log entry sa EventType=VALIDATION_FAILURE
```

**Provera u kodu:**
- Lokacija: `services/users-service/internal/handler/register.go`
- `h.Logger.LogValidationFailure()` poziva se za svaku validaciju

### Test 7.2: Logovanje Pokušaja Prijave

```powershell
Write-Host "`n=== TEST 7.2: Logovanje Pokušaja Prijave ===" -ForegroundColor Cyan

# Uspešna prijava
$body = @{
    username = "testuser"
    password = "Test1234!"
} | ConvertTo-Json

Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/login/request-otp" -Method "POST" -Body $body -ContentType "application/json"

# Proverite logove
docker logs projekat-2025-2-users-service-1 | Select-String "LOGIN_SUCCESS" | Select-Object -Last 1

# Neuspešna prijava
$body = @{
    username = "testuser"
    password = "WrongPassword"
} | ConvertTo-Json

Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/login/request-otp" -Method "POST" -Body $body -ContentType "application/json"

# Proverite logove
docker logs projekat-2025-2-users-service-1 | Select-String "LOGIN_FAILURE" | Select-Object -Last 1

# Očekivano: Log entry sa EventType=LOGIN_SUCCESS ili LOGIN_FAILURE
```

**Provera u kodu:**
- Lokacija: `services/users-service/internal/handler/login_handler.go`
- `LogLoginSuccess()` i `LogLoginFailure()` pozivaju se za svaki pokušaj

### Test 7.3: Logovanje Neuspeha Kontrole Pristupa

```powershell
Write-Host "`n=== TEST 7.3: Logovanje Neuspeha Kontrole Pristupa ===" -ForegroundColor Cyan

# Pokušaj pristupa bez tokena
Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/logout" -Method "POST" -Body "{}" -ContentType "application/json"

# Proverite logove
docker logs projekat-2025-2-api-gateway-1 | Select-String "ACCESS_CONTROL_FAILURE" | Select-Object -Last 1

# Očekivano: Log entry sa EventType=ACCESS_CONTROL_FAILURE
```

**Provera u kodu:**
- Lokacija: `services/api-gateway/internal/middleware/auth.go`
- `log.LogAccessControlFailure()` poziva se za neuspehe autorizacije

### Test 7.4: Rotacija Logova

```powershell
Write-Host "`n=== TEST 7.4: Rotacija Logova ===" -ForegroundColor Cyan

# Proverite log fajlove
ls services/users-service/logs/

# Očekivano: 
# - app-YYYY-MM-DD.log (trenutni fajl)
# - app-YYYY-MM-DD.log.YYYYMMDD-HHMMSS (rotirani fajlovi)
# - Maksimalno 5 rotiranih fajlova
# - Maksimalna veličina: 10MB po fajlu

Write-Host "Provera implementacije u kodu:" -ForegroundColor Yellow
Write-Host "- Lokacija: services/shared/logger/logger.go" -ForegroundColor Gray
Write-Host "- rotateLog() - automatska rotacija na 10MB" -ForegroundColor Gray
Write-Host "- cleanupOldFiles() - briše stare fajlove (max 5)" -ForegroundColor Gray
```

**Provera u kodu:**
- Lokacija: `services/shared/logger/logger.go`
- `rotateLog()` - rotacija na 10MB
- `cleanupOldFiles()` - zadržava max 5 fajlova

### Test 7.5: Zaštita i Integritet Log-Datoteka

```powershell
Write-Host "`n=== TEST 7.5: Zaštita i Integritet Log-Datoteka ===" -ForegroundColor Cyan

# Proverite permisije log fajlova
ls services/users-service/logs/ | Get-Item | Select-Object Name, Mode

# Očekivano: Permisije 0640 (samo vlasnik i grupa)

# Proverite checksum fajlove
ls services/users-service/logs/*.checksum

# Očekivano: .checksum fajlovi za svaki log fajl

Write-Host "Provera implementacije u kodu:" -ForegroundColor Yellow
Write-Host "- Lokacija: services/shared/logger/logger.go" -ForegroundColor Gray
Write-Host "- openLogFile() - permissions 0640" -ForegroundColor Gray
Write-Host "- updateChecksum() - SHA256 checksum" -ForegroundColor Gray
Write-Host "- VerifyIntegrity() - provera integriteta" -ForegroundColor Gray
```

**Provera u kodu:**
- Lokacija: `services/shared/logger/logger.go`
- Permisije: 0640
- SHA256 checksums za integritet

---

## ✅ TEST 8: Analiza Ranjivosti (2.21)

### Zahtevi:
- ✅ Izveštaj o nivou bezbednosti
- ✅ Korišćeni alati
- ✅ Identifikovane ranjivosti
- ✅ Preporuke za prevazilaženje

### Test 8.1: Provera Izveštaja

```powershell
Write-Host "=== TEST 8.1: Provera Izveštaja ===" -ForegroundColor Cyan

# Proverite da li postoji izveštaj
ls IZVESTAJ_ANALIZA_RANJIVOSTI_2.21.md

# Pročitajte izveštaj
Get-Content IZVESTAJ_ANALIZA_RANJIVOSTI_2.21.md | Select-Object -First 50

Write-Host "`nIzveštaj treba da sadrži:" -ForegroundColor Yellow
Write-Host "- Korišćene alate (Gosec, GolangCI-Lint, Semgrep, Snyk)" -ForegroundColor Gray
Write-Host "- Identifikovane ranjivosti (Critical, High, Medium, Low)" -ForegroundColor Gray
Write-Host "- Kako se mogu eksploatisati" -ForegroundColor Gray
Write-Host "- Preporuke za prevazilaženje" -ForegroundColor Gray
Write-Host "- Zaštita od eksploatacije" -ForegroundColor Gray
```

**Provera:**
- Lokacija: `IZVESTAJ_ANALIZA_RANJIVOSTI_2.21.md`
- Treba da sadrži sve navedene sekcije

---

## ✅ TEST 9: Demonstracija Pokušaja Napada (2.22)

### Zahtevi:
- ✅ XSS napad
- ✅ SQL Injection napad
- ✅ Brute-force napad
- ✅ DoS napad

### Test 9.1: XSS Napad

```powershell
Write-Host "=== TEST 9.1: XSS Napad ===" -ForegroundColor Cyan

.\test-xss-attack.ps1

# Očekivano: Svi napadi blokirani (HTTP 400 ili 429)
```

### Test 9.2: SQL Injection Napad

```powershell
Write-Host "`n=== TEST 9.2: SQL Injection Napad ===" -ForegroundColor Cyan

.\test-sql-injection-attack.ps1

# Očekivano: Svi napadi blokirani (HTTP 400)
```

### Test 9.3: Brute-force Napad

```powershell
Write-Host "`n=== TEST 9.3: Brute-force Napad ===" -ForegroundColor Cyan

.\test-brute-force-attack.ps1

# Očekivano: 
# - Prvih 5 pokušaja neuspešni (HTTP 401)
# - Nakon 5. pokušaja, nalog zaključan (HTTP 403)
```

### Test 9.4: DoS Napad

```powershell
Write-Host "`n=== TEST 9.4: DoS Napad ===" -ForegroundColor Cyan

.\test-dos-attack.ps1

# Očekivano:
# - Prvih ~100 zahteva prolazi (HTTP 200)
# - Preko 100 zahteva blokirano (HTTP 429)
```

### Test 9.5: Svi Napadi Odjednom

```powershell
Write-Host "`n=== TEST 9.5: Svi Napadi Odjednom ===" -ForegroundColor Cyan

.\test-all-attacks.ps1

# Očekivano: Svi napadi testirani i blokirani
```

---

## 📊 Finalni Rezime Testiranja

### Checklist

- [ ] **1.1 Registracija naloga** - ✅ Testirano
- [ ] **1.2 Prijava na sistem** - ✅ Testirano
- [ ] **1.3 Povraćaj naloga** - ✅ Testirano
- [ ] **2.17 Kontrola pristupa** - ✅ Testirano
- [ ] **2.18 Validacija podataka** - ✅ Testirano
- [ ] **2.19 Zaštita podataka** - ✅ Testirano
- [ ] **2.20 Logovanje** - ✅ Testirano
- [ ] **2.21 Analiza ranjivosti** - ✅ Provereno
- [ ] **2.22 Demonstracija napada** - ✅ Testirano

### Komande za Brzu Proveru

```powershell
# Pokrenite sve testove odjednom
.\test-all-attacks.ps1

# Proverite logove
docker logs projekat-2025-2-users-service-1 | Select-String "VALIDATION_FAILURE|LOGIN_FAILURE|LOGIN_SUCCESS"
docker logs projekat-2025-2-api-gateway-1 | Select-String "ACCESS_CONTROL_FAILURE"

# Proverite HTTPS
Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/health" -Method "GET"

# Proverite hash lozinki u bazi
docker exec projekat-2025-2-mongodb-users-1 mongosh --quiet --eval "db.users.findOne({}, {passwordHash: 1, email: 1, _id: 0})"
```

---

## 🎯 Za Odbranu

### Kako Demonstrirati:

1. **Pokrenite sistem** (`docker-compose up -d`)
2. **Pokrenite test skripte** (`.\test-all-attacks.ps1`)
3. **Pokažite logove** (`docker logs ...`)
4. **Pokažite kod** (otvorite relevantne fajlove u IDE-u)
5. **Objasnite mehanizme** (kako funkcioniše svaka zaštita)

### Ključne Tačke za Objašnjenje:

- **Gde je implementirano:** Navedite lokaciju fajla
- **Kako radi:** Objasnite algoritam/kod
- **Zašto je važno:** Objasnite sigurnosni aspekt
- **Primer:** Pokažite konkretan primer koda

---

**Srećno na testiranju i odbrani! 🎓**
