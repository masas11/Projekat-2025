# 📋 Vodič za Proveru Validacije Podataka, Upravljanja Datotekama i Output Encoding-a

## ✅ Status Implementacije

**Sve funkcionalnosti su POTPUNO IMPLEMENTIRANE:**

### 1. ✅ Validacija Ulaznih Podataka (Input Validation)

#### Server-Side:
- **String provere**: Regex validacija za email, username, name
- **Whitelisting**: Samo dozvoljeni karakteri (username: slova, brojevi, underscore)
- **Boundary checking**: Provera dužine stringova (email: max 254, username: 3-20, name: max 100)
- **Character escaping**: SanitizeString uklanja null bytes i control karaktere
- **Numeric validation**: ValidateNumeric proverava da li je broj u granicama
- **Specijalni karakteri**: Provera SQL Injection i XSS pattern-a

#### Client-Side:
- **HTML5 validacija**: `required`, `type="email"`, `type="password"`
- **JavaScript validacija**: Regex provere za lozinku, email format
- **Real-time validacija**: Provera pre slanja forme

### 2. ✅ Bezbedno Upravljanje Datotekama

- **Provera privilegija**: JWT autorizacija za upload operacije
- **Provera tipa datoteke (whitelisting)**: MIME type whitelist (image/jpeg, image/png, audio/mpeg, itd.)
- **Provera ekstenzije**: Dodatna provera file extension-a
- **Provera veličine**: Max 10MB
- **Provera integriteta**: MD5 hash za verifikaciju integriteta datoteke

### 3. ✅ Enkodovanje Izlaza (Output Encoding)

- **HTML escaping**: `html.EscapeString()` za XSS zaštitu
- **URL encoding**: `url.QueryEscape()` za URL parametre
- **JSON encoding**: `json.NewEncoder()` automatski escape-uje specijalne karaktere

---

## 🔍 1. Validacija Ulaznih Podataka

### Server-Side Implementacija

#### Lokacija:
- `services/users-service/internal/validation/input.go`

#### Implementirane Funkcije:

##### 1. ValidateEmail
```go
// Regex validacija + boundary check (max 254 karaktera)
ValidateEmail(email string) error
```

##### 2. ValidateUsername
```go
// Whitelist: samo slova, brojevi, underscore
// Boundary: 3-20 karaktera
ValidateUsername(username string) error
```

##### 3. ValidateName
```go
// Whitelist: samo slova i razmaci
// Boundary: max 100 karaktera
ValidateName(name string) error
```

##### 4. SanitizeString
```go
// Uklanja null bytes i control karaktere
SanitizeString(input string) string
```

##### 5. CheckSQLInjection
```go
// Detektuje SQL injection pattern-e
CheckSQLInjection(input string) error
```

##### 6. CheckXSS
```go
// Detektuje XSS pattern-e
CheckXSS(input string) error
```

##### 7. ValidateNumeric
```go
// Proverava da li je broj u granicama
ValidateNumeric(value string, min, max int) error
```

##### 8. ValidateStringLength
```go
// Proverava dužinu stringa
ValidateStringLength(input string, min, max int) error
```

### Client-Side Implementacija

#### Lokacija:
- `frontend/src/components/Register.js`
- `frontend/src/components/ChangePassword.js`
- `frontend/src/components/ResetPassword.js`

#### Primer:
```javascript
const validatePassword = (password) => {
  if (password.length < 8) {
    return 'Lozinka mora imati najmanje 8 karaktera';
  }
  if (!/[A-Z]/.test(password)) {
    return 'Lozinka mora sadržati najmanje jedno veliko slovo';
  }
  if (!/[0-9]/.test(password)) {
    return 'Lozinka mora sadržati najmanje jedan broj';
  }
  return null;
};
```

---

## 🧪 Test 1: Validacija Ulaznih Podataka

### Test 1.1: Email Validacija

