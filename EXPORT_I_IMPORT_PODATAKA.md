# Export i Import Podataka

## Za tebe (koji exportuješ):

### 1. Export podataka:
```powershell
.\scripts\export-data.ps1
```

### 2. Commit i push na Git:
```powershell
git add scripts/seed-data/*.json
git commit -m "Update seed data - HDFS paths added"
git push
```

---

## Za kolege (koji importuju):

### 1. Git pull:
```powershell
git pull
```

### 2. Import podataka:
```powershell
.\scripts\import-data.ps1
```

### 3. (Opciono) Upload audio fajlova na HDFS:
- Ako su pesme imale audio fajlove, kolege moraju ponovo da ih upload-uju preko frontenda
- Ili koriste postojeće ako su već na HDFS-u

---

## Napomena:
- **HDFS fajlovi se NE exportuju** - ostaju samo na tvom HDFS-u
- **Samo metadata** (audioFileURL) se čuva u MongoDB i exportuje
- Kolege će imati sve informacije o pesmama, ali će morati ponovo da upload-uju audio fajlove
