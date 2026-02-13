# SSL Sertifikati - Objašnjenje Logike

## 🔐 Šta su SSL/TLS Sertifikati?

**SSL/TLS sertifikati** su digitalni dokumenti koji:
- **Šifruju komunikaciju** između klijenta i servera
- **Potvrđuju identitet** servera (da je stvarno taj server)
- **Sprečavaju "man-in-the-middle" napade** (presretanje podataka)

## 🏗️ Arhitektura Sistema

### Portovi i Servisi:

```
Frontend (React)          →  localhost:3000  (HTTP)
    ↓
API Gateway              →  localhost:8081  (HTTPS) ← SSL sertifikat ovde
    ↓
Backend Servisi          →  localhost:8001-8007 (HTTPS) ← SSL sertifikati ovde
```

### Zašto Odvojeni Portovi?

1. **Frontend (3000)** - React dev server
   - Koristi HTTP (nema sertifikat)
   - Komunicira sa API Gateway-em preko HTTPS

2. **API Gateway (8081)** - Jedinstveni ulazni punkt
   - **HTTPS sa sertifikatom** - šifruje komunikaciju sa frontend-om
   - Prima zahteve od klijenta
   - Prosleđuje zahteve backend servisima (takođe HTTPS)

3. **Backend Servisi (8001-8007)** - Interna komunikacija
   - **HTTPS sa sertifikatima** - šifruje inter-service komunikaciju
   - Ne pristupa se direktno iz browser-a

## 🔄 Kako Radi HTTPS Komunikacija?

### 1. Frontend → API Gateway
```
Browser (localhost:3000)
    ↓ HTTPS zahtev
    ↓ (šifrovan sa sertifikatom)
API Gateway (localhost:8081)
```

**Zašto sertifikat ovde?**
- Šifruje podatke (email, password, token) između browser-a i API Gateway-a
- Sprečava presretanje podataka na mreži

### 2. API Gateway → Backend Servisi
```
API Gateway
    ↓ HTTPS zahtev
    ↓ (šifrovan sa sertifikatom)
Users Service (localhost:8001)
```

**Zašto sertifikati ovde?**
- Šifruje podatke između servisa
- Zaštita od presretanja u Docker mreži

## ⚠️ Zašto Browser Kaže "Not Secure"?

**Self-Signed Sertifikati:**
- Sertifikati su **kreirani lokalno** (ne od Certificate Authority)
- Browser **ne veruje** self-signed sertifikate
- **Normalno za development** - za production koristiti validne sertifikate

**Šta to znači?**
- HTTPS **radi** (komunikacija je šifrovana)
- Browser **upozorava** jer sertifikat nije od poznatog CA
- **Sigurno je za development** - samo prihvatite sertifikat

## 📋 Šta Sertifikat Sadrži?

1. **Public Key** - za šifrovanje podataka
2. **Informacije o serveru** - CN=localhost
3. **Potpis** - potvrda identiteta (self-signed = potpisali smo sami)

## 🎯 Zašto Ovakva Arhitektura?

### API Gateway kao Jedinstveni Ulaz:
- **Jedan sertifikat** za sve zahteve od klijenta
- **Centralizovana autentifikacija** i autorizacija
- **Rate limiting** na jednom mestu
- **CORS** konfiguracija na jednom mestu

### Backend Servisi Interno:
- **HTTPS između servisa** - zaštita podataka u Docker mreži
- **Svaki servis ima sertifikat** - za inter-service komunikaciju

## 🔍 Primer Komunikacije:

```
1. Korisnik unosi email/password u frontend (localhost:3000)
   ↓
2. Frontend šalje HTTPS zahtev API Gateway-u (localhost:8081)
   - Podaci su ŠIFROVANI sertifikatom
   ↓
3. API Gateway proverava autentifikaciju
   ↓
4. API Gateway šalje HTTPS zahtev Users Service-u (localhost:8001)
   - Podaci su ŠIFROVANI sertifikatom
   ↓
5. Users Service proverava password (heširan bcrypt-om)
   ↓
6. Odgovor se vraća nazad (takođe šifrovan)
```

## ✅ Zaključak

**Sertifikati služe za:**
- ✅ Šifrovanje komunikacije (HTTPS)
- ✅ Zaštitu senzitivnih podataka (password, token)
- ✅ Sprečavanje presretanja podataka

**Odvojeni portovi jer:**
- Frontend (3000) - React aplikacija
- API Gateway (8081) - Jedinstveni ulaz sa HTTPS
- Backend (8001-8007) - Interni servisi sa HTTPS

**"Not Secure" je normalno:**
- Self-signed sertifikati za development
- HTTPS radi, samo browser upozorava
- Za production koristiti validne sertifikate od CA
