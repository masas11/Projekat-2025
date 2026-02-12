# 📋 Vodič za Proveru Autorizacije, Šifrovanja i DoS Zaštite

## ✅ Status Implementacije

**Sve funkcionalnosti su POTPUNO IMPLEMENTIRANE:**

### 1. ✅ Autorizacija za Svaki Zahtev
- **Server-side**: JWT middleware na API Gateway i svim servisima
- **Client-side**: ProtectedRoute komponenta za zaštitu ruta
- **Provera tokena**: Za svaki zahtev koji zahteva autentifikaciju
- **Role-based access**: RequireRole middleware za ADMIN pristup

### 2. ✅ Šifrovanje i Provera Integriteta State Podataka
- **Šifrovanje**: Korisnički podaci se šifruju pre čuvanja u localStorage
- **Provera integriteta**: Checksum za detekciju manipulacije podacima
- **Automatska provera**: Pri učitavanju podataka se proverava integritet

### 3. ✅ Zaštita od DoS - Rate Limiting
- **API Gateway**: 100 zahteva po minuti po IP adresi
- **Users Service**: 10 zahteva po minuti za osetljive endpoint-e
- **Per-IP ograničenje**: Ograničava broj transakcija po korisniku/IP
- **HTTP 429**: Vraća "too many requests" kada je limit prekoračen

---

## 🔐 1. Autorizacija

### Server-Side Autorizacija

#### Implementacija:
- **API Gateway**: `services/api-gateway/internal/middleware/auth.go`
- **Content Service**: `services/content-service/internal/middleware/jwt.go`
- **Users Service**: JWT validacija u svim handlerima

#### Middleware Funkcije:

1. **RequireAuth** - Zahteva validan JWT token
   ```go
   middleware.RequireAuth(cfg)(handler)
   ```

2. **RequireRole** - Zahteva određenu ulogu (npr. ADMIN)
   ```go
   middleware.RequireRole("ADMIN", cfg)(handler)
   ```

3. **OptionalAuth** - Opciona autentifikacija (dodaje user info ako postoji)
   ```go
   middleware.OptionalAuth(cfg)(handler)
   ```

#### Zaštićeni Endpoint-i:

**Zahtevaju autentifikaciju:**
- `/api/users/logout` - RequireAuth
- `/api/users/password/change` - RequireAuth
- `/api/notifications` - RequireAuth
- `/api/subscriptions` - RequireAuth
- `/api/subscriptions/subscribe-artist` - RequireAuth
- `/api/subscriptions/subscribe-genre` - RequireAuth
- `/api/ratings/rate-song` - RequireAuth (non-admin)
- `/api/ratings/delete-rating` - RequireAuth (non-admin)

**Zahtevaju ADMIN ulogu:**
- `POST /api/content/artists` - RequireRole("ADMIN")
- `PUT /api/content/artists/{id}` - RequireRole("ADMIN")
- `DELETE /api/content/artists/{id}` - RequireRole("ADMIN")
- `POST /api/content/albums` - RequireRole("ADMIN")
- `PUT /api/content/albums/{id}` - RequireRole("ADMIN")
- `DELETE /api/content/albums/{id}` - RequireRole("ADMIN")
- `POST /api/content/songs` - RequireRole("ADMIN")
- `PUT /api/content/songs/{id}` - RequireRole("ADMIN")
- `DELETE /api/content/songs/{id}` - RequireRole("ADMIN")

### Client-Side Autorizacija

#### Implementacija:
- **ProtectedRoute**: `frontend/src/components/ProtectedRoute.js`
- **AuthContext**: `frontend/src/context/AuthContext.js`
- **API Service**: Automatski dodaje Authorization header

#### Kako Radi:

1. **ProtectedRoute** proverava da li je korisnik prijavljen
2. Ako nije prijavljen → preusmeravanje na `/login`
3. **API Service** automatski dodaje token u header:
   ```javascript
   const token = localStorage.getItem('token');
   if (token) {
     config.headers.Authorization = `Bearer ${token}`;
   }
   ```

---

## 🔒 2. Šifrovanje i Provera Integriteta

