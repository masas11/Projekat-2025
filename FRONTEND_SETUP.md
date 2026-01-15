# Uputstvo za pokretanje Frontend aplikacije

## Korak 1: Pokreni Backend servise

Prvo moraš pokrenuti backend mikroservise preko Docker Compose-a:

```powershell
# U root direktorijumu projekta
docker-compose up -d
```

Ovo će pokrenuti sve servise:
- API Gateway na portu 8081
- Users Service na portu 8001
- Content Service na portu 8002
- MongoDB na portu 27017
- i ostale servise

Proveri da li su servisi pokrenuti:
```powershell
docker-compose ps
```

## Korak 2: Instaliraj Frontend zavisnosti

```powershell
cd frontend
npm install
```

## Korak 3: Pokreni Frontend

```powershell
npm start
```

Frontend će se automatski otvoriti u browseru na `http://localhost:3000`

## Testiranje aplikacije

### 1. Registracija korisnika
- Idi na `/register`
- Popuni formu i registruj se
- Nakon registracije, bićeš preusmeren na login stranicu

### 2. Prijava
- Idi na `/login`
- Unesi korisničko ime i lozinku
- Klikni "Zatraži OTP"
- **VAŽNO**: OTP kod će biti poslat na email, ali za testiranje proveri konzolu backend servisa (users-service) gde će se ispisati OTP kod
- Unesi OTP kod i klikni "Verifikuj OTP"
- Nakon uspešne verifikacije, bićeš prijavljen

### 3. Pregled sadržaja
- Klikni na "Izvođači" u navigaciji da vidiš listu izvođača
- Klikni na "Albumi" da vidiš albume
- Klikni na "Pesme" da vidiš pesme

### 4. Admin funkcionalnosti (ako si admin)
- Klikni "Dodaj izvođača" da kreiraš novog izvođača
- Klikni "Dodaj album" da kreiraš novi album
- Klikni "Dodaj pesmu" da kreiraš novu pesmu
- Klikni "Izmeni" na izvođaču da ga ažuriraš

### 5. Notifikacije
- Klikni na "Notifikacije" u navigaciji
- Videćeš notifikacije za trenutno prijavljenog korisnika

## Troubleshooting

### Backend servisi ne rade
```powershell
# Proveri status
docker-compose ps

# Proveri logove
docker-compose logs api-gateway
docker-compose logs users-service
docker-compose logs content-service
```

### Frontend se ne povezuje sa backend-om
- Proveri da li je API Gateway pokrenut na `http://localhost:8081`
- Otvori browser konzolu (F12) i proveri greške
- Proveri da li postoji CORS problem (trebalo bi da radi jer je proxy podešen)

### OTP kod ne stiže
- Proveri logove users-service servisa:
```powershell
docker-compose logs users-service
```
- OTP kod se obično ispisuje u konzoli servisa za testiranje

## Zaustavljanje servisa

```powershell
# Zaustavi sve servise
docker-compose down

# Zaustavi i obriši volume-ove (briše podatke)
docker-compose down -v
```
