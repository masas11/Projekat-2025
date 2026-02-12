# 📋 Vodič za Proveru Registracije Naloga

## ✅ Status Implementacije

**Registracija je POTPUNO IMPLEMENTIRANA** sa svim zahtevanim funkcionalnostima:

### Implementirane Funkcionalnosti:

1. ✅ **Jedinstven username** - Proverava se da li username već postoji
2. ✅ **Obavezna polja**: Ime, Prezime, Email, Username, Lozinka, Potvrda lozinke
3. ✅ **Jaka lozinka** - Validacija na frontendu i backendu:
   - Najmanje 8 karaktera
   - Najmanje jedno veliko slovo
   - Najmanje jedan broj
4. ✅ **Periodična promena lozinke** - Lozinka ističe nakon 60 dana (konfigurabilno)
5. ✅ **Potvrda registracije** - Email verifikacija sa tokenom

---

## 🧪 Kako Proveriti Registraciju

### Metoda 1: Preko Frontend Aplikacije (Preporučeno)

#### Korak 1: Pokrenite Sistem
```powershell
# Pokrenite sve servise
docker-compose up -d

# Sačekajte da se servisi pokrenu
Start-Sleep -Seconds 20
```

#### Korak 2: Pokrenite Frontend
```powershell
cd frontend
npm install  # samo prvi put
npm start
```

#### Korak 3: Otvorite Registraciju
1. Otvorite browser: `http://localhost:3000/register`
2. Ili kliknite na "Registruj se" link u navigaciji

#### Korak 4: Testirajte Različite Scenarije

**Test 1: Uspešna Registracija**
- Ime: `Marko`
- Prezime: `Marković`
- Email: `marko@example.com`
- Username: `marko123` (mora biti jedinstven)
- Lozinka: `Test1234` (ispunjava kriterijume)
- Potvrdi lozinku: `Test1234`
- **Očekivano**: Poruka "Uspešna registracija! Email za verifikaciju je poslat..."

**Test 2: Slaba Lozinka**
- Lozinka: `test` (prekratka)
- **Očekivano**: Greška "Lozinka mora imati najmanje 8 karaktera"

- Lozinka: `testtest` (bez velikog slova i broja)
- **Očekivano**: Greška "Lozinka mora sadržati najmanje jedno veliko slovo i jedan broj"

**Test 3: Lozinke se ne Poklapaju**
- Lozinka: `Test1234`
- Potvrdi lozinku: `Test12345`
- **Očekivano**: Greška "Lozinke se ne poklapaju"

**Test 4: Duplikat Username**
- Registrujte korisnika sa username `marko123`
- Pokušajte ponovo sa istim username-om
- **Očekivano**: Greška "user already exists"

**Test 5: Email Verifikacija**
- Nakon registracije, proverite konzolu (F12 → Console)
- Trebalo bi da se vidi poruka o slanju email-a
- Email link bi trebao biti: `http://localhost:3000/verify-email?token=...`
- Kliknite na link ili otvorite direktno u browseru
- **Očekivano**: Poruka "Email je uspešno verifikovan!"

---

### Metoda 2: Preko API-ja (curl/Postman)

