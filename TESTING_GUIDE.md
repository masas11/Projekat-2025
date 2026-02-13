# Vodič za Testiranje Sistema

Ovaj dokument objašnjava kako da testirate sve funkcionalnosti sistema, uključujući HTTPS sertifikate, komunikaciju između servisa, MailHog, i sigurnosne mehanizme.

## 📋 Preduslovi

1. **Docker i Docker Compose su pokrenuti**
2. **Svi servisi su pokrenuti**: `docker-compose up -d`
3. **Sertifikati su generisani**: `certs/server.crt` i `certs/server.key` postoje

## 🔍 1. Provera Statusa Servisa

### Provera da li su svi servisi pokrenuti:
```powershell
docker-compose ps
```

Svi servisi treba da budu u statusu "Up".

### Provera logova:
```powershell
# API Gateway
docker logs projekat-2025-1-api-gateway-1 --tail 20

# Users Service
docker logs projekat-2025-1-users-service-1 --tail 20

# MailHog
docker logs projekat-2025-1-mailhog-1 --tail 10
```

## 🔐 2. Testiranje HTTPS Komunikacije Između Servisa

### Test 1: Provera da li servisi koriste HTTPS interno

Proverite logove servisa - trebalo bi da vidite:
- `Starting HTTPS server on port XXXX` (za servise koji imaju sertifikate)
- `Starting HTTP server on port XXXX` (za servise bez sertifikata)

```powershell
# Provera Users Service
docker logs projekat-2025-1-users-service-1 | Select-String "HTTPS|HTTP"

# Provera Content Service  
docker logs projekat-2025-1-content-service-1 | Select-String "HTTPS|HTTP"

# Provera API Gateway
docker logs projekat-2025-1-api-gateway-1 | Select-String "HTTPS|HTTP"
```

**Očekivani rezultat:**
- Users Service: `Starting HTTPS server on port 8001`
- Content Service: `Starting HTTPS server on port 8002`
- API Gateway: `Starting HTTP server on port 8080` (za eksterni pristup)

### Test 2: Testiranje HTTPS komunikacije između servisa

```powershell
# Test između API Gateway-a i Users Service-a
docker exec projekat-2025-1-api-gateway-1 wget --no-check-certificate -O- https://users-service:8001/health

# Test između API Gateway-a i Content Service-a
docker exec projekat-2025-1-api-gateway-1 wget --no-check-certificate -O- https://content-service:8002/health
```

**Očekivani rezultat:** Status 200 OK sa health check porukom.

## 🌐 3. Testiranje API Gateway Endpoint-a

### Test 1: Health Check Endpoint

```powershell
# API Gateway health (ako je registrovan)
Invoke-WebRequest -Uri "http://localhost:8081/health" -UseBasicParsing

# Users Service preko API Gateway-a
Invoke-WebRequest -Uri "http://localhost:8081/api/users/health" -UseBasicParsing

# Content Service preko API Gateway-a
Invoke-WebRequest -Uri "http://localhost:8081/api/content/health" -UseBasicParsing
```

**Očekivani rezultat:** Status 200 OK sa odgovarajućom porukom.

### Test 2: CORS Headers

```powershell
$response = Invoke-WebRequest -Uri "http://localhost:8081/api/users/health" -UseBasicParsing
$response.Headers["Access-Control-Allow-Origin"]
```

**Očekivani rezultat:** `*` ili origin frontend aplikacije.

## 📧 4. Testiranje MailHog Funkcionalnosti

### Test 1: Provera da li MailHog radi

```powershell
# Provera MailHog Web UI
Start-Process "http://localhost:8025"

# Provera SMTP porta
Test-NetConnection -ComputerName localhost -Port 1025
```

**Očekivani rezultat:** 
- Web UI se otvara na `http://localhost:8025`
- Port 1025 je otvoren

### Test 2: Testiranje slanja email-a (OTP za admin login)

1. **Otvorite frontend aplikaciju**: `http://localhost:3000`
2. **Pokušajte da se ulogujete kao admin**:
   - Email: `admin@musicstreaming.com`
   - Kliknite na "Request OTP"
3. **Proverite MailHog Web UI**: `http://localhost:8025`
4. **Trebalo bi da vidite email sa OTP kodom**

**Očekivani rezultat:** Email sa OTP kodom se pojavljuje u MailHog-u.

### Test 3: Provera logova Users Service-a

```powershell
docker logs projekat-2025-1-users-service-1 | Select-String "EMAIL"
```

**Očekivani rezultat:** 
- `[EMAIL] SMTP configured: mailhog:1025`
- `[EMAIL] Sent successfully via MailHog to admin@musicstreaming.com`

## 🔒 5. Testiranje Sigurnosnih Mehanizama

### Test 1: Password Hashing (Hash & Salt)

```powershell
# Registrujte novog korisnika preko API-ja
$body = @{
    email = "test@example.com"
    password = "TestPassword123!"
    firstName = "Test"
    lastName = "User"
} | ConvertTo-Json

$response = Invoke-WebRequest -Uri "http://localhost:8081/api/users/register" `
    -Method POST `
    -Body $body `
    -ContentType "application/json" `
    -UseBasicParsing

# Proverite MongoDB da vidite da li je password heširan
docker exec projekat-2025-1-mongodb-users-1 mongosh --eval "db.users.findOne({email: 'test@example.com'}, {password: 1})"
```

