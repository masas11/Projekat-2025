# Izveštaj o Analizi Ranjivosti - 2.21

**Datum:** 12. februar 2025  
**Aplikacija:** Music Streaming Platform  
**Verzija:** 1.0

---

## 📋 Sadržaj

1. [Uvod](#uvod)
2. [Korišćeni Alati za Identifikaciju Ranjivosti](#korišćeni-alati)
3. [Identifikovane Ranjivosti](#identifikovane-ranjivosti)
4. [Analiza Ranjivosti](#analiza-ranjivosti)
5. [Preporuke za Prevazilaženje](#preporuke-za-prevazilaženje)
6. [Zaštita od Eksploatacije](#zaštita-od-eksploatacije)
7. [Zaključak](#zaključak)

---

## 1. Uvod

Ovaj izveštaj predstavlja analizu bezbednosti Music Streaming Platform aplikacije. Analiza je sprovedena kroz pregled koda, identifikaciju potencijalnih ranjivosti i preporuke za njihovo prevazilaženje.

### Metodologija

- **Static Code Analysis (SAST)** - Pregled izvornog koda
- **Manual Code Review** - Ručna provera kritičnih delova
- **Security Best Practices Review** - Provera prema OWASP Top 10
- **Dependency Analysis** - Provera korišćenih biblioteka

---

## 2. Korišćeni Alati za Identifikaciju Ranjivosti

### 2.1. Static Application Security Testing (SAST)

#### Preporučeni alati:
1. **Gosec** - Go Security Checker
   - Detektuje SQL injection, XSS, hardcoded secrets
   - Komanda: `gosec ./...`

2. **GolangCI-Lint** - Linter sa security plugin-ima
   - Detektuje common security issues
   - Komanda: `golangci-lint run`

3. **Semgrep** - Pattern-based security scanner
   - Detektuje security anti-patterns
   - Komanda: `semgrep --config=auto`

4. **SonarQube** - Comprehensive code analysis
   - Detektuje security vulnerabilities, code smells

### 2.2. Dependency Scanning

#### Preporučeni alati:
1. **Snyk** - Dependency vulnerability scanner
   - Komanda: `snyk test`
   - Detektuje poznate ranjivosti u dependencies

2. **OWASP Dependency-Check**
   - Komanda: `dependency-check --scan .`
   - Detektuje CVEs u dependencies

3. **Go Modules Security Checker**
   - `govulncheck` - Go vulnerability database

### 2.3. Manual Code Review

- Pregled authentication/authorization logike
- Pregled input validation implementacije
- Pregled error handling-a
- Pregled logging implementacije

### 2.4. Penetration Testing Tools

#### Preporučeni alati:
1. **OWASP ZAP** - Web application security scanner
2. **Burp Suite** - Web vulnerability scanner
3. **Nmap** - Network scanning
4. **SQLMap** - SQL injection testing

---

## 3. Identifikovane Ranjivosti

### 3.1. Kategorizacija Ranjivosti

Ranjivosti su kategorisane prema:
- **Severity** (Kritičnost): Critical, High, Medium, Low
- **CVSS Score** (ako je primenljivo)
- **OWASP Top 10** kategorija

---

### 🔴 CRITICAL (Kritične)

#### VULN-001: Hardcoded Admin Credentials
**Severity:** CRITICAL  
**CVSS Score:** 9.8 (Critical)  
**OWASP Category:** A07:2021 – Identification and Authentication Failures

**Lokacija:**
- `services/users-service/cmd/main.go:31`

**Opis:**
```go
hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
```

**Eksploatacija:**
- Admin korisnik se automatski kreira sa poznatom lozinkom `admin123`
- Napadač može lako pristupiti admin nalogu
- Potpuna kontrola nad sistemom

**Impact:**
- Kompromitacija celog sistema
- Neovlašćene administratorske aktivnosti
- Brisanje/promena podataka

---

#### VULN-002: InsecureSkipVerify za TLS
**Severity:** CRITICAL  
**CVSS Score:** 7.5 (High)  
**OWASP Category:** A02:2021 – Cryptographic Failures

**Lokacija:**
- `services/api-gateway/cmd/main.go:79`
- `services/content-service/internal/events/emitter.go:68`

**Opis:**
```go
TLSClientConfig: &tls.Config{InsecureSkipVerify: true}
```

**Eksploatacija:**
- Man-in-the-Middle (MITM) napadi
- Interceptovanje komunikacije između servisa
- Čitanje/modifikacija podataka u tranzitu

**Impact:**
- Kompromitacija inter-service komunikacije
- Krađa osetljivih podataka
- Modifikacija podataka u tranzitu

**Napomena:** Ovo je prihvatljivo za development, ali **NE** za produkciju!

---

### 🟠 HIGH (Visoke)

#### VULN-003: Pattern-Based SQL Injection Detection
**Severity:** HIGH  
**CVSS Score:** 8.6 (High)  
**OWASP Category:** A03:2021 – Injection

**Lokacija:**
- `services/users-service/internal/validation/input.go:98-119`

**Opis:**
```go
func CheckSQLInjection(input string) error {
    sqlPatterns := []string{
        "' OR '1'='1",
        "' OR '1'='1'--",
        // ...
    }
    // Pattern matching
}
```

**Eksploatacija:**
- Pattern matching može biti obeštećen sa:
  - Encoding: `%27 OR %271%27=%271`
  - Case variations: `' Or '1'='1`
  - Whitespace variations: `'OR'1'='1`
  - Comments: `'/**/OR/**/'1'='1`
- Napadač može zaobići detekciju

**Impact:**
- SQL injection napadi
- Neovlašćen pristup bazi podataka
- Krađa/manipulacija podataka

---

#### VULN-004: Pattern-Based XSS Detection
**Severity:** HIGH  
**CVSS Score:** 7.2 (High)  
**OWASP Category:** A03:2021 – Injection

**Lokacija:**
- `services/users-service/internal/validation/input.go:121-144`

**Opis:**
```go
func CheckXSS(input string) error {
    xssPatterns := []string{
        "<script",
        "</script>",
        // ...
    }
}
```

**Eksploatacija:**
- Pattern matching može biti obeštećen sa:
  - Encoding: `&lt;script&gt;`, `%3Cscript%3E`
  - Case variations: `<ScRiPt>`
  - Event handlers: `onmouseover=`, `onfocus=`
  - SVG/HTML5 vectors: `<svg onload=alert(1)>`

**Impact:**
- XSS napadi
- Krađa session tokena
- Phishing napadi

---

#### VULN-005: Rate Limiting po IP Adresi
**Severity:** HIGH  
**CVSS Score:** 6.5 (Medium-High)  
**OWASP Category:** A04:2021 – Insecure Design

**Lokacija:**
- `services/api-gateway/internal/middleware/rate_limit.go:94-98`

**Opis:**
```go
ip := r.RemoteAddr
if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
    ip = forwarded
}
```

**Eksploatacija:**
- IP adresa može biti lažirana (`X-Forwarded-For` header)
- Napadač može koristiti proxy/VPN za promenu IP adrese
- Distributed DoS (DDoS) napadi sa više IP adresa

**Impact:**
- DoS napadi
- Brute force napadi
- Preopterećenje servera

---

#### VULN-006: MD5 za File Integrity
**Severity:** HIGH  
**CVSS Score:** 5.3 (Medium)  
**OWASP Category:** A02:2021 – Cryptographic Failures

**Lokacija:**
- `services/users-service/internal/validation/file.go:86-92`

**Opis:**
```go
func CalculateFileHash(reader io.Reader) (string, error) {
    hash := md5.New()
    // ...
}
```

**Eksploatacija:**
- MD5 je kriptografski slaba hash funkcija
- Collision attacks su mogući
- Napadač može kreirati fajl sa istim MD5 hash-om

**Impact:**
- File integrity compromise
- Malicious file upload
- File tampering

---

### 🟡 MEDIUM (Srednje)

#### VULN-007: Prazna sanitizeMessage Funkcija
**Severity:** MEDIUM  
**CVSS Score:** 4.3 (Medium)  
**OWASP Category:** A03:2021 – Injection

**Lokacija:**
- `services/shared/logger/logger.go:213-222`

**Opis:**
```go
func sanitizeMessage(message string) string {
    // Remove passwords (common patterns)
    // This is a simple implementation - in production, use more sophisticated sanitization
    return message  // Ne radi ništa!
}
```

**Eksploatacija:**
- Osetljivi podaci se loguju u plain text-u
- Stack traces se loguju (može otkriti internu strukturu)
- Passwords, tokens, secrets se mogu pojaviti u logovima

**Impact:**
- Information disclosure kroz logove
- Kompromitacija credentials-a
- Debugging informacije u produkciji

---

#### VULN-008: JWT Secret Hardcoding
**Severity:** MEDIUM  
**CVSS Score:** 5.9 (Medium)  
**OWASP Category:** A02:2021 – Cryptographic Failures

**Lokacija:**
- `services/api-gateway/internal/middleware/auth.go:75`
- `services/users-service/internal/security/jwt.go`

**Opis:**
```go
jwtSecret := cfg.JWTSecret
if jwtSecret == "" {
    jwtSecret = "your-secret-key-change-in-production" // Default!
}
```

**Eksploatacija:**
- Ako JWT secret nije postavljen, koristi se default vrednost
- Napadač može generisati validne tokene
- Potpuna kompromitacija autentifikacije

**Impact:**
- Token forgery
- Neovlašćen pristup
- Session hijacking

---

#### VULN-009: SQL Injection Pattern Matching - Nedovoljno
**Severity:** MEDIUM  
**CVSS Score:** 6.1 (Medium)  
**OWASP Category:** A03:2021 – Injection

**Opis:**
- Pattern matching pokriva samo osnovne SQL injection pattern-e
- Ne pokriva:
  - Time-based SQL injection
  - Boolean-based blind SQL injection
  - Union-based SQL injection sa encoding-om
  - NoSQL injection (MongoDB)

**Eksploatacija:**
- Napadač može koristiti naprednije SQL injection tehnike
- Encoding/obfuscation za zaobilaženje detekcije

**Impact:**
- SQL injection napadi
- Neovlašćen pristup bazi

---

#### VULN-010: XSS Pattern Matching - Nedovoljno
**Severity:** MEDIUM  
**CVSS Score:** 5.8 (Medium)  
**OWASP Category:** A03:2021 – Injection

**Opis:**
- Pattern matching pokriva samo osnovne XSS pattern-e
- Ne pokriva:
  - DOM-based XSS
  - Reflected XSS sa encoding-om
  - Stored XSS sa obfuscation-om
  - HTML5/SVG vectors

**Eksploatacija:**
- Napadač može koristiti naprednije XSS vektore
- Encoding/obfuscation za zaobilaženje detekcije

**Impact:**
- XSS napadi
- Session hijacking
- Phishing

---

### 🟢 LOW (Niske)

#### VULN-011: Error Messages - Information Disclosure
**Severity:** LOW  
**CVSS Score:** 3.1 (Low)  
**OWASP Category:** A04:2021 – Insecure Design

**Lokacija:**
- Razni handler-i

**Opis:**
```go
http.Error(w, "failed to create user: "+err.Error(), http.StatusInternalServerError)
```

**Eksploatacija:**
- Error poruke mogu otkriti internu strukturu
- Stack traces u error porukama
- Database error messages

**Impact:**
- Information disclosure
- Reconnaissance za napadača

---

#### VULN-012: CORS Configuration
**Severity:** LOW  
**CVSS Score:** 2.9 (Low)  
**OWASP Category:** A05:2021 – Security Misconfiguration

**Lokacija:**
- `services/api-gateway/cmd/main.go:18-27`

**Opis:**
```go
origin := r.Header.Get("Origin")
if origin == "" {
    origin = "*"  // Allow all origins if no Origin header
}
w.Header().Set("Access-Control-Allow-Origin", origin)
```

**Eksploatacija:**
- Ako nema Origin header-a, dozvoljava sve origin-e
- Može dozvoliti neovlašćene origin-e

**Impact:**
- Cross-origin attacks
- CSRF napadi

---

## 4. Analiza Ranjivosti

### 4.1. Kritične Ranjivosti - Detaljna Analiza

#### VULN-001: Hardcoded Admin Credentials

**Kako se može eksploatisati:**

1. **Direktan pristup:**
   ```
   POST /api/users/login/request-otp
   {
     "username": "admin",
     "password": "admin123"
   }
   ```

2. **Nakon prijave, napadač dobija admin token:**
   ```
   POST /api/users/login/verify-otp
   {
     "username": "admin",
     "otp": "<otp_code>"
   }
   ```

3. **Korišćenje admin tokena za:**
   - Brisanje korisnika
   - Kreiranje/ažuriranje/brisanje content-a
   - Pristup svim podacima

**Rizik:**
- **VERY HIGH** - Potpuna kompromitacija sistema

---

#### VULN-002: InsecureSkipVerify

**Kako se može eksploatisati:**

1. **Man-in-the-Middle napad:**
   ```
   Napadač postavlja proxy između API Gateway-a i backend servisa
   - Interceptuje HTTPS komunikaciju
   - Čita/modifikuje podatke
   - Može inject-ovati malicious kod
   ```

2. **Certificate spoofing:**
   - Napadač generiše lažni sertifikat
   - API Gateway prihvata lažni sertifikat
   - Komunikacija se preusmerava na napadačev server

**Rizik:**
- **HIGH** - Kompromitacija inter-service komunikacije

---

### 4.2. Visoke Ranjivosti - Detaljna Analiza

#### VULN-003: Pattern-Based SQL Injection Detection

**Kako se može eksploatisati:**

1. **Encoding obfuscation:**
   ```
   Input: %27%20OR%20%271%27=%271
   Decoded: ' OR '1'='1
   Pattern matching ne detektuje jer traži plain text
   ```

2. **Case/whitespace variations:**
   ```
   Input: 'Or'1'='1
   Pattern matching ne detektuje jer traži tačan case
   ```

3. **Comments:**
   ```
   Input: '/*comment*/OR/*comment*/'1'='1
   Pattern matching ne detektuje
   ```

**Rizik:**
- **HIGH** - SQL injection napadi mogu proći neprimećeni

---

#### VULN-004: Pattern-Based XSS Detection

**Kako se može eksploatisati:**

1. **HTML encoding:**
   ```
   Input: &lt;script&gt;alert(1)&lt;/script&gt;
   Pattern matching traži <script>, ne detektuje
   ```

2. **Event handlers:**
   ```
   Input: <img src=x onerror=alert(1)>
   Pattern matching ne pokriva sve event handler-e
   ```

3. **SVG vectors:**
   ```
   Input: <svg onload=alert(1)>
   Pattern matching ne detektuje
   ```

**Rizik:**
- **HIGH** - XSS napadi mogu proći neprimećeni

---

#### VULN-005: Rate Limiting po IP Adresi

**Kako se može eksploatisati:**

1. **IP spoofing:**
   ```
   Napadač lažira X-Forwarded-For header:
   X-Forwarded-For: 192.168.1.100
   Rate limiter koristi lažnu IP adresu
   ```

2. **Distributed attacks:**
   ```
   Napadač koristi botnet sa više IP adresa
   - Svaki bot ima svoj rate limit
   - Ukupno prekoračenje limita
   ```

3. **Proxy/VPN rotation:**
   ```
   Napadač rotira IP adrese kroz proxy/VPN
   - Svaka nova IP ima svoj limit
   - Brute force napadi
   ```

**Rizik:**
- **HIGH** - DoS i brute force napadi

---

## 5. Preporuke za Prevazilaženje

### 5.1. VULN-001: Hardcoded Admin Credentials

**Rešenje:**

1. **Ukloniti hardcoded credentials:**
   ```go
   // Ukloniti initAdminUser() funkciju
   // Ili koristiti environment variable
   adminPassword := os.Getenv("ADMIN_PASSWORD")
   if adminPassword == "" {
       log.Fatal("ADMIN_PASSWORD environment variable required")
   }
   ```

2. **Koristiti secret management:**
   - HashiCorp Vault
   - AWS Secrets Manager
   - Kubernetes Secrets

3. **Prvo pokretanje:**
   - Admin korisnik se kreira samo pri prvom pokretanju
   - Lozinka se generiše i prikazuje samo jednom
   - Lozinka se traži pri prvom pokretanju

**Implementacija:**
```go
func initAdminUser(ctx context.Context, userRepo *store.UserRepository, cfg *config.Config) {
    adminPassword := os.Getenv("ADMIN_PASSWORD")
    if adminPassword == "" {
        // Generate random password
        adminPassword = generateSecurePassword()
        log.Printf("Admin password generated: %s (SAVE THIS SECURELY!)", adminPassword)
    }
    // ...
}
```

---

### 5.2. VULN-002: InsecureSkipVerify

**Rešenje:**

1. **Za Development:**
   - Ovo je prihvatljivo, ali dodati warning:
   ```go
   if os.Getenv("ENV") == "development" {
       log.Warn("WARNING: InsecureSkipVerify enabled - NOT FOR PRODUCTION!")
       tr := &http.Transport{
           TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
       }
   }
   ```

2. **Za Produkciju:**
   - Koristiti validne sertifikate
   - Ukloniti `InsecureSkipVerify: true`
   - Koristiti certificate pinning
   ```go
   certPool := x509.NewCertPool()
   certPool.AppendCertsFromPEM(caCert)
   tr := &http.Transport{
       TLSClientConfig: &tls.Config{
           RootCAs: certPool,
       },
   }
   ```

---

### 5.3. VULN-003: Pattern-Based SQL Injection Detection

**Rešenje:**

1. **Koristiti Prepared Statements:**
   - MongoDB driver već koristi parameterized queries (bezbedno)
   - Ako se koristi SQL, uvek koristiti prepared statements

2. **Poboljšati pattern matching:**
   ```go
   func CheckSQLInjection(input string) error {
       // Normalize input
       normalized := strings.ToLower(input)
       normalized = url.QueryUnescape(normalized) // Decode URL encoding
       normalized = html.UnescapeString(normalized) // Decode HTML entities
       
       // Remove comments
       normalized = regexp.MustCompile(`/\*.*?\*/`).ReplaceAllString(normalized, "")
       normalized = regexp.MustCompile(`--.*`).ReplaceAllString(normalized, "")
       
       // Check patterns
       sqlPatterns := []string{
           "or '1'='1",
           "or '1'='1'",
           "union select",
           "drop table",
           // ...
       }
       
       for _, pattern := range sqlPatterns {
           if strings.Contains(normalized, pattern) {
               return ErrSQLInjection
           }
       }
       
       return nil
   }
   ```

3. **Koristiti whitelisting umesto blacklisting:**
   - Dozvoliti samo poznate dobre vrednosti
   - Validirati format, ne samo proveravati zle pattern-e

---

### 5.4. VULN-004: Pattern-Based XSS Detection

**Rešenje:**

1. **Poboljšati pattern matching:**
   ```go
   func CheckXSS(input string) error {
       // Normalize input
       normalized := strings.ToLower(input)
       normalized = url.QueryUnescape(normalized)
       normalized = html.UnescapeString(normalized)
       
       // Check for script tags (various encodings)
       scriptPatterns := []*regexp.Regexp{
           regexp.MustCompile(`<script`),
           regexp.MustCompile(`&lt;script`),
           regexp.MustCompile(`%3Cscript`),
           regexp.MustCompile(`&#60;script`),
       }
       
       for _, pattern := range scriptPatterns {
           if pattern.MatchString(normalized) {
               return ErrXSS
           }
       }
       
       // Check for event handlers
       eventHandlers := []string{
           "onerror=", "onload=", "onclick=", "onmouseover=",
           "onfocus=", "onblur=", "onchange=",
       }
       
       for _, handler := range eventHandlers {
           if strings.Contains(normalized, handler) {
               return ErrXSS
           }
       }
       
       return nil
   }
   ```

2. **Output encoding (već implementirano):**
   - `html.EscapeString()` za HTML output
   - JSON encoding automatski escape-uje
   - React automatski escape-uje

---

### 5.5. VULN-005: Rate Limiting po IP Adresi

**Rešenje:**

1. **Poboljšati IP extraction:**
   ```go
   func getClientIP(r *http.Request) string {
       // Proveri X-Forwarded-For (uzmi prvi IP, ne poslednji)
       forwarded := r.Header.Get("X-Forwarded-For")
       if forwarded != "" {
           ips := strings.Split(forwarded, ",")
           return strings.TrimSpace(ips[0]) // Prvi IP je originalni
       }
       
       // Proveri X-Real-IP
       realIP := r.Header.Get("X-Real-IP")
       if realIP != "" {
           return realIP
       }
       
       // Fallback na RemoteAddr
       ip := r.RemoteAddr
       if idx := strings.LastIndex(ip, ":"); idx != -1 {
           ip = ip[:idx]
       }
       return ip
   }
   ```

2. **Dodati user-based rate limiting:**
   ```go
   // Rate limiting po user ID-u (ako je autentifikovan)
   if userID := getUserIDFromToken(r); userID != "" {
       if !limiter.Allow("user:"+userID) {
           return http.StatusTooManyRequests
       }
   }
   ```

3. **Dodati CAPTCHA za osetljive endpoint-e:**
   - Registracija
   - Login
   - Password reset

4. **Dodati progressive delays:**
   ```go
   // Nakon 5 neuspešnih pokušaja, dodati delay
   if failedAttempts > 5 {
       time.Sleep(time.Duration(failedAttempts-5) * time.Second)
   }
   ```

---

### 5.6. VULN-006: MD5 za File Integrity

**Rešenje:**

1. **Zameniti MD5 sa SHA256:**
   ```go
   func CalculateFileHash(reader io.Reader) (string, error) {
       hash := sha256.New()
       if _, err := io.Copy(hash, reader); err != nil {
           return "", ErrFileReadError
       }
       return hex.EncodeToString(hash.Sum(nil)), nil
   }
   ```

2. **Ili koristiti SHA512 za još veću sigurnost**

---

### 5.7. VULN-007: Prazna sanitizeMessage Funkcija

**Rešenje:**

1. **Implementirati sanitizaciju:**
   ```go
   func sanitizeMessage(message string) string {
       // Remove passwords (common patterns)
       passwordPattern := regexp.MustCompile(`(?i)(password|pwd|pass)\s*[:=]\s*[^\s]+`)
       message = passwordPattern.ReplaceAllString(message, "$1=***")
       
       // Remove tokens
       tokenPattern := regexp.MustCompile(`(?i)(token|bearer|jwt)\s*[:=]\s*[^\s]+`)
       message = tokenPattern.ReplaceAllString(message, "$1=***")
       
       // Remove email addresses (keep domain)
       emailPattern := regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@([A-Za-z0-9.-]+\.[A-Z|a-z]{2,})\b`)
       message = emailPattern.ReplaceAllString(message, "***@$1")
       
       // Remove stack traces
       lines := strings.Split(message, "\n")
       var sanitizedLines []string
       for _, line := range lines {
           if strings.HasPrefix(strings.TrimSpace(line), "goroutine") ||
              strings.HasPrefix(strings.TrimSpace(line), "panic:") ||
              strings.Contains(line, "runtime.") {
               continue
           }
           sanitizedLines = append(sanitizedLines, line)
       }
       
       return strings.Join(sanitizedLines, "\n")
   }
   ```

---

### 5.8. VULN-008: JWT Secret Hardcoding

**Rešenje:**

1. **Ukloniti default vrednost:**
   ```go
   jwtSecret := cfg.JWTSecret
   if jwtSecret == "" {
       log.Fatal("JWT_SECRET environment variable is required")
   }
   ```

2. **Generisati jak secret:**
   ```go
   // Pri prvom pokretanju, generisati secret
   if jwtSecret == "" {
       secret := make([]byte, 32)
       if _, err := rand.Read(secret); err != nil {
           log.Fatal("Failed to generate JWT secret")
       }
       jwtSecret = base64.URLEncoding.EncodeToString(secret)
       log.Printf("Generated JWT secret (save this!): %s", jwtSecret)
   }
   ```

3. **Koristiti secret management:**
   - Environment variables
   - Secret management servisi
   - Kubernetes secrets

---

### 5.9. VULN-009 i VULN-010: Poboljšanje SQL/XSS Detection

**Rešenje:**

1. **Koristiti biblioteke:**
   - `github.com/sonatype-nexus-community/nancy` za dependency scanning
   - `github.com/securego/gosec` za security scanning

2. **Defense in Depth:**
   - Input validation (već implementirano)
   - Output encoding (već implementirano)
   - Content Security Policy (CSP) headers
   - Parameterized queries (MongoDB driver)

---

### 5.10. VULN-011: Error Messages

**Rešenje:**

1. **Generičke error poruke:**
   ```go
   // Umesto:
   http.Error(w, "failed to create user: "+err.Error(), http.StatusInternalServerError)
   
   // Koristiti:
   http.Error(w, "internal server error", http.StatusInternalServerError)
   log.Printf("Failed to create user: %v", err) // Loguj detalje
   ```

2. **Error handling middleware:**
   ```go
   func ErrorHandler(next http.HandlerFunc) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
           defer func() {
               if err := recover(); err != nil {
                   log.Printf("Panic: %v", err)
                   http.Error(w, "internal server error", http.StatusInternalServerError)
               }
           }()
           next(w, r)
       }
   }
   ```

---

### 5.11. VULN-012: CORS Configuration

**Rešenje:**

1. **Whitelist origin-e:**
   ```go
   allowedOrigins := []string{
       "https://musicstreaming.com",
       "https://www.musicstreaming.com",
   }
   
   origin := r.Header.Get("Origin")
   allowed := false
   for _, allowedOrigin := range allowedOrigins {
       if origin == allowedOrigin {
           allowed = true
           break
       }
   }
   
   if !allowed && origin != "" {
       http.Error(w, "origin not allowed", http.StatusForbidden)
       return
   }
   
   if allowed {
       w.Header().Set("Access-Control-Allow-Origin", origin)
   }
   ```

---

## 6. Zaštita od Eksploatacije

### 6.1. Defense in Depth Strategija

#### Sloj 1: Network Layer
- **Firewall** - Blokiranje nepotrebnih portova
- **DDoS Protection** - Cloudflare, AWS Shield
- **WAF (Web Application Firewall)** - ModSecurity, AWS WAF

#### Sloj 2: Application Layer
- **Input Validation** - ✅ Implementirano
- **Output Encoding** - ✅ Implementirano
- **Authentication/Authorization** - ✅ Implementirano
- **Rate Limiting** - ✅ Implementirano (poboljšati)

#### Sloj 3: Data Layer
- **Encryption at Rest** - MongoDB encryption
- **Encryption in Transit** - HTTPS ✅
- **Access Control** - Database user permissions

#### Sloj 4: Monitoring & Detection
- **Logging** - ✅ Implementirano
- **Intrusion Detection** - SIEM sistem
- **Anomaly Detection** - Machine learning

---

### 6.2. Security Headers

**Dodati security headers:**

```go
func SecurityHeaders(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Content Security Policy
        w.Header().Set("Content-Security-Policy", 
            "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'")
        
        // X-Frame-Options (clickjacking protection)
        w.Header().Set("X-Frame-Options", "DENY")
        
        // X-Content-Type-Options (MIME sniffing protection)
        w.Header().Set("X-Content-Type-Options", "nosniff")
        
        // X-XSS-Protection
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        
        // Strict-Transport-Security (HSTS)
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        
        // Referrer-Policy
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        
        next(w, r)
    }
}
```

---

### 6.3. Regular Security Audits

1. **Monthly Code Reviews**
   - Pregled novog koda
   - Provera security best practices

2. **Quarterly Penetration Testing**
   - External penetration testing
   - Internal security assessment

3. **Dependency Updates**
   - Weekly dependency scanning
   - Monthly dependency updates
   - Immediate updates za critical vulnerabilities

4. **Security Training**
   - Developer security training
   - OWASP Top 10 awareness
   - Secure coding practices

---

### 6.4. Incident Response Plan

1. **Detection**
   - Automated monitoring
   - Anomaly detection
   - Alert system

2. **Response**
   - Immediate containment
   - Investigation
   - Remediation

3. **Recovery**
   - System restoration
   - Data recovery
   - Post-incident review

---

## 7. Zaključak

### 7.1. Rezime Ranjivosti

| Severity | Broj | Status |
|----------|------|--------|
| Critical | 2 | Zahtevaju hitno rešavanje |
| High | 4 | Zahtevaju prioritetno rešavanje |
| Medium | 4 | Zahtevaju planirano rešavanje |
| Low | 2 | Zahtevaju monitoring |

### 7.2. Preporuke po Prioritetu

#### Prioritet 1 (Hitno - 1 nedelja):
1. ✅ Ukloniti hardcoded admin credentials (VULN-001)
2. ✅ Ukloniti InsecureSkipVerify za produkciju (VULN-002)
3. ✅ Implementirati sanitizeMessage (VULN-007)
4. ✅ Ukloniti default JWT secret (VULN-008)

#### Prioritet 2 (Visok - 1 mesec):
1. ✅ Poboljšati SQL injection detection (VULN-003)
2. ✅ Poboljšati XSS detection (VULN-004)
3. ✅ Poboljšati rate limiting (VULN-005)
4. ✅ Zameniti MD5 sa SHA256 (VULN-006)

#### Prioritet 3 (Srednji - 3 meseca):
1. ✅ Dodati security headers
2. ✅ Poboljšati error handling
3. ✅ Poboljšati CORS konfiguraciju
4. ✅ Implementirati CAPTCHA

### 7.3. Ukupna Ocena Bezbednosti

**Trenutna ocena:** 6.5/10 (Medium)

**Nakon implementacije preporuka:** 8.5/10 (High)

### 7.4. Finalne Napomene

- Aplikacija ima dobru osnovu za bezbednost
- Većina kritičnih zaštita je implementirana
- Potrebno je poboljšati neke implementacije
- Regular security audits su neophodni

---

## Dodatak A: Komande za Security Scanning

### A.1. Gosec Scanning
```bash
# Instalacija
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Scanning
gosec ./...

# Sa detaljnim izveštajem
gosec -fmt json -out results.json ./...
```

### A.2. Dependency Scanning
```bash
# Snyk
snyk test

# OWASP Dependency-Check
dependency-check --scan . --format JSON --out dependency-report.json

# Go vulnerability check
govulncheck ./...
```

### A.3. Code Quality
```bash
# GolangCI-Lint
golangci-lint run

# SonarQube
sonar-scanner
```

---

## Dodatak B: Security Checklist

- [ ] Hardcoded credentials uklonjeni
- [ ] InsecureSkipVerify uklonjen za produkciju
- [ ] JWT secret iz environment variable
- [ ] SQL injection detection poboljšan
- [ ] XSS detection poboljšan
- [ ] Rate limiting poboljšan
- [ ] MD5 zamenjen sa SHA256
- [ ] sanitizeMessage implementiran
- [ ] Security headers dodati
- [ ] Error messages generički
- [ ] CORS whitelist konfigurisan
- [ ] Dependency scanning automatski
- [ ] Penetration testing sproveden
- [ ] Incident response plan kreiran

---

**Izveštaj pripremio:** Security Analysis Team  
**Datum:** 12. februar 2025  
**Verzija:** 1.0
