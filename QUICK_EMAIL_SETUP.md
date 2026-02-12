# Brza Konfiguracija Email-a

## 🚀 Brzo Pokretanje (Gmail)

1. **Kreirajte Gmail App Password:**
   - Idite na https://myaccount.google.com/apppasswords
   - Generišite App Password za "Mail"
   - Kopirajte password (16 karaktera, bez razmaka)

2. **Dodajte u `docker-compose.yml` u `users-service` sekciju:**
   ```yaml
   environment:
     - SMTP_HOST=smtp.gmail.com
     - SMTP_PORT=587
     - SMTP_USERNAME=your-email@gmail.com
     - SMTP_PASSWORD=xxxx xxxx xxxx xxxx
     - SMTP_FROM=your-email@gmail.com
     - FRONTEND_URL=https://localhost:3000
   ```

3. **Restartujte servis:**
   ```bash
   docker-compose restart users-service
   ```

4. **Proverite logove:**
   ```bash
   docker logs projekat-2025-2-users-service-1 | grep EMAIL
   ```
   
   Trebalo bi da vidite:
   ```
   [EMAIL] SMTP configured: smtp.gmail.com:587 (from: your-email@gmail.com)
   ```

## ✅ Test

1. Registrujte novog korisnika
2. Proverite inbox email adrese
3. Trebalo bi da primite verification email

## 📚 Detaljna Dokumentacija

Pogledajte `EMAIL_KONFIGURACIJA.md` za detaljne uputstva i troubleshooting.
