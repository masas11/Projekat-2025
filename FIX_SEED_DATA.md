# 🔧 Rešenje: Dodavanje Toše Proeskog u Seed Skriptu

## Problem

Dodali ste Tošu Proeskog, album "Ledena" i pesmu "Mesečina" preko frontenda, ali kada je drugarica klonirala projekat i pokrenula seed skriptu, nije dobila te podatke.

## Rešenje

Dodao sam podatke u `scripts/seed-content.js`:

### ✅ Dodato:

1. **Umetnik: Toše Proeski** (artist6)
2. **Album: Ledena** (album6)
3. **Pesma: Mesečina** (song7)

---

## 📝 Šta Treba Uraditi

### 1. Commit-ujte ažuriranu seed skriptu:

```powershell
git add scripts/seed-content.js
git commit -m "Add Toše Proeski, Ledena album and Mesečina song to seed data"
git push
```

### 2. Drugarica treba da:

```powershell
# Pull-uje najnovije promene
git pull

# Pokrene seed skriptu
.\scripts\seed-all.ps1
```

---

## ⚠️ Važno: Duplikati

Ako već postoje podaci u bazi, seed skripta može da fail-uje zbog duplikata.

### Opcija 1: Obrišite postojeće podatke pre seed-ovanja

U `scripts/seed-content.js` na početku dodajte:

```javascript
// Clear existing data
db.artists.deleteMany({});
db.albums.deleteMany({});
db.songs.deleteMany({});
```

### Opcija 2: Koristite insertOne sa upsert (za pojedinačne podatke)

Umesto `insertMany`, možete koristiti:

```javascript
db.artists.updateOne(
  { _id: "artist6" },
  { $set: { name: "Toše Proeski", ... } },
  { upsert: true }
);
```

---

## 🧪 Testiranje

Nakon što drugarica pokrene seed skriptu, proverite:

```powershell
# Proveri umetnike
docker exec projekat-2025-mongodb-content-1 mongosh music_streaming --quiet --eval "db.artists.find({}, {name: 1}).pretty()"

# Proveri albume
docker exec projekat-2025-mongodb-content-1 mongosh music_streaming --quiet --eval "db.albums.find({}, {name: 1}).pretty()"

# Proveri pesme
docker exec projekat-2025-mongodb-content-1 mongosh music_streaming --quiet --eval "db.songs.find({}, {name: 1}).pretty()"
```

Trebalo bi da vidi:
- ✅ Toše Proeski (artist6)
- ✅ Ledena (album6)
- ✅ Mesečina (song7)

---

## 📋 Dodati Podaci u Seed Skripti

### Umetnik:
```javascript
{
  _id: "artist6",
  name: "Toše Proeski",
  biography: "Makedonski pop izvođač.",
  genres: ["Pop"],
  createdAt: new Date()
}
```

### Album:
```javascript
{
  _id: "album6",
  name: "Ledena",
  releaseDate: new Date("2001-12-01"),
  genre: "Pop",
  artistIds: ["artist6"],
  createdAt: new Date(),
  updatedAt: new Date()
}
```

### Pesma:
```javascript
{
  _id: "song7",
  name: "Mesečina",
  duration: 182,
  genre: "Pop",
  albumId: "album6",
  artistIds: ["artist6"],
  audioFileUrl: "/music/Mesecina.mp3",
  createdAt: new Date(),
  updatedAt: new Date()
}
```

---

## ✅ Finalni Koraci

1. ✅ Seed skripta je ažurirana
2. ⏳ Commit-ujte promene
3. ⏳ Push-ujte na Git
4. ⏳ Drugarica pull-uje i seed-uje

---

## 💡 Za Budućnost

Kada dodajete podatke preko frontenda i želite da ih kolege imaju:

1. **Dodajte podatke preko frontenda** → Podaci su u vašoj lokalnoj bazi
2. **Ažurirajte seed skriptu** sa tim podacima
3. **Commit-ujte seed skriptu** → `git add scripts/seed-content.js && git commit -m "..." && git push`
4. **Kolege pull-uju i seed-uju** → `git pull && .\scripts\seed-all.ps1`

**VAŽNO:** Seed skripta je u Git-u, ali podaci iz `data/` foldera NISU (zbog `.gitignore`).
