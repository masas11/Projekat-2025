# 🚀 Brzi Start - Pokretanje Projekta

## Korak 1: Proverite Docker

```bash
docker --version
```

Ako nema Docker-a, preuzmite: https://www.docker.com/products/docker-desktop/

## Korak 2: Pokrenite sve (uključujući MongoDB)

```bash
cd Projekat-2025
docker-compose up
```

**To je sve!** 🎉

MongoDB će se automatski pokrenuti i svi servisi će se povezati.

## Korak 3: Proverite da li radi

U drugom terminalu:

```bash
# Proverite MongoDB
docker ps | findstr mongo

# Proverite servise
curl http://localhost:8081/api/users/health
curl http://localhost:8002/health
```

## Korak 4: Pokrenite Frontend

U novom terminalu:

```bash
cd Projekat-2025/frontend
npm install
npm start
```

Frontend će se otvoriti na: http://localhost:3000

---

## 🛑 Zaustavljanje

```bash
# U terminalu gde je docker-compose up pokrenut, pritisnite:
Ctrl + C

# Zatim:
docker-compose down
```

---

## ❓ Problemi?

Pogledajte `MONGODB_UPUTSTVO.md` za detaljna uputstva!


