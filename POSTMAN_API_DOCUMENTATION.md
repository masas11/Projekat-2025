# Postman API Documentation

## Base URLs

### Direktno ka servisima:
- **Users Service**: `http://localhost:8001`
- **Content Service**: `http://localhost:8002`
- **Ratings Service**: `http://localhost:8003`
- **Subscriptions Service**: `http://localhost:8004`
- **Notifications Service**: `http://localhost:8005`
- **Recommendation Service**: `http://localhost:8006`
- **Analytics Service**: `http://localhost:8007`

### Preko API Gateway-ja:
- **API Gateway**: `http://localhost:8081`

---

## 1. Health Check Endpoints

### 1.1 Users Service Health (Direktno)
```
GET http://localhost:8001/health
```

### 1.2 Users Service Health (API Gateway)
```
GET http://localhost:8081/api/users/health
```

### 1.3 Content Service Health (Direktno)
```
GET http://localhost:8002/health
```

### 1.4 Content Service Health (API Gateway)
```
GET http://localhost:8081/api/content/health
```

---

## 2. Users Service Endpoints

### 2.1 Register User

**Direktno:**
```
POST http://localhost:8001/register
```

**API Gateway:**
```
POST http://localhost:8081/api/users/register
```

**Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "firstName": "John",
  "lastName": "Doe",
  "email": "john.doe@example.com",
  "username": "johndoe",
  "password": "StrongP@ss123",
  "confirmPassword": "StrongP@ss123"
}
```

**Response (200 OK):**
```json
{
  "id": "user-id",
  "username": "johndoe",
  "email": "john.doe@example.com"
}
```

---

### 2.2 Request OTP (Login)

**Direktno:**
```
POST http://localhost:8001/login/request-otp
```

**API Gateway:**
```
POST http://localhost:8081/api/users/login/request-otp
```

**Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "username": "johndoe",
  "password": "StrongP@ss123"
}
```

**Response (200 OK):**
```
(Empty body - OTP sent to email)
```

**Note:** OTP će biti poslat na email adresu korisnika. Za testiranje, proverite konzolu ili logove servisa.

---

### 2.3 Verify OTP (Login)

**Direktno:**
```
POST http://localhost:8001/login/verify-otp
```

**API Gateway:**
```
POST http://localhost:8081/api/users/login/verify-otp
```

**Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "username": "johndoe",
  "otp": "123456"
}
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "id": "user-id",
  "username": "johndoe",
  "email": "john.doe@example.com",
  "firstName": "John",
  "lastName": "Doe",
  "role": "user"
}
```

**Note:** Sačuvajte `token` za autentifikaciju u narednim zahtevima!

---

### 2.4 Change Password

**Direktno:**
```
POST http://localhost:8001/password/change
```

**API Gateway:**
```
POST http://localhost:8081/api/users/password/change
```

**Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "username": "johndoe",
  "oldPassword": "StrongP@ss123",
  "newPassword": "NewStrongP@ss456"
}
```

**Response (200 OK):**
```
(Empty body)
```

---

### 2.5 Reset Password

**Direktno:**
```
POST http://localhost:8001/password/reset
```

**API Gateway:**
```
POST http://localhost:8081/api/users/password/reset
```

**Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "username": "johndoe",
  "newPassword": "NewStrongP@ss456"
}
```

**Response (200 OK):**
```
(Empty body)
```

---

## 3. Content Service Endpoints

### 3.1 Get All Artists

**Direktno:**
```
GET http://localhost:8002/artists
```

**API Gateway:**
```
GET http://localhost:8081/api/content/artists
```

**Headers:**
```
(none required)
```

**Response (200 OK):**
```json
[
  {
    "id": "artist-id",
    "name": "Artist Name",
    "biography": "Artist biography...",
    "genres": ["Rock", "Pop"],
    "createdAt": "2025-01-14T23:00:00Z",
    "updatedAt": "2025-01-14T23:00:00Z"
  }
]
```

---

### 3.2 Create Artist (Admin Only)

**Direktno:**
```
POST http://localhost:8002/artists
```

**API Gateway:**
```
POST http://localhost:8081/api/content/artists
```

**Headers:**
```
Content-Type: application/json
Authorization: Bearer YOUR_JWT_TOKEN
```

**Request Body:**
```json
{
  "name": "The Beatles",
  "biography": "English rock band formed in Liverpool in 1960.",
  "genres": ["Rock", "Pop", "Psychedelic Rock"]
}
```

**Response (201 Created):**
```json
{
  "id": "artist-id",
  "name": "The Beatles",
  "biography": "English rock band formed in Liverpool in 1960.",
  "genres": ["Rock", "Pop", "Psychedelic Rock"],
  "createdAt": "2025-01-14T23:00:00Z",
  "updatedAt": "2025-01-14T23:00:00Z"
}
```

**Note:** Zahteva JWT token sa `role: "admin"` u tokenu.

---

### 3.3 Get Artist by ID

**Direktno:**
```
GET http://localhost:8002/artists/{id}
```

**API Gateway:**
```
GET http://localhost:8081/api/content/artists/{id}
```

**Example:**
```
GET http://localhost:8081/api/content/artists/507f1f77bcf86cd799439011
```

**Headers:**
```
(none required)
```

**Response (200 OK):**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "name": "The Beatles",
  "biography": "English rock band...",
  "genres": ["Rock", "Pop"],
  "createdAt": "2025-01-14T23:00:00Z",
  "updatedAt": "2025-01-14T23:00:00Z"
}
```

