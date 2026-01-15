# 🎯 Pokretanje Projekta - SAMO ZA OCENU 6

## ✅ Servisi potrebni za ocenu 6:
- ✅ **mongodb** - Baza podataka
- ✅ **users-service** - Registracija i login
- ✅ **content-service** - Artists, Albums, Songs
- ✅ **notifications-service** - Notifikacije
- ✅ **api-gateway** - Ulazna tačka

---

## 🚀 KORAK 1: Popravite users-service

Prvo, ažurirajte `go.mod` u users-service:

```cmd
cd D:\projekat\Projekat-2025\services\users-service
go mod tidy
```

Ovo je već urađeno! ✅

---

## 🚀 KORAK 2: Pokrenite samo servise za ocenu 6

Idite u glavni folder projekta:

```cmd
cd D:\projekat\Projekat-2025
```

Pokrenite samo potrebne servise:

```cmd
docker-compose -f docker-compose.ocena6.yml up --build
```

**Šta se dešava:**
- Build-uje samo 4 servisa (users, content, notifications, api-gateway)
- Pokreće MongoDB
- Pokreće sve servise
- **Brže je jer ne build-uje nepotrebne servise!**

---

## 🔍 Kako znati da radi:

Trebalo bi da vidite:

```
mongodb_1              | Listening on 0.0.0.0:27017
users-service_1        | Connected to MongoDB
users-service_1        | Users service running on port 8001
content-service_1      | Connected to MongoDB
content-service_1      | Content service running on port 8002
notifications-service_1 | Connected to MongoDB
notifications-service_1 | Notifications service running on port 8005
api-gateway_1          | API Gateway running on port 8081
```

**Ako vidite sve ovo - SVE RADI!** ✅

---

## 🧪 Testiranje:

U novom CMD prozoru:

```cmd
# Test users-service
curl http://localhost:8001/health

# Test content-service
curl http://localhost:8002/health

# Test api-gateway
curl http://localhost:8081/api/users/health
```

---

## 🛑 Zaustavljanje:

U CMD prozoru gde je `docker-compose` pokrenut, pritisnite:

```
Ctrl + C
```

Zatim:

```cmd
docker-compose -f docker-compose.ocena6.yml down
```

---

## ⚠️ Ako i dalje imate greške:

### Problem: "go mod tidy" greška

Pokrenite u svakom servisu:

```cmd
cd D:\projekat\Projekat-2025\services\users-service
go mod tidy

cd D:\projekat\Projekat-2025\services\content-service
go mod tidy

cd D:\projekat\Projekat-2025\services\notifications-service
go mod tidy
```

### Problem: Build greške

Pokušajte da build-ujete jedan po jedan:

```cmd
docker-compose -f docker-compose.ocena6.yml build users-service
docker-compose -f docker-compose.ocena6.yml build content-service
docker-compose -f docker-compose.ocena6.yml build notifications-service
docker-compose -f docker-compose.ocena6.yml build api-gateway
```

---

## 📊 Razlika između fajlova:

- **docker-compose.yml** - SVI servisi (8 servisa)
- **docker-compose.ocena6.yml** - SAMO servisi za ocenu 6 (4 servisa + MongoDB)

Za ocenu 6, koristite `docker-compose.ocena6.yml` - brže je! ⚡

