# Brzi Vodič za Testiranje

## 🚀 Pokretanje Test Skripte

```powershell
.\test-system.ps1
```

## ✅ Osnovni Testovi

### 1. Provera da li servisi rade

```powershell
# Provera statusa svih servisa
docker-compose ps

# Provera API Gateway-a
Invoke-WebRequest -Uri "http://localhost:8081/api/users/health" -UseBasicParsing
```

**Očekivani rezultat:** Status 200 OK

### 2. Testiranje MailHog-a

1. **Otvorite MailHog Web UI**: http://localhost:8025
2. **Otvorite frontend**: http://localhost:3000
3. **Pokušajte admin login**:
   - Email: `admin@musicstreaming.com`
   - Kliknite "Request OTP"
4. **Proverite MailHog** - trebalo bi da vidite email sa OTP kodom

### 3. Provera HTTPS Sertifikata

```powershell
# Provera da li sertifikati postoje
ls certs\

# Trebalo bi da vidite:
# - server.crt
# - server.key
```

### 4. Provera HTTPS Komunikacije Između Servisa

```powershell
# Provera environment varijabli u API Gateway-u
docker exec projekat-2025-1-api-gateway-1 env | Select-String "SERVICE_URL"

# Trebalo bi da vidite:
# USERS_SERVICE_URL=https://users-service:8001
# CONTENT_SERVICE_URL=https://content-service:8002
# itd.
```

### 5. Provera Password Hashing-a

```powershell
# Registrujte novog korisnika preko frontend-a ili API-ja
# Zatim proverite MongoDB:

docker exec projekat-2025-1-mongodb-users-1 mongosh --quiet --eval "db.users.findOne({email: 'test@example.com'}, {password: 1})"

# Password treba da bude heširan (počinje sa $2a$ ili $2b$)
```

## 📋 Checklist

- [x] Svi Docker servisi su pokrenuti
- [x] API Gateway odgovara na zahteve
- [x] MailHog Web UI je dostupan
- [x] HTTPS sertifikati postoje
- [x] Servisi koriste HTTPS za inter-service komunikaciju
- [x] Email se šalje kada se traži OTP
- [x] Password se čuva kao hash (ne plain text)

## 🎯 Glavni Endpoint-i za Testiranje

- **API Gateway**: http://localhost:8081
- **Users Health**: http://localhost:8081/api/users/health
- **Content Health**: http://localhost:8081/api/content/health
- **MailHog UI**: http://localhost:8025
- **Frontend**: http://localhost:3000

## 📧 Testiranje Email Funkcionalnosti

1. Otvorite frontend: http://localhost:3000
2. Pokušajte da se ulogujete kao admin: `admin@musicstreaming.com`
3. Kliknite "Request OTP"
4. Otvorite MailHog: http://localhost:8025
5. Trebalo bi da vidite email sa OTP kodom

## 🔒 Testiranje Sigurnosti

### Password Hashing
- Registrujte novog korisnika
- Proverite MongoDB - password treba da bude bcrypt hash

### HTTPS
- Proverite logove servisa - trebalo bi da vidite "Starting HTTPS server"
- Proverite environment varijable - trebalo bi da koriste `https://`

### POST Metoda
- Senzitivni podaci (login, registration) se šalju preko POST metode
- Proverite Network tab u browser developer tools-u

## 🐛 Troubleshooting

**Problem:** MailHog ne prima email-e
- Proverite da li je MailHog pokrenut: `docker ps | Select-String mailhog`
- Proverite logove: `docker logs projekat-2025-1-users-service-1 | Select-String EMAIL`

**Problem:** API Gateway vraća 404
- Proverite da li je API Gateway pokrenut: `docker ps | Select-String api-gateway`
- Koristite `/api/*` endpoint-e umesto root endpoint-a

**Problem:** HTTPS sertifikati nisu pronađeni
- Generišite sertifikate: `.\generate-certs.ps1`
- Proverite da li postoje: `ls certs\`
