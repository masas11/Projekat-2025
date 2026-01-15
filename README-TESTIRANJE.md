# Uputstva za Testiranje - Ocena 6

## Pokretanje sistema

1. **Pokrenite Docker Compose:**
   ```bash
   cd Projekat-2025
   docker-compose up --build
   ```

2. **Sačekajte da se svi servisi pokrenu** (može potrajati nekoliko minuta, posebno za Cassandra)

3. **Proverite da li su svi servisi pokrenuti:**
   - API Gateway: http://localhost:8081
   - Users Service: http://localhost:8001
   - Content Service: http://localhost:8002
   - Notifications Service: http://localhost:8005

## Uvoz Postman kolekcije

1. Otvorite Postman
2. Kliknite na **Import**
3. Izaberite fajl `Test-Collection-Ocena-6.postman_collection.json`
4. Kolekcija će biti uvezena sa svim zahtevima

## Redosled testiranja

### 1. Registracija i Prijava (1.1, 1.2)

1. **Registracija - Uspešna**
   - Endpoint: `POST /api/users/register`
   - Očekivani status: `201 Created`
   - Proverite da li je korisnik kreiran

2. **Registracija - Slaba lozinka**
   - Endpoint: `POST /api/users/register`
   - Očekivani status: `400 Bad Request`
   - Proverite da li se vraća poruka o grešci

3. **Prijava - Zahtev OTP**
   - Endpoint: `POST /api/users/login/request-otp`
   - Očekivani status: `200 OK`
   - OTP će biti poslat na email (mock - proverite logove)

4. **Prijava - Verifikacija OTP**
   - Endpoint: `POST /api/users/login/verify-otp`
   - Očekivani status: `200 OK`
   - **Napomena:** OTP se generiše u logovima users-service

### 2. Umetnici (1.3)

1. **Kreiranje umetnika**
   - Endpoint: `POST /api/content/artists`
   - **Važno:** Zabeležite `id` umetnika iz odgovora
   - Očekivani status: `201 Created`

2. **Pregled svih umetnika**
   - Endpoint: `GET /api/content/artists`
   - Očekivani status: `200 OK`
   - Proverite da li se vraća lista umetnika

3. **Pregled pojedinačnog umetnika**
   - Endpoint: `GET /api/content/artists/:artistId`
   - Zamenite `:artistId` sa ID-jem iz koraka 1
   - Očekivani status: `200 OK`

4. **Izmena umetnika**
   - Endpoint: `PUT /api/content/artists/:artistId`
   - Očekivani status: `200 OK`

### 3. Albumi i Pesme (1.4, 1.5)

1. **Kreiranje albuma**
   - Endpoint: `POST /api/content/albums`
   - **Važno:** Zabeležite `id` albuma i `artistIds` iz odgovora
   - Očekivani status: `201 Created`

2. **Kreiranje pesme - Album postoji**
   - Endpoint: `POST /api/content/songs`
   - Zamenite `ALBUM_ID_HERE` sa ID-jem albuma iz koraka 1
   - Zamenite `ARTIST_ID_HERE` sa ID-jem umetnika
   - Očekivani status: `201 Created`

3. **Kreiranje pesme - Album ne postoji**
   - Endpoint: `POST /api/content/songs`
   - Koristite `NON_EXISTENT_ALBUM_ID` za albumId
   - Očekivani status: `400 Bad Request`
   - Proverite da li se vraća poruka "album not found"

4. **Pregled svih albuma**
   - Endpoint: `GET /api/content/albums`
   - Očekivani status: `200 OK`

5. **Pregled albuma umetnika**
   - Endpoint: `GET /api/content/albums/by-artist?artistId=...`
   - Zamenite `artistId` sa ID-jem umetnika
   - Očekivani status: `200 OK`

6. **Pregled pesama albuma**
   - Endpoint: `GET /api/content/songs/by-album?albumId=...`
   - Zamenite `albumId` sa ID-jem albuma
   - Očekivani status: `200 OK`
   - Proverite da li se vraćaju pesme koje pripadaju tom albumu

### 4. Notifikacije (1.11)

1. **Pregled notifikacija korisnika**
   - Endpoint: `GET /api/notifications?userId=user1`
   - Očekivani status: `200 OK`
   - Proverite da li se vraćaju sample notifikacije (inicijalizovane pri pokretanju servisa)

2. **Pregled notifikacija - Bez userId**
   - Endpoint: `GET /api/notifications`
   - Očekivani status: `400 Bad Request`
   - Proverite da li se vraća poruka o grešci

## Napomene

### JWT Token za Admin operacije

Za kreiranje umetnika, albuma i pesama potreban je JWT token sa admin rolom. Trenutno:
- API Gateway prosleđuje Authorization header ka backend servisima
- Za testiranje možete koristiti placeholder token: `Bearer admin-token`
- **Za produkciju:** Implementirati stvarnu JWT autentifikaciju

### Cassandra baza podataka

- Notifications service sada koristi **Cassandra** umesto MongoDB
- Keyspace i tabela se automatski kreiraju pri pokretanju servisa
- Sample notifikacije se inicijalizuju pri pokretanju

### Promenljive u Postman zahtevima

U Postman kolekciji postoje placeholder vrednosti:
- `ARTIST_ID_HERE` - zamenite sa stvarnim ID-jem umetnika
- `ALBUM_ID_HERE` - zamenite sa stvarnim ID-jem albuma
- `SONG_ID_HERE` - zamenite sa stvarnim ID-jem pesme

**Preporuka:** Koristite Postman environment variables za ove vrednosti.

## Troubleshooting

### Cassandra se ne pokreće

```bash
# Proverite logove
docker-compose logs cassandra

# Restartujte Cassandra
docker-compose restart cassandra
```

### Servisi se ne povezuju

1. Proverite da li su svi servisi pokrenuti: `docker-compose ps`
2. Proverite logove: `docker-compose logs [service-name]`
3. Proverite mrežu: `docker network ls`

### Notifikacije se ne vraćaju

1. Proverite da li je Cassandra pokrenuta
2. Proverite logove notifications-service: `docker-compose logs notifications-service`
3. Proverite da li su sample notifikacije inicijalizovane (treba da vidite "Sample notifications initialized" u logovima)

## Testiranje kroz cURL (alternativa)

```bash
# Registracija
curl -X POST http://localhost:8081/api/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "firstName": "Marko",
    "lastName": "Marković",
    "email": "marko@example.com",
    "username": "marko123",
    "password": "StrongPass123!",
    "confirmPassword": "StrongPass123!"
  }'

# Pregled umetnika
curl http://localhost:8081/api/content/artists

# Pregled notifikacija
curl "http://localhost:8081/api/notifications?userId=user1"
```
