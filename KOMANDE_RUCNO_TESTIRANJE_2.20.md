# 🔧 Komande za Ručno Testiranje - Zahtev 2.20 Logovanje

## 📋 PRIprema

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

---

## ✅ TEST 1: Logovanje Neuspeha Validacije

### Komanda za Test:

```powershell
# Pokušaj registracije sa XSS payload-om
$body = @{
    firstName = "<script>alert('XSS')</script>"
    lastName = "Test"
    email = "xss-test@example.com"
    username = "xsstest"
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"
```

### Provera Logova:

```powershell
# Provera VALIDATION_FAILURE logova
docker exec projekat-2025-1-users-service-1 cat /app/logs/app-$(Get-Date -Format 'yyyy-MM-dd').log | Select-String "VALIDATION_FAILURE"
```

**Očekivano:** Vidite log entry sa `EventType=VALIDATION_FAILURE` i razlogom "XSS attempt detected"

---

## ✅ TEST 2: Logovanje Pokušaja Prijave

### Komanda za Neuspešnu Prijavu:

```powershell
# Neuspešna prijava
$body = @{
    username = "nonexistent"
    password = "wrongpassword"
} | ConvertTo-Json

Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/login/request-otp" -Method "POST" -Body $body -ContentType "application/json"
```

### Provera Logova:

```powershell
# Provera LOGIN_FAILURE logova
docker exec projekat-2025-1-users-service-1 cat /app/logs/app-$(Get-Date -Format 'yyyy-MM-dd').log | Select-String "LOGIN_FAILURE"
```

**Očekivano:** Vidite log entry sa `EventType=LOGIN_FAILURE` i razlogom "user not found"

### Komanda za Uspešnu Prijavu (ako admin postoji):

```powershell
# Uspešna prijava (request OTP)
$body = @{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json

Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/login/request-otp" -Method "POST" -Body $body -ContentType "application/json"
```

### Provera OTP-a:

```powershell
# Provera OTP-a u mock email logovima
docker logs projekat-2025-1-users-service-1 | Select-String "Sending OTP" | Select-Object -Last 1
```

**Očekivano:** Vidite "Sending OTP XXXXXX to admin@..."

---

## ✅ TEST 3: Logovanje Neuspeha Kontrole Pristupa

### Komanda za Test (bez tokena):

```powershell
# Pokušaj pristupa zaštićenom endpoint-u bez tokena
Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/logout" -Method "POST" -Body "{}" -ContentType "application/json"
```

### Provera Logova:

```powershell
# Provera ACCESS_CONTROL_FAILURE logova
docker exec projekat-2025-1-api-gateway-1 cat /app/logs/app-$(Get-Date -Format 'yyyy-MM-dd').log | Select-String "ACCESS_CONTROL_FAILURE"
```

**Očekivano:** Vidite log entry sa `EventType=ACCESS_CONTROL_FAILURE` i razlogom "missing authorization header"

### Komanda za Test (nevažeći token):

```powershell
# Pokušaj sa nevažećim tokenom
$headers = @{ Authorization = "Bearer invalid-token-12345" }
Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/logout" -Method "POST" -Body "{}" -ContentType "application/json" -Headers $headers
```

### Provera Logova:

```powershell
# Provera INVALID_TOKEN logova
docker exec projekat-2025-1-api-gateway-1 cat /app/logs/app-$(Get-Date -Format 'yyyy-MM-dd').log | Select-String "INVALID_TOKEN"
```

**Očekivano:** Vidite log entry sa `EventType=INVALID_TOKEN` i razlogom "token is malformed"

---

## ✅ TEST 4: Provera Rotacije Logova

### Komanda:

```powershell
# Lista log fajlova u kontejneru
docker exec projekat-2025-1-users-service-1 ls -la /app/logs/
```

**Očekivano:** Vidite:
- `app-YYYY-MM-DD.log` - trenutni fajl
- `app-YYYY-MM-DD.log.YYYYMMDD-HHMMSS` - rotirani fajlovi (ako postoje)

### Provera Veličine Fajlova:

```powershell
# Provera veličine log fajlova
docker exec projekat-2025-1-users-service-1 sh -c "ls -lh /app/logs/*.log"
```

**Očekivano:** Vidite veličinu fajlova (rotacija se dešava na 10MB)

---

## ✅ TEST 5: Provera Zaštite Log-Datoteka

### Komanda:

```powershell
# Provera permisija log fajlova
docker exec projekat-2025-1-users-service-1 ls -la /app/logs/
```

**Očekivano:** Vidite permisije `-rw-r-----` (0640) - samo vlasnik i grupa mogu čitati

---

## ✅ TEST 6: Provera Integriteta Log-Datoteka

### Komanda:

```powershell
# Lista checksum fajlova
docker exec projekat-2025-1-users-service-1 ls -la /app/logs/*.checksum
```

**Očekivano:** Vidite `.checksum` fajlove za svaki log fajl

### Provera Sadržaja Checksum Fajla:

```powershell
# Provera SHA256 checksum-a
docker exec projekat-2025-1-users-service-1 cat /app/logs/app-$(Get-Date -Format 'yyyy-MM-dd').log.checksum
```

**Očekivano:** Vidite SHA256 hash vrednost

---

## ✅ TEST 7: Provera Filtriranja Osetljivih Podataka

### Komanda:

