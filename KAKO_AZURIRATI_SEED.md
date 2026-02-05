# 📝 Kako Ažurirati Seed Skriptu - Kratak Vodič

## 🎯 Kada Treba Ažurirati Seed Skriptu?

**Kada dodajete podatke preko frontenda** i želite da ih kolege imaju.

---

## ✅ Koraci (3 koraka)

### **1. Dodajte podatke preko frontenda**
- Frontend → Artists/Albums/Songs → "Dodaj novi"
- Unesite podatke i sačuvajte
- ✅ Podaci su sada u vašoj lokalnoj bazi

### **2. Ažurirajte seed skriptu**

Otvorite `scripts/seed-content.js` i dodajte podatke:

#### **Za umetnika:**
```javascript
{
  _id: "artist6",  // ili sledeći broj
  name: "Ime Umetnika",
  biography: "Biografija...",
  genres: ["Pop", "Rock"],
  createdAt: new Date()
}
```

#### **Za album:**
```javascript
{
  _id: "album6",  // ili sledeći broj
  name: "Ime Albuma",
  releaseDate: new Date("2024-01-01"),
  genre: "Pop",
  artistIds: ["artist6"],  // ID umetnika
  createdAt: new Date(),
  updatedAt: new Date()
}
```

#### **Za pesmu:**
```javascript
{
  _id: "song7",  // ili sledeći broj
  name: "Ime Pesme",
  duration: 180,  // u sekundama
  genre: "Pop",
  albumId: "album6",  // ID albuma
  artistIds: ["artist6"],  // ID umetnika
  audioFileUrl: "/music/pesma.mp3",
  createdAt: new Date(),
  updatedAt: new Date()
}
```

### **3. Commit-ujte i push-ujte**

```powershell
git add scripts/seed-content.js
git commit -m "Add [opis podataka] to seed data"
git push
```

---

## 🔍 Kako Pronaći Podatke iz Baze?

Ako ne znate tačne podatke, proverite u bazi:

```powershell
# Pronađi umetnika
docker exec projekat-2025-mongodb-content-1 mongosh music_streaming --quiet --eval "db.artists.find().forEach(function(a) { print(JSON.stringify(a)); })"

# Pronađi album
docker exec projekat-2025-mongodb-content-1 mongosh music_streaming --quiet --eval "db.albums.find().forEach(function(a) { print(JSON.stringify(a)); })"

# Pronađi pesmu
docker exec projekat-2025-mongodb-content-1 mongosh music_streaming --quiet --eval "db.songs.find().forEach(function(s) { print(JSON.stringify(s)); })"
```

---

## 📋 Primer: Dodavanje Toše Proeskog

### **1. Dodato preko frontenda:**
- Umetnik: Toše Proeski
- Album: Ledena
- Pesma: Mesečina

### **2. Ažurirana seed skripta:**

```javascript
// Umetnik
{
  _id: "artist6",
  name: "Toše Proeski",
  biography: "Makedonski pop izvođač.",
  genres: ["Pop"],
  createdAt: new Date()
}

// Album
{
  _id: "album6",
  name: "Ledena",
  releaseDate: new Date("2001-12-01"),
  genre: "Pop",
  artistIds: ["artist6"],
  createdAt: new Date(),
  updatedAt: new Date()
}

// Pesma
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

### **3. Commit-ovano:**
```powershell
git add scripts/seed-content.js
git commit -m "Add Toše Proeski, Ledena album and Mesečina song to seed data"
git push
```

---

## ⚠️ Važno

- **ID-jevi moraju biti jedinstveni** - koristite `artist6`, `album6`, `song7` itd.
- **Povezivanje:** Album mora imati `artistIds`, pesma mora imati `albumId` i `artistIds`
- **Commit-ujte seed skriptu**, ne `data/` folder (on je u `.gitignore`)

---

## 🎯 Brzi Checklist

- [ ] Dodao/la podatke preko frontenda
- [ ] Otvorio/la `scripts/seed-content.js`
- [ ] Dodao/la podatke u odgovarajuće sekcije (artists, albums, songs)
- [ ] Proverio/la da su ID-jevi jedinstveni i povezani
- [ ] Commit-ovao/la: `git add scripts/seed-content.js && git commit -m "..." && git push`

---

## 💡 Savet

**Kada dodajete podatke preko frontenda, odmah ažurirajte seed skriptu!**
Tako nećete zaboraviti podatke i kolege će ih lako dobiti.
