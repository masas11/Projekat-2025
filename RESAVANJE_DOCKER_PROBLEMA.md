# 🔧 Rešavanje Docker problema

## ❌ Greške koje se pojavljuju

### Greška 1: Docker engine nije spreman
```
unable to get image 'projekat-2025-1-users-service': 
error during connect: Get "http://%2F%2F.%2Fpipe%2FdockerDesktopLinuxEngine/...": 
open //./pipe/dockerDesktopLinuxEngine: The system cannot find the file specified.
```

### Greška 2: Ne može da pristupi Docker Hub-u
```
failed to resolve source metadata for docker.io/library/alpine:latest: 
failed to do request: Head "https://registry-1.docker.io/v2/library/alpine/manifests/latest": 
dial tcp: lookup registry-1.docker.io: no such host
```

**Uzrok:** Docker unutar kontejnera/WSL ne može da pristupi Docker Hub-u, iako Windows može.

## ✅ Rešenja

### 1. Proveri da li Docker Desktop radi

```powershell
# Proveri da li je Docker Desktop pokrenut
Get-Process -Name "Docker Desktop" -ErrorAction SilentlyContinue

# Ako nije pokrenut, pokreni Docker Desktop ručno
```

### 2. Sačekaj da se Docker engine inicijalizuje

Docker engine može da traje 30-60 sekundi da se potpuno inicijalizuje nakon pokretanja.

**Proveri status:**
```powershell
# Proveri da li Docker engine radi
docker ps

# Ako vraća grešku, sačekaj još malo i probaj ponovo
```

### 3. Restartuj Docker Desktop

```powershell
# 1. Zatvori Docker Desktop
# 2. Otvori Docker Desktop ponovo
# 3. Sačekaj da se potpuno pokrene (zeleni indikator)
# 4. Probaj ponovo
docker-compose up --build -d
```

### 4. Restartuj Docker engine

```powershell
# Restartuj Docker Desktop servis
Restart-Service -Name "com.docker.service" -ErrorAction SilentlyContinue

# Ili restartuj Docker Desktop aplikaciju
Stop-Process -Name "Docker Desktop" -Force
Start-Process "C:\Program Files\Docker\Docker\Docker Desktop.exe"
```

### 5. Proveri Docker Desktop status

U Docker Desktop aplikaciji:
- Proveri da li piše "Docker Desktop is running"
- Proveri da li je zeleni indikator aktivan
- Proveri Settings → General → da li je "Use the WSL 2 based engine" omogućeno (ako koristiš WSL)

### 6. Rešavanje problema sa Docker Hub pristupom

Ako dobijaš grešku "lookup registry-1.docker.io: no such host":

**Rešenje 1: Restartuj Docker Desktop**
```powershell
# 1. Zatvori Docker Desktop
# 2. Otvori Docker Desktop ponovo
# 3. Sačekaj da se potpuno pokrene
# 4. Probaj ponovo
docker pull alpine:latest
docker-compose up --build -d
```

**Rešenje 2: Proveri Docker Desktop DNS Settings**
1. Otvori Docker Desktop
2. Settings → Resources → Network
3. Proveri DNS settings
4. Ako koristiš VPN/proxy, možda treba da konfigurišeš

**Rešenje 3: Restartuj WSL (ako koristiš WSL 2)**
```powershell
wsl --shutdown
# Sačekaj 10 sekundi
# Restartuj Docker Desktop
```

**Rešenje 4: Proveri Docker Desktop Network Settings**
1. Docker Desktop → Settings → Resources → Network
2. Proveri da li je "Enable host networking" omogućeno
3. Restartuj Docker Desktop

### 7. Ako ništa ne pomaže

```powershell
# 1. Zatvori Docker Desktop
# 2. Očisti Docker cache (opciono)
docker system prune -a --volumes

# 3. Restartuj računar
# 4. Pokreni Docker Desktop
# 5. Sačekaj da se potpuno inicijalizuje
# 6. Probaj ponovo
docker-compose up --build -d
```

## 🔍 Debugging koraci

### Korak 1: Proveri Docker status
```powershell
docker info
docker version
```

### Korak 2: Proveri da li postoje kontejneri
```powershell
docker ps -a
```

### Korak 3: Proveri da li postoje image-i
```powershell
docker images
```

### Korak 4: Proveri Docker network
```powershell
docker network ls
```

## ⚠️ Česti problemi

1. **Docker Desktop nije potpuno pokrenut**
   - Rešenje: Sačekaj 30-60 sekundi nakon pokretanja

2. **WSL 2 nije omogućen**
   - Rešenje: Omogući WSL 2 u Docker Desktop Settings

3. **Docker engine je zamrznut**
   - Rešenje: Restartuj Docker Desktop

4. **Nedovoljno memorije**
   - Rešenje: Povećaj RAM za Docker Desktop u Settings

## ✅ Provera da li je sve spremno

```powershell
# 1. Proveri Docker
docker ps

# 2. Proveri docker-compose
docker-compose --version

# 3. Probaj da pokreneš servise
docker-compose up --build -d

# 4. Proveri status
docker-compose ps
```

## 🎯 Nakon što Docker radi

Kada Docker radi, možeš testirati:
1. **2.6 Asinhrona komunikacija** - vidi `BRZI_TEST_2.6_I_NOTIFIKACIJE.md`
2. **Notifikacije** - vidi `TEST_NOTIFIKACIJE_503.md`
