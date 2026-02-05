# 📝 Workflow: Dodavanje Podataka Preko Frontenda

## 🎯 Vaš Scenario

**Vi:** Dodajete podatke preko frontenda → Pokrećete sa `docker-compose up -d`

**Pitanje:** Šta kolege treba da urade kada pull-uju vaše promene?

---

## ✅ Odgovor: ZAVISI OD TOGA ŠTA COMMIT-UJETE

### **Scenario 1: Samo kod (bez podataka) - TRENUTNO STANJE**

**Vi:**
```powershell
# Dodajete podatke preko frontenda
# → Podaci su u data/mongodb-content/ folderu
# → Commit-ujete samo kod (bez data/ foldera)
git add .
git commit -m "Add new feature"
git push
```

**Kolege:**
```powershell
# Pull-uju promene
git pull

# Pokreću servise
docker-compose up -d

# ⚠️ PROBLEM: Oni NEĆE imati podatke koje ste vi dodali preko frontenda!
# → Moraju seed-ovati ili dodati podatke sami
```

**Rešenje za kolege:**
```powershell
# Opcija 1: Seed-ujte postojeće podatke
.\scripts\seed-all.ps1

# Opcija 2: Dodajte podatke preko frontenda sami
```

---

### **Scenario 2: Commit-ujete i podatke (data/ folder)**

**Vi:**
```powershell
# Dodajete podatke preko frontenda
# → Podaci su u data/mongodb-content/ folderu

# Promenite .gitignore da NE ignorira data/ folder
# Ili ručno commit-ujte:
git add data/
git commit -m "Update database data"
git push
```

**Kolege:**
```powershell
# Pull-uju promene (uključujući data/ folder)
git pull

# Restart-uju servise da učitate nove podatke
docker-compose restart

# ✅ Sada imaju iste podatke kao vi!
```

**⚠️ PAŽNJA:** 
- Data folder može biti veliki
- Može uzrokovati Git konflikte ako više ljudi radi istovremeno
- Nije preporučeno za velike projekte

---

### **Scenario 3: Ažurirate seed skripte (PREPORUČENO!)**

**Vi:**
```powershell
# 1. Dodajete podatke preko frontenda
# → Podaci su u vašoj lokalnoj bazi

# 2. Ažurirate seed skriptu sa tim podacima
# → Editujete scripts/seed-content.js i dodate nove podatke

# 3. Commit-ujete seed skriptu
git add scripts/seed-content.js
git commit -m "Add new artists/albums/songs to seed data"
git push
```

**Kolege:**
```powershell
# Pull-uju promene
git pull

# Pokreću servise
docker-compose up -d

# Seed-uju podatke (uključujući nove koje ste vi dodali)
.\scripts\seed-all.ps1

# ✅ Sada imaju iste podatke kao vi!
```

---

## 🎯 PREPORUČENI WORKFLOW

### **Za Vas (koji dodajete podatke preko frontenda):**

```powershell
# 1. Dodajte podatke preko frontenda
# → Otvorite frontend, prijavite se kao admin
# → Dodajte umetnike/albume/pesme

# 2. Ažurirajte seed skriptu sa tim podacima
# → Otvorite scripts/seed-content.js
# → Dodajte nove podatke u odgovarajuće kolekcije

# 3. Commit-ujte seed skriptu
git add scripts/seed-content.js
git commit -m "Add [description] to seed data"
git push
```

### **Za Vaše Kolege:**

```powershell
# 1. Pull-uju promene
git pull

# 2. Pokreću servise (ako već nisu pokrenuti)
docker-compose up -d

# 3. Seed-uju podatke (uključujući nove)
.\scripts\seed-all.ps1

# ✅ Gotovo! Imaju iste podatke kao vi!
```

---

## 📋 Detaljni Primer

### **Primer: Dodavanje novog umetnika**

**Vi:**
1. Otvorite frontend → Artists → "Dodaj novi"
2. Unesete: Name: "Taylor Swift", Biography: "...", Genres: ["Pop", "Country"]
3. Sačuvate → Podatak je u vašoj bazi
4. Otvorite `scripts/seed-content.js`
5. Dodate u `db.artists.insertMany([...])`:
```javascript
{
  _id: "artist6",
  name: "Taylor Swift",
  biography: "...",
  genres: ["Pop", "Country"],
  createdAt: new Date()
}
```
6. Commit-ujte:
```powershell
git add scripts/seed-content.js
git commit -m "Add Taylor Swift to seed data"
git push
```

**Kolege:**
```powershell
git pull
docker-compose up -d
.\scripts\seed-all.ps1
# → Sada i oni imaju Taylor Swift!
```

---

## ⚠️ VAŽNE NAPOMENE

### **1. Seed skripta će pokušati da doda duplikate**
- Ako umetnik već postoji (isti `_id`), seed će fail-ovati za taj umetnik
- To je OK - ostali podaci će biti dodati
- Ili možete koristiti `insertOne` umesto `insertMany` sa proverom

### **2. Ako ne ažurirate seed skriptu**
- Kolege NEĆE imati podatke koje ste vi dodali preko frontenda
- Moraju dodati podatke sami preko frontenda ili seed-ovati ručno

### **3. Ako commit-ujete data/ folder**
- Kolege će imati podatke automatski
- Ali može biti problematično za Git (veliki fajlovi, konflikti)

---

## 🎯 FINALNI ODGOVOR NA VAŠE PITANJE

**Pitanje:** Da li kolege treba da urade neku komandu zbog seeda?

**Odgovor:** 
- ✅ **DA** - Moraju pokrenuti `.\scripts\seed-all.ps1` nakon `git pull`
- **ALI** samo ako ste vi ažurirali seed skripte sa novim podacima
- Ako NISTE ažurirali seed skripte → kolege neće imati vaše podatke

**Preporuka:**
- Uvek ažurirajte seed skripte kada dodajete podatke preko frontenda
- Tako će svi imati iste podatke nakon seed-ovanja

---

## 📝 Brzi Checklist

**Vi (nakon dodavanja podataka preko frontenda):**
- [ ] Ažurirate `scripts/seed-content.js` sa novim podacima
- [ ] Commit-ujete seed skriptu
- [ ] Push-ujete promene

**Kolege (nakon git pull):**
- [ ] Pull-uju promene (`git pull`)
- [ ] Pokreću servise (`docker-compose up -d`)
- [ ] Seed-uju podatke (`.\scripts\seed-all.ps1`)

---

## 🆘 Troubleshooting

### Problem: Seed skripta ne dodaje podatke koje sam dodao preko frontenda
**Rešenje:** Morate ručno dodati te podatke u seed skriptu

### Problem: Kolege nemaju moje podatke nakon seed-ovanja
**Rešenje:** Proverite da li ste commit-ovali seed skriptu sa novim podacima

### Problem: Duplikati u bazi
**Rešenje:** Seed skripta će fail-ovati za duplikate, ali to je OK - ostali podaci će biti dodati
