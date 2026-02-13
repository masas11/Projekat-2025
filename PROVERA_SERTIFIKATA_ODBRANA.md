# Provera Sertifikata za Odbranu

## 📁 Gde Se Nalaze Sertifikati?

### Lokacija Fajlova:

```powershell
certs/
├── server.crt  (SSL sertifikat)
└── server.key  (Privatni ključ)
```

**Provera:**
```powershell
ls certs\
```

**Trebalo bi da vidite:**
- `server.crt` - SSL sertifikat (Certificate)
- `server.key` - Privatni ključ (Private Key)

## 🔍 Kako da Proverite Sertifikat

### 1. Provera da li sertifikati postoje:

```powershell
ls certs\
```

**Očekivano:**
```
server.crt
server.key
```

### 2. Detalji sertifikata (ako imate OpenSSL):

```powershell
# Windows (ako imate Git instaliran)
& "C:\Program Files\Git\mingw64\bin\openssl.exe" x509 -in certs\server.crt -text -noout
```

**Šta ćete videti:**
- Certificate information
- Issuer (izdavalac)
- Subject (CN=localhost)
- Validity period
- Public key info

### 3. Provera u Browser-u:

1. Otvorite: `https://localhost:8081/api/users/health`
2. Kliknite na **"🔒"** ili **"Not Secure"** ikonu u address bar-u
3. Kliknite na **"Certificate"**
4. Videćete detalje sertifikata:
   - Issued to: localhost
   - Issued by: (self-signed)
   - Valid from: (datum)
   - Valid to: (datum)

## ✅ Provera da li HTTPS Radi

### U Network Tab-u (DevTools):

1. **Otvorite DevTools:** `F12`
2. **Idite na Network tab**
3. **Kliknite na zahtev** (npr. "notifications")
4. **Idite na "Headers" tab**
5. **Proverite "General" sekciju:**

**Trebalo bi da vidite:**
```
Request URL: https://localhost:8081/api/notifications
Request Method: GET
Status Code: 200 OK
```

**Ako vidite `https://` u Request URL → HTTPS radi! ✅**

### Provera Protocol:

U **"Headers" tab-u**, proverite:
- **Protocol:** `h2` (HTTP/2 over HTTPS) ili `http/1.1` (HTTPS)
- Ako vidite `h2` ili `http/1.1` → HTTPS radi ✅

## 📋 Šta Pokazati na Odbrani

### 1. Sertifikati kao fajlovi:

```powershell
# Pokazati da sertifikati postoje
ls certs\
```

**Rezultat:**
```
server.crt
server.key
```

### 2. Konfiguracija u docker-compose.yml:

```yaml
api-gateway:
  environment:
    - TLS_CERT_FILE=/app/certs/server.crt
    - TLS_KEY_FILE=/app/certs/server.key
  volumes:
    - ./certs:/app/certs:ro
```

### 3. Logovi koji pokazuju HTTPS:

```powershell
docker logs projekat-2025-1-api-gateway-1 --tail 3
```

**Očekivano:**
```
Starting HTTPS server on port 8080
```

### 4. Environment varijable:

```powershell
docker exec projekat-2025-1-api-gateway-1 env | Select-String "TLS|SERVICE_URL"
```

**Očekivano:**
```
TLS_CERT_FILE=/app/certs/server.crt
TLS_KEY_FILE=/app/certs/server.key
USERS_SERVICE_URL=https://users-service:8001
CONTENT_SERVICE_URL=https://content-service:8002
...
```

### 5. Network Tab u Browser-u:

- Otvorite DevTools → Network tab
- Kliknite na bilo koji zahtev ka `localhost:8081`
- Proverite Headers → Request URL treba da počinje sa `https://`

## 🎯 Checklist za Odbranu

- [ ] Sertifikati postoje u `certs/` direktorijumu
- [ ] `docker-compose.yml` ima TLS varijable za API Gateway
- [ ] Logovi pokazuju "Starting HTTPS server"
- [ ] Environment varijable pokazuju HTTPS URL-ove
- [ ] Network tab pokazuje `https://` u Request URL
- [ ] Browser može da pristupi `https://localhost:8081`

## 📝 Napomene

**Self-Signed Sertifikati:**
- Kreirani lokalno za development
- Browser pokazuje "Not Secure" (normalno)
- HTTPS **radi** - komunikacija je šifrovana
- Za production koristiti validne sertifikate od CA

**Za Odbranu:**
- Pokazati sertifikate u `certs/` direktorijumu
- Pokazati konfiguraciju u `docker-compose.yml`
- Pokazati logove koji potvrđuju HTTPS
- Pokazati Network tab sa `https://` zahtevima
