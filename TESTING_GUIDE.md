# Detaljni Vodič za Testiranje - Ocena 7

## 🚀 Priprema

### 1.1 Pokretanje svih servisa
```bash
# U root folderu projekta
docker-compose up -d

# Proveri da li svi rade
docker-compose ps
```

### 1.2 Proveri health endpoint-e
```bash
# API Gateway
curl http://localhost:8080/health

# Content Service
curl http://localhost:8081/health

# Users Service
curl http://localhost:8082/health

# Ratings Service
curl http://localhost:8083/health

# Subscriptions Service
curl http://localhost:8084/health

# Notifications Service
curl http://localhost:8085/health
```

---

## 📋 Testiranje Funkcionalnosti

### ✅ Zahtevi za Ocenu 6

#### 1.1 Registracija naloga
```bash
# POST request
curl -X POST http://localhost:8080/api/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "StrongPass123!",
    "email": "test@example.com",
    "firstName": "Test",
    "lastName": "User"
  }'

# Očekivani odgovor: 201 Created
# Proveri email za verifikaciju
```

#### 1.2 Prijava na sistem (OTP)
```bash
# 1. Zahtevaj OTP
curl -X POST http://localhost:8080/api/users/login/request-otp \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser"
  }'

# 2. Proveri email i dobij OTP kod
# 3. Verifikuj OTP
curl -X POST http://localhost:8080/api/users/login/verify-otp \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "otp": "123456"
  }'

# Očekivani odgovor: JWT token
```

#### 1.3 Kreiranje umetnika (Admin)
```bash
# Prvo prijavi se kao admin
# Zatim kreiraj umetnika
curl -X POST http://localhost:8080/api/content/artists \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "Test Artist",
    "biography": "Test biography",
    "genres": ["Pop", "Rock"]
  }'
```

#### 1.4 Kreiranje albuma i pesama
```bash
# Kreiraj album
curl -X POST http://localhost:8080/api/content/albums \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "Test Album",
    "releaseDate": "2024-01-01",
    "genre": "Pop",
    "artistIds": ["ARTIST_ID_FROM_PREVIOUS_STEP"]
  }'

# Kreiraj pesmu
curl -X POST http://localhost:8080/api/content/songs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "Test Song",
    "duration": 180,
    "genre": "Pop",
    "albumId": "ALBUM_ID_FROM_PREVIOUS_STEP",
    "artistIds": ["ARTIST_ID"],
    "audioFileUrl": "https://www.soundhelix.com/examples/mp3/SoundHelix-Song-1.mp3"
  }'
```

#### 1.5 Pregled sadržaja
```bash
# Lista svih umetnika
curl http://localhost:8080/api/content/artists

# Detalji umetnika
curl http://localhost:8080/api/content/artists/{artistId}

# Lista svih albuma
curl http://localhost:8080/api/content/albums

# Lista svih pesama
curl http://localhost:8080/api/content/songs

# Detalji pesme
curl http://localhost:8080/api/content/songs/{songId}
```

#### 1.11 Pregled notifikacija
```bash
curl http://localhost:8080/api/notifications \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

### 🎵 Dodatno za Ocenu 7

#### 1.6 Reprodukcija pesme
```bash
# Test streaming endpoint-a
curl -I http://localhost:8081/api/content/songs/{songId}/stream

# Frontend test:
# 1. Otvori http://localhost:3000
# 2. Idi na /songs/{songId}
# 3. Proveri AudioPlayer komponentu
# 4. Klikni play dugme
```

#### 1.7 Filtriranje i pretraga
```bash
# Pretraga pesama po nazivu
curl "http://localhost:8080/api/content/songs?search=Test"

# Filtriranje umetnika po žanru
curl "http://localhost:8080/api/content/artists?genre=Pop"
```

#### 1.8 Ocenjivanje pesama
```bash
# Oceni pesmu (sa sinhronom validacijom)
curl -X POST "http://localhost:8083/rate-song?songId={songId}&rating=5&userId={userId}"

# Proveri da li radi sinhrona komunikacija:
# - Ako pesme ne postoji: treba vratiti 404
# - Ako content service nije dostupan: treba aktivirati fallback
```

#### 1.9 Pretplata na umetnike
```bash
# Pretplati se na umetnika (sa sinhronom validacijom)
curl -X POST "http://localhost:8084/subscribe-artist?artistId={artistId}&userId={userId}"

