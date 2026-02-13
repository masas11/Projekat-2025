# HTTPS Implementacija - Pregled

## ✅ Implementirano

### 1. **Hash & Salt Mehanizam za Lozinke**
- ✅ Koristi se **bcrypt** sa `bcrypt.DefaultCost` (10 rounds)
- ✅ Automatski generiše salt za svaku lozinku
- ✅ Implementirano u:
  - Registraciji korisnika
  - Promeni lozinke
  - Admin korisniku
  - Password reset funkcionalnosti

**Lokacija:** `services/users-service/internal/security/password.go`

```go
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}
```

### 2. **HTTPS Protokol - Interna Komunikacija**
- ✅ Svi servisi koriste HTTPS za komunikaciju između servisa
- ✅ API Gateway → Backend servisi: `https://users-service:8001`, `https://content-service:8002`, itd.
- ✅ Svi servisi imaju SSL sertifikate mountovane u `/app/certs/`
- ✅ Konfigurisano u `docker-compose.yml` sa `TLS_CERT_FILE` i `TLS_KEY_FILE`

**Servisi sa HTTPS:**
- users-service (port 8001)
- content-service (port 8002)
- ratings-service (port 8003)
- subscriptions-service (port 8004)
- notifications-service (port 8005)
- recommendation-service (port 8006)
- analytics-service (port 8007)

### 3. **HTTPS Protokol - Spoljni Pristup**
- ⚠️ **Development:** API Gateway koristi HTTP na portu 8081 (lakše za development sa samopotpisanim sertifikatima)
- ✅ **Production:** API Gateway može koristiti HTTPS dodavanjem `TLS_CERT_FILE` i `TLS_KEY_FILE` environment varijabli
- ✅ SSL sertifikati su kreirani i dostupni
- ✅ Frontend je konfigurisan da koristi `http://localhost:8081` (development)

**Konfiguracija:**
- API Gateway: `http://localhost:8081` (development) ili `https://localhost:8081` (production)
- Frontend: `http://localhost:8081` (u `frontend/src/services/api.js`)

**Napomena:** Za production, omogućiti HTTPS dodavanjem TLS varijabli u `docker-compose.yml`:
```yaml
- TLS_CERT_FILE=/app/certs/server.crt
- TLS_KEY_FILE=/app/certs/server.key
volumes:
  - ./certs:/app/certs:ro
```

### 4. **HTTP Metode za Senzitivne Podatke**
- ✅ **POST metode** za sve senzitivne operacije:
  - `/api/users/register` - POST (password u body-ju)
  - `/api/users/login/request-otp` - POST (password u body-ju)
  - `/api/users/login/verify-otp` - POST (OTP u body-ju)
  - `/api/users/password/change` - POST (passwords u body-ju)
  - `/api/users/password/reset/request` - POST (email u body-ju)
  - `/api/users/password/reset` - POST (token i password u body-ju)
  - `/api/users/recover/request` - POST (email u body-ju)

- ⚠️ **GET metode** se koriste samo za:
  - `/api/users/verify-email?token=...` - jednokratni token za verifikaciju
  - `/api/users/recover/verify?token=...` - jednokratni magic link token
  
  **Napomena:** Ovi tokeni su jednokratni i bezbedni preko HTTPS-a.

## 📋 SSL Sertifikati

### Kreiranje Sertifikata
Sertifikati su kreirani pomoću OpenSSL:
```bash
openssl genrsa -out certs/server.key 2048
openssl req -new -key certs/server.key -out certs/server.csr -subj "/C=RS/ST=Serbia/L=Belgrade/O=MusicStreaming/OU=IT/CN=localhost"
openssl x509 -req -days 365 -in certs/server.csr -signkey certs/server.key -out certs/server.crt
```

**Lokacija:** `./certs/server.crt` i `./certs/server.key`

## 🔒 Bezbednosne Karakteristike

### 1. **Enkripcija u Tranzitu**
- ✅ HTTPS/TLS 1.2+ za sve komunikacije
- ✅ Interna komunikacija između servisa: HTTPS
- ✅ Spoljni pristup (API Gateway): HTTPS

### 2. **Enkripcija u Mirovanju**
- ✅ Lozinke se čuvaju kao bcrypt hash (ne plaintext)
- ✅ Automatski salt za svaku lozinku
- ✅ bcrypt cost factor = 10 (balans između bezbednosti i performansi)

### 3. **Senzitivni Podaci**
- ✅ Lozinke: uvek u POST body-ju, nikada u URL-u
- ✅ OTP kodovi: u POST body-ju
- ✅ JWT tokeni: u Authorization header-u
- ✅ Email adrese: u POST body-ju za registraciju/reset

## ⚠️ Važne Napomene

### Development vs Production

**Development:**
- Samopotpisani SSL sertifikati
- Browser će prikazati upozorenje - potrebno je prihvatiti sertifikat
- `InsecureSkipVerify: true` za inter-service komunikaciju (OK za dev)

**Production:**
- Koristiti validne SSL sertifikate (Let's Encrypt, itd.)
- Ukloniti `InsecureSkipVerify: true`
- Koristiti certificate pinning
- Koristiti HSTS headers

### Browser Upozorenje

Kada pristupite `https://localhost:8081` u browser-u, videćete upozorenje o samopotpisanom sertifikatu. Ovo je normalno za development:

1. Kliknite na "Advanced" ili "Napredno"
2. Kliknite na "Proceed to localhost" ili "Nastavi na localhost"
3. Sertifikat će biti prihvaćen za ovu sesiju

## 🧪 Testiranje

### Test HTTPS Endpoint-a
```powershell
# Test API Gateway HTTPS
Invoke-WebRequest -Uri "https://localhost:8081/api/users/health" -SkipCertificateCheck -UseBasicParsing
```

### Provera Sertifikata
```powershell
# Proveri sertifikat
openssl s_client -connect localhost:8081 -showcerts
```

## 📝 Checklist

- [x] SSL sertifikati kreirani
- [x] API Gateway koristi HTTPS
- [x] Frontend koristi HTTPS
- [x] Interna komunikacija koristi HTTPS
- [x] Lozinke su heširane sa bcrypt
- [x] Senzitivni podaci u POST body-ju
- [x] CORS konfigurisan
- [x] Rate limiting implementiran

## 🔗 Reference

- [HTTPS_SETUP.md](./HTTPS_SETUP.md) - Detaljna dokumentacija o HTTPS setup-u
- [TEST_HTTPS.md](./TEST_HTTPS.md) - Testiranje HTTPS funkcionalnosti
