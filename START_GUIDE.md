# 🚀 Vodič za Pokretanje Sistema - Za Sve Članove Tima

## 📋 Odgovori na Vaša Pitanja

### 1. ✅ Da li svi imamo iste podatke u bazi?

**ODGOVOR: ZAVISI OD PRISTUPA**

#### **Opcija A: Seed Skripte (PREPORUČENO za timski rad)**
- ✅ **DA** - Svi će imati iste podatke
- Svaki član tima pokreće `.\scripts\seed-all.ps1` nakon `git pull`
- Seed skripte su u Git-u, tako da su svi sinhronizovani
- **Prednost:** Lako ažuriranje - samo commit-ujte nove seed skripte

#### **Opcija B: Bind Mount Podaci**
- ⚠️ **MOŽDA** - Zavisi od Git konfiguracije
- Podaci su u `data/` folderu
- Ako commit-ujete `data/` folder → svi imaju iste podatke
- Ako NE commit-ujete → svako ima svoje podatke
- **Prednost:** Brže (nema seed-ovanja)
- **Mana:** Veći Git repo, mogući konflikti

**PREPORUKA:** Koristite seed skripte! Commit-ujte samo `scripts/seed-*.js` fajlove.

---

### 2. 🔧 Kako da pokrećem program?

#### **Prvi put (ili nakon promene koda):**
```powershell
docker-compose up -d --build
```
- `--build` rebuild-uje Docker image-e sa novim kodom
- Koristite kada menjate Go kod ili Dockerfile-ove

#### **Svaki sledeći put (ako niste menjali kod):**
```powershell
docker-compose up -d
```
- Koristi postojeće Docker image-e
- Brže pokretanje

#### **Nakon git pull (ako su seed skripte ažurirane):**
```powershell
docker-compose up -d
Start-Sleep -Seconds 20
.\scripts\seed-all.ps1
```

---

### 3. 💾 Kako se upisuju podaci u bazu?

#### **✅ DA, radi preko frontenda!**

Frontend automatski šalje podatke kroz API Gateway:

**Kreiranje umetnika (Admin):**
- Frontend → API Gateway → Content Service → MongoDB
- Podaci se **automatski čuvaju** u `data/mongodb-content/`

**Kreiranje albuma (Admin):**
- Frontend → API Gateway → Content Service → MongoDB
- Podaci se **automatski čuvaju**

**Kreiranje pesme (Admin):**
- Frontend → API Gateway → Content Service → MongoDB
- Podaci se **automatski čuvaju**

**Ocenjivanje pesme (Korisnik):**
- Frontend → API Gateway → Ratings Service → MongoDB
- Podaci se **automatski čuvaju** u `data/mongodb-ratings/`

**Pretplata na umetnika/žanr:**
- Frontend → API Gateway → Subscriptions Service → MongoDB
- Podaci se **automatski čuvaju**

---

## 🎯 Standardni Workflow za Timski Rad

### **Prvi put (novi član tima):**
```powershell
# 1. Kloniraj repo
git clone <repo-url>
cd Projekat-2025

# 2. Pokreni servise
docker-compose up -d --build

# 3. Sačekaj da se servisi pokrenu
Start-Sleep -Seconds 20

# 4. Seed-uj podatke
.\scripts\seed-all.ps1

# 5. Pokreni frontend (opciono)
cd frontend
npm install
npm start
```

### **Svakodnevni rad:**
```powershell
# 1. Pull najnovije promene
git pull

# 2. Pokreni servise (bez build ako nema promena u kodu)
docker-compose up -d

# 3. Ako su seed skripte ažurirane, seed-uj ponovo
.\scripts\seed-all.ps1

# 4. Pokreni frontend
cd frontend
npm start
```

### **Nakon promene koda:**
```powershell
# 1. Commit-uj promene
git add .
git commit -m "Opis promena"
git push

# 2. Drugi članovi: Pull i rebuild
git pull
docker-compose up -d --build
```

---

## 📝 Dodavanje Novih Podataka

### **Metoda 1: Preko Frontenda (PREPORUČENO)**
1. Prijavite se kao admin (username: `admin`, password: `admin123`)
2. Idite na Artists/Albums/Songs stranicu
3. Kliknite "Dodaj novi"
4. Popunite formu i sačuvajte
5. ✅ **Podaci su automatski sačuvani u bazi!**

### **Metoda 2: Preko Seed Skripti**
1. Otvorite `scripts/seed-content.js`
2. Dodajte nove podatke
3. Commit-ujte: `git add scripts/seed-content.js && git commit -m "Add new data" && git push`
4. Drugi članovi: `git pull && .\scripts\seed-all.ps1`

### **Metoda 3: Preko API-ja (za testiranje)**
```powershell
# Dodaj umetnika
$token = "YOUR_JWT_TOKEN"
$body = @{
    name = "Novi Umetnik"
    biography = "Biografija..."
    genres = @("Pop", "Rock")
} | ConvertTo-Json

Invoke-WebRequest -Uri http://localhost:8081/api/content/artists `
    -Method POST `
    -Headers @{"Authorization"="Bearer $token"; "Content-Type"="application/json"} `
    -Body $body
```

---

## 🔄 Sinhronizacija Podataka između Članova

### **Scenario 1: Seed Skripte (PREPORUČENO)**
```powershell
# Član A: Dodaje novog umetnika preko frontenda
# → Podaci su u njegovoj lokalnoj bazi

# Član A: Ažurira seed skriptu sa novim podacima
# → Edituje scripts/seed-content.js
git add scripts/seed-content.js
git commit -m "Add new artist to seed data"
git push

# Član B: Pull i seed
git pull
.\scripts\seed-all.ps1
# → Sada i Član B ima novog umetnika!
```

### **Scenario 2: Direktno Deljenje Podataka**
```powershell
# Član A: Commit-uje data folder (ako je u .gitignore dozvoljeno)
git add data/
git commit -m "Update database data"
git push

# Član B: Pull
git pull
docker-compose restart
# → Sada i Član B ima iste podatke!
```

**⚠️ PAŽNJA:** Scenario 2 može uzrokovati Git konflikte ako više ljudi radi istovremeno!

---

## ✅ Provera da li Sve Radi

```powershell
# 1. Proveri da li su servisi pokrenuti
docker-compose ps

# 2. Proveri API
Invoke-WebRequest -Uri http://localhost:8081/api/content/artists -UseBasicParsing

# 3. Proveri da li postoje podaci
docker exec projekat-2025-mongodb-content-1 mongosh music_streaming --eval "db.artists.countDocuments()" --quiet
```

---

## 🆘 Česti Problemi

### Problem: "Port already in use"
```powershell
# Zaustavi sve Docker kontejnere
docker-compose down

# Proveri koji proces koristi port
netstat -ano | findstr :8081

# Pokreni ponovo
docker-compose up -d
```

### Problem: "MongoDB connection refused"
```powershell
# Sačekaj duže da se MongoDB pokrene
Start-Sleep -Seconds 30
docker-compose restart
```

### Problem: "Podaci se ne čuvaju"
```powershell
# Proveri da li data/ folderi postoje
ls data/

# Proveri prava za pisanje
# Windows: Trebalo bi da radi automatski
```

---

## 📚 Dodatni Resursi

- [TEAM_DATA_SHARING.md](./TEAM_DATA_SHARING.md) - Detaljno o deljenju podataka
- [README_DATA.md](./README_DATA.md) - Brzi vodič za rad sa bazama
- [QUICK_START.md](./QUICK_START.md) - Brzi start vodič
