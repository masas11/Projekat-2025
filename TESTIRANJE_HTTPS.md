# Kako da Testirate da li HTTPS Radi

## ✅ Test 1: Provera u Browser-u (Network Tab)

### Koraci:

1. **Otvorite DevTools:**
   - Pritisnite `F12` ili `Ctrl+Shift+I`
   - Ili desni klik → "Inspect"

2. **Idite na Network tab:**
   - Kliknite na "Network" tab

3. **Osvežite stranicu:**
   - Pritisnite `F5` ili kliknite refresh
   - Ili otvorite: `https://localhost:8081/api/users/health`

4. **Proverite zahteve:**
   - Kliknite na zahtev ka `localhost:8081`
   - Idite na **"Headers"** tab
   - Proverite **"General"** sekciju:

**Trebalo bi da vidite:**
```
Request URL: https://localhost:8081/api/users/health
Request Method: GET
Status Code: 200 OK
```

**Ako vidite `https://` u Request URL → HTTPS radi! ✅**

## ✅ Test 2: Provera Logova

```powershell
docker logs projekat-2025-1-api-gateway-1 --tail 10 --since 2m
```

**Trebalo bi da vidite:**
- `Starting HTTPS server on port 8080` ✅
- **Nema TLS handshake grešaka** (ili mnogo manje) ✅
- Uspešne zahteve

**Ako vidite TLS greške:**
- To je normalno dok browser ne prihvati sertifikat
- Nakon prihvatanja sertifikata, greške nestaju

## ✅ Test 3: Test Frontend Aplikacije

1. **Otvorite frontend:**
   ```
   http://localhost:3000
   ```

2. **Otvorite DevTools:**
   - Pritisnite `F12`
   - Idite na **Network** tab

3. **Uradite neku akciju:**
   - Login
   - Pregled notifikacija
   - Pregled pesama
   - Bilo koja akcija koja komunicira sa API-jem

4. **Proverite zahteve:**
   - Filtrirajte: `localhost:8081`
   - Kliknite na Fetch/XHR filter
   - Kliknite na bilo koji zahtev

**Trebalo bi da vidite:**
- Request URL: `https://localhost:8081/api/...`
- Protocol: `h2` ili `http/1.1` (HTTPS)

## ✅ Test 4: Provera Environment Varijabli

```powershell
docker exec projekat-2025-1-api-gateway-1 env | Select-String "TLS|SERVICE_URL"
```

**Trebalo bi da vidite:**
```
TLS_CERT_FILE=/app/certs/server.crt
TLS_KEY_FILE=/app/certs/server.key
USERS_SERVICE_URL=https://users-service:8001
CONTENT_SERVICE_URL=https://content-service:8002
...
```

**Sve URL-ove treba da počinju sa `https://` ✅**

## ✅ Test 5: Direktan Test Endpoint-a

**U browser-u:**
1. Otvorite: `https://localhost:8081/api/users/health`
2. Trebalo bi da vidite: `users-service is running`
3. Address bar treba da pokazuje `https://` (ne `http://`)

**Ako vidite "Not Secure":**
- To je normalno za self-signed sertifikate
- HTTPS **radi** - samo browser upozorava
- Kliknite "Advanced" → "Proceed" ako već niste

## 📊 Checklist

- [ ] Network tab pokazuje `https://` u Request URL
- [ ] Logovi pokazuju "Starting HTTPS server"
- [ ] Nema (ili malo) TLS handshake grešaka
- [ ] Frontend komunicira preko HTTPS
- [ ] Environment varijable pokazuju `https://` URL-ove
- [ ] Browser može da pristupi `https://localhost:8081`

## 🎯 Zaključak

**Ako sve testove prođete:**
- ✅ HTTPS radi ispravno
- ✅ Komunikacija je šifrovana
- ✅ Sertifikati su pravilno konfigurisani

**Za odbranu:**
- Pokazati Network tab sa `https://` zahtevima
- Pokazati logove bez TLS grešaka (ili sa malo grešaka)
- Pokazati environment varijable sa `https://` URL-ovima
