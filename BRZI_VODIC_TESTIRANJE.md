# ⚡ Brzi Vodič za Testiranje - Informaciona Bezbednost

## 🚀 Brzo Pokretanje (3 Koraka)

### Korak 1: Pokrenite Sistem

```powershell
cd C:\Users\ivana\Desktop\Projekat-2025-1
docker-compose up -d
Start-Sleep -Seconds 30
```

### Korak 2: Učitajte Helper Funkcije

```powershell
. .\https-helper.ps1
```

### Korak 3: Pokrenite Kompletno Testiranje

```powershell
.\test-informaciona-bezbednost.ps1
```

**To je sve!** Skripta će automatski testirati sve zahteve i generisati izveštaj.

---

## 📋 Pojedinačno Testiranje

### Test 1: Registracija (1.1)

```powershell
# Uspešna registracija
$body = @{
    firstName = "Test"
    lastName = "User"
    email = "test@example.com"
    username = "testuser"
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"
# Očekivano: HTTP 201
```

### Test 2: Prijava (1.2)

```powershell
# Request OTP
$body = @{username = "admin"; password = "admin123"} | ConvertTo-Json
Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/login/request-otp" -Method "POST" -Body $body -ContentType "application/json"

# Proverite logove za OTP
docker logs projekat-2025-2-users-service-1 | Select-String "OTP"

# Verify OTP (zamenite sa stvarnim OTP kodom)
$body = @{username = "admin"; otp = "123456"} | ConvertTo-Json
Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/login/verify-otp" -Method "POST" -Body $body -ContentType "application/json"
# Očekivano: HTTP 200 sa tokenom
```

### Test 3: Magic Link (1.3)

```powershell
$body = @{email = "admin@example.com"} | ConvertTo-Json
Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/recover/request" -Method "POST" -Body $body -ContentType "application/json"
# Očekivano: HTTP 200
```

### Test 4: Autorizacija (2.17)

```powershell
# Bez tokena
Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/logout" -Method "POST" -Body "{}" -ContentType "application/json"
# Očekivano: HTTP 401

# Sa tokenom (zamenite sa stvarnim tokenom)
Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/logout" -Method "POST" -Body "{}" -ContentType "application/json" -Headers @{Authorization = "Bearer YOUR_TOKEN"}
# Očekivano: HTTP 200
```

### Test 5: Validacija (2.18)

```powershell
# SQL Injection
.\test-sql-injection-attack.ps1

# XSS
.\test-xss-attack.ps1
```

### Test 6: HTTPS (2.19)

```powershell
# Provera HTTPS
Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/health" -Method "GET"
# Očekivano: HTTP 200 (HTTPS radi)

# Provera hash lozinki
docker exec projekat-2025-2-mongodb-users-1 mongosh --quiet --eval "db.users.findOne({}, {passwordHash: 1, email: 1, _id: 0})"
# Očekivano: passwordHash počinje sa $2a$ ili $2b$
```

### Test 7: Logovanje (2.20)

```powershell
# Provera log fajlova
ls services/users-service/logs/

# Provera logovanja
docker logs projekat-2025-2-users-service-1 | Select-String "VALIDATION_FAILURE|LOGIN_FAILURE|LOGIN_SUCCESS"
```

### Test 8: Demonstracija Napada (2.22)

```powershell
# Svi napadi odjednom
.\test-all-attacks.ps1

# Ili pojedinačno:
.\test-xss-attack.ps1
.\test-sql-injection-attack.ps1
.\test-brute-force-attack.ps1
.\test-dos-attack.ps1
```

---

## ✅ Checklist Pre Odbrane

- [ ] Sistem je pokrenut (`docker-compose ps`)
- [ ] HTTPS radi (`Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/health"`)
- [ ] Test skripte rade (`.\test-all-attacks.ps1`)
- [ ] Logovi se generišu (`ls services/users-service/logs/`)
- [ ] Izveštaj o ranjivostima postoji (`ls IZVESTAJ_ANALIZA_RANJIVOSTI_2.21.md`)

---

## 🎯 Za Odbranu

### Demonstracija:

1. **Pokrenite sistem** (`docker-compose up -d`)
2. **Pokrenite test skripte** (`.\test-all-attacks.ps1`)
3. **Pokažite logove** (`docker logs ...`)
4. **Pokažite kod** (otvorite fajlove u IDE-u)
5. **Objasnite mehanizme** (kako funkcioniše zaštita)

### Ključne Komande:

```powershell
# Kompletno testiranje
.\test-informaciona-bezbednost.ps1

# Demonstracija napada
.\test-all-attacks.ps1

# Provera logova
docker logs projekat-2025-2-users-service-1 | Select-String "VALIDATION_FAILURE|LOGIN_FAILURE"

# Provera HTTPS
Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/health" -Method "GET"
```

---

**Detaljna dokumentacija:** `TEST_PLAN_INFORMACIONA_BEZBEDNOST.md`
