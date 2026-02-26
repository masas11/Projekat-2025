# 🚀 Vodič za Push na Git - Bez Konflikata

## ⚠️ VAŽNO: Pre Push-a

### Korak 1: Proverite da li ima novih izmena na serveru
```powershell
git fetch origin
git status
```

Ako vidite "Your branch is behind 'origin/main'", prvo uradite:
```powershell
git pull origin main
```

### Korak 2: Rešite konflikte (ako ih ima)
Ako `git pull` pokaže konflikte, rešite ih pre push-a.

---

## 📤 Push Novih Izmena

### Korak 1: Dodajte sve izmene
```powershell
# Dodaj sve nove i izmenjene fajlove
git add .

# ILI dodajte pojedinačno:
git add KAKO_DELITI_PODATKE.md
git add scripts/export-data.ps1
git add scripts/import-data.ps1
git add scripts/seed-data/
git add RESILIENCE_IMPLEMENTATION.md
```

### Korak 2: Commit-ujte izmene
```powershell
git commit -m "Add data export/import scripts and fix resilience implementation"
```

### Korak 3: Pull-ujte najnovije izmene (ako ih ima)
```powershell
git pull origin main
```

Ako ima konflikata, rešite ih, pa:
```powershell
git add .
git commit -m "Resolve merge conflicts"
```

### Korak 4: Push-ujte
```powershell
git push origin main
```

---

## 🔄 Kompletan Workflow (Bez Konflikata)

```powershell
# 1. Proveri status
git status

# 2. Fetch najnovije izmene (bez merge)
git fetch origin

# 3. Proveri da li si iza
git status

# 4. Ako si iza, pull-uj
git pull origin main

# 5. Reši konflikte ako ih ima
# (otvori fajlove sa konfliktima i reši ručno)

# 6. Dodaj izmene
git add .

# 7. Commit
git commit -m "Opis izmena"

# 8. Push
git push origin main
```

---

## 👥 Za Druge Članove Tima

### Kako da preuzmu novu verziju:

#### Prvi put (git clone):
```powershell
git clone https://github.com/masas11/Projekat-2025.git
cd Projekat-2025
docker-compose up -d
Start-Sleep -Seconds 20
.\scripts\import-data.ps1
```

#### Sledeći puti (git pull):
```powershell
# 1. Sačuvaj svoje lokalne izmene (ako ih ima)
git stash

# 2. Pull najnovije izmene
git pull origin main

# 3. Vrati svoje izmene (ako si ih sačuvao)
git stash pop

# 4. Reši konflikte ako ih ima

# 5. Restartuj servise i importuj podatke
docker-compose up -d
Start-Sleep -Seconds 20
.\scripts\import-data.ps1
```

---

## ⚠️ Ako Dođe do Konflikata

### Scenario 1: Konflikti pri pull-u
```powershell
# Git će reći koje fajlove imaju konflikte
# Otvori fajlove i vidi:
<<<<<<< HEAD
# Tvoja verzija
=======
# Njihova verzija
>>>>>>> origin/main

# Reši ručno - zadrži obe izmene ili izaberi jednu
# Zatim:
git add .
git commit -m "Resolve merge conflicts"
git push origin main
```

### Scenario 2: Neko drugi push-ovao pre tebe
```powershell
# Git će reći da ne može push jer si iza
# Uradi:
git pull origin main
# Reši konflikte
git push origin main
```

---

## 🎯 Najbezbedniji Način (Preporučeno)

```powershell
# 1. Sačuvaj svoje izmene
git add .
git commit -m "WIP: Local changes"

# 2. Fetch najnovije
git fetch origin

# 3. Pull i merge
git pull origin main

# 4. Reši konflikte ako ih ima

# 5. Push
git push origin main
```

---

## 📋 Checklist Pre Push-a

- [ ] `git status` - proverio/la sam status
- [ ] `git fetch origin` - proverio/la sam da li ima novih izmena
- [ ] `git pull origin main` - preuzeo/la sam najnovije izmene
- [ ] Rešio/la sam sve konflikte
- [ ] `git add .` - dodao/la sam sve izmene
- [ ] `git commit -m "..."` - commit-ovao/la sam sa smislenom porukom
- [ ] `git push origin main` - push-ovao/la sam

---

## 💡 Saveti

1. **Često commit-ujte** - lakše je rešavati konflikte
2. **Koristite smislene commit poruke** - lakše je pratiti izmene
3. **Komunikujte sa timom** - javite kada push-ujete velike izmene
4. **Pull pre push-a** - uvek pull-ujte pre push-a da izbegnete konflikte

---

## 🆘 Hitno: Ako Nešto Pođe Po Zlu

### Reset na poslednji commit (OPASNO - gubi izmene):
```powershell
git reset --hard HEAD
```

### Reset na remote verziju (OPASNO - gubi sve lokalne izmene):
```powershell
git reset --hard origin/main
```

### Vrati se na prethodni commit:
```powershell
git log --oneline  # vidi commit hash
git reset --hard <commit-hash>
```

---

**Napomena:** Uvek backup-ujte važne izmene pre reset operacija!
