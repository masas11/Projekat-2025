# 👥 Kako Videti Sve Korisnike u Bazi

## Metoda 1: Direktno preko MongoDB (PREPORUČENO)

### Komanda za CMD:
```cmd
docker exec -it projekat-2025-mongodb-users-1 mongosh users_db --eval "db.users.find().pretty()"
```

### Komanda za PowerShell:
```powershell
docker exec -it projekat-2025-mongodb-users-1 mongosh users_db --eval "db.users.find().pretty()"
```

### Jednostavnija verzija (bez pretty):
```cmd
docker exec -it projekat-2025-mongodb-users-1 mongosh users_db --eval "db.users.find()"
```

### Broj korisnika:
```cmd
docker exec -it projekat-2025-mongodb-users-1 mongosh users_db --eval "db.users.countDocuments()"
```

### Samo username i email:
```cmd
docker exec -it projekat-2025-mongodb-users-1 mongosh users_db --eval "db.users.find({}, {username: 1, email: 1, role: 1, verified: 1}).pretty()"
```

---

## Metoda 2: Interaktivni MongoDB Shell

### Pokrenite interaktivni shell:
```cmd
docker exec -it projekat-2025-mongodb-users-1 mongosh users_db
```

### Zatim u shell-u unesite:
```javascript
// Pregled svih korisnika
db.users.find().pretty()

// Broj korisnika
db.users.countDocuments()

// Samo username i email
db.users.find({}, {username: 1, email: 1, role: 1, verified: 1}).pretty()

// Pronađi određenog korisnika
db.users.findOne({username: "admin"})

// Izlaz iz shell-a
exit
```

---

## Metoda 3: Preko PowerShell (Formatirano)

### Komanda za PowerShell:
```powershell
docker exec projekat-2025-mongodb-users-1 mongosh users_db --quiet --eval "db.users.find().forEach(function(user) { print('Username: ' + user.username + ', Email: ' + user.email + ', Role: ' + user.role); })"
```

### Ili jednostavnije:
```powershell
docker exec projekat-2025-mongodb-users-1 mongosh users_db --eval "db.users.find().pretty()" | Out-String
```

---

## Metoda 4: Export u JSON fajl

### Export svih korisnika u fajl:
```cmd
docker exec projekat-2025-mongodb-users-1 mongosh users_db --quiet --eval "db.users.find().toArray()" > users.json
```

---

## Brzi Primeri

### 1. Pregled svih korisnika (formatirano):
```cmd
docker exec -it projekat-2025-mongodb-users-1 mongosh users_db --eval "db.users.find().pretty()"
```

### 2. Broj korisnika:
```cmd
docker exec projekat-2025-mongodb-users-1 mongosh users_db --quiet --eval "db.users.countDocuments()"
```

### 3. Pronađi admin korisnika:
```cmd
docker exec -it projekat-2025-mongodb-users-1 mongosh users_db --eval "db.users.findOne({username: 'admin'})"
```

### 4. Samo username i email:
```cmd
docker exec -it projekat-2025-mongodb-users-1 mongosh users_db --eval "db.users.find({}, {username: 1, email: 1, role: 1}).pretty()"
```

---

## Ako Ne Znaš Ime Kontejnera

### Pronađi ime MongoDB kontejnera:
```cmd
docker ps | findstr mongo
```

Ili:
```cmd
docker ps --format "table {{.Names}}\t{{.Image}}" | findstr mongo
```

---

## Troubleshooting

### Problem: "container not found"
**Rešenje:** Proveri da li je kontejner pokrenut:
```cmd
docker ps
```

### Problem: "mongosh: command not found"
**Rešenje:** Koristi `mongo` umesto `mongosh` (za starije verzije):
```cmd
docker exec -it projekat-2025-mongodb-users-1 mongo users_db --eval "db.users.find().pretty()"
```

### Problem: "cannot connect"
**Rešenje:** Sačekaj da se MongoDB potpuno pokrene:
```cmd
docker-compose restart mongodb-users
Start-Sleep -Seconds 10
```

---

## Najbrži Način (Copy-Paste)

**Za CMD:**
```cmd
docker exec -it projekat-2025-mongodb-users-1 mongosh users_db --eval "db.users.find().pretty()"
```

**Za PowerShell:**
```powershell
docker exec projekat-2025-mongodb-users-1 mongosh users_db --eval "db.users.find().pretty()"
```
