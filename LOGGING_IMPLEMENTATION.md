# Implementacija Logovanja (2.20)

## ✅ Implementirano

### 1. Strukturirani Logger (`services/shared/logger/logger.go`)

**Funkcionalnosti:**
- ✅ Nivoi logovanja: `INFO`, `WARN`, `ERROR`, `AUDIT`
- ✅ Logovanje u fajlove sa rotacijom
- ✅ Maksimalna veličina fajla: 10MB (konfigurabilno)
- ✅ Zadržava 5 rotiranih fajlova (konfigurabilno)
- ✅ SHA256 checksums za integritet log-datoteka
- ✅ Filtriranje osetljivih podataka (passwords, tokens, OTP)
- ✅ Zaštita log-datoteka (permissions: 0640)

**Tipovi događaja:**
- `VALIDATION_FAILURE` - Neuspeh validacije
- `LOGIN_SUCCESS` - Uspešna prijava
- `LOGIN_FAILURE` - Neuspešna prijava
- `ACCESS_CONTROL_FAILURE` - Neuspeh kontrole pristupa
- `STATE_CHANGE` - Neočekivana promena state podataka
- `INVALID_TOKEN` - Nevalidan token
- `EXPIRED_TOKEN` - Istekao token
- `ADMIN_ACTIVITY` - Administratorska aktivnost
- `TLS_FAILURE` - Neuspešna TLS konekcija

### 2. Integracija u Users Service

**Logovani događaji:**
- ✅ Neuspehe validacije ulaznih podataka (email, username, password, SQL injection, XSS)
- ✅ Uspešne prijave (sa username i IP adresom)
- ✅ Neuspešne prijave (sa razlogom: invalid password, user not found, email not verified, account locked, password expired, invalid OTP, expired OTP)

**Lokacija logova:**
- `services/users-service/logs/app-YYYY-MM-DD.log`
- Checksum fajlovi: `app-YYYY-MM-DD.log.checksum`

### 3. Zaštita Log-Datoteka

- ✅ Permissions: 0640 (samo vlasnik i grupa mogu čitati)
- ✅ SHA256 checksums za verifikaciju integriteta
- ✅ Rotacija automatski sprečava prevelike fajlove

### 4. Filtriranje Osetljivih Podataka

- ✅ Passwords se maskiraju kao `***`
- ✅ Tokens se maskiraju (samo prefix se loguje)
- ✅ OTP se maskiraju
- ✅ Secrets se maskiraju

## ⚠️ Delimično Implementirano

### 1. API Gateway Logovanje

**Treba dodati:**
- Logovanje neuspeha kontrole pristupa (`RequireAuth`, `RequireRole`)
- Logovanje nevalidnih tokena
- Logovanje isteklih tokena
- Logovanje TLS grešaka

**Lokacija:** `services/api-gateway/internal/middleware/auth.go`

### 2. Administratorske Aktivnosti

**Treba dodati:**
- Logovanje kreiranja/izmene/brisanja umetnika (Content Service)
- Logovanje kreiranja/izmene/brisanja albuma (Content Service)
- Logovanje kreiranja/izmene/brisanja pesama (Content Service)
- Logovanje promene korisničkih uloga (Users Service)

**Lokacija:** 
- `services/content-service/internal/handler/*_handler.go`
- `services/users-service/internal/handler/*_handler.go`

### 3. Neočekivane Promene State Podataka

**Treba dodati:**
- Detekcija neočekivanih promena u korisničkim podacima
- Detekcija neočekivanih promena u state-u sesije
- Logovanje promena koje nisu inicirane od strane korisnika

### 4. TLS Greške

**Treba dodati:**
- Logovanje TLS handshake grešaka u API Gateway
- Logovanje TLS grešaka u inter-service komunikaciji
- Logovanje sertifikatnih grešaka

**Lokacija:** 
- `services/api-gateway/cmd/main.go` (u `proxyRequest` funkciji)
- `services/*/cmd/main.go` (u `ListenAndServeTLS` error handler-ima)

## 📝 Primeri Korišćenja

### Logovanje Validacione Greške

```go
logger.LogValidationFailure("email", "invalid format", "invalid@email")
```

### Logovanje Uspešne Prijave

```go
logger.LogLoginSuccess("username", "192.168.1.1")
```

### Logovanje Neuspešne Prijave

```go
logger.LogLoginFailure("username", "invalid password", "192.168.1.1")
```

### Logovanje Neuspeha Kontrole Pristupa

```go
logger.LogAccessControlFailure("user123", "/api/admin/users", "DELETE", "insufficient permissions")
```

### Logovanje Administratorske Aktivnosti

```go
logger.LogAdminActivity("admin123", "CREATE_ARTIST", "artists", map[string]interface{}{
    "artistId": "artist1",
    "name": "New Artist",
})
```

### Logovanje TLS Greške

```go
logger.LogTLSFailure("users-service", "certificate verification failed", "172.22.0.5:8001")
```

## 🔧 Konfiguracija

### Environment Variables

```bash
LOG_DIR=./logs  # Direktorijum za log fajlove (default: ./logs)
```

### Rotacija Logova

- Maksimalna veličina: 10MB (konfigurabilno u `logger.go`)
- Broj rotiranih fajlova: 5 (konfigurabilno u `logger.go`)
- Format rotiranih fajlova: `app-YYYY-MM-DD.log.YYYYMMDD-HHMMSS`

## 🔒 Bezbednost

1. **Zaštita Log-Datoteka:**
   - Permissions: 0640 (samo vlasnik i grupa)
   - Logovi se čuvaju u zaštićenom direktorijumu

2. **Integritet:**
   - SHA256 checksums za svaki log fajl
   - Checksum fajlovi se čuvaju odvojeno
   - `VerifyIntegrity()` metoda za verifikaciju

3. **Filtriranje Osetljivih Podataka:**
   - Passwords se nikad ne loguju
   - Tokens se loguju samo sa prefixom
   - Stack trace-ovi se ne loguju u production modu

## 📊 Format Log Entries

```
[LEVEL] EventType=EVENT_TYPE Message=MESSAGE Fields=key1=value1 key2=value2
```

**Primer:**
```
[WARN] EventType=LOGIN_FAILURE Message=Login failed Fields=username=testuser reason=invalid password ip=192.168.1.1 timestamp=1705123456
[AUDIT] EventType=LOGIN_SUCCESS Message=Login successful Fields=username=testuser ip=192.168.1.1 timestamp=1705123456
```

## 🚀 Sledeći Koraci

1. Integrisati logger u API Gateway middleware
2. Dodati logovanje administratorskih aktivnosti u Content Service
3. Implementirati detekciju neočekivanih promena state podataka
4. Dodati logovanje TLS grešaka u sve servise
5. Dodati monitoring i alerting na osnovu logova

## 📚 Testiranje

```bash
# Proveri logove
tail -f services/users-service/logs/app-*.log

# Proveri integritet
cd services/users-service
go run -c 'import "users-service/internal/logger"; logger.GetLogger().VerifyIntegrity()'
```
