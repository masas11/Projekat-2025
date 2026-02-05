# 🚀 Brzi Vodič za Pokretanje Sistema

## Korak 1: Pokrenite Sve Servise

```powershell
docker-compose up -d
```

Ovo će pokrenuti:
- MongoDB kontejnere (users, content, ratings)
- Cassandra kontejner
- Sve mikroservise
- API Gateway

## Korak 2: Sačekajte da se Servisi Pokrenu

```powershell
# Sačekajte 15-20 sekundi da se svi servisi potpuno pokrenu
Start-Sleep -Seconds 20

# Proverite status
docker-compose ps
```

## Korak 3: Seed-ujte Podatke

```powershell
.\scripts\seed-all.ps1
```

Ovo će dodati:
- 5 umetnika (Michael Jackson, The Beatles, Lady Gaga, The Weeknd, Ed Sheeran)
- 5 albuma
- 6 pesama
- 4 ocene
- Admin korisnika (username: `admin`, password: `admin123`)

## Korak 4: Proverite da li Sve Radi

```powershell
# Proverite API Gateway
Invoke-WebRequest -Uri http://localhost:8081/api/content/artists -UseBasicParsing

# Proverite albume
Invoke-WebRequest -Uri http://localhost:8081/api/content/albums -UseBasicParsing

# Proverite pesme
Invoke-WebRequest -Uri http://localhost:8081/api/content/songs/by-album?albumId=album1 -UseBasicParsing
```

## Korak 5: Pokrenite Frontend (opciono)

```powershell
cd frontend
npm install
npm start
```

Frontend će biti dostupan na `http://localhost:3000`

---

## 🛑 Zaustavljanje Sistema

```powershell
docker-compose down
```

**Napomena:** Podaci će ostati sačuvani u `data/` folderu!

---

## 🔄 Restart Sistema

```powershell
docker-compose restart
```

---

## 📊 Pregled Servisa

- **API Gateway**: http://localhost:8081
- **Users Service**: http://localhost:8001
- **Content Service**: http://localhost:8002
- **Ratings Service**: http://localhost:8003
- **Subscriptions Service**: http://localhost:8004
- **Notifications Service**: http://localhost:8005
- **Recommendation Service**: http://localhost:8006
- **Analytics Service**: http://localhost:8007

---

## 🆘 Troubleshooting

### Problem: Servisi se ne pokreću
```powershell
# Proverite logove
docker-compose logs

# Restart-ujte sve
docker-compose down
docker-compose up -d
```

### Problem: MongoDB ne može da se poveže
```powershell
# Sačekajte malo duže
Start-Sleep -Seconds 30
docker-compose restart
```

### Problem: Port je zauzet
```powershell
# Proverite koji proces koristi port
netstat -ano | findstr :8081

# Zaustavite proces ili promenite port u docker-compose.yml
```
