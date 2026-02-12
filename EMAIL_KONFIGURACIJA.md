# Konfiguracija Email Servisa

## 📧 Pregled

Aplikacija sada podržava slanje pravih email poruka preko SMTP protokola. Ako SMTP nije konfigurisan, aplikacija automatski koristi mock mode (loguje email poruke umesto slanja).

## 🔧 Konfiguracija

### Environment Varijable

Dodajte sledeće environment varijable u `docker-compose.yml` ili `.env` fajl:

```yaml
SMTP_HOST=smtp.gmail.com          # SMTP server adresa
SMTP_PORT=587                     # SMTP port (587 za TLS, 465 za SSL)
SMTP_USERNAME=your-email@gmail.com # Vaša email adresa
SMTP_PASSWORD=your-app-password   # App Password (ne obična lozinka!)
SMTP_FROM=your-email@gmail.com   # From adresa (može biti ista kao username)
FRONTEND_URL=https://localhost:3000 # Frontend URL za linkove u email-ovima
```

### Primer za Gmail

1. **Kreirajte App Password:**
   - Idite na [Google Account Settings](https://myaccount.google.com/)
   - Security → 2-Step Verification → App passwords
   - Generišite App Password za "Mail"
   - Kopirajte generisani password (16 karaktera)

2. **Dodajte u docker-compose.yml:**
   ```yaml
   environment:
     - SMTP_HOST=smtp.gmail.com
     - SMTP_PORT=587
     - SMTP_USERNAME=your-email@gmail.com
     - SMTP_PASSWORD=xxxx xxxx xxxx xxxx  # App Password bez razmaka
     - SMTP_FROM=your-email@gmail.com
     - FRONTEND_URL=https://localhost:3000
   ```

3. **Ili kreirajte `.env` fajl:**
   ```env
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USERNAME=your-email@gmail.com
   SMTP_PASSWORD=your-app-password
   SMTP_FROM=your-email@gmail.com
   FRONTEND_URL=https://localhost:3000
   ```

   I u `docker-compose.yml`:
   ```yaml
   environment:
     - SMTP_HOST=${SMTP_HOST}
     - SMTP_PORT=${SMTP_PORT}
     - SMTP_USERNAME=${SMTP_USERNAME}
     - SMTP_PASSWORD=${SMTP_PASSWORD}
     - SMTP_FROM=${SMTP_FROM}
     - FRONTEND_URL=${FRONTEND_URL}
   ```

### Primer za Outlook/Hotmail

```yaml
environment:
  - SMTP_HOST=smtp-mail.outlook.com
  - SMTP_PORT=587
  - SMTP_USERNAME=your-email@outlook.com
  - SMTP_PASSWORD=your-password
  - SMTP_FROM=your-email@outlook.com
```

### Primer za Yahoo

```yaml
environment:
  - SMTP_HOST=smtp.mail.yahoo.com
  - SMTP_PORT=587
  - SMTP_USERNAME=your-email@yahoo.com
  - SMTP_PASSWORD=your-app-password
  - SMTP_FROM=your-email@yahoo.com
```

## 📝 Tipovi Email Poruka

Aplikacija šalje sledeće tipove email poruka:

1. **OTP Email** - Jednokratna lozinka za prijavu
2. **Verification Email** - Potvrda registracije
3. **Magic Link Email** - Link za povraćaj naloga
4. **Password Reset Email** - Link za reset lozinke

## 🎨 Email Template-i

Svi email-ovi koriste HTML template-e sa:
- Profesionalnim dizajnom
- Responsive layout-om
- Jasnim pozivima na akciju (CTA buttons)
- Informacijama o isteku linkova/tokena

## 🔍 Provera Funkcionalnosti

### 1. Provera da li je SMTP konfigurisan

Kada pokrenete aplikaciju, proverite logove:

```
[EMAIL] SMTP configured: smtp.gmail.com:587 (from: your-email@gmail.com)
```

Ako vidite:
```
[EMAIL] SMTP not configured - using mock mode
```

To znači da SMTP nije konfigurisan i aplikacija koristi mock mode.

### 2. Test Slanja Email-a

1. Registrujte novog korisnika
2. Proverite inbox email adrese koju ste koristili
3. Trebalo bi da primite verification email

### 3. Provera Logova

**PowerShell:**
```powershell
docker logs projekat-2025-2-users-service-1 | Select-String "EMAIL"
```

**Ili:**
```powershell
docker logs projekat-2025-2-users-service-1 | findstr EMAIL
```

U Docker logovima možete videti:
```
[EMAIL] Sent successfully to user@example.com: Verify Your Email Address
```

Ili u slučaju greške:
```
[EMAIL ERROR] Failed to send verification email to user@example.com: ...
```

## ⚠️ Bezbednosne Napomene

1. **Nikada ne commit-ujte SMTP credentials u git!**
   - Koristite `.env` fajl i dodajte ga u `.gitignore`
   - Ili koristite Docker secrets za produkciju

2. **Koristite App Passwords:**
   - Za Gmail, Yahoo i slične servise, koristite App Passwords umesto običnih lozinki
   - Ovo je bezbednije i omogućava revokaciju bez menjanja glavne lozinke

3. **HTTPS za Frontend URL:**
   - Uvek koristite HTTPS za `FRONTEND_URL` u produkciji
   - Linkovi u email-ovima će biti sigurni

## 🐛 Troubleshooting

### Problem: Email-ovi se ne šalju

**Rešenje:**
1. Proverite da li su SMTP credentials ispravni
2. Proverite da li je 2-Step Verification omogućen (za Gmail)
3. Proverite da li koristite App Password (ne običnu lozinku)
4. Proverite firewall i da li port 587/465 nije blokiran

### Problem: "Authentication failed"

**Rešenje:**
- Za Gmail: Proverite da li koristite App Password
- Proverite da li je "Less secure app access" omogućen (starije Gmail naloge)
- Proverite da li su username i password ispravni

### Problem: "Connection timeout"

**Rešenje:**
- Proverite da li je SMTP_HOST ispravan
- Proverite da li je SMTP_PORT ispravan (587 za TLS, 465 za SSL)
- Proverite firewall i mrežne postavke

## 📚 Dodatni Resursi

- [Gmail App Passwords](https://support.google.com/accounts/answer/185833)
- [Outlook SMTP Settings](https://support.microsoft.com/en-us/office/pop-imap-and-smtp-settings-8361e398-8af4-4e97-b147-6c6c4ac95353)
- [Yahoo SMTP Settings](https://help.yahoo.com/kb/SLN4725.html)

---

**Napomena:** Ako ne konfigurišete SMTP, aplikacija će i dalje raditi u mock mode-u (loguje email poruke umesto slanja).
