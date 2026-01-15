# Postman Setup Instructions

## Quick Start

### 1. Import Collection
1. Otvorite Postman
2. Kliknite na **Import** (gore levo)
3. Drag & drop fajl `Music-Streaming-API.postman_collection.json`
4. Ili kliknite **Upload Files** i izaberite fajl

### 2. Import Environment
1. U Postman-u, kliknite na **Environments** (levo sidebar)
2. Kliknite **Import**
3. Drag & drop fajl `Music-Streaming-API.postman_environment.json`
4. Selektujte environment **"Music Streaming API - Local"** (gore desno)

### 3. Testiranje

#### Korak 1: Health Checks
1. Pokrenite request **"Users Service Health (Gateway)"**
2. Očekivani odgovor: `200 OK` sa porukom `"users-service is running"`
3. Pokrenite request **"Content Service Health (Gateway)"**
4. Očekivani odgovor: `200 OK` sa porukom `"content-service is running"`

#### Korak 2: Registracija korisnika
1. Otvorite **"Users Service" > "Register User"**
2. Request body je već popunjen sa primerom
3. Kliknite **Send**
4. Očekivani odgovor: `200 OK` sa user ID-jem

#### Korak 3: Login (OTP)
1. Otvorite **"Users Service" > "Request OTP (Login)"**
2. Unesite username i password (isti kao u registraciji)
3. Kliknite **Send**
4. Očekivani odgovor: `200 OK` (OTP je poslat na email)

**Napomena**: Za testiranje, OTP kod možete pronaći u logovima servisa. Proverite konzolu gde je pokrenut `users-service`.

5. Otvorite **"Users Service" > "Verify OTP (Login)"**
6. Unesite username i OTP kod
7. Kliknite **Send**
8. **VAŽNO**: JWT token će se automatski sačuvati u environment varijablu `jwt_token`

#### Korak 4: Content Service (Public)
1. Otvorite **"Content Service" > "Get All Artists"**
2. Kliknite **Send**
3. Očekivani odgovor: `200 OK` sa listom izvođača (može biti prazna lista `[]`)

#### Korak 5: Content Service (Admin - zahteva JWT)
1. Otvorite **"Content Service" > "Create Artist (Admin)"**
2. Request već ima `Authorization: Bearer {{jwt_token}}` header
3. Request body je već popunjen sa primerom
4. Kliknite **Send**
5. Očekivani odgovor: `201 Created` sa podacima novog izvođača

**Napomena**: Ako dobijete `403 Forbidden`, proverite da li je korisnik sa `role: "admin"` u JWT tokenu.

---

## Environment Variables

Nakon import-a, environment sadrži sledeće varijable:

| Varijabla | Opis | Primer |
|-----------|------|--------|
| `base_url_gateway` | API Gateway URL | `http://localhost:8081` |
| `jwt_token` | JWT token (automatski se postavlja nakon login-a) | `eyJhbGciOiJIUzI1NiIs...` |
| `user_id` | ID ulogovanog korisnika | `user-123` |
| `username` | Username ulogovanog korisnika | `johndoe` |
| `artist_id` | ID izvođača (možete ručno postaviti) | `507f1f77bcf86cd799439011` |

---

## Troubleshooting

### Problem: "Connection refused" ili "ECONNREFUSED"
**Rešenje**: Proverite da li su svi Docker kontejneri pokrenuti:
```powershell
docker ps
```

### Problem: "401 Unauthorized" na Create Artist
**Rešenje**: 
1. Proverite da li ste prvo izvršili "Verify OTP" request
2. Proverite da li je `jwt_token` postavljen u environment varijablama
3. Proverite da li je korisnik admin (proverite JWT token na jwt.io)

### Problem: "403 Forbidden" na Create Artist
**Rešenje**: Korisnik mora imati `role: "admin"` u JWT tokenu. Proverite registraciju korisnika.

### Problem: OTP kod ne radi
**Rešenje**: 
- Za testiranje, proverite logove `users-service` kontejnera:
```powershell
docker logs projekat-2025-1-users-service-1
```
- OTP se šalje na email, ali u development modu možete videti u logovima

### Problem: "404 Not Found" na API Gateway endpoint-ima
**Rešenje**: 
1. Proverite da li je API Gateway pokrenut: `docker ps | findstr api-gateway`
2. Proverite da li koristite tačan URL: `http://localhost:8081/api/...`

---

## Test Scenarios

### Scenario 1: Kompletan Flow
1. ✅ Health check (Users Service)
2. ✅ Health check (Content Service)
3. ✅ Register User
4. ✅ Request OTP
5. ✅ Verify OTP (sačuva token)
6. ✅ Get All Artists (public)
7. ✅ Create Artist (sa JWT tokenom)
8. ✅ Get Artist by ID
9. ✅ Update Artist (sa JWT tokenom)

### Scenario 2: Error Handling
1. ✅ Register sa nevažećim podacima (prazan email)
2. ✅ Login sa pogrešnim passwordom
3. ✅ Verify OTP sa pogrešnim kodom
4. ✅ Create Artist bez JWT tokena
5. ✅ Get Artist sa nepostojećim ID-jem

---

## Tips & Tricks

### 1. Automatsko čuvanje JWT Tokena
Request **"Verify OTP (Login)"** automatski čuva JWT token u environment varijablu. Ne morate ručno da kopirate token!

### 2. Korišćenje Varijabli
U request URL-ovima možete koristiti:
- `{{base_url_gateway}}/api/users/health`
- `{{jwt_token}}` u Authorization header-u

### 3. Pre-request Scripts
Možete dodati pre-request script da automatski proveri da li postoji JWT token pre slanja zahteva koji ga zahteva.

### 4. Test Scripts
Dodajte test scripts da automatski proverite status kod i response body:
```javascript
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});
```

---

## Collection Structure

```
📁 Music Streaming API
  📁 Health Checks
    - Users Service Health (Direct)
    - Users Service Health (Gateway)
    - Content Service Health (Direct)
    - Content Service Health (Gateway)
  📁 Users Service
    - Register User
    - Request OTP (Login)
    - Verify OTP (Login) ⭐ (čuva JWT token)
    - Change Password
    - Reset Password
  📁 Content Service
    - Get All Artists
    - Create Artist (Admin) 🔒 (zahteva JWT)
    - Get Artist by ID
    - Update Artist (Admin) 🔒 (zahteva JWT)
    - Check Song Exists
  📁 Other Services
    - Ratings Service Health
    - Subscriptions Service Health
    - Notifications Service Health
    - Recommendation Service Health
    - Analytics Service Health
```

---

## Next Steps

1. ✅ Import Collection
2. ✅ Import Environment
3. ✅ Test Health Checks
4. ✅ Test User Registration & Login
5. ✅ Test Content Service Endpoints
6. 🎉 Svi endpoint-i su spremni za testiranje!

Za detaljnu dokumentaciju svih endpoint-a, pogledajte `POSTMAN_API_DOCUMENTATION.md`.