### Implementacija:
- **Encryption Utility**: `frontend/src/utils/encryption.js`
- **Korišćenje**: `frontend/src/context/AuthContext.js`

### Kako Radi:

#### Šifrovanje:
```javascript
setEncryptedItem('user', userData);
```

**Proces:**
1. JSON stringify korisničkih podataka
2. Base64 encoding sa XOR obfuscation
3. Čuvanje šifrovanih podataka u localStorage
4. Generisanje checksum-a za proveru integriteta
5. Čuvanje checksum-a (`key + '_checksum'`)

#### Dešifrovanje sa Proverom Integriteta:
```javascript
const userData = getEncryptedItem('user');
```

**Proces:**
1. Učitavanje šifrovanih podataka iz localStorage
2. Dešifrovanje podataka
3. Parsiranje JSON-a
4. **Provera integriteta**: Upoređivanje checksum-a
5. Ako checksum ne odgovara → podaci su manipulirani → brisanje i vraćanje null

### Zaštićeni Podaci:
- **Korisnički podaci** (`user`) - šifrovani
- **Token** - čuva se kao plain text (JWT je već encoded)

---

## 🛡️ 3. Zaštita od DoS - Rate Limiting

### Implementacija:
- **API Gateway**: `services/api-gateway/internal/middleware/rate_limit.go`
- **Users Service**: `services/users-service/internal/middleware/rate_limit.go`

### Rate Limiting Strategija:

#### API Gateway (Globalni):
- **Limit**: 100 zahteva po minuti po IP adresi
- **Primenjeno na**: Sve endpoint-e
- **HTTP Status**: 429 (Too Many Requests)

#### Users Service (Osetljivi Endpoint-i):
- **Limit**: 10 zahteva po minuti po IP adresi
- **Primenjeno na**:
  - `/register`
  - `/login/request-otp`
  - `/login/verify-otp`
  - `/logout`
  - `/password/change`
  - `/password/reset/request`
  - `/password/reset`
  - `/verify-email`
  - `/recover/request`
  - `/recover/verify`

### Kako Radi:

1. **Identifikacija klijenta**: Po IP adresi (`RemoteAddr` ili `X-Forwarded-For`)
2. **Praćenje zahteva**: Čuvanje timestamp-a zahteva u mapi
3. **Provera limita**: Brojanje zahteva u vremenskom prozoru
4. **Ograničenje**: Ako limit prekoračen → HTTP 429
5. **Čišćenje**: Automatsko brisanje starih zapisa (svakih 1 minut)

---

## 🧪 Kako Proveriti

### Test 1: Autorizacija - Zaštićeni Endpoint

#### Test bez Tokena:
```powershell
# Pokušaj pristupa zaštićenom endpoint-u bez tokena
Invoke-RestMethod -Uri "http://localhost:8081/api/users/logout" `
    -Method POST `
    -ContentType "application/json"
```

**Očekivani odgovor:** HTTP 401 "authorization header required"

#### Test sa Nevažećim Tokenom:
```powershell
Invoke-RestMethod -Uri "http://localhost:8081/api/users/logout" `
    -Method POST `
    -Headers @{Authorization = "Bearer invalid-token"} `
    -ContentType "application/json"