#### Test 1: Uspešna Registracija
```powershell
$body = @{
    firstName = "Jovan"
    lastName = "Jovanović"
    email = "jovan@example.com"
    username = "jovan123"
    password = "Test1234"
    confirmPassword = "Test1234"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8081/api/users/register" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

**Očekivani odgovor:**
```json
{
  "message": "registration successful, verification email sent"
}
```

#### Test 2: Slaba Lozinka
```powershell
$body = @{
    firstName = "Test"
    lastName = "Test"
    email = "test@example.com"
    username = "testuser"
    password = "weak"
    confirmPassword = "weak"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8081/api/users/register" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

**Očekivani odgovor:** HTTP 400 sa porukom o slaboj lozinci

#### Test 3: Duplikat Username
```powershell
# Prvo registrujte korisnika
$body = @{
    firstName = "Petar"
    lastName = "Petrović"
    email = "petar@example.com"
    username = "petar123"
    password = "Test1234"
    confirmPassword = "Test1234"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8081/api/users/register" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body

# Pokušajte ponovo sa istim username-om
Invoke-RestMethod -Uri "http://localhost:8081/api/users/register" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

**Očekivani odgovor:** HTTP 409 sa porukom "user already exists"

#### Test 4: Lozinke se ne Poklapaju
```powershell
$body = @{
    firstName = "Ana"
    lastName = "Anić"
    email = "ana@example.com"
    username = "ana123"
    password = "Test1234"
    confirmPassword = "Test12345"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8081/api/users/register" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

**Očekivani odgovor:** HTTP 400 sa porukom "passwords do not match"

---

### Metoda 3: Provera u Bazi Podataka

#### Pregled Registrovanih Korisnika
```powershell
# Povežite se na MongoDB
docker exec -it mongodb-users mongosh

# U MongoDB shell-u:
use users_db
db.users.find().pretty()
```

**Proverite polja:**
- `username` - jedinstven
- `email` - jedinstven
- `verified` - false (dok se ne verifikuje email)
- `passwordExpiresAt` - datum kada lozinka ističe (60 dana od kreiranja)
- `passwordChangedAt` - datum poslednje promene lozinke

---

## 📁 Relevantni Fajlovi

### Frontend:
- `frontend/src/components/Register.js` - Forma za registraciju
- `frontend/src/components/VerifyEmail.js` - Komponenta za verifikaciju email-a
- `frontend/src/services/api.js` - API pozivi

### Backend:
- `services/users-service/internal/handler/register.go` - Handler za registraciju
- `services/users-service/internal/handler/verification_handler.go` - Handler za verifikaciju
- `services/users-service/internal/validation/password.go` - Validacija lozinke
- `services/users-service/internal/store/user_repository.go` - Provera jedinstvenosti username-a
- `services/users-service/config/config.go` - Konfiguracija (password expiration days)

---

## 🔍 Dodatne Provere

### Provera Periodične Promene Lozinke

Lozinka se automatski postavlja da ističe nakon određenog broja dana. Podrazumevano je 60 dana, ali može se promeniti:

```powershell
# U docker-compose.yml ili .env fajlu
PASSWORD_EXPIRATION_DAYS=60
```

Kada korisnik pokuša da se prijavi sa isteklom lozinkom, dobijaće grešku i moraće da promeni lozinku.

### Provera Email Verifikacije

Token za verifikaciju se čuva u MongoDB kolekciji `magic_links` sa tipom `verification`:

```javascript
// U MongoDB shell-u:
db.magic_links.find({type: "verification"}).pretty()
```

Token ističe nakon 24 sata.

---

## ✅ Checklist za Proveru

- [ ] Frontend forma prikazuje sva obavezna polja
- [ ] Validacija lozinke radi na frontendu
- [ ] Backend validacija lozinke radi
- [ ] Duplikat username vraća grešku
- [ ] Duplikat email vraća grešku
- [ ] Lozinke se proveravaju da li se poklapaju
- [ ] Email za verifikaciju se šalje nakon registracije
- [ ] Verifikacioni link radi
- [ ] Korisnik se ne može prijaviti dok ne verifikuje email
- [ ] PasswordExpiresAt se postavlja pri registraciji

---

## 🐛 Troubleshooting

### Problem: Email se ne šalje
- Proverite logove: `docker-compose logs users-service`
- Email funkcionalnost možda koristi mock implementaciju (proverite `services/users-service/internal/mail/mailer.go`)

### Problem: Verifikacioni link ne radi
- Proverite da li je frontend pokrenut na `http://localhost:3000`
- Proverite da li je token ispravno URL-encoded u linku
- Proverite logove u browser konzoli (F12)

### Problem: Duplikat username ne vraća grešku
- Proverite MongoDB konekciju
- Proverite da li se provera izvršava u `user_repository.go` Create metodi

---

## 📝 Napomene

- Email verifikacija koristi token koji ističe nakon 24 sata
- Lozinka ističe nakon 60 dana (konfigurabilno)
- Korisnik mora verifikovati email pre nego što može da se prijavi
- Sve lozinke se čuvaju kao bcrypt hash (ne u plain text-u)
