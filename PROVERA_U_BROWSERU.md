# 🌐 Kako Proveriti da li Projekat Radi u Browseru

## ✅ KORAK 1: Proverite da li servisi rade

### Test 1: API Gateway Health Check

Otvorite browser i idite na:

```
http://localhost:8081/api/users/health
```

**Šta treba da vidite:**
- Tekst: `users-service is running` ✅

### Test 2: Content Service Health Check

```
http://localhost:8002/health
```

**Šta treba da vidite:**
- Tekst: `content-service is running` ✅

### Test 3: Users Service Health Check

```
http://localhost:8001/health
```

**Šta treba da vidite:**
- Tekst: `users-service is running` ✅

### Test 4: Notifications Service Health Check

```
http://localhost:8005/health
```

**Šta treba da vidite:**
- Tekst: `notifications-service is running` ✅

---

## 🎨 KORAK 2: Pokrenite Frontend

### Opcija A: Ako frontend NIJE pokrenut

1. Otvorite NOVI CMD prozor
2. Idite u frontend folder:

```cmd
cd D:\projekat\Projekat-2025\frontend
```

3. Instalirajte dependencies (ako nije već urađeno):

```cmd
npm install
```

4. Pokrenite frontend:

```cmd
npm start
```

5. Browser će se automatski otvoriti na: `http://localhost:3000`

### Opcija B: Ako frontend VEĆ radi

Samo otvorite browser i idite na:

```
http://localhost:3000
```

---

## 🧪 KORAK 3: Testiranje API Endpoints u Browseru

### Test 1: Lista umetnika (Artists)

```
http://localhost:8081/api/content/artists
```

**Šta treba da vidite:**
- JSON sa listom umetnika (može biti prazan `[]` ako nema podataka) ✅

### Test 2: Lista albuma (Albums)

```
http://localhost:8081/api/content/albums
```

**Šta treba da vidite:**
- JSON sa listom albuma (može biti prazan `[]`) ✅

### Test 3: Lista pesama (Songs)

```
http://localhost:8081/api/content/songs
```

**Šta treba da vidite:**
- JSON sa listom pesama (može biti prazan `[]`) ✅

---

## 🎯 KORAK 4: Testiranje Frontend Aplikacije

### 1. Početna stranica

Idite na: `http://localhost:3000`

**Šta treba da vidite:**
- Početnu stranicu sa opcijama za Login/Register ✅

### 2. Registracija

1. Kliknite na "Registruj se"
2. Popunite formu:
   - Ime: Test
   - Prezime: User
   - Email: test@test.com
   - Username: testuser
   - Password: Test123!
   - Confirm Password: Test123!
3. Kliknite "Register"

**Šta treba da vidite:**
- Poruku o uspešnoj registraciji ✅
- Ili grešku ako korisnik već postoji (to je OK)

### 3. Login

1. Kliknite na "Prijavi se"
2. Unesite:
   - Username: `admin` (ili korisnika koji ste kreirali)
   - Password: `admin123` (za admin nalog)
3. Kliknite "Request OTP"
4. Proverite CMD prozor gde je `users-service` pokrenut - videćete OTP kod u logovima
5. Unesite OTP kod
6. Kliknite "Verify OTP"

**Šta treba da vidite:**
- Uspešan login ✅
- Preusmeravanje na početnu stranicu sa korisničkim imenom ✅

### 4. Pregled umetnika

1. Kliknite na "Izvođači" (Artists)
2. Trebalo bi da vidite listu umetnika

**Šta treba da vidite:**
- Listu umetnika (može biti prazna ako nema podataka) ✅

### 5. Notifikacije

1. Kliknite na "Notifikacije"
2. Trebalo bi da vidite notifikacije za korisnika

**Šta treba da vidite:**
- Listu notifikacija (test podaci su već kreirani) ✅

---

## 🔍 KORAK 5: Provera Developer Console

### Otvorite Developer Tools

1. U browseru pritisnite `F12` ili `Ctrl + Shift + I`
2. Idite na tab "Console"
3. Idite na tab "Network"

### Proverite da li ima grešaka

**U Console tab-u:**
- Ne bi trebalo da vidite crvene greške ✅
- Ako vidite greške, proverite da li su servisi pokrenuti

**U Network tab-u:**
- Kada kliknete na neki link, videćete HTTP zahteve
- Proverite da li su status kodovi `200 OK` ✅

---

## ❓ Česti Problemi

### Problem 1: "Cannot GET /api/..."

**Rešenje:**
- Proverite da li je `api-gateway` pokrenut
- Proverite CMD prozor gde je `docker-compose` pokrenut
- Trebalo bi da vidite: `API Gateway running on port 8081`

### Problem 2: Frontend se ne učitava

**Rešenje:**
1. Proverite da li je frontend pokrenut:
   ```cmd
   # U CMD prozoru gde ste pokrenuli npm start
   # Trebalo bi da vidite: "webpack compiled successfully"
   ```

2. Proverite da li je port 3000 slobodan

3. Restartujte frontend:
   ```cmd
   # Pritisnite Ctrl + C
   # Zatim: npm start
   ```

### Problem 3: "Network Error" ili CORS greške

**Rešenje:**
- Proverite da li su svi servisi pokrenuti
- Proverite da li API Gateway radi na portu 8081
- Frontend koristi proxy na `http://localhost:8081` (proverite `package.json`)

### Problem 4: Prazne liste (artists, albums, songs)

**To je OK!** ✅
- Ako nema podataka u bazi, liste će biti prazne `[]`
- To znači da servisi rade, ali baza je prazna
- Možete dodati podatke preko API-ja ili direktno u MongoDB

---

## 🎯 Brzi Test Checklist

- [ ] `http://localhost:8081/api/users/health` → `users-service is running`
- [ ] `http://localhost:8002/health` → `content-service is running`
- [ ] `http://localhost:3000` → Frontend se učitava
- [ ] `http://localhost:8081/api/content/artists` → JSON odgovor (može biti `[]`)
- [ ] Login funkcioniše
- [ ] Registracija funkcioniše
- [ ] Notifikacije se prikazuju

---

## 📊 Dodatne Provere

### Provera MongoDB konekcije

Ako želite da proverite da li MongoDB radi:

1. Otvorite MongoDB Compass (ako je instaliran)
2. Connection string: `mongodb://localhost:27017`
3. Kliknite "Connect"
4. Trebalo bi da vidite baze: `users_db`, `music_streaming`, `notifications_db`

---

## ✅ Rezime

**Ako vidite:**
- ✅ Health check endpoints vraćaju "is running"
- ✅ Frontend se učitava na `localhost:3000`
- ✅ API endpoints vraćaju JSON (čak i prazan `[]`)
- ✅ Login/Register funkcionišu

**SVE RADI!** 🎉🎉🎉

