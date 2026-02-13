# Kako da Testirate HTTPS - Kratak Vodič

## ✅ Brzi Testovi

### 1. Provera da li API Gateway koristi HTTPS

```powershell
docker logs projekat-2025-1-api-gateway-1 --tail 3
```

**Očekivano:** `Starting HTTPS server on port 8080`

### 2. Test u Browser-u

1. Otvorite: `https://localhost:8081/api/users/health`
2. Browser će pokazati "Not Secure" (normalno za self-signed sertifikate)
3. Kliknite **"Advanced"** → **"Proceed to localhost"**
4. Trebalo bi da vidite: `users-service is running`

### 3. Test Frontend-a

```powershell
cd frontend
npm start
```

1. Otvorite: `http://localhost:3000`
2. Frontend će komunicirati sa API Gateway preko HTTPS
3. Proverite Network tab u browser DevTools (F12)
4. Trebalo bi da vidite HTTPS zahteve ka `https://localhost:8081`

### 4. Provera HTTPS između Servisa

```powershell
docker exec projekat-2025-1-api-gateway-1 env | Select-String SERVICE_URL
```

**Očekivano:** Sve URL-ove počinju sa `https://`

## 🔍 Šta Proveriti

- ✅ API Gateway log pokazuje "Starting HTTPS server"
- ✅ Browser može da pristupi `https://localhost:8081`
- ✅ Frontend komunicira preko HTTPS (proverite Network tab)
- ✅ Svi servisi koriste HTTPS za inter-service komunikaciju

## ⚠️ Ako Ne Radi

1. **Proverite sertifikate:**
   ```powershell
   ls certs\
   ```
   Trebalo bi da vidite: `server.crt` i `server.key`

2. **Proverite docker-compose.yml:**
   - API Gateway treba da ima `TLS_CERT_FILE` i `TLS_KEY_FILE`
   - Volume za certs treba da postoji

3. **Restartujte API Gateway:**
   ```powershell
   docker-compose restart api-gateway
   ```