# Proveri sinhronu komunikaciju:
# - Ako umetnik ne postoji: treba vratiti 404
# - Ako content service nije dostupan: treba aktivirati fallback
```

---

## 🔧 Testiranje Otpornosti na Otkaze (2.7)

### 2.7.1 HTTP Client konfiguracija
```bash
# Proveri timeout - zaustavi content service
docker-compose stop content-service

# Pokušaj ocenjivanje - treba pasti nakon 2 sekunde
curl -X POST "http://localhost:8083/rate-song?songId=test&rating=5&userId=test"
```

### 2.7.2 Timeout testiranje
```bash
# Sledi logove - treba videti timeout poruke
docker-compose logs ratings-service
```

### 2.7.3 Fallback logika
```bash
# Sa content service down, proveri fallback poruke
curl -X POST "http://localhost:8084/subscribe-artist?artistId=test&userId=test"

# Očekivano: "Content-service unavailable, fallback activated"
```

### 2.7.4 Circuit Breaker testiranje
```bash
# 1. Pokreni content service
docker-compose start content-service

# 2. Napravi 3 neuspešna poziva redom (sa nevalidnim songId)
for i in {1..3}; do
  curl -X POST "http://localhost:8083/rate-song?songId=invalid&rating=5&userId=test"
done

# 3. Četvrti poziv treba da vrati "circuit breaker open"
curl -X POST "http://localhost:8083/rate-song?songId=valid&rating=5&userId=test"

# 4. Sačekaj 5+ sekundi i probaj opet - treba raditi
sleep 6
curl -X POST "http://localhost:8083/rate-song?songId=valid&rating=5&userId=test"
```

---

## 🌐 Frontend Testiranje

### Pokretanje frontend-a
```bash
cd frontend
npm start
```

### Test rute u browseru:
- `http://localhost:3000` - Početna stranica
- `http://localhost:3000/login` - Prijava
- `http://localhost:3000/register` - Registracija
- `http://localhost:3000/artists` - Lista umetnika
- `http://localhost:3000/artists/{id}` - Detalji umetnika
- `http://localhost:3000/albums` - Lista albuma
- `http://localhost:3000/songs` - Lista pesama
- `http://localhost:3000/songs/{id}` - Detalji pesme sa AudioPlayer
- `http://localhost:3000/url-tester` - URL tester za audio fajlove

---

## 📊 Provera Logova

### Svi servisi
```bash
# Proveri logove svih servisa
docker-compose logs -f

# Specifični servis
docker-compose logs -f content-service
docker-compose logs -f ratings-service
docker-compose logs -f subscriptions-service
```

### Traži ključne poruke:
- "Circuit breaker opened"
- "Retrying call to content-service"
- "Content-service unavailable, fallback activated"
- "Song rated successfully"
- "Subscribed to artist successfully"

---

## ✅ Checklist za Ocenu 7

- [ ] Registracija radi (email verifikacija)
- [ ] Prijava sa OTP radi
- [ ] Kreiranje umetnika radi
- [ ] Kreiranje albuma i pesama radi
- [ ] Pregled sadržaja radi
- [ ] Notifikacje se prikazuju
- [ ] AudioPlayer reprodukuje pesme
- [ ] Pretraga i filtriranje rade
- [ ] Ocenjivanje pesama radi (sa validacijom)
- [ ] Pretplata na umetnike radi (sa validacijom)
- [ ] Timeout se primenjuje (2 sekunde)
- [ ] Retry mehanizam radi (2 puta)
- [ ] Fallback logika radi
- [ ] Circuit breaker otvara/se zatvara
- [ ] Frontend prikazuje sve funkcionalnosti

---

## 🐛 Debug Tips

### Ako nešto ne radi:
1. Proveri da li su svi servisi pokrenuti: `docker-compose ps`
2. Proveri logove: `docker-compose logs {service-name}`
3. Proveri portove: `netstat -tulpn | grep :808`
4. Očisti i restartuj: `docker-compose down && docker-compose up -d`

### Najčešći problemi:
- Port konflikti - promeni portove u docker-compose.yml
- MongoDB ne startuje - proveri docker volume
- Frontend ne može da se poveže - proveri API proxy u package.json