#### Test Nevažećeg Email-a:
```powershell
$body = @{
    firstName = "Test"
    lastName = "User"
    email = "invalid-email"
    username = "testuser"
    password = "Test1234"
    confirmPassword = "Test1234"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8081/api/users/register" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

**Očekivani odgovor:** HTTP 400 "invalid email format"

#### Test Predugačkog Email-a (>254 karaktera):
```powershell
$longEmail = "a" * 255 + "@example.com"
$body = @{
    email = $longEmail
    # ... ostali podaci
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8081/api/users/register" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

**Očekivani odgovor:** HTTP 400 "input length exceeds maximum allowed"

### Test 1.2: Username Validacija (Whitelisting)

#### Test sa Specijalnim Karakterima:
```powershell
$body = @{
    firstName = "Test"
    lastName = "User"
    email = "test@example.com"
    username = "test@user!"  # Nevažeći karakteri
    password = "Test1234"
    confirmPassword = "Test1234"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8081/api/users/register" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

**Očekivani odgovor:** HTTP 400 "username must be 3-20 characters and contain only letters, numbers, and underscores"

#### Test Prekratkog Username-a (<3 karaktera):
```powershell
$body = @{
    username = "ab"  # Prekratko
    # ... ostali podaci
} | ConvertTo-Json
```

**Očekivani odgovor:** HTTP 400 "username must be 3-20 characters..."

### Test 1.3: SQL Injection Provera

#### Test SQL Injection Pattern-a:
```powershell
$body = @{
    firstName = "Test' OR '1'='1"  # SQL injection pattern
    lastName = "User"
    email = "test@example.com"
    username = "testuser"
    password = "Test1234"
    confirmPassword = "Test1234"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8081/api/users/register" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

**Očekivani odgovor:** HTTP 400 "invalid input"

### Test 1.4: XSS Provera

#### Test XSS Pattern-a:
```powershell
$body = @{
    firstName = "<script>alert('XSS')</script>"  # XSS pattern
    lastName = "User"
    email = "test@example.com"
    username = "testuser"
    password = "Test1234"
    confirmPassword = "Test1234"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8081/api/users/register" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

**Očekivani odgovor:** HTTP 400 "invalid input"

### Test 1.5: Numeric Validacija

#### Test sa Nevažećim Brojem:
```powershell
# Pretpostavimo da imamo endpoint koji prima rating (1-5)
$body = @{
    rating = "10"  # Prekoračenje maksimuma
} | ConvertTo-Json

# Očekivano: HTTP 400 "value exceeds maximum"
```

### Test 1.6: Client-Side Validacija

#### Test u Browser-u:
1. Otvorite `http://localhost:3000/register`
2. Pokušajte uneti:
   - Email bez @ → HTML5 validacija blokira
   - Username sa specijalnim karakterima → JavaScript validacija
   - Prekratku lozinku → JavaScript validacija
3. **Očekivano**: Poruke o greškama prije slanja forme

---

## 📁 2. Bezbedno Upravljanje Datotekama

### Implementacija

#### Lokacija:
- `services/users-service/internal/validation/file.go`

#### Implementirane Funkcije:

##### 1. ValidateFileType (Whitelisting)
```go
// Proverava MIME type prema whitelist-u
AllowedFileTypes = map[string]bool{
    "image/jpeg":      true,
    "image/png":       true,
    "image/gif":       true,
    "application/pdf": true,
    "text/plain":      true,
    "audio/mpeg":      true,
    "audio/wav":       true,
}
```

##### 2. ValidateFileSize
```go
// Max 10MB
const MaxFileSize = 10 * 1024 * 1024
```

##### 3. ValidateFileExtension
```go
// Dodatna provera ekstenzije
ValidateFileExtension(filename string, allowedExtensions []string) error
```

##### 4. CalculateFileHash
```go
// MD5 hash za proveru integriteta
CalculateFileHash(reader io.Reader) (string, error)
```

##### 5. VerifyFileIntegrity
```go
// Proverava integritet datoteke
VerifyFileIntegrity(expectedHash string, fileReader io.Reader) error
```

### Provera Privilegija

#### Implementacija:
- JWT autorizacija za upload operacije
- `middleware.RequireAuth(cfg)` ili `middleware.RequireRole("ADMIN", cfg)`

---

## 🧪 Test 2: Bezbedno Upravljanje Datotekama

### Test 2.1: Provera Tipa Datoteke (Whitelisting)

#### Test sa Nevažećim MIME Type-om:
```powershell
# Simulacija file upload-a sa nevažećim tipom
$headers = @{
    "Content-Type" = "application/x-executable"  # Nevažeći tip
    "Authorization" = "Bearer $token"
}

# Očekivano: HTTP 400 "file type not allowed"
```

### Test 2.2: Provera Veličine Datoteke

#### Test sa Prevelikom Datotekom (>10MB):
```powershell
# Simulacija upload-a prevelike datoteke
# Očekivano: HTTP 400 "file size exceeds maximum allowed"
```

### Test 2.3: Provera Ekstenzije

#### Test sa Nevažećom Ekstenzijom:
```powershell
# Simulacija upload-a sa .exe ekstenzijom
# Očekivano: HTTP 400 "file type not allowed"
```

### Test 2.4: Provera Integriteta

#### Test Manipulacije Datoteke:
```go
// 1. Upload datoteke → dobijete hash
// 2. Promenite datoteku
// 3. Pokušajte verifikaciju
// Očekivano: ErrIntegrityCheckFailed
```

### Test 2.5: Provera Privilegija

#### Test Upload-a bez Autentifikacije:
```powershell
# Pokušaj upload-a bez tokena
# Očekivano: HTTP 401 "authorization header required"
```

#### Test Upload-a kao Regular User (ako je potrebno ADMIN):
```powershell
# Upload sa regular user token-om
# Očekivano: HTTP 403 "forbidden: ADMIN access required"
```

---

## 🔒 3. Enkodovanje Izlaza (Output Encoding)

### Implementacija

#### Lokacija:
- `services/users-service/internal/security/encoding.go`

#### Implementirane Funkcije:

##### 1. EscapeHTML
```go
// Escape-uje HTML specijalne karaktere
EscapeHTML(input string) string
// Primer: <script> → &lt;script&gt;
```

##### 2. EscapeURL
```go
// Escape-uje URL specijalne karaktere
EscapeURL(input string) string
// Primer: space → %20
```

### JSON Encoding

#### Automatsko Escape-ovanje:
- `json.NewEncoder(w).Encode(data)` automatski escape-uje specijalne karaktere
- JSON standard obezbeđuje sigurno enkodovanje

### Korišćenje u Handler-ima

#### Primer:
```go
// U register handler-u
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(response)  // Automatski escape-uje
```

---

## 🧪 Test 3: Output Encoding

### Test 3.1: HTML Escape-ovanje

#### Test sa HTML Karakterima u Odgovoru:
```powershell
# Pretpostavimo da korisnik unese "<script>alert('XSS')</script>" kao ime
# Kada se prikaže u frontendu, trebalo bi biti escape-ovano
# Očekivano: &lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;
```

### Test 3.2: URL Encoding

#### Test sa URL Karakterima:
```go
// Token se URL encode-uje pre slanja u email linku
encodedToken := url.QueryEscape(verificationToken)
// Očekivano: Specijalni karakteri su escape-ovani
```

### Test 3.3: JSON Encoding

#### Test sa Specijalnim Karakterima u JSON Odgovoru:
```powershell
# Registrujte korisnika sa imenom koje sadrži specijalne karaktere
# Proverite JSON odgovor
# Očekivano: Svi specijalni karakteri su escape-ovani
```

---

## 📁 Relevantni Fajlovi

### Validacija:
- `services/users-service/internal/validation/input.go` - Input validacija
- `services/users-service/internal/validation/file.go` - File validacija
- `services/users-service/internal/validation/password.go` - Password validacija
- `frontend/src/components/Register.js` - Client-side validacija
- `frontend/src/components/ChangePassword.js` - Client-side validacija

### Output Encoding:
- `services/users-service/internal/security/encoding.go` - HTML/URL encoding
- `services/users-service/internal/handler/register.go` - Korišćenje encoding-a

### File Management:
- `services/users-service/internal/validation/file.go` - File validacija
- `services/content-service/cmd/main.go` - File serving sa whitelisting-om

---

## ✅ Checklist za Proveru

### Input Validacija:
- [ ] Email validacija (regex + boundary check)
- [ ] Username validacija (whitelisting + boundary check)
- [ ] Name validacija (whitelisting + boundary check)
- [ ] String sanitization (null bytes, control characters)
- [ ] SQL Injection provera
- [ ] XSS provera
- [ ] Numeric validacija (boundary check)
- [ ] String length validacija
- [ ] Client-side validacija (HTML5 + JavaScript)
- [ ] Server-side validacija (sve input-e)

### File Management:
- [ ] MIME type whitelisting
- [ ] File extension provera
- [ ] File size provera (max 10MB)
- [ ] File integrity provera (MD5 hash)
- [ ] Provera privilegija (JWT autorizacija)
- [ ] Admin-only upload (ako je potrebno)

### Output Encoding:
- [ ] HTML escaping za XSS zaštitu
- [ ] URL encoding za URL parametre
- [ ] JSON encoding (automatski escape)
- [ ] Content-Type header postavljen

---

## 🐛 Troubleshooting

### Problem: Validacija ne radi
- Proverite da li se poziva `validation.Validate*()` funkcija
- Proverite da li se `SanitizeString()` poziva pre validacije
- Proverite logove za greške

### Problem: File upload ne radi
- Proverite MIME type datoteke
- Proverite veličinu datoteke (max 10MB)
- Proverite da li je korisnik autentifikovan
- Proverite da li korisnik ima potrebne privilegije

### Problem: XSS napad prolazi
- Proverite da li se koristi `html.EscapeString()` pri prikazu podataka
- Proverite da li se JSON encoding koristi za API odgovore
- Proverite da li React automatski escape-uje (default ponašanje)

---

## 📝 Napomene

- **Whitelisting**: Koristi se umesto blacklisting-a (sigurnije)
- **Boundary checking**: Svi input-i imaju min/max granice
- **Character escaping**: Sanitizacija pre validacije
- **File integrity**: MD5 hash za proveru integriteta
- **Output encoding**: Automatski u JSON, ručno za HTML/URL
- **Client + Server**: Validacija na obe strane (defense in depth)

---

## 🎯 Demonstracija za Odbranu

### Scenario 1: SQL Injection Napad
1. Pokušaj registracije sa: `firstName = "Test' OR '1'='1"`
2. **Očekivano**: HTTP 400 "invalid input"
3. Proverite da li se SQL injection pattern detektuje

### Scenario 2: XSS Napad
1. Pokušaj registracije sa: `firstName = "<script>alert('XSS')</script>"`
2. **Očekivano**: HTTP 400 "invalid input"
3. Proverite da li se XSS pattern detektuje

### Scenario 3: File Upload Validacija
1. Pokušaj upload-a sa nevažećim tipom (npr. .exe)
2. **Očekivano**: HTTP 400 "file type not allowed"
3. Proverite MIME type whitelisting

### Scenario 4: Output Encoding
1. Unesite HTML karaktere u formu
2. Proverite da li su escape-ovani u odgovoru
3. Proverite da li React prikazuje escape-ovane karaktere

---

## 🔐 Sigurnosne Best Practices

1. **Whitelisting > Blacklisting**: Dozvolite samo poznate dobre vrednosti
2. **Defense in Depth**: Validacija na client i server strani
3. **Sanitize → Validate**: Prvo sanitizujte, zatim validirajte
4. **Output Encoding**: Uvek encode-ujte output pre prikaza
5. **File Validation**: Proverite tip, veličinu i integritet
6. **Privilege Check**: Proverite privilegije pre file operacija