```powershell
# Provera da li se lozinke, tokeni, OTP maskiraju
docker exec projekat-2025-1-users-service-1 cat /app/logs/app-$(Get-Date -Format 'yyyy-MM-dd').log | Select-String "password|token|otp"
```

**Očekivano:** 
- Ako se osetljivi podaci loguju, treba da budu maskirani kao `***`
- Ili se uopšte ne loguju

### Provera Stack Trace-a:

```powershell
# Provera da li se stack trace loguje
docker exec projekat-2025-1-users-service-1 cat /app/logs/app-$(Get-Date -Format 'yyyy-MM-dd').log | Select-String "goroutine|panic|stack trace"
```

**Očekivano:** Nema rezultata (stack trace se ne loguje)

---

## 📊 Kompletan Pregled Logova

### Svi Logovi za Users Service:

```powershell
# Svi logovi
docker exec projekat-2025-1-users-service-1 cat /app/logs/app-$(Get-Date -Format 'yyyy-MM-dd').log
```

### Filtrirani Logovi po Tipu:

```powershell
# Samo VALIDATION_FAILURE
docker exec projekat-2025-1-users-service-1 cat /app/logs/app-$(Get-Date -Format 'yyyy-MM-dd').log | Select-String "VALIDATION_FAILURE"

# Samo LOGIN_FAILURE
docker exec projekat-2025-1-users-service-1 cat /app/logs/app-$(Get-Date -Format 'yyyy-MM-dd').log | Select-String "LOGIN_FAILURE"

# Samo LOGIN_SUCCESS
docker exec projekat-2025-1-users-service-1 cat /app/logs/app-$(Get-Date -Format 'yyyy-MM-dd').log | Select-String "LOGIN_SUCCESS"
```

### Svi Logovi za API Gateway:

```powershell
# Svi logovi
docker exec projekat-2025-1-api-gateway-1 cat /app/logs/app-$(Get-Date -Format 'yyyy-MM-dd').log

# Samo ACCESS_CONTROL_FAILURE
docker exec projekat-2025-1-api-gateway-1 cat /app/logs/app-$(Get-Date -Format 'yyyy-MM-dd').log | Select-String "ACCESS_CONTROL_FAILURE"

# Samo INVALID_TOKEN
docker exec projekat-2025-1-api-gateway-1 cat /app/logs/app-$(Get-Date -Format 'yyyy-MM-dd').log | Select-String "INVALID_TOKEN"
```

---

## 🎯 Brzi Test - Svi Tipovi Logovanja Odjednom

### Komanda:

```powershell
# Pokrenite sve testove odjednom
.\test-logging-2.20.ps1
```

**Očekivano:** Vidite rezultate svih testova i procenat uspešnosti

---

## 📋 Checklist za Ručno Testiranje

- [ ] **TEST 1:** VALIDATION_FAILURE logovanje radi
- [ ] **TEST 2:** LOGIN_FAILURE logovanje radi
- [ ] **TEST 3:** ACCESS_CONTROL_FAILURE logovanje radi
- [ ] **TEST 4:** INVALID_TOKEN logovanje radi
- [ ] **TEST 5:** Rotacija logova radi
- [ ] **TEST 6:** Permisije log fajlova su 0640
- [ ] **TEST 7:** Checksum fajlovi postoje
- [ ] **TEST 8:** Osetljivi podaci se maskiraju ili ne loguju
- [ ] **TEST 9:** Stack trace se ne loguje

---

## 🔍 Korisne Komande za Debugging

### Provera da li servisi rade:

```powershell
docker-compose ps
```

### Provera logova u realnom vremenu:

```powershell
# Users Service
docker logs projekat-2025-1-users-service-1 -f

# API Gateway
docker logs projekat-2025-1-api-gateway-1 -f
```

### Kopiranje Log Fajlova na Host Sistem:

```powershell
# Kopiraj log fajl na host sistem
docker cp projekat-2025-1-users-service-1:/app/logs/app-$(Get-Date -Format 'yyyy-MM-dd').log ./temp-log.log

# Proveri sadržaj
Get-Content ./temp-log.log | Select-String "LOGIN_FAILURE"
```

### Brisanje Starih Logova (opciono):

```powershell
# Brisanje log fajlova u kontejneru (oprezno!)
docker exec projekat-2025-1-users-service-1 sh -c "rm /app/logs/*.log"
```

---

## 📊 Format Log Entry-ja

Svaki log entry ima sledeći format:

```
[LEVEL] YYYY/MM/DD HH:MM:SS [LEVEL] EventType=EVENT_TYPE Message=... Fields=...
```

**Primer:**
```
[WARN] 2026/02/17 16:19:27 [WARN] EventType=LOGIN_FAILURE Message=Login failed Fields=username=nonexistent reason=user not found ip=172.20.0.15 timestamp=1771345167
```

---

## ✅ Finalni Test

### Pokrenite kompletan test:

```powershell
.\test-logging-2.20.ps1
```

**Očekivano rezultat:**
- ✅ SVI TESTOVI PROŠLI
- Procenat uspešnosti: 100%

---

**Napomena:** Ako neki test ne prođe, proverite:
1. Da li su servisi pokrenuti (`docker-compose ps`)
2. Da li su log fajlovi kreirani (`docker exec ... ls /app/logs/`)
3. Da li je datum u imenu fajla tačan (`Get-Date -Format 'yyyy-MM-dd'`)
