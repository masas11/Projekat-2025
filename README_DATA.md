# 🗄️ Brzi Vodič za Rad sa Bazama Podataka

## 🚀 Brzi Start

### 1. Pokrenite Servise
```bash
docker-compose up -d
```

### 2. Seed-ujte Podatke
**Windows:**
```powershell
.\scripts\seed-all.ps1
```

**Linux/Mac:**
```bash
chmod +x scripts/seed-all.sh
./scripts/seed-all.sh
```

### 3. Proverite Podatke
```bash
# Pregled umetnika
curl http://localhost:8081/api/content/artists

# Pregled albuma
curl http://localhost:8081/api/content/albums
```

## 📝 Dodavanje Novih Podataka

### Kroz API (PREPORUČENO)
```bash
# Prijavite se kao admin (username: admin, password: admin123)
# Zatim dodajte umetnika:
curl -X POST http://localhost:8081/api/content/artists \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"name": "Novi Umetnik", "biography": "...", "genres": ["Pop"]}'
```

### Kroz Seed Skripte
1. Otvorite `scripts/seed-content.js`
2. Dodajte nove podatke
3. Pokrenite: `.\scripts\seed-all.ps1`

## 💾 Perzistencija Podataka

✅ **Podaci se automatski čuvaju!**
- Podaci su u `data/` folderu
- Nakon `docker-compose down`, podaci ostaju
- Nakon `docker-compose up`, podaci se vraćaju

## 🔄 Deljenje između Članova Tima

Detaljna uputstva: [TEAM_DATA_SHARING.md](./TEAM_DATA_SHARING.md)

**Kratko:**
1. Commit-ujte seed skripte u git
2. Drugi članovi: `git pull` → `.\scripts\seed-all.ps1`

## 📊 Pregled Podataka

```bash
# MongoDB Shell
docker exec -it projekat-2025-mongodb-content-1 mongosh music_streaming
> db.artists.find().pretty()
> db.albums.find().pretty()
> db.songs.find().pretty()
```

## 🆘 Troubleshooting

**Problem:** Podaci se ne čuvaju
- Proverite da li `data/` folderi postoje
- Proverite prava za pisanje

**Problem:** MongoDB ne može da se poveže
```bash
docker-compose down
docker-compose up -d
# Sačekajte 10 sekundi
```

---

Za više detalja, pogledajte [TEAM_DATA_SHARING.md](./TEAM_DATA_SHARING.md)
