# SMTP Setup - Brzo Uputstvo

## 📧 Kako da konfigurišete SMTP

### Opcija 1: Korišćenje .env fajla (Preporučeno)

1. **Kreirajte `.env` fajl u root direktorijumu projekta:**
   ```env
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USERNAME=your-email@gmail.com
   SMTP_PASSWORD=your-app-password
   SMTP_FROM=your-email@gmail.com
   FRONTEND_URL=https://localhost:3000
   ```

2. **Docker Compose automatski će učitati vrednosti iz `.env` fajla**

### Opcija 2: Direktno u docker-compose.yml

1. **Otvorite `docker-compose.yml`**
2. **Pronađite `users-service` sekciju**
3. **Zamenite placeholder vrednosti sa vašim SMTP credentials:**

```yaml
environment:
  - SMTP_HOST=smtp.gmail.com
  - SMTP_PORT=587
  - SMTP_USERNAME=your-email@gmail.com
  - SMTP_PASSWORD=your-app-password
  - SMTP_FROM=your-email@gmail.com
  - FRONTEND_URL=https://localhost:3000
```

## 🔑 Gmail App Password

Za Gmail, morate kreirati App Password:

1. Idite na https://myaccount.google.com/apppasswords
2. Izaberite "Mail" i "Other (Custom name)"
3. Unesite ime (npr. "Music Streaming App")
4. Kliknite "Generate"
5. Kopirajte 16-karaktarni password (bez razmaka)
6. Koristite taj password u `SMTP_PASSWORD`

## ✅ Provera

Nakon konfiguracije, restartujte servis:

```bash
docker-compose restart users-service
```

Proverite logove:

**PowerShell:**
```powershell
docker logs projekat-2025-2-users-service-1 | Select-String "EMAIL"
```

**Ili:**
```powershell
docker logs projekat-2025-2-users-service-1 | findstr EMAIL
```

Trebalo bi da vidite:
```
[EMAIL] SMTP configured: smtp.gmail.com:587 (from: your-email@gmail.com)
```

## ⚠️ Napomena

- **Nikada ne commit-ujte `.env` fajl u git!** (Već je u .gitignore)
- Ako ne konfigurišete SMTP, aplikacija će raditi u mock mode-u
