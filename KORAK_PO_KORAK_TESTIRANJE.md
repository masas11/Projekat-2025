# Korak po Korak - Testiranje Logovanja

## 📋 Preduslovi

1. **Svi servisi moraju biti pokrenuti:**
   ```powershell
   docker-compose ps
   ```
   Svi servisi treba da budu u statusu "Up".

2. **Proverite da li servisi rade:**
   ```powershell
   # Test API Gateway
   Invoke-HTTPSRequest
   ```

---

## 🧪 TEST 1: Neuspeh Kontrole Pristupa (bez tokena)

### Korak 1: Učitajte helper funkciju
```powershell
. .\https-helper.ps1
```

### Korak 2: Pošaljite zahtev bez tokena
```powershell
$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/logout" -Method "GET"
Write-Host "Status Code: $($result.StatusCode)"
```

**Očekivani rezultat:** Greška 401 (Unauthorized)

### Korak 2: Proverite logove
```powershell
# Opcija 1: Docker logs (ako logger koristi stdout)
docker logs projekat-2025-2-api-gateway-1 --tail 20 | Select-String "ACCESS_CONTROL_FAILURE"

# Opcija 2: Log fajl u kontejneru
docker exec projekat-2025-2-api-gateway-1 sh -c "find /app/logs -name '*.log' -exec grep -l 'ACCESS_CONTROL_FAILURE' {} \; 2>/dev/null | head -1 | xargs tail -10"
```

**Očekivani log:**
```
[AUDIT] EventType=ACCESS_CONTROL_FAILURE Message=... Resource=/api/users/logout Action=GET Reason=missing authorization header
```

---

## 🧪 TEST 2: Nevalidni Tokeni

### Korak 1: Pošaljite zahtev sa nevalidnim tokenom
```powershell
$headers = @{
    "Authorization" = "Bearer invalid_token_12345"
}
$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/logout" -Method "GET" -Headers $headers
Write-Host "Status Code: $($result.StatusCode)"
```

**Očekivani rezultat:** Greška 401 (Unauthorized)

### Korak 2: Proverite logove
```powershell
# Docker logs
docker logs projekat-2025-2-api-gateway-1 --tail 20 | Select-String "INVALID_TOKEN"

# Ili log fajl
docker exec projekat-2025-2-api-gateway-1 sh -c "find /app/logs -name '*.log' -exec grep -l 'INVALID_TOKEN' {} \; 2>/dev/null | head -1 | xargs tail -10"
```

**Očekivani log:**
```
[AUDIT] EventType=INVALID_TOKEN Message=... TokenPrefix=invalid_tok... Reason=...
```

---

## 🧪 TEST 3: Neuspeh Kontrole Pristupa - RequireRole

### Korak 1: Pokušaj pristupa admin endpoint-u bez admin tokena
```powershell
$headers = @{
    "Authorization" = "Bearer invalid_token"
    "Content-Type" = "application/json"
}
$body = '{"name":"Test Artist","biography":"Test biography","genres":["Rock"]}'

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/content/artists" -Method "POST" -Headers $headers -Body $body
Write-Host "Status Code: $($result.StatusCode)"
```

**Očekivani rezultat:** Greška 401 ili 403 (Unauthorized/Forbidden)

### Korak 2: Proverite logove
```powershell
docker logs projekat-2025-2-api-gateway-1 --tail 30 | Select-String "ACCESS_CONTROL_FAILURE|insufficient permissions"
```

**Očekivani log:**
```
[AUDIT] EventType=ACCESS_CONTROL_FAILURE Message=... Reason=insufficient permissions: required role ADMIN, user role USER
```

---

## 🧪 TEST 4: Administratorske Aktivnosti - CREATE Artist

### Korak 1: Prijavite se kao admin korisnik

**Prvo, registrujte admin korisnika (ako ne postoji):**
```powershell
# Registracija
$registerBody = @{
    firstName = "Admin"
    lastName = "User"
    email = "admin@test.com"
    username = "admin"
    password = "Admin123!"
    confirmPassword = "Admin123!"
} | ConvertTo-Json

$response = Invoke-HTTPSRequest
```

