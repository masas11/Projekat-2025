# Testiranje Generisanja Notifikacija (1.11)

## Brzi test korak po korak

### 1. Pokreni sistem
```bash
docker-compose up -d
```

### 2. Prijavi se kao redovan korisnik (ne admin)
- Otvori frontend: http://localhost:3000
- Prijavi se sa nalogom redovnog korisnika

### 3. Pretplati se na umetnika ili žanr
- Idi na stranicu umetnika i klikni "Pretplati se"
- ILI pretplati se na žanr sa liste pesama

### 4. Prijavi se kao admin (u drugom browseru ili incognito)
- Prijavi se sa admin nalogom

### 5. Test slučaj 1: Novi album umetnika na koji si pretplaćen
- Kao admin, kreiraj novi album umetnika na čiji sadržaj si pretplaćen
- Vrati se na profil redovnog korisnika
- Proveri notifikacije - trebalo bi da vidiš: "New album 'Album Name' by Artist Name has been released"

### 6. Test slučaj 2: Nova pesma umetnika na koji si pretplaćen
- Kao admin, kreiraj novu pesmu umetnika na čiji sadržaj si pretplaćen
- Vrati se na profil redovnog korisnika
- Proveri notifikacije - trebalo bi da vidiš: "New song 'Song Name' by Artist Name has been added"

### 7. Test slučaj 3: Novi umetnik žanra na koji si pretplaćen
- Kao admin, kreiraj novog umetnika sa žanrom na koji si pretplaćen
- Vrati se na profil redovnog korisnika
- Proveri notifikacije - trebalo bi da vidiš: "New artist 'Artist Name' in genre Genre has been added"

## Provera logova

### Subscriptions service logovi
```bash
docker-compose logs subscriptions-service | grep -i "event\|notification"
```

Trebalo bi da vidiš:
- "Processed new_album event for album..."
- "Processed new_song event for song..."
- "Processed new_artist event for artist..."
- "Notification created for user..."

### Notifications service logovi
```bash
docker-compose logs notifications-service | tail -20
```

## API test (alternativa)

### 1. Pretplati se na umetnika
```bash
curl -X POST "http://localhost:8081/api/subscriptions/subscribe-artist?artistId=ARTIST_ID&userId=USER_ID" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 2. Kao admin, kreiraj novi album
```bash
curl -X POST "http://localhost:8081/api/content/albums" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Album",
    "releaseDate": "2025-01-01",
    "genre": "Pop",
    "artistIDs": ["ARTIST_ID"]
  }'
```

### 3. Proveri notifikacije
```bash
curl "http://localhost:8081/api/notifications?userId=USER_ID" \
  -H "Authorization: Bearer USER_TOKEN"
```

## Očekivani rezultati

✅ Notifikacije se kreiraju automatski kada se doda novi sadržaj
✅ Poruke sadrže imena umetnika (ne samo "by artist")
✅ Notifikacije se prikazuju u frontendu na profilu korisnika
✅ Logovi pokazuju uspešnu obradu eventa

## Troubleshooting

**Problem: Notifikacije se ne kreiraju**
- Proveri da li je korisnik stvarno pretplaćen: `GET /api/subscriptions?userId=USER_ID`
- Proveri logove subscriptions-service za greške
- Proveri da li notifications-service radi: `GET /health`

**Problem: Poruke ne sadrže imena umetnika**
- Proveri da li su artistNames u eventu: logovi subscriptions-service
- Proveri da li artistRepo radi u content-service
