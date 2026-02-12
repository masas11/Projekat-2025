# Testiranje Logovanja

Ovaj dokument opisuje kako da testirate sve implementirane funkcionalnosti logovanja.

## Priprema

1. Pokrenite sve servise:
```bash
docker-compose up -d
```

2. Proverite da li su log direktorijumi kreirani:
```bash
ls -la services/*/logs/
```

## 1. Testiranje API Gateway Logovanja

### 1.1. Testiranje neuspeha kontrole pristupa (RequireAuth)

**Test scenarij:** Pokušaj pristupa zaštićenom endpoint-u bez tokena

```bash
# Test bez tokena
curl -X GET https://localhost:8081/api/users/logout -k

# Proverite logove
cat services/api-gateway/logs/app.log | grep "ACCESS_CONTROL_FAILURE"
```

**Očekivani rezultat:**
- Log sa tipom `ACCESS_CONTROL_FAILURE`
- Razlog: "missing authorization header"

### 1.2. Testiranje nevalidnih tokena

**Test scenarij:** Pokušaj pristupa sa nevalidnim tokenom

```bash
# Test sa nevalidnim tokenom
curl -X GET https://localhost:8081/api/users/logout \
  -H "Authorization: Bearer invalid_token_123" \
  -k

# Proverite logove
cat services/api-gateway/logs/app.log | grep "INVALID_TOKEN"
```

**Očekivani rezultat:**
- Log sa tipom `INVALID_TOKEN`
- Token prefix i razlog greške

### 1.3. Testiranje isteklih tokena

**Test scenarij:** Pokušaj pristupa sa isteklim tokenom

```bash
# Prvo se prijavite i dobijte token
# Zatim sačekajte da token istekne (ili koristite stari token)

curl -X GET https://localhost:8081/api/users/logout \
  -H "Authorization: Bearer <expired_token>" \
  -k

# Proverite logove
cat services/api-gateway/logs/app.log | grep "EXPIRED_TOKEN"
```

**Očekivani rezultat:**
- Log sa tipom `EXPIRED_TOKEN`
- UserID i IP adresa

### 1.4. Testiranje neuspeha kontrole pristupa (RequireRole)

**Test scenarij:** Pokušaj pristupa admin endpoint-u sa non-admin tokenom

```bash
# Prijavite se kao običan korisnik (ne admin)
# Zatim pokušajte da kreirate artist

curl -X POST https://localhost:8081/api/content/artists \
  -H "Authorization: Bearer <non_admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Artist","biography":"Test","genres":["Rock"]}' \
  -k

# Proverite logove
cat services/api-gateway/logs/app.log | grep "ACCESS_CONTROL_FAILURE" | grep "insufficient permissions"
```

**Očekivani rezultat:**
- Log sa tipom `ACCESS_CONTROL_FAILURE`
- Razlog: "insufficient permissions: required role ADMIN, user role USER"

### 1.5. Testiranje TLS grešaka u inter-service komunikaciji

**Test scenarij:** Simulacija TLS greške (npr. neispravan sertifikat)

```bash
# Ovo će se desiti automatski ako postoji problem sa sertifikatima
# Proverite logove
cat services/api-gateway/logs/app.log | grep "TLS_FAILURE"
```

## 2. Testiranje Content Service Logovanja

### 2.1. Testiranje logovanja administratorskih aktivnosti - CREATE

**Test scenarij:** Kreiranje novog artist-a kao admin

```bash
# Prijavite se kao admin i dobijte token
# Zatim kreirajte artist

curl -X POST https://localhost:8081/api/content/artists \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Artist",
    "biography": "Test biography",
    "genres": ["Rock", "Pop"]
  }' \
  -k

# Proverite logove
cat services/content-service/logs/app.log | grep "ADMIN_ACTIVITY" | grep "CREATE_ARTIST"
```

**Očekivani rezultat:**
- Log sa tipom `ADMIN_ACTIVITY`
- Action: "CREATE_ARTIST"
- Resource: "artists"
- Detalji: artistId, name, genres

### 2.2. Testiranje logovanja administratorskih aktivnosti - UPDATE

**Test scenarij:** Ažuriranje artist-a kao admin

```bash
# Ažurirajte postojeći artist
curl -X PUT https://localhost:8081/api/content/artists/<artist_id> \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Artist",
    "biography": "Updated biography",
    "genres": ["Jazz"]
  }' \
  -k

# Proverite logove
cat services/content-service/logs/app.log | grep "ADMIN_ACTIVITY" | grep "UPDATE_ARTIST"
cat services/content-service/logs/app.log | grep "STATE_CHANGE" | grep "artist"
```

**Očekivani rezultat:**
- Log sa tipom `ADMIN_ACTIVITY` za UPDATE_ARTIST
- Log sa tipom `STATE_CHANGE` sa starim i novim stanjem

### 2.3. Testiranje logovanja administratorskih aktivnosti - DELETE

