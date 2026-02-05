# 📊 Deljenje Podataka između Članova Tima

Ovaj dokument objašnjava kako svi članovi tima mogu da imaju iste podatke u bazama i kako da dodaju nove podatke koji će se sačuvati.

## 🎯 Kako Funkcioniše

### 1. Bind Mount Volumes
Umesto Docker named volumes, koristimo **bind mount-ove** koji mapiraju Docker volume-e na lokalne foldere u projektu:

```
./data/mongodb-users    → MongoDB podaci za users-service
./data/mongodb-content  → MongoDB podaci za content-service  
./data/mongodb-ratings  → MongoDB podaci za ratings-service
./data/cassandra        → Cassandra podaci za notifications-service
```

### 2. Deljenje Podataka kroz Git

**Opcija A: Commit-ujte podatke u Git (PREPORUČENO za timski rad)**
- Podaci će biti u `data/` folderu
- Svaki član tima će imati iste podatke nakon `git pull`
- `.gitignore` je konfigurisan da ignorira database fajlove, ali možete promeniti

**Opcija B: Koristite seed skripte**
- Podaci se generišu automatski pri pokretanju
- Svaki član tima pokreće seed skripte lokalno

## 🚀 Početni Setup

### Korak 1: Klonirajte Repozitorijum
```bash
git clone <your-repo-url>
cd Projekat-2025
```

### Korak 2: Pokrenite Servise
```bash
docker-compose up -d
```

### Korak 3: Seed-ujte Podatke

**Na Windows-u:**
```powershell
.\scripts\seed-all.ps1
```

**Na Linux/Mac:**
```bash
chmod +x scripts/seed-all.sh
./scripts/seed-all.sh
```

**Ili ručno:**
```bash
# Seed content database
docker exec -i projekat-2025-mongodb-content-1 mongosh music_streaming < scripts/seed-content.js

# Seed ratings database  
docker exec -i projekat-2025-mongodb-ratings-1 mongosh ratings_db < scripts/seed-ratings.js
```

## 📝 Dodavanje Novih Podataka

### Metoda 1: Kroz API (PREPORUČENO)
Koristite API endpoint-e da dodate podatke:

```bash
# Dodaj umetnika (kao admin)
curl -X POST http://localhost:8081/api/content/artists \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "Novi Umetnik",
    "biography": "Biografija...",
    "genres": ["Pop", "Rock"]
  }'

# Dodaj album
curl -X POST http://localhost:8081/api/content/albums \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Novi Album",
    "releaseDate": "2024-01-01",
    "genre": "Pop",
    "artistIds": ["artist1"]
  }'
```

### Metoda 2: Direktno u MongoDB
```bash
# Povežite se na MongoDB
docker exec -it projekat-2025-mongodb-content-1 mongosh music_streaming

# Dodajte podatke
db.artists.insertOne({
  name: "Novi Umetnik",
  biography: "Biografija...",
  genres: ["Pop"],
  createdAt: new Date()
})
```

### Metoda 3: Ažuriranje Seed Skripti
1. Otvorite `scripts/seed-content.js`
2. Dodajte nove podatke u odgovarajuće kolekcije
3. Pokrenite seed skriptu ponovo:
```bash
.\scripts\seed-all.ps1
```

## 💾 Perzistencija Podataka

### Podaci se Automatski Čuvaju!
- Kada zatvorite Docker kontejnere (`docker-compose down`), podaci ostaju u `data/` folderu
- Kada ponovo pokrenete (`docker-compose up`), podaci se vraćaju
- **VAŽNO:** Podaci se čuvaju lokalno u `data/` folderu

### Backup Podataka
```bash
# Napravite backup
tar -czf backup-$(date +%Y%m%d).tar.gz data/

# Restore backup
tar -xzf backup-YYYYMMDD.tar.gz
```

## 🔄 Sinhronizacija između Članova Tima

### Scenario 1: Commit-ujte Podatke u Git
1. Dodajte podatke kroz API ili seed skripte
2. Commit-ujte `data/` folder:
```bash
git add data/
git commit -m "Add new artists and albums"
git push
```

3. Drugi članovi tima:
```bash
git pull
docker-compose restart
```

### Scenario 2: Koristite Seed Skripte
1. Ažurirajte seed skripte sa novim podacima
2. Commit-ujte seed skripte:
```bash
git add scripts/seed-*.js
git commit -m "Add new seed data"
git push
```

3. Drugi članovi tima:
```bash
git pull
docker-compose restart
.\scripts\seed-all.ps1
```

## 🗑️ Brisanje Podataka

### Reset Baze na Početno Stanje
```bash
# Zaustavite servise
docker-compose down

# Obrišite data foldere
Remove-Item -Recurse -Force data/mongodb-*
Remove-Item -Recurse -Force data/cassandra

# Kreirajte prazne foldere
New-Item -ItemType Directory -Force data/mongodb-users
New-Item -ItemType Directory -Force data/mongodb-content
New-Item -ItemType Directory -Force data/mongodb-ratings
New-Item -ItemType Directory -Force data/cassandra

# Pokrenite servise i seed-ujte
docker-compose up -d
.\scripts\seed-all.ps1
```

## 📊 Pregled Podataka

### Pregled kroz MongoDB Shell
```bash
# Users database
docker exec -it projekat-2025-mongodb-users-1 mongosh users_db
> db.users.find().pretty()

# Content database
docker exec -it projekat-2025-mongodb-content-1 mongosh music_streaming
> db.artists.find().pretty()
> db.albums.find().pretty()
> db.songs.find().pretty()

# Ratings database
docker exec -it projekat-2025-mongodb-ratings-1 mongosh ratings_db
> db.ratings.find().pretty()
```

### Pregled kroz API
```bash
# Lista umetnika
curl http://localhost:8081/api/content/artists

# Lista albuma
curl http://localhost:8081/api/content/albums

# Lista pesama
curl http://localhost:8081/api/content/songs
```

## ⚠️ Važne Napomene

1. **Ne commit-ujte velike audio fajlove** - koristite samo reference na fajlove
2. **Sinhronizujte seed skripte** - ako menjate seed skripte, commit-ujte ih
3. **Backup pre većih promena** - napravite backup pre brisanja podataka
4. **Koristite API za dodavanje** - API automatski validira podatke

## 🆘 Troubleshooting

### Problem: Podaci se ne čuvaju
**Rešenje:** Proverite da li `data/` folderi postoje i imaju prava za pisanje

### Problem: Konflikti između članova tima
**Rešenje:** Koristite seed skripte umesto direktnog commit-ovanja podataka

### Problem: MongoDB ne može da se poveže
**Rešenje:** 
```bash
docker-compose down
docker-compose up -d
# Sačekajte 10 sekundi da se MongoDB pokrene
```

## 📚 Dodatni Resursi

- [MongoDB Dokumentacija](https://docs.mongodb.com/)
- [Docker Volumes Dokumentacija](https://docs.docker.com/storage/volumes/)
- [API Dokumentacija](./POSTMAN_API_DOCUMENTATION.md)