**Zatim, prijavite se:**
```powershell
# 1. Zatražite OTP
$otpRequest = @{
    username = "admin"
    password = "Admin123!"
} | ConvertTo-Json

Invoke-HTTPSRequest

# 2. Proverite email za OTP kod (u produkciji bi stigao email)
# Za test, možete proveriti u bazi podataka ili koristiti test OTP

# 3. Verifikujte OTP i dobijte token
$otpVerify = @{
    username = "admin"
    otp = "123456"  # Zamenite sa stvarnim OTP kodom
} | ConvertTo-Json

$loginResponse = Invoke-HTTPSRequest
$token = ($loginResponse.Content | ConvertFrom-Json).token
```

**Napomena:** Ako imate već postojećeg admin korisnika, samo se prijavite.

### Korak 2: Kreirajte artist kao admin
```powershell
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}
$body = @{
    name = "Test Artist"
    biography = "Test biography for artist"
    genres = @("Rock", "Pop")
} | ConvertTo-Json

$response = Invoke-HTTPSRequest
```

**Očekivani rezultat:** Status 201 (Created) sa artist podacima

### Korak 3: Proverite logove
```powershell
# Docker logs
docker logs projekat-2025-2-content-service-1 --tail 30 | Select-String "ADMIN_ACTIVITY"

# Ili log fajl
docker exec projekat-2025-2-content-service-1 sh -c "find /app/logs -name '*.log' -exec grep -l 'ADMIN_ACTIVITY' {} \; 2>/dev/null | head -1 | xargs tail -10"
```

**Očekivani log:**
```
[AUDIT] EventType=ADMIN_ACTIVITY Message=... Action=CREATE_ARTIST Resource=artists AdminID=... artistId=... name=Test Artist
```

---

## 🧪 TEST 5: Administratorske Aktivnosti - UPDATE Artist

### Korak 1: Ažurirajte artist
```powershell
# Prvo dobijte artist ID iz prethodnog koraka
$artistId = "artist_id_here"  # Zamenite sa stvarnim ID-jem

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}
$body = @{
    name = "Updated Artist Name"
    biography = "Updated biography"
    genres = @("Jazz", "Blues")
} | ConvertTo-Json

$response = Invoke-HTTPSRequest
```

**Očekivani rezultat:** Status 200 (OK) sa ažuriranim podacima

### Korak 2: Proverite logove
```powershell
# Admin aktivnost
docker logs projekat-2025-2-content-service-1 --tail 30 | Select-String "ADMIN_ACTIVITY.*UPDATE_ARTIST"

# Promena state podataka
docker logs projekat-2025-2-content-service-1 --tail 30 | Select-String "STATE_CHANGE"
```

**Očekivani logovi:**
```
[AUDIT] EventType=ADMIN_ACTIVITY ... Action=UPDATE_ARTIST ...
[AUDIT] EventType=STATE_CHANGE ... Entity=artist OldState=... NewState=...
```

---

## 🧪 TEST 6: Administratorske Aktivnosti - DELETE Artist

### Korak 1: Obrišite artist
```powershell
$headers = @{
    "Authorization" = "Bearer $token"
}

Invoke-HTTPSRequest
```

**Očekivani rezultat:** Status 204 (No Content)

### Korak 2: Proverite logove
```powershell
docker logs projekat-2025-2-content-service-1 --tail 30 | Select-String "ADMIN_ACTIVITY.*DELETE_ARTIST"
```

**Očekivani log:**
```
[AUDIT] EventType=ADMIN_ACTIVITY ... Action=DELETE_ARTIST ... artistId=... name=...
```

---

## 🧪 TEST 7: TLS Greške

### Korak 1: Simulacija TLS greške
TLS greške se automatski loguju kada:
- API Gateway ne može da se poveže sa backend servisom preko HTTPS
- Events Emitter ne može da pošalje event preko HTTPS

### Korak 2: Proverite logove
```powershell
# API Gateway
docker logs projekat-2025-2-api-gateway-1 --tail 50 | Select-String "TLS_FAILURE"

# Content Service (Events Emitter)
docker logs projekat-2025-2-content-service-1 --tail 50 | Select-String "TLS_FAILURE"
```

---

## 📊 Kompletan Pregled Logova

### Pregled svih logova po tipu:

```powershell
# ACCESS_CONTROL_FAILURE
docker logs projekat-2025-2-api-gateway-1 2>&1 | Select-String "ACCESS_CONTROL_FAILURE"

# INVALID_TOKEN
docker logs projekat-2025-2-api-gateway-1 2>&1 | Select-String "INVALID_TOKEN"

# EXPIRED_TOKEN
docker logs projekat-2025-2-api-gateway-1 2>&1 | Select-String "EXPIRED_TOKEN"

# ADMIN_ACTIVITY
docker logs projekat-2025-2-content-service-1 2>&1 | Select-String "ADMIN_ACTIVITY"

# STATE_CHANGE
docker logs projekat-2025-2-content-service-1 2>&1 | Select-String "STATE_CHANGE"

# TLS_FAILURE (iz svih servisa)
docker logs projekat-2025-2-api-gateway-1 2>&1 | Select-String "TLS_FAILURE"
docker logs projekat-2025-2-content-service-1 2>&1 | Select-String "TLS_FAILURE"
```

### Pregled poslednjih logova:

```powershell
# API Gateway - poslednjih 50 linija
docker logs projekat-2025-2-api-gateway-1 --tail 50

# Content Service - poslednjih 50 linija
docker logs projekat-2025-2-content-service-1 --tail 50
```

---

## 🔍 Provera Log Fajlova u Kontejnerima

### Provera da li postoje log fajlovi:

```powershell
# API Gateway
docker exec projekat-2025-2-api-gateway-1 sh -c "ls -lh /app/logs/ 2>/dev/null || echo 'Log direktorijum ne postoji'"

# Content Service
docker exec projekat-2025-2-content-service-1 sh -c "ls -lh /app/logs/ 2>/dev/null || echo 'Log direktorijum ne postoji'"
```

### Čitanje log fajlova:

```powershell
# API Gateway
docker exec projekat-2025-2-api-gateway-1 sh -c "find /app/logs -name '*.log' -type f -exec cat {} \;"

# Content Service
docker exec projekat-2025-2-content-service-1 sh -c "find /app/logs -name '*.log' -type f -exec cat {} \;"
```

---

## ⚠️ Rešavanje Problema

### Problem: Logovi se ne vide u fajlovima

**Rešenje 1:** Logger možda koristi stdout (Docker logs)
```powershell
docker logs projekat-2025-2-api-gateway-1 --tail 100
```

**Rešenje 2:** Kreirajte log direktorijum ručno
```powershell
docker exec projekat-2025-2-api-gateway-1 sh -c "mkdir -p /app/logs && chmod 755 /app/logs"
docker restart projekat-2025-2-api-gateway-1
```

### Problem: Nema logova nakon testiranja

**Rešenje:** Proverite da li se logovanje uopšte poziva
```powershell
# Proverite da li servisi rade
docker ps

# Proverite greške u kontejnerima
docker logs projekat-2025-2-api-gateway-1 --tail 50 | Select-String "ERROR|WARN"
```

---

## ✅ Checklist Testiranja

- [ ] Test 1: Neuspeh kontrole pristupa (bez tokena) - ✅
- [ ] Test 2: Nevalidni tokeni - ✅
- [ ] Test 3: Neuspeh kontrole pristupa - RequireRole - ✅
- [ ] Test 4: Administratorske aktivnosti - CREATE - ⏳ (zahteva admin login)
- [ ] Test 5: Administratorske aktivnosti - UPDATE - ⏳ (zahteva admin login)
- [ ] Test 6: Administratorske aktivnosti - DELETE - ⏳ (zahteva admin login)
- [ ] Test 7: TLS greške - ✅ (automatski)

---

## 📝 Napomene

1. **Logger koristi stdout ako ne može da kreira fajl** - u tom slučaju koristite `docker logs`
2. **Admin korisnik mora postojati** - za testiranje ADMIN_ACTIVITY
3. **Tokeni imaju expiration** - za testiranje EXPIRED_TOKEN, sačekajte da token istekne
4. **Logovi se rotiraju** - kada fajl dostigne 10MB, kreira se novi

---

## 🚀 Brzi Test (Sve odjednom)

```powershell
# 1. Test bez tokena
Invoke-HTTPSRequest

# 2. Test sa nevalidnim tokenom
$headers = @{ "Authorization" = "Bearer invalid" }
Invoke-HTTPSRequest

# 3. Pregled logova
docker logs projekat-2025-2-api-gateway-1 --tail 30 | Select-String "ACCESS_CONTROL|INVALID_TOKEN"
```
