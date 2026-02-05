# HTTPS Setup za Ocena 7

## 🔐 SSL Sertifikati

### Kreiranje sertifikata:
```bash
# Linux/MacOS
chmod +x generate-certs.sh
./generate-certs.sh

# Windows sa OpenSSL
openssl genrsa -out certs/server.key 2048
openssl req -new -key certs/server.key -out certs/server.csr -subj "/C=RS/ST=Serbia/L=Belgrade/O=MusicStreaming/OU=IT/CN=localhost"
openssl x509 -req -days 365 -in certs/server.csr -signkey certs/server.key -out certs/server.crt
del certs/server.csr
```

## 🚀 Pokretanje sa HTTPS

### 1. Kreiraj sertifikate:
```bash
./generate-certs.sh
```

### 2. Pokreni servise:
```bash
docker-compose -f docker-compose.https.yml up -d
```

### 3. Frontend promena:
```javascript
// U frontend/src/services/api.js
const API_BASE_URL = process.env.REACT_APP_API_URL || 'https://localhost:8081';
```

## 🔍 Bezbednosne karakteristike:

### ✅ **Implementirano:**
- **Hash & Salt**: bcrypt za lozinke
- **HTTPS**: SSL/TLS enkripcija svih servisa
- **POST metode**: Senzitivni podaci u body (ne URL)
- **Input validacija**: XSS i SQL injection zaštita

### 🛡️ **Sigurnosni mehanizmi:**
- **Enkripcija u tranzitu**: TLS 1.2+
- **Enkripcija u mirovanju**: MongoDB heširane lozinke
- **JWT tokeni**: Sigurna autentifikacija
- **CORS**: Zaštita cross-origin zahteva

## 📋 Testiranje:
```bash
# Proveri HTTPS endpoint
curl -k https://localhost:8081/health

# Proveri sertifikat
openssl s_client -connect localhost:8081 -showcerts
```

## ⚠️ **Napomena:**
- Samopotpisani sertifikati za development
- Za production koristiti Let's Encrypt
- `-k` flag za curl zbog samopotpisanih sertifikata