```

**Očekivani odgovor:** HTTP 401 "invalid or expired token"

#### Test sa Validnim Tokenom:
```powershell
# Prvo se prijavite da dobijete token
$loginBody = @{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json

$otpResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/users/login/request-otp" `
    -Method POST `
    -ContentType "application/json" `
    -Body $loginBody

# Unesite OTP iz logova
$verifyBody = @{
    username = "admin"
    otp = "123456"  # Zamenite sa stvarnim OTP kodom
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8081/api/users/login/verify-otp" `
    -Method POST `
    -ContentType "application/json" `
    -Body $verifyBody

$token = $response.token

# Sada testirajte zaštićeni endpoint
Invoke-RestMethod -Uri "http://localhost:8081/api/users/logout" `
    -Method POST `
    -Headers @{Authorization = "Bearer $token"} `
    -ContentType "application/json"
```

**Očekivani odgovor:** HTTP 200 "logged out successfully"

---

### Test 2: Role-Based Access Control (ADMIN)

#### Test kao Regular User:
```powershell
# Prijavite se kao regular user (ne admin)
# Pokušajte kreirati artist-a
$body = @{
    name = "Test Artist"
    bio = "Test bio"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8081/api/content/artists" `
    -Method POST `
    -Headers @{Authorization = "Bearer $regularUserToken"} `
    -ContentType "application/json" `
    -Body $body
```

**Očekivani odgovor:** HTTP 403 "forbidden: ADMIN access required"

#### Test kao Admin:
```powershell
# Prijavite se kao admin
# Kreirajte artist-a
Invoke-RestMethod -Uri "http://localhost:8081/api/content/artists" `
    -Method POST `
    -Headers @{Authorization = "Bearer $adminToken"} `
    -ContentType "application/json" `
    -Body $body
```

**Očekivani odgovor:** HTTP 200/201 sa kreiranim artist-om

---

### Test 3: Šifrovanje i Integritet Podataka

#### Provera u Browser-u:

1. **Prijavite se** na aplikaciju
2. **Otvorite Developer Tools** (F12)
3. **Idite na Application → Local Storage**
4. **Proverite ključeve**:
   - `token` - plain text (JWT)
   - `user` - šifrovani podaci (base64 encoded)
   - `user_checksum` - checksum za proveru integriteta

#### Test Manipulacije Podataka:

1. **U localStorage**, promenite vrednost `user` ključa
2. **Osvežite stranicu**
3. **Očekivano**: 
   - Checksum ne odgovara
   - Podaci se brišu
   - Korisnik se odjavljuje
   - Preusmeravanje na login

#### Provera u Kodu:
```javascript
// U browser konzoli
const user = localStorage.getItem('user');
console.log('Encrypted user data:', user);

const checksum = localStorage.getItem('user_checksum');
console.log('Checksum:', checksum);

// Pokušajte promeniti user podatke
localStorage.setItem('user', 'tampered-data');

// Osvežite stranicu - podaci će biti obrisani
```

---

### Test 4: Rate Limiting (DoS Zaštita)

#### Test Prekoračenja Limita:

```powershell
# Napravite više od 100 zahteva u 1 minuti
for ($i = 1; $i -le 110; $i++) {
    try {
        Invoke-RestMethod -Uri "http://localhost:8081/api/users/health" `
            -Method GET `
            -ErrorAction Stop
        Write-Host "Request $i: OK"
    } catch {
        if ($_.Exception.Response.StatusCode.value__ -eq 429) {
            Write-Host "Request $i: Rate limit exceeded (429)"
            break
        }
    }
    Start-Sleep -Milliseconds 100
}
```

**Očekivano**: 
- Prvih ~100 zahteva: HTTP 200
- Nakon limita: HTTP 429 "too many requests"

#### Test Osetljivih Endpoint-a (10 req/min):

```powershell
# Napravite više od 10 zahteva za registraciju u 1 minuti
for ($i = 1; $i -le 15; $i++) {
    $body = @{
        firstName = "Test$i"
        lastName = "User$i"
        email = "test$i@example.com"
        username = "testuser$i"
        password = "Test1234"
        confirmPassword = "Test1234"
    } | ConvertTo-Json
    
    try {
        Invoke-RestMethod -Uri "http://localhost:8081/api/users/register" `
            -Method POST `
            -ContentType "application/json" `
            -Body $body `
            -ErrorAction Stop
        Write-Host "Request $i: OK"
    } catch {
        if ($_.Exception.Response.StatusCode.value__ -eq 429) {
            Write-Host "Request $i: Rate limit exceeded (429)"
            break
        }
    }
}
```

**Očekivano**: 
- Prvih ~10 zahteva: HTTP 200/201 ili 400 (validation errors)
- Nakon limita: HTTP 429 "too many requests"

---

## 📁 Relevantni Fajlovi

### Autorizacija:
- `services/api-gateway/internal/middleware/auth.go` - JWT middleware
- `services/content-service/internal/middleware/jwt.go` - JWT middleware
- `services/users-service/internal/security/jwt.go` - JWT generisanje/validacija
- `frontend/src/components/ProtectedRoute.js` - Client-side zaštita ruta
- `frontend/src/services/api.js` - Automatsko dodavanje Authorization header-a

### Šifrovanje:
- `frontend/src/utils/encryption.js` - Encryption/decryption funkcije
- `frontend/src/context/AuthContext.js` - Korišćenje šifrovanja

### Rate Limiting:
- `services/api-gateway/internal/middleware/rate_limit.go` - Rate limiter
- `services/users-service/internal/middleware/rate_limit.go` - Rate limiter
- `services/api-gateway/cmd/main.go` - Primena rate limiting-a
- `services/users-service/cmd/main.go` - Primena rate limiting-a

---

## ✅ Checklist za Proveru

### Autorizacija:
- [ ] JWT middleware proverava token za svaki zahtev
- [ ] RequireAuth middleware blokira neautentifikovane zahteve
- [ ] RequireRole middleware proverava ulogu korisnika
- [ ] ProtectedRoute komponenta zaštitiće rute na frontendu
- [ ] API Service automatski dodaje Authorization header
- [ ] Greška 401 za nevažeće token-e
- [ ] Greška 403 za nedovoljne privilegije

### Šifrovanje:
- [ ] Korisnički podaci se šifruju pre čuvanja
- [ ] Checksum se generiše za proveru integriteta
- [ ] Provera integriteta pri učitavanju podataka
- [ ] Manipulisani podaci se brišu automatski
- [ ] Token se čuva kao plain text (JWT je već encoded)

### Rate Limiting:
- [ ] API Gateway ima globalni rate limit (100 req/min)
- [ ] Users Service ima rate limit za osetljive endpoint-e (10 req/min)
- [ ] Rate limiting radi po IP adresi
- [ ] HTTP 429 vraća se kada je limit prekoračen
- [ ] Stari zapisi se automatski brišu

---

## 🐛 Troubleshooting

### Problem: "authorization header required"
- Proverite da li se token šalje u Authorization header-u
- Format: `Authorization: Bearer <token>`
- Proverite da li je token validan i nije istekao

### Problem: "forbidden: ADMIN access required"
- Korisnik nema ADMIN ulogu
- Proverite `user.role` u tokenu
- Prijavite se kao admin korisnik

### Problem: Podaci se ne šifruju
- Proverite da li se koristi `setEncryptedItem()` umesto `localStorage.setItem()`
- Proverite da li je `encryption.js` importovan

### Problem: Rate limit se ne primenjuje
- Proverite da li je rate limiting middleware primenjen na endpoint
- Proverite logove za rate limit poruke
- Proverite IP adresu klijenta

---

## 📝 Napomene

- **JWT Token**: Ističe nakon 24 sata
- **Rate Limit Window**: 1 minut (rolling window)
- **Encryption**: Client-side obfuscation (za production koristiti Web Crypto API)
- **Checksum**: Jednostavna provera integriteta (za production koristiti HMAC)
- **IP Detection**: Koristi `RemoteAddr` ili `X-Forwarded-For` header

---

## 🎯 Demonstracija za Odbranu

### Scenario 1: Neautorizovan Pristup
1. Pokušaj pristupa `/api/users/logout` bez tokena
2. **Očekivano**: HTTP 401
3. Dodajte token i pokušajte ponovo
4. **Očekivano**: HTTP 200

### Scenario 2: Nedovoljne Privilegije
1. Prijavite se kao regular user
2. Pokušajte kreirati artist-a (`POST /api/content/artists`)
3. **Očekivano**: HTTP 403
4. Prijavite se kao admin
5. Pokušajte ponovo
6. **Očekivano**: HTTP 200/201

### Scenario 3: Manipulacija Podataka
1. Prijavite se
2. Otvorite Developer Tools → Local Storage
3. Promenite `user` vrednost
4. Osvežite stranicu
5. **Očekivano**: Podaci se brišu, korisnik se odjavljuje

### Scenario 4: DoS Zaštita
1. Napravite 110 zahteva za `/api/users/health` u 1 minuti
2. **Očekivano**: Prvih ~100 OK, ostali HTTP 429
3. Sačekajte 1 minut
4. Pokušajte ponovo
5. **Očekivano**: Zahtevi ponovo prolaze
