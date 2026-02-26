# 🔧 Rešavanje Docker Image Greške

## Problem
```
unable to get image 'projekat-2025-1-api-gateway': unexpected end of JSON input
```

## Rešenje

### Korak 1: Pokreni Docker Desktop
- Otvori Docker Desktop aplikaciju
- Sačekaj da se potpuno pokrene (30-60 sekundi)
- Proveri da li je status "Running"

### Korak 2: Očisti oštećene image-e
```powershell
docker-compose down
docker rmi projekat-2025-1-api-gateway
docker rmi projekat-2025-1-ratings-service
```

### Korak 3: Rebuild-uj image-e
```powershell
docker-compose build --no-cache
```

### Korak 4: Pokreni servise
```powershell
docker-compose up -d
```

---

## Alternativno: Brzo rešenje

```powershell
# 1. Zaustavi sve
docker-compose down

# 2. Očisti sve image-e projekta
docker images | findstr projekat-2025-1 | ForEach-Object { docker rmi $_.Split()[2] -f }

# 3. Rebuild i pokreni
docker-compose up --build -d
```

---

## Ako Docker Desktop ne može da se pokrene

1. Restartuj Docker Desktop
2. Proveri da li ima dovoljno RAM memorije
3. Proveri Windows Services:
   ```powershell
   Get-Service | Where-Object {$_.Name -like "*docker*"}
   ```
