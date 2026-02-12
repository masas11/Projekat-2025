# Demonstracija Pokušaja Napada - 2.22

**Datum:** 12. februar 2025  
**Aplikacija:** Music Streaming Platform  
**Svrha:** Demonstracija zaštite od XSS, SQL Injection, Brute-force i DoS napada

---

## 📋 Sadržaj

1. [Uvod](#uvod)
2. [Napad 1: XSS (Cross-Site Scripting)](#napad-1-xss)
3. [Napad 2: SQL Injection](#napad-2-sql-injection)
4. [Napad 3: Brute-force Attack](#napad-3-brute-force)
5. [Napad 4: DoS (Denial of Service)](#napad-4-dos)
6. [Rezime Demonstracije](#rezime)

---

## 1. Uvod

Ova demonstracija pokazuje kako aplikacija brani od četiri glavna tipa napada:
- **XSS (Cross-Site Scripting)** - Pokušaj inject-ovanja malicioznog JavaScript koda
- **SQL Injection** - Pokušaj inject-ovanja SQL komandi
- **Brute-force Attack** - Pokušaj probijanja lozinke kroz više pokušaja
- **DoS (Denial of Service)** - Pokušaj preopterećenja servera

Za svaki napad ćemo:
1. Pokazati kako bi napad izgledao
2. Pokrenuti napad
3. Demonstrirati kako aplikacija brani od napada
4. Objasniti mehanizme zaštite

---

## 2. Napad 1: XSS (Cross-Site Scripting)

### 2.1. Opis Napada

**XSS napad** pokušava da inject-uje maliciozni JavaScript kod u aplikaciju koji se izvršava u browser-u žrtve.

### 2.2. Vrste XSS Napada

#### 2.2.1. Stored XSS
- Maliciozni kod se čuva u bazi podataka
- Izvršava se svaki put kada se podaci prikažu

#### 2.2.2. Reflected XSS
- Maliciozni kod se reflektuje u odgovoru servera
- Izvršava se jednom, kada korisnik klikne na link

#### 2.2.3. DOM-based XSS
- Maliciozni kod se inject-uje direktno u DOM
- Izvršava se u browser-u bez komunikacije sa serverom

### 2.3. Pokušaj Napada

#### Test 1: Osnovni XSS Pattern
```powershell
# Pokušaj registracije sa XSS payload-om
$body = @{
    firstName = "<script>alert('XSS')</script>"
    lastName = "User"
    email = "xss@test.com"
    username = "xssuser"
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

. .\https-helper.ps1
$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"
Write-Host "Status: $($result.StatusCode)"
Write-Host "Response: $($result.Content)"
```

**Očekivani rezultat:** HTTP 400 "invalid input" - XSS pattern je detektovan i blokiran

#### Test 2: XSS sa Encoding-om
```powershell
# Pokušaj sa HTML encoding-om
$body = @{
    firstName = "&lt;script&gt;alert('XSS')&lt;/script&gt;"
    lastName = "User"
    email = "xss2@test.com"
    username = "xssuser2"
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"
Write-Host "Status: $($result.StatusCode)"
```

#### Test 3: Event Handler XSS
```powershell
# Pokušaj sa event handler-om
$body = @{
    firstName = "<img src=x onerror=alert('XSS')>"
    lastName = "User"
    email = "xss3@test.com"
    username = "xssuser3"
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"
Write-Host "Status: $($result.StatusCode)"
```

### 2.4. Mehanizmi Zaštite

#### 2.4.1. Input Validation
**Lokacija:** `services/users-service/internal/validation/input.go:121-144`

```go
func CheckXSS(input string) error {
    xssPatterns := []string{
        "<script",
        "</script>",
        "javascript:",
        "onerror=",
        "onload=",
        "onclick=",
        "<iframe",
        "<img",
        "<svg",
    }
    // Pattern matching
}
```

**Kako radi:**
- Proverava input pre čuvanja u bazi
- Detektuje osnovne XSS pattern-e
- Blokira zahtev ako se detektuje XSS

#### 2.4.2. Output Encoding
**Lokacija:** `services/users-service/internal/security/encoding.go`

```go
func EscapeHTML(input string) string {
    return html.EscapeString(input)
}
```

**Kako radi:**
- Escape-uje HTML specijalne karaktere pri prikazu
- `<script>` postaje `&lt;script&gt;`
- Browser prikazuje tekst umesto izvršavanja koda

#### 2.4.3. React Automatski Escape
**Lokacija:** `frontend/src/components/*.js`

**Kako radi:**
- React automatski escape-uje sve vrednosti pri render-ovanju
- `{user.firstName}` je automatski escape-ovano
- Dodatna zaštita na client strani

### 2.5. Demonstracija Zaštite

**Scenario:**
1. Napadač pokušava da registruje korisnika sa XSS payload-om
2. Server detektuje XSS pattern u `CheckXSS()` funkciji
3. Server vraća HTTP 400 "invalid input"
4. Zahtev je blokiran, XSS napad neuspešan

**Logovanje:**
```
[AUDIT] EventType=VALIDATION_FAILURE Message=... Field=firstName Reason=XSS attempt detected
```

---

## 3. Napad 2: SQL Injection

### 3.1. Opis Napada

**SQL Injection napad** pokušava da inject-uje SQL komande u input polja kako bi manipulisao bazom podataka.

### 3.2. Vrste SQL Injection Napada

#### 3.2.1. Classic SQL Injection
```sql
' OR '1'='1
```

#### 3.2.2. Union-based SQL Injection
```sql
' UNION SELECT * FROM users--
```

#### 3.2.3. Time-based Blind SQL Injection
```sql
'; WAITFOR DELAY '00:00:05'--
```

### 3.3. Pokušaj Napada

#### Test 1: Osnovni SQL Injection Pattern
```powershell
# Pokušaj registracije sa SQL injection payload-om
$body = @{
    firstName = "Test' OR '1'='1"
    lastName = "User"
    email = "sqli@test.com"
    username = "sqliuser"
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

. .\https-helper.ps1
$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"
Write-Host "Status: $($result.StatusCode)"
Write-Host "Response: $($result.Content)"
```

**Očekivani rezultat:** HTTP 400 "invalid input" - SQL injection pattern je detektovan i blokiran

#### Test 2: SQL Injection sa UNION
```powershell
$body = @{
    firstName = "Test"
    lastName = "User' UNION SELECT * FROM users--"
    email = "sqli2@test.com"
    username = "sqliuser2"
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"
Write-Host "Status: $($result.StatusCode)"
```

#### Test 3: SQL Injection sa DROP TABLE
```powershell
$body = @{
    firstName = "Test'; DROP TABLE users--"
    lastName = "User"
    email = "sqli3@test.com"
    username = "sqliuser3"
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" -Method "POST" -Body $body -ContentType "application/json"
Write-Host "Status: $($result.StatusCode)"
```

### 3.4. Mehanizmi Zaštite

#### 3.4.1. Input Validation
**Lokacija:** `services/users-service/internal/validation/input.go:98-119`

```go
func CheckSQLInjection(input string) error {
    sqlPatterns := []string{
        "' OR '1'='1",
        "' OR '1'='1'--",
        "'; DROP TABLE",
        "UNION SELECT",
        "'; INSERT INTO",
        "'; UPDATE",
        "'; DELETE FROM",
    }
    // Pattern matching
}
```

**Kako radi:**
- Proverava input pre obrade
- Detektuje osnovne SQL injection pattern-e
- Blokira zahtev ako se detektuje SQL injection

#### 3.4.2. Parameterized Queries (MongoDB)
**Lokacija:** `services/users-service/internal/store/user_repository.go`

**Kako radi:**
- MongoDB driver koristi parameterized queries
- Input se tretira kao podatak, ne kao kod
- SQL injection nije moguć jer nema SQL sintaksu

**Primer:**
```go
// MongoDB query - bezbedno
filter := bson.M{"username": username}  // username je parameter, ne kod
user := &model.User{}
err := collection.FindOne(ctx, filter).Decode(user)
```

### 3.5. Demonstracija Zaštite

**Scenario:**
1. Napadač pokušava da registruje korisnika sa SQL injection payload-om
2. Server detektuje SQL injection pattern u `CheckSQLInjection()` funkciji
3. Server vraća HTTP 400 "invalid input"
4. Zahtev je blokiran, SQL injection napad neuspešan

**Logovanje:**
```
[AUDIT] EventType=VALIDATION_FAILURE Message=... Field=firstName Reason=SQL injection attempt detected
```

---

## 4. Napad 3: Brute-force Attack

### 4.1. Opis Napada

**Brute-force napad** pokušava da probije lozinku kroz više pokušaja sa različitim kombinacijama.

### 4.2. Vrste Brute-force Napada

#### 4.2.1. Dictionary Attack
- Pokušaj sa listom čestih lozinki

#### 4.2.2. Credential Stuffing
- Pokušaj sa ukradenim credentials-ima

#### 4.2.3. Password Spraying
- Pokušaj sa malim brojem čestih lozinki na više naloga

### 4.3. Pokušaj Napada

#### Test 1: Višestruki Neuspešni Pokušaji
```powershell
# Simulacija brute-force napada
. .\https-helper.ps1

$username = "testuser"
$passwords = @("password", "123456", "admin", "test", "qwerty", "password123")

Write-Host "=== BRUTE-FORCE NAPAD SIMULACIJA ===" -ForegroundColor Red
Write-Host "Pokušaj probijanja lozinke za korisnika: $username" -ForegroundColor Yellow

$attempt = 0
foreach ($password in $passwords) {
    $attempt++
    Write-Host "`nPokušaj $attempt : $password" -ForegroundColor Gray
    
    $body = @{
        username = $username
        password = $password
    } | ConvertTo-Json
    
    $result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/login/request-otp" -Method "POST" -Body $body -ContentType "application/json"
    
    if ($result.StatusCode -eq 200) {
        Write-Host "  ✓ USPEH! Lozinka probijena: $password" -ForegroundColor Green
        break
    } else {
        Write-Host "  ✗ Neuspešno (Status: $($result.StatusCode))" -ForegroundColor Red
    }
    
    Start-Sleep -Milliseconds 500
}

Write-Host "`n=== NAPAD ZAVRŠEN ===" -ForegroundColor Red
```

**Očekivani rezultat:**
- Prvih nekoliko pokušaja prolazi (ali vraća grešku za lozinku)
- Nakon 5 neuspešnih pokušaja, nalog se zaključava
- Dalji pokušaji vraćaju HTTP 403 "account locked"

#### Test 2: Account Locking
```powershell
# Test account locking mehanizma
$body = @{
    username = "testuser"
    password = "wrongpassword"
} | ConvertTo-Json

# 5 neuspešnih pokušaja
for ($i = 1; $i -le 5; $i++) {
    Write-Host "Pokušaj $i..." -ForegroundColor Yellow
    $result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/login/request-otp" -Method "POST" -Body $body -ContentType "application/json"
    Write-Host "  Status: $($result.StatusCode)" -ForegroundColor Gray
    Start-Sleep -Seconds 1
}

# 6. pokušaj - trebalo bi biti blokirano
Write-Host "`n6. pokušaj (nakon 5 neuspešnih)..." -ForegroundColor Red
$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/login/request-otp" -Method "POST" -Body $body -ContentType "application/json"
Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 403) { "Green" } else { "Yellow" })
Write-Host "Response: $($result.Content)" -ForegroundColor Gray
```

### 4.4. Mehanizmi Zaštite

#### 4.4.1. Account Locking
**Lokacija:** `services/users-service/internal/handler/login_handler.go:78-84`

```go
if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
    user.FailedLoginAttempts++
    if user.FailedLoginAttempts >= 5 {
        user.LockedUntil = time.Now().Add(15 * time.Minute)
    }
    h.Repo.Update(ctx, user)
    // ...
}
```

**Kako radi:**
- Broji neuspešne pokušaje prijave
- Nakon 5 neuspešnih pokušaja, zaključava nalog na 15 minuta
- Blokira dalje pokušaje prijave

#### 4.4.2. Rate Limiting
**Lokacija:** `services/users-service/internal/middleware/rate_limit.go`

```go
// 10 zahteva po minuti po IP adresi za osetljive endpoint-e
rateLimit := middleware.RateLimit(10, 1*time.Minute)
```

**Kako radi:**
- Ograničava broj zahteva po IP adresi
- Za login endpoint: 10 zahteva po minuti
- Blokira prekomerno slanje zahteva

#### 4.4.3. Logovanje Neuspešnih Pokušaja
**Lokacija:** `services/users-service/internal/handler/login_handler.go:85-87`

```go
h.Logger.LogLoginFailure(req.Username, "invalid password", ipAddress)
```

**Kako radi:**
- Loguje svaki neuspešan pokušaj prijave
- Omogućava detekciju brute-force napada
- Pomaže u forensics analizi

### 4.5. Demonstracija Zaštite

**Scenario:**
1. Napadač pokušava više lozinki za isti nalog
2. Prvih 5 pokušaja prolazi (ali vraćaju grešku)
3. Nakon 5. pokušaja, nalog se zaključava
4. 6. i dalji pokušaji vraćaju HTTP 403 "account locked"
5. Nalog je zaključan 15 minuta

**Logovanje:**
```
[AUDIT] EventType=LOGIN_FAILURE Message=... Username=testuser Reason=invalid password IP=...
[AUDIT] EventType=LOGIN_FAILURE Message=... Username=testuser Reason=account locked IP=...
```

---

## 5. Napad 4: DoS (Denial of Service)

### 5.1. Opis Napada

**DoS napad** pokušava da preoptereti server velikim brojem zahteva, čineći ga nedostupnim.

### 5.2. Vrste DoS Napada

#### 5.2.1. Volume-based DoS
- Veliki broj zahteva
- Preopterećenje bandwidth-a

#### 5.2.2. Protocol-based DoS
- Eksploatacija protokola (npr. TCP SYN flood)

#### 5.2.3. Application-layer DoS
- Napad na aplikacijski sloj
- Eksploatacija resource-intensive operacija

### 5.3. Pokušaj Napada

#### Test 1: Volume-based DoS
```powershell
# Simulacija DoS napada - veliki broj zahteva
. .\https-helper.ps1

Write-Host "=== DoS NAPAD SIMULACIJA ===" -ForegroundColor Red
Write-Host "Slanje velikog broja zahteva..." -ForegroundColor Yellow

$successCount = 0
$blockedCount = 0
$totalRequests = 150  # Više od limita (100/min)

for ($i = 1; $i -le $totalRequests; $i++) {
    $result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/health" -Method "GET"
    
    if ($result.StatusCode -eq 200) {
        $successCount++
    } elseif ($result.StatusCode -eq 429) {
        $blockedCount++
        Write-Host "Zahtev $i : BLOKIRAN (429 Too Many Requests)" -ForegroundColor Red
    }
    
    if ($i % 20 -eq 0) {
        Write-Host "Progres: $i/$totalRequests (Uspešno: $successCount, Blokirano: $blockedCount)" -ForegroundColor Gray
    }
    
    Start-Sleep -Milliseconds 100
}

Write-Host "`n=== REZULTATI ===" -ForegroundColor Cyan
Write-Host "Ukupno zahteva: $totalRequests" -ForegroundColor White
Write-Host "Uspešno: $successCount" -ForegroundColor Green
Write-Host "Blokirano: $blockedCount" -ForegroundColor Red
Write-Host "Procenat blokiranih: $([math]::Round(($blockedCount/$totalRequests)*100, 2))%" -ForegroundColor Yellow
```

**Očekivani rezultat:**
- Prvih ~100 zahteva prolazi
- Preko 100 zahteva se blokira (HTTP 429)
- Rate limiting zaštiti server od preopterećenja

#### Test 2: Distributed DoS (DDoS) Simulacija
```powershell
# Simulacija DDoS sa više "IP adresa" (simulirano kroz različite identifikatore)
Write-Host "=== DDoS SIMULACIJA ===" -ForegroundColor Red

$ips = @("192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.4", "192.168.1.5")
$requestsPerIP = 30

foreach ($ip in $ips) {
    Write-Host "`nNapad sa IP: $ip" -ForegroundColor Yellow
    $blocked = 0
    
    for ($i = 1; $i -le $requestsPerIP; $i++) {
        # Simulacija: svaki "IP" šalje zahteve
        $result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/health" -Method "GET"
        
        if ($result.StatusCode -eq 429) {
            $blocked++
        }
        
        Start-Sleep -Milliseconds 50
    }
    
    Write-Host "  Blokirano: $blocked/$requestsPerIP" -ForegroundColor $(if ($blocked -gt 0) { "Red" } else { "Green" })
}
```

### 5.4. Mehanizmi Zaštite

#### 5.4.1. Rate Limiting
**Lokacija:** `services/api-gateway/internal/middleware/rate_limit.go`

```go
// Global rate limiting: 100 requests per minute per IP
globalRateLimit := middleware.RateLimit(100, 1*time.Minute)
```

**Kako radi:**
- Ograničava broj zahteva po IP adresi
- Globalni limit: 100 zahteva po minuti
- Blokira prekomerno slanje zahteva (HTTP 429)

#### 5.4.2. Per-Endpoint Rate Limiting
**Lokacija:** `services/users-service/internal/middleware/rate_limit.go`

```go
// Osetljivi endpoint-i: 10 requests per minute
rateLimit := middleware.RateLimit(10, 1*time.Minute)
```

**Kako radi:**
- Stricte limit za osetljive endpoint-e
- Login, register, password reset: 10 zahteva po minuti
- Dodatna zaštita od brute-force i DoS

#### 5.4.3. Request Timeout
**Lokacija:** `services/api-gateway/cmd/main.go:79-81`

```go
client := &http.Client{
    Timeout: 5 * time.Second,  // Timeout za pozive backend servisa
}
```

**Kako radi:**
- Ograničava vreme izvršavanja zahteva
- Prekida dugotrajne zahteve
- Sprečava resource exhaustion

### 5.5. Demonstracija Zaštite

**Scenario:**
1. Napadač šalje veliki broj zahteva (>100/min)
2. Prvih 100 zahteva prolazi normalno
3. Preko 100 zahteva se blokira (HTTP 429)
4. Server ostaje dostupan za legitimne zahteve
5. Rate limiting zaštiti server od preopterećenja

**Logovanje:**
```
[AUDIT] EventType=ACCESS_CONTROL_FAILURE Message=... Reason=rate limit exceeded
```

---

## 6. Rezime Demonstracije

### 6.1. Tabela Rezultata

| Napad | Status | Zaštita | Rezultat |
|-------|--------|---------|----------|
| XSS | ✅ Blokiran | Input validation + Output encoding | Neuspešan |
| SQL Injection | ✅ Blokiran | Input validation + Parameterized queries | Neuspešan |
| Brute-force | ✅ Blokiran | Account locking + Rate limiting | Neuspešan |
| DoS | ✅ Blokiran | Rate limiting + Timeout | Neuspešan |

### 6.2. Zaključak

**Svi napadi su uspešno blokirani!**

Aplikacija ima višeslojnu zaštitu:
1. **Input Validation** - Blokira maliciozne input-e
2. **Output Encoding** - Sprečava XSS pri prikazu
3. **Account Locking** - Zaštita od brute-force
4. **Rate Limiting** - Zaštita od DoS
5. **Logovanje** - Detekcija i monitoring napada

### 6.3. Preporuke za Produkciju

1. **WAF (Web Application Firewall)** - Dodatna zaštita
2. **DDoS Protection** - Cloudflare, AWS Shield
3. **CAPTCHA** - Za osetljive endpoint-e
4. **IP Whitelisting** - Za administratorske funkcije
5. **Monitoring & Alerting** - Automatska detekcija napada

---

## Dodatak A: Test Skripte

Sve test skripte su dostupne u:
- `test-xss-attack.ps1`
- `test-sql-injection-attack.ps1`
- `test-brute-force-attack.ps1`
- `test-dos-attack.ps1`

---

**Izveštaj pripremio:** Security Testing Team  
**Datum:** 12. februar 2025  
**Verzija:** 1.0
