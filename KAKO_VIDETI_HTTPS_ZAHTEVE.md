# Kako da Vidite HTTPS Zahteve u Browser-u

## 🔍 Objašnjenje

### Sertifikat NIJE na `/api/users/health`

**Sertifikat je na API Gateway-u:**
- ✅ Sertifikat je na **`localhost:8081`** (API Gateway)
- ✅ `/api/users/health` je samo **endpoint (putanja)**
- ✅ **Svi zahtevi** ka `https://localhost:8081/*` koriste **isti sertifikat**

**Primer:**
- `https://localhost:8081/api/users/health` → koristi sertifikat
- `https://localhost:8081/api/content/health` → koristi isti sertifikat
- `https://localhost:8081/api/users/register` → koristi isti sertifikat

### Zašto Već Vidite Odgovor?

**Mogući razlozi:**
1. Browser je **već prihvatio sertifikat** (možda ranije)
2. Browser **automatski prihvata** self-signed sertifikate za `localhost`
3. Sertifikat je **već dodat** u browser trust store

## 📊 Kako da Vidite HTTPS Zahteve

### Metoda 1: Network Tab u DevTools

1. **Otvorite DevTools:**
   - Pritisnite `F12` ili `Ctrl+Shift+I`
   - Ili desni klik → "Inspect"

2. **Idite na Network tab:**
   - Kliknite na "Network" tab

3. **Filtrirajte zahteve:**
   - U **filter polju** (gore desno) unesite: `localhost:8081`
   - Ili kliknite na **"Fetch/XHR"** filter (samo API zahtevi)

4. **Osvežite stranicu:**
   - Pritisnite `F5` ili kliknite refresh
   - Trebalo bi da vidite zahteve ka `https://localhost:8081`

5. **Proverite detalje:**
   - Kliknite na zahtev
   - Proverite **"Headers"** tab
   - Trebalo bi da vidite:
     - **Request URL:** `https://localhost:8081/api/...`
     - **Protocol:** `h2` ili `http/1.1` (HTTPS)

### Metoda 2: Console Tab

1. **Otvorite Console tab:**
   - U DevTools, kliknite na "Console"

2. **Proverite zahteve:**
   - Ako ima greške, videćete ih ovde
   - HTTPS greške će biti vidljive

### Metoda 3: Provera u Browser-u

1. **Kliknite na "Not Secure" ili "🔒" ikonu** u address bar-u
2. **Kliknite na "Certificate"**
3. Trebalo bi da vidite sertifikat informacije

## ✅ Šta Treba da Vidite

### U Network Tab-u:

**Zahtev ka API Gateway-u:**
```
Name: /api/users/health (ili drugi endpoint)
Status: 200
Type: fetch
Protocol: h2 (HTTP/2 over HTTPS) ili http/1.1
Request URL: https://localhost:8081/api/users/health
```

**Headers:**
```
Request URL: https://localhost:8081/api/users/health
Request Method: GET
Status Code: 200 OK
Protocol: h2 (HTTPS)
```

### Ako Ne Vidite HTTPS:

1. **Proverite filter:**
   - Uklonite filter `/revi` ili bilo koji drugi filter
   - Kliknite "All" filter

2. **Osvežite stranicu:**
   - Pritisnite `F5`
   - Ili `Ctrl+R`

3. **Proverite da li frontend koristi HTTPS:**
   - Otvorite `frontend/src/services/api.js`
   - Trebalo bi da vidite: `https://localhost:8081`

## 🔍 Primer Testiranja

1. **Otvorite frontend:** `http://localhost:3000`
2. **Otvorite DevTools:** `F12`
3. **Idite na Network tab**
4. **Filtrirajte:** `localhost:8081`
5. **Uradite neku akciju** (npr. login, pregled notifikacija)
6. **Trebalo bi da vidite HTTPS zahteve** ka `https://localhost:8081`

## ⚠️ Ako Ne Vidite HTTPS Zahteve

**Proverite:**
1. Da li je frontend restartovan nakon promene na HTTPS?
2. Da li `API_BASE_URL` u `frontend/src/services/api.js` je `https://localhost:8081`?
3. Da li je filter u Network tab-u pravilno postavljen?
