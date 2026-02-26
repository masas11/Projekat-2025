# 📦 Kako Deliti Podatke sa Timom - Vodič

## 🎯 Problem

Kada dodate nove podatke (korisnike, pesme, albume, pretplate) kroz frontend ili API, oni se čuvaju lokalno u `data/` folderu. Kada drugi član tima klonira repozitorijum, ne dobija te podatke jer su oni ignorisani u `.gitignore`.

## ✅ Rešenje

Koristite **eksport/import skripte** koje čuvaju podatke u JSON formatu koji se može commit-ovati u git.

---

## 📤 Kada Dodate Nove Podatke - EKSPORT

### Korak 1: Dodajte podatke kroz frontend/API
- Kreirajte nove umetnike, albume, pesme
- Ocenite pesme
- Kreirajte pretplate
- itd.

### Korak 2: Eksportujte podatke
```powershell
.\scripts\export-data.ps1
```

Ovo će kreirati JSON fajlove u `scripts/seed-data/`:
- `artists.json`
- `albums.json`
- `songs.json`
- `ratings.json`
- `subscriptions.json`

### Korak 3: Commit-ujte u git
```powershell
git add scripts/seed-data/*.json
git commit -m "Update seed data - dodati novi podaci"
git push
```

---

## 📥 Kada Preuzmete Nove Podatke - IMPORT

### Korak 1: Preuzmite najnovije izmene
```powershell
git pull
```

### Korak 2: Pokrenite servise (ako nisu pokrenuti)
```powershell
docker-compose up -d
```

### Korak 3: Sačekajte da se servisi pokrenu
```powershell
Start-Sleep -Seconds 20
```

### Korak 4: Importujte podatke
```powershell
.\scripts\import-data.ps1
```

**Gotovo!** Sada imate sve najnovije podatke.

---

## 🔄 Workflow za Tim

### Scenario 1: Vi dodajete nove podatke

1. Dodajte podatke kroz frontend/API
2. Pokrenite: `.\scripts\export-data.ps1`
3. Commit-ujte: `git add scripts/seed-data/*.json && git commit -m "Update data" && git push`

### Scenario 2: Drugi član tima dodaje podatke

1. Preuzmite izmene: `git pull`
2. Pokrenite servise: `docker-compose up -d`
3. Sačekajte 20 sekundi
4. Importujte: `.\scripts\import-data.ps1`

---

## ⚠️ Važne Napomene

### 1. Eksportujte pre svakog push-a
Ako ste dodali nove podatke, **uvek** eksportujte ih pre push-a:
```powershell
.\scripts\export-data.ps1
git add scripts/seed-data/*.json
git commit -m "Update data"
git push
```

### 2. Importujte nakon svakog pull-a
Ako ste preuzeli nove izmene, **uvek** importujte podatke:
```powershell
git pull
docker-compose up -d
Start-Sleep -Seconds 20
.\scripts\import-data.ps1
```

### 3. Ako nema eksportovanih fajlova
Ako `scripts/seed-data/` folder ne postoji ili je prazan, koristite osnovne seed skripte:
```powershell
.\scripts\seed-all.ps1
```

---

## 🛠️ Troubleshooting

### Problem: "MongoDB kontejneri nisu pokrenuti"
**Rešenje:**
```powershell
docker-compose up -d
Start-Sleep -Seconds 20
.\scripts\export-data.ps1  # ili import-data.ps1
```

### Problem: "Folder scripts/seed-data ne postoji"
**Rešenje:** Skripta će automatski kreirati folder. Ako ne, kreirajte ručno:
```powershell
New-Item -ItemType Directory -Path "scripts\seed-data"
```

### Problem: Import ne radi
**Rešenje:** Proverite da li su JSON fajlovi validni:
```powershell
Get-Content scripts/seed-data/artists.json | ConvertFrom-Json
```

### Problem: Duplikati podataka
**Rešenje:** Import skripta koristi `--drop` flag koji briše postojeće podatke pre importa. Ako želite da zadržite postojeće, izmenite skriptu.

---

## 📋 Checklist

- [ ] Dodao/la sam nove podatke kroz frontend/API
- [ ] Pokrenuo/la sam `.\scripts\export-data.ps1`
- [ ] Commit-ovao/la sam JSON fajlove u git
- [ ] Push-ovao/la sam izmene

**Za drugi član tima:**
- [ ] Preuzeo/la sam najnovije izmene (`git pull`)
- [ ] Pokrenuo/la sam servise (`docker-compose up -d`)
- [ ] Sačekao/la sam 20 sekundi
- [ ] Pokrenuo/la sam `.\scripts\import-data.ps1`

---

## 💡 Saveti

1. **Eksportujte često** - što češće eksportujete, to je lakše deljenje
2. **Commit-ujte JSON fajlove** - oni su mali i brzi za commit
3. **Koristite smislene commit poruke** - npr. "Add new artists and songs"
4. **Komunikujte sa timom** - javite kada dodajete nove podatke

---

## 🎓 Primer

```powershell
# Vi dodajete nove podatke
# 1. Dodajte umetnika, album, pesme kroz frontend
# 2. Eksportujte
.\scripts\export-data.ps1

# 3. Commit i push
git add scripts/seed-data/*.json
git commit -m "Add new artist: Taylor Swift with album and songs"
git push

# Drugi član tima
# 1. Pull
git pull

# 2. Pokreni servise
docker-compose up -d
Start-Sleep -Seconds 20

# 3. Import
.\scripts\import-data.ps1

# Sada ima sve vaše podatke!
```

---

**Napomena:** Ovo je najbolji način za deljenje podataka u timu. Podaci se čuvaju u JSON formatu koji je lako čitljiv i može se commit-ovati u git.