**Test scenarij:** Brisanje artist-a kao admin

```bash
# Obrišite artist
curl -X DELETE https://localhost:8081/api/content/artists/<artist_id> \
  -H "Authorization: Bearer <admin_token>" \
  -k

# Proverite logove
cat services/content-service/logs/app.log | grep "ADMIN_ACTIVITY" | grep "DELETE_ARTIST"
```

**Očekivani rezultat:**
- Log sa tipom `ADMIN_ACTIVITY`
- Action: "DELETE_ARTIST"
- Detalji: artistId, name

### 2.4. Testiranje za Album i Song

Ponovite iste testove za album i song:

```bash
# CREATE ALBUM
curl -X POST https://localhost:8081/api/content/albums \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Album",
    "releaseDate": "2024-01-01",
    "genre": "Rock",
    "artistIDs": ["<artist_id>"]
  }' \
  -k

# CREATE SONG
curl -X POST https://localhost:8081/api/content/songs \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Song",
    "duration": 180,
    "genre": "Rock",
    "albumID": "<album_id>",
    "artistIDs": ["<artist_id>"],
    "audioFileURL": "http://example.com/song.mp3"
  }' \
  -k

# Proverite logove
cat services/content-service/logs/app.log | grep "ADMIN_ACTIVITY"
```

## 3. Testiranje Events Emitter TLS Logovanja

**Test scenarij:** TLS greška pri slanju eventa

```bash
# Kreirajte artist (što će pokrenuti event)
curl -X POST https://localhost:8081/api/content/artists \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Artist",
    "biography": "Test",
    "genres": ["Rock"]
  }' \
  -k

# Proverite logove (ako subscriptions-service nije dostupan, biće TLS greška)
cat services/content-service/logs/app.log | grep "TLS_FAILURE"
```

## 4. Testiranje Users Service TLS Logovanja

**Test scenarij:** TLS greška pri pokretanju servera

```bash
# Ako sertifikati nisu ispravni, proverite logove
cat services/users-service/logs/app.log | grep "TLS_FAILURE"
```

## 5. Kompletan Test Scenarij

Evo kompletnog test scenarija koji pokriva sve:

```bash
#!/bin/bash

echo "=== TESTIRANJE LOGOVANJA ==="

# 1. Test bez tokena
echo "1. Test bez tokena..."
curl -X GET https://localhost:8081/api/users/logout -k -s > /dev/null
echo "   Proverite: cat services/api-gateway/logs/app.log | grep ACCESS_CONTROL_FAILURE"

# 2. Test sa nevalidnim tokenom
echo "2. Test sa nevalidnim tokenom..."
curl -X GET https://localhost:8081/api/users/logout \
  -H "Authorization: Bearer invalid_token" \
  -k -s > /dev/null
echo "   Proverite: cat services/api-gateway/logs/app.log | grep INVALID_TOKEN"

# 3. Test kreiranja artist-a kao admin
echo "3. Test kreiranja artist-a..."
# Prvo se prijavite i dobijte admin token
# TOKEN="your_admin_token_here"
# curl -X POST https://localhost:8081/api/content/artists \
#   -H "Authorization: Bearer $TOKEN" \
#   -H "Content-Type: application/json" \
#   -d '{"name":"Test","biography":"Test","genres":["Rock"]}' \
#   -k -s > /dev/null
echo "   Proverite: cat services/content-service/logs/app.log | grep ADMIN_ACTIVITY"

echo "=== TESTIRANJE ZAVRŠENO ==="
```

## 6. Pregled Logova

### Pregled svih logova po tipu:

```bash
# Access control failures
grep "ACCESS_CONTROL_FAILURE" services/api-gateway/logs/app.log

# Invalid tokens
grep "INVALID_TOKEN" services/api-gateway/logs/app.log

# Expired tokens
grep "EXPIRED_TOKEN" services/api-gateway/logs/app.log

# Admin activities
grep "ADMIN_ACTIVITY" services/content-service/logs/app.log

# State changes
grep "STATE_CHANGE" services/content-service/logs/app.log

# TLS failures
grep "TLS_FAILURE" services/*/logs/app.log
```

### Pregled logova po vremenu:

```bash
# Poslednjih 50 linija
tail -50 services/api-gateway/logs/app.log

# Logovi iz poslednjih 5 minuta
find services/*/logs -name "app.log" -mmin -5 -exec tail -20 {} \;
```

## 7. Verifikacija Integriteta Logova

```bash
# Proverite checksums
cat services/*/logs/*.checksum

# Proverite rotaciju logova
ls -lh services/*/logs/
```

## Napomene

- Svi logovi se čuvaju u `services/<service-name>/logs/app.log`
- Logovi se rotiraju kada dostignu 10MB
- Zadržava se poslednjih 5 log fajlova
- Checksums se čuvaju u `.checksum` fajlovima
- Osetljivi podaci (lozinke, tokeni) se filtriraju automatski
