# 🔧 Rešenje Problema sa Logovima

## Problem

Log direktorijum `services/users-service/logs` ne postoji na host sistemu jer se logovi čuvaju **unutar Docker kontejnera**.

## Rešenje

### Opcija 1: Pristup Logovima iz Docker Kontejnera (Preporučeno)

Logovi se čuvaju unutar kontejnera. Koristite Docker komande za pristup:

```powershell
# Proverite logove direktno iz kontejnera
docker logs projekat-2025-2-users-service-1 | Select-String "LOGIN"

# Ili proverite log fajlove unutar kontejnera
docker exec projekat-2025-2-users-service-1 ls -la /app/logs

# Kopirajte log fajl iz kontejnera na host
docker cp projekat-2025-2-users-service-1:/app/logs/app-2025-02-13.log ./logs/
```

### Opcija 2: Dodati Volume Mount za Logove

Dodajte volume mount u `docker-compose.yml` da bi logovi bili dostupni na host sistemu:

**U `docker-compose.yml`, sekcija `users-service`:**

```yaml
users-service:
  # ... postojeća konfiguracija ...
  environment:
    - LOG_DIR=/app/logs
  volumes:
    - ./certs:/app/certs:ro
    - ./logs/users-service:/app/logs  # DODATI OVO
```

**Kreiranje direktorijuma:**

```powershell
# Kreirajte direktorijum za logove
mkdir -p logs/users-service
mkdir -p logs/api-gateway
mkdir -p logs/content-service
```

**Nakon izmene, restartujte kontejnere:**

```powershell
docker-compose down
docker-compose up -d
```

---

## ✅ HTTPS Status

**HTTPS je OK!** 

Iz slika vidim:
- ✅ Sertifikat postoji (`localhost` sertifikat)
- ✅ HTTPS radi (`https://localhost:8081`)
- ⚠️ "Not secure" upozorenje je **normalno** za self-signed sertifikate u developmentu

### Objašnjenje:

**Za Development:**
- Self-signed sertifikati su prihvatljivi
- Browser prikazuje "Not secure" jer sertifikat nije potpisan od strane poznatog CA
- To je očekivano ponašanje

**Za Produkciju:**
- Trebalo bi koristiti validne SSL sertifikate od CA (npr. Let's Encrypt)
- "Not secure" upozorenje ne bi trebalo da se pojavi

### Kako Demonstrirati HTTPS na Odbrani:

1. **Pokažite sertifikat:**
   - Kliknite na "Not secure" u browser-u
   - Otvorite "Certificate Viewer"
   - Pokažite detalje sertifikata (Issued To/By, Validity, Fingerprints)

2. **Pokažite HTTPS u Developer Tools:**
   - F12 → Network tab
   - Kliknite na bilo koji API zahtev
   - Pokažite da URL počinje sa `https://`
   - Pokažite da Protocol je `h2` (HTTP/2) ili `h3`

3. **Pokažite kod:**
   - `docker-compose.yml` - TLS_CERT_FILE i TLS_KEY_FILE
   - `services/api-gateway/cmd/main.go` - ListenAndServeTLS()

---

## 📋 Ažurirani Vodič za Pristup Logovima

### Metoda 1: Docker Logs (Najbrže)

```powershell
# Svi logovi servisa
docker logs projekat-2025-2-users-service-1

# Filtrirano po tipu događaja
docker logs projekat-2025-2-users-service-1 | Select-String "LOGIN"
docker logs projekat-2025-2-users-service-1 | Select-String "VALIDATION_FAILURE"
docker logs projekat-2025-2-users-service-1 | Select-String "ACCESS_CONTROL_FAILURE"

# Poslednjih 50 linija
docker logs projekat-2025-2-users-service-1 --tail 50

# Praćenje logova u realnom vremenu
docker logs projekat-2025-2-users-service-1 -f
```

### Metoda 2: Pristup Log Fajlovima u Kontejneru

```powershell
# Lista log fajlova
docker exec projekat-2025-2-users-service-1 ls -la /app/logs

# Čitanje log fajla
docker exec projekat-2025-2-users-service-1 cat /app/logs/app-2025-02-13.log

# Kopiranje log fajla na host
docker cp projekat-2025-2-users-service-1:/app/logs/app-2025-02-13.log ./logs/
```

### Metoda 3: Volume Mount (Nakon izmene docker-compose.yml)

```powershell
# Kreirajte direktorijum
mkdir -p logs/users-service

# Restartujte kontejnere
docker-compose down
docker-compose up -d

# Sada su logovi dostupni na host sistemu
ls logs/users-service/
Get-Content logs/users-service/app-*.log | Select-String "LOGIN"
```

---

## 🎯 Za Ručno Testiranje

### Ažurirani Koraci za Test 7 (Logovanje):

**Umesto:**
```powershell
Get-Content services/users-service/logs/app-*.log | Select-String "LOGIN"
```

**Koristite:**
```powershell
# Metoda 1: Docker logs (preporučeno)
docker logs projekat-2025-2-users-service-1 | Select-String "LOGIN"

# Metoda 2: Kopiranje iz kontejnera
docker cp projekat-2025-2-users-service-1:/app/logs/app-$(Get-Date -Format 'yyyy-MM-dd').log ./temp-log.log
Get-Content ./temp-log.log | Select-String "LOGIN"
```

---

## ✅ Finalni Checklist

- [ ] HTTPS radi (`https://localhost:8081` odgovara)
- [ ] Sertifikat postoji (može se videti u browser-u)
- [ ] "Not secure" upozorenje je normalno za development
- [ ] Logovi su dostupni preko `docker logs` komande
- [ ] Volume mount je dodat u `docker-compose.yml` (opciono)

---

**Napomena:** Za odbranu, možete koristiti `docker logs` komandu za prikaz logova - to je potpuno validno i pokazuje da znate kako pristupiti logovima u Docker okruženju.