**Očekivani rezultat:** 
- Password je heširan (bcrypt hash, počinje sa `$2a$` ili `$2b$`)
- Password NIJE u plain text formatu

### Test 2: POST Metoda za Senzitivne Podatke

```powershell
# Test login request (treba da koristi POST)
$body = @{
    email = "admin@musicstreaming.com"
} | ConvertTo-Json

$response = Invoke-WebRequest -Uri "http://localhost:8081/api/users/login/request-otp" `
    -Method POST `
    -Body $body `
    -ContentType "application/json" `
    -UseBasicParsing

Write-Host "Status: $($response.StatusCode)"
```

**Očekivani rezultat:** 
- Status 200 OK
- Email je poslat (proverite MailHog)

### Test 3: HTTPS za Inter-Service Komunikaciju

```powershell
# Provera da li API Gateway koristi HTTPS za komunikaciju sa backend servisima
docker exec projekat-2025-1-api-gateway-1 cat /proc/self/environ | Select-String "USERS_SERVICE_URL|CONTENT_SERVICE_URL"
```

**Očekivani rezultat:** 
- `USERS_SERVICE_URL=https://users-service:8001`
- `CONTENT_SERVICE_URL=https://content-service:8002`

## 🧪 6. Kompletan Test Scenarijo

### Scenario: Registracija i Login Novog Korisnika

```powershell
# 1. Registracija
$registerBody = @{
    email = "newuser@test.com"
    password = "SecurePass123!"
    firstName = "New"
    lastName = "User"
} | ConvertTo-Json

$registerResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/users/register" `
    -Method POST `
    -Body $registerBody `
    -ContentType "application/json" `
    -UseBasicParsing

Write-Host "Registration Status: $($registerResponse.StatusCode)"

# 2. Request OTP
$otpBody = @{
    email = "newuser@test.com"
} | ConvertTo-Json

$otpResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/users/login/request-otp" `
    -Method POST `
    -Body $otpBody `
    -ContentType "application/json" `
    -UseBasicParsing

Write-Host "OTP Request Status: $($otpResponse.StatusCode)"

# 3. Proverite MailHog za OTP kod
Write-Host "`nProverite MailHog na http://localhost:8025 za OTP kod"

# 4. Verify OTP (zamenite OTP_CODE sa stvarnim kodom)
# $verifyBody = @{
#     email = "newuser@test.com"
#     otp = "OTP_CODE"
# } | ConvertTo-Json
# 
# $verifyResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/users/login/verify-otp" `
#     -Method POST `
#     -Body $verifyBody `
#     -ContentType "application/json" `
#     -UseBasicParsing
# 
# Write-Host "Login Status: $($verifyResponse.StatusCode)"
# Write-Host "Token: $($verifyResponse.Content)"
```

## 📊 7. Checklist za Testiranje

- [ ] Svi Docker kontejneri su pokrenuti
- [ ] API Gateway odgovara na `http://localhost:8081/api/users/health`
- [ ] MailHog Web UI je dostupan na `http://localhost:8025`
- [ ] HTTPS sertifikati postoje u `certs/` direktorijumu
- [ ] Servisi koriste HTTPS za inter-service komunikaciju (proverite logove)
- [ ] Password se čuva kao hash u MongoDB-u (ne plain text)
- [ ] Senzitivni podaci se šalju preko POST metode
- [ ] Email se šalje kada se traži OTP
- [ ] CORS headers su postavljeni pravilno
- [ ] Frontend može da komunicira sa API Gateway-em

## 🐛 Troubleshooting

### Problem: "404 page not found" na `/health`
**Rešenje:** API Gateway možda nema root endpoint. Koristite `/api/users/health` umesto toga.

### Problem: Email se ne šalje
**Rešenje:** 
1. Proverite da li je MailHog pokrenut: `docker ps | Select-String mailhog`
2. Proverite logove Users Service-a: `docker logs projekat-2025-1-users-service-1 | Select-String EMAIL`
3. Proverite environment varijable: `docker exec projekat-2025-1-users-service-1 env | Select-String SMTP`

### Problem: "unencrypted connection" greška
**Rešenje:** MailHog ne koristi TLS. Proverite da li je `SMTP_HOST=mailhog` u docker-compose.yml.

### Problem: HTTPS sertifikati nisu pronađeni
**Rešenje:** 
1. Generišite sertifikate: `.\generate-certs.ps1`
2. Proverite da li postoje: `ls certs/`

## 📝 Napomene

- **Development Mode**: API Gateway koristi HTTP za eksterni pristup (između frontenda i API Gateway-a) radi lakšeg razvoja
- **Production Mode**: Za produkciju, trebalo bi da koristite HTTPS i za eksterni pristup
- **Self-Signed Certificates**: Sertifikati su self-signed i neće biti verifikovani od strane browsera. To je normalno za development.