---

### 3.4 Update Artist (Admin Only)

**Direktno:**
```
PUT http://localhost:8002/artists/{id}
```

**API Gateway:**
```
PUT http://localhost:8081/api/content/artists/{id}
```

**Example:**
```
PUT http://localhost:8081/api/content/artists/507f1f77bcf86cd799439011
```

**Headers:**
```
Content-Type: application/json
Authorization: Bearer YOUR_JWT_TOKEN
```

**Request Body:**
```json
{
  "name": "The Beatles (Updated)",
  "biography": "Updated biography...",
  "genres": ["Rock", "Pop", "Classic Rock"]
}
```

**Response (200 OK):**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "name": "The Beatles (Updated)",
  "biography": "Updated biography...",
  "genres": ["Rock", "Pop", "Classic Rock"],
  "createdAt": "2025-01-14T23:00:00Z",
  "updatedAt": "2025-01-14T23:30:00Z"
}
```

---

### 3.5 Check Song Exists (Dummy Endpoint)

**Direktno:**
```
GET http://localhost:8002/songs/exists
```

**Headers:**
```
(none required)
```

**Response (200 OK):**
```
true
```

---

## 4. Other Services (Basic Endpoints)

### 4.1 Ratings Service
```
GET http://localhost:8003/health
```

### 4.2 Subscriptions Service
```
GET http://localhost:8004/health
```

### 4.3 Notifications Service
```
GET http://localhost:8005/health
```

### 4.4 Recommendation Service
```
GET http://localhost:8006/health
```

### 4.5 Analytics Service
```
GET http://localhost:8007/health
```

---

## 5. Postman Collection Setup Tips

### 5.1 Environment Variables

Kreirajte Postman Environment sa sledećim varijablama:

```
base_url_direct: http://localhost
base_url_gateway: http://localhost:8081
jwt_token: (prazno - postavite nakon login-a)
```

### 5.2 Collection Structure

Organizujte requestove u foldere:
```
📁 Music Streaming API
  📁 Health Checks
    - Users Service Health
    - Content Service Health
  📁 Users Service
    - Register
    - Request OTP
    - Verify OTP
    - Change Password
    - Reset Password
  📁 Content Service
    - Get All Artists
    - Create Artist
    - Get Artist by ID
    - Update Artist
```

### 5.3 Automatsko čuvanje JWT Tokena

U **Verify OTP** requestu, dodajte Test script:

```javascript
if (pm.response.code === 200) {
    var jsonData = pm.response.json();
    pm.environment.set("jwt_token", jsonData.token);
    console.log("JWT token saved:", jsonData.token);
}
```

Zatim u **Create Artist** i **Update Artist** requestovima, koristite:
```
Authorization: Bearer {{jwt_token}}
```

---

## 6. Test Scenarios

### Scenario 1: Registracija i Login
1. Registrujte novog korisnika (`POST /api/users/register`)
2. Zatražite OTP (`POST /api/users/login/request-otp`)
3. Verifikujte OTP (`POST /api/users/login/verify-otp`) - sačuvajte token

### Scenario 2: Content Management (Admin)
1. Login kao admin korisnik
2. Kreirajte novog izvođača (`POST /api/content/artists`)
3. Preuzmite sve izvođače (`GET /api/content/artists`)
4. Ažurirajte izvođača (`PUT /api/content/artists/{id}`)

### Scenario 3: Public Content Access
1. Preuzmite sve izvođače bez autentifikacije (`GET /api/content/artists`)
2. Preuzmite pojedinačnog izvođača (`GET /api/content/artists/{id}`)

---

## 7. Common Error Responses

### 400 Bad Request
```json
{
  "error": "invalid JSON body"
}
```

### 401 Unauthorized
```json
{
  "error": "invalid credentials"
}
```

### 403 Forbidden
```json
{
  "error": "admin access required"
}
```

### 404 Not Found
```json
{
  "error": "artist not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "internal server error"
}
```

---

## 8. Notes

- **JWT Token**: Token se dobija nakon uspešne OTP verifikacije
- **Admin Role**: Za kreiranje/ažuriranje izvođača potreban je admin token
- **OTP**: Za testiranje, proverite logove servisa ili konzolu za OTP kod
- **MongoDB**: Content service koristi MongoDB za čuvanje podataka
- **In-Memory Store**: Users service koristi in-memory store (podaci se gube nakon restart-a)

---

## 9. Quick Test Commands (cURL)

Za brzo testiranje iz terminala:

```bash
# Health check
curl http://localhost:8081/api/users/health

# Register
curl -X POST http://localhost:8081/api/users/register \
  -H "Content-Type: application/json" \
  -d '{"firstName":"John","lastName":"Doe","email":"john@example.com","username":"johndoe","password":"StrongP@ss123","confirmPassword":"StrongP@ss123"}'

# Get Artists
curl http://localhost:8081/api/content/artists
```

---

**Napomena**: Svi endpoint-i su dostupni i direktno ka servisima i preko API Gateway-ja. Preporučeno je korišćenje API Gateway-ja za produkciju.
