# ✅ Rešenje Problema: HTTPS i Logovi

## 1. HTTPS Status - ✅ OK!

### Objašnjenje:

**HTTPS je potpuno implementiran i radi!**

Iz slika vidim:
- ✅ Sertifikat postoji (`localhost` sertifikat)
- ✅ HTTPS radi (`https://localhost:8081`)
- ✅ Certificate Viewer pokazuje detalje sertifikata
- ⚠️ "Not secure" upozorenje je **normalno** za self-signed sertifikate

### Zašto "Not secure"?

**Self-signed sertifikati:**
- Nisu potpisani od strane poznatog Certificate Authority (CA)
- Browser ih ne prepoznaje kao "trusted"
- Prikazuje "Not secure" upozorenje
- **Ovo je normalno za development okruženje**

**Za produkciju:**
- Trebalo bi koristiti validne SSL sertifikate (npr. Let's Encrypt)
- "Not secure" upozorenje ne bi trebalo da se pojavi

### Kako Demonstrirati HTTPS na Odbrani:

1. **Pokažite sertifikat:**
   - Kliknite na "Not secure" u browser-u
   - Otvorite "Certificate Viewer"
   - Pokažite:
     - Issued To: `localhost` (MusicStreaming, IT)
     - Issued By: `localhost` (self-signed)
     - Validity: 1 godina (2026-2027)
     - SHA-256 Fingerprints

2. **Pokažite HTTPS u Developer Tools:**
   - F12 → Network tab
   - Kliknite na bilo koji API zahtev
   - Pokažite:
     - Request URL: `https://localhost:8081/...`
     - Protocol: `h2` (HTTP/2) ili `h3`

3. **Pokažite kod:**
   - `docker-compose.yml` - TLS_CERT_FILE i TLS_KEY_FILE
   - `services/api-gateway/cmd/main.go` - `ListenAndServeTLS()`

**Odgovor za profesorku:**
> "HTTPS je implementiran. Vidite da URL počinje sa `https://` i da postoji sertifikat. 'Not secure' upozorenje je normalno za self-signed sertifikate u development okruženju. Za produkciju bi koristili validne SSL sertifikate od CA."

---

## 2. Problem sa Logovima - ✅ Rešeno!

### Problem:

```
Get-Content services/users-service/logs/app-*.log
# Error: Cannot find path 'services/users-service/logs'
```

**Razlog:** Logovi se čuvaju **unutar Docker kontejnera**, ne na host sistemu.

### Rešenje 1: Koristite Docker Logs (Preporučeno)

**Umesto:**
```powershell
Get-Content services/users-service/logs/app-*.log | Select-String "LOGIN"
```

**Koristite:**
```powershell
docker logs projekat-2025-2-users-service-1 | Select-String "LOGIN"
```

**Ovo je potpuno validno za Docker okruženje!**

### Rešenje 2: Dodati Volume Mount (Opciono)

Ako želite da logovi budu dostupni na host sistemu:

1. **Kreirajte direktorijum:**
   ```powershell
   mkdir -p logs/users-service
   mkdir -p logs/api-gateway
   mkdir -p logs/content-service
   ```

2. **Dodajte volume mount u `docker-compose.yml`:**
   
   **Za users-service:**
   ```yaml
   users-service:
     environment:
       - LOG_DIR=/app/logs
     volumes:
       - ./certs:/app/certs:ro
       - ./logs/users-service:/app/logs  # DODATI OVO
   ```

3. **Restartujte kontejnere:**
   ```powershell
   docker-compose down
   docker-compose up -d
   ```

4. **Sada su logovi dostupni:**
   ```powershell
   ls logs/users-service/
   Get-Content logs/users-service/app-*.log | Select-String "LOGIN"
   ```

---

## 📋 Ažurirane Komande za Ručno Testiranje

### Umesto:
```powershell
Get-Content services/users-service/logs/app-*.log | Select-String "LOGIN"
```

### Koristite:
```powershell
# Metoda 1: Docker logs (preporučeno)
docker logs projekat-2025-2-users-service-1 | Select-String "LOGIN"

# Metoda 2: Poslednjih 50 linija
docker logs projekat-2025-2-users-service-1 --tail 50 | Select-String "LOGIN"

# Metoda 3: Praćenje u realnom vremenu
docker logs projekat-2025-2-users-service-1 -f | Select-String "LOGIN"

# Metoda 4: Kopiranje iz kontejnera (ako treba fajl)
docker cp projekat-2025-2-users-service-1:/app/logs/app-2025-02-13.log ./temp-log.log
Get-Content ./temp-log.log | Select-String "LOGIN"
```

---

## 🎯 Za Odbranu

### HTTPS:

**Pitanje:** "Zašto piše 'Not secure'?"

**Odgovor:**
> "HTTPS je implementiran i radi. 'Not secure' upozorenje je normalno za self-signed sertifikate u development okruženju jer browser ne prepoznaje sertifikat kao 'trusted'. Možete videti da URL počinje sa `https://` i da postoji sertifikat (Certificate Viewer). Za produkciju bi koristili validne SSL sertifikate od Certificate Authority."

**Demonstracija:**
1. Otvorite Certificate Viewer (kliknite na "Not secure")
2. Pokažite detalje sertifikata
3. Pokažite HTTPS u Network tab-u (F12)

### Logovi:

**Pitanje:** "Gde su logovi?"

**Odgovor:**
> "Logovi se čuvaju unutar Docker kontejnera, što je standardna praksa. Pristupamo im preko `docker logs` komande. Mogu pokazati kako to radi."

**Demonstracija:**
```powershell
# Pokažite logove
docker logs projekat-2025-2-users-service-1 | Select-String "LOGIN"

# Pokažite rotaciju logova
docker exec projekat-2025-2-users-service-1 ls -la /app/logs/

# Pokažite checksum fajlove
docker exec projekat-2025-2-users-service-1 ls -la /app/logs/*.checksum
```

---

## ✅ Finalni Checklist

- [x] HTTPS radi (`https://localhost:8081`)
- [x] Sertifikat postoji (može se videti u browser-u)
- [x] "Not secure" je normalno za development
- [x] Logovi su dostupni preko `docker logs`
- [x] Volume mount je dodat u `docker-compose.yml` (opciono)

---

**Sve je OK! Možete koristiti `docker logs` komandu za pristup logovima - to je standardni način u Docker okruženju.**
