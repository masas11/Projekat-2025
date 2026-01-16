# Modeli podataka - Ocena 6

## 1. Users Service

**Baza podataka:** MongoDB (dokument-orijentisana)

### Entitet: User

**Kolekcija:** `users`

**Atributi:**
- `_id` (string) - jedinstveni identifikator korisnika (UUID)
- `firstName` (string) - ime korisnika
- `lastName` (string) - prezime korisnika
- `email` (string) - email adresa korisnika (jedinstvena)
- `username` (string) - korisničko ime (jedinstveno)
- `passwordHash` (string) - heširana lozinka (bcrypt)
- `role` (string) - uloga korisnika ("USER" ili "ADMIN")
- `verified` (boolean) - da li je email verifikovan
- `passwordChangedAt` (datetime) - datum i vreme poslednje promene lozinke
- `passwordExpiresAt` (datetime) - datum i vreme isteka lozinke (60 dana od promene)
- `failedLoginAttempts` (integer) - broj neuspešnih pokušaja prijave
- `lockedUntil` (datetime) - datum i vreme do kog je nalog zaključan
- `createdAt` (datetime) - datum i vreme kreiranja naloga

**Indeksi:**
- Jedinstveni indeks na `username`
- Jedinstveni indeks na `email`

### Privremene kolekcije za tokenizaciju:

**Kolekcija:** `otps` - OTP kodovi za prijavu
- `username` (string) - korisničko ime
- `code` (string) - OTP kod
- `expiresAt` (datetime) - datum isteka (5 minuta)

**Kolekcija:** `magic_links` - Magic link tokeni za povraćaj naloga
- `token` (string) - magic link token
- `email` (string) - email adresa korisnika
- `type` (string) - tip tokena ("magic_link", "verification", "password_reset")
- `expiresAt` (datetime) - datum isteka:
  - Magic link: 15 minuta
  - Email verifikacija: 24 sata
  - Password reset: 1 sat

---

## 2. Content Service

**Baza podataka:** MongoDB (dokument-orijentisana)

### Entitet: Artist (Umetnik)

**Kolekcija:** `artists`

**Atributi:**
- `_id` (string) - jedinstveni identifikator umetnika (UUID)
- `name` (string) - ime umetnika
- `biography` (string) - biografija umetnika
- `genres` (array<string>) - lista žanrova sa kojima je umetnik povezan
- `createdAt` (datetime) - datum i vreme kreiranja
- `updatedAt` (datetime) - datum i vreme poslednje izmene

### Entitet: Album

**Kolekcija:** `albums`

**Atributi:**
- `_id` (string) - jedinstveni identifikator albuma (UUID)
- `name` (string) - naziv albuma
- `releaseDate` (datetime) - datum izdavanja albuma
- `genre` (string) - žanr albuma
- `artistIds` (array<string>) - lista ID-jeva umetnika koji su uradili album
- `createdAt` (datetime) - datum i vreme kreiranja
- `updatedAt` (datetime) - datum i vreme poslednje izmene

**Veze:**
- `artistIds` -> Artist._id (many-to-many)

### Entitet: Song (Pesma)

**Kolekcija:** `songs`

**Atributi:**
- `_id` (string) - jedinstveni identifikator pesme (UUID)
- `name` (string) - naziv pesme
- `duration` (integer) - trajanje pesme u sekundama
- `genre` (string) - žanr pesme
- `albumId` (string) - ID albuma kojem pesma pripada
- `artistIds` (array<string>) - lista ID-jeva umetnika koji su izvodili pesmu
- `createdAt` (string/datetime) - datum i vreme kreiranja
- `updatedAt` (string/datetime) - datum i vreme poslednje izmene

**Veze:**
- `albumId` -> Album._id (many-to-one)
- `artistIds` -> Artist._id (many-to-many)

---

## 3. Notifications Service

**Baza podataka:** Wide-column baza (za ocenu 6 može biti MongoDB, ali za više ocene Cassandra)

### Entitet: Notification

**Kolekcija/Tabela:** `notifications`

**Atributi:**
- `_id` (string) - jedinstveni identifikator notifikacije (UUID)
- `userId` (string) - ID korisnika koji prima notifikaciju
- `title` (string) - naslov notifikacije
- `message` (string) - poruka notifikacije
- `type` (string) - tip notifikacije ("album_added", "song_added", "artist_added")
- `relatedId` (string) - ID povezanog entiteta (album/song/artist ID)
- `read` (boolean) - da li je notifikacija pročitana
- `createdAt` (datetime) - datum i vreme kreiranja notifikacije

**Indeksi:**
- Indeks na `userId` za brzo filtriranje po korisniku

**Veze:**
- `userId` -> User._id (many-to-one)

---

## 4. API Gateway

**Tip:** HTTP Proxy/Reverse Proxy

**Funkcionalnosti:**
- Rutiranje zahteva ka odgovarajućim servisima
- CORS handling
- Rate limiting (DoS zaštita)
- Authorization middleware (JWT provera)
- Prosleđivanje query parametara i headers

**Nema bazu podataka** - samo prosleđuje zahteve

---

## Komunikacija između servisa

### Stil komunikacije:

1. **Sinhrona komunikacija:**
   - REST API preko HTTP
   - Koristi se za:
     - API Gateway -> Users Service (autentifikacija/autorizacija)
     - API Gateway -> Content Service (CRUD operacije)
     - API Gateway -> Notifications Service (dobavljanje notifikacija)

2. **HTTP metode:**
   - GET - čitanje podataka
   - POST - kreiranje podataka
   - PUT - ažuriranje podataka
   - DELETE - brisanje podataka

---

## Relacije između entiteta

```
User (1) -----< (N) Notification
  |
  |
  
Artist (N) -----< (N) Album (many-to-many preko artistIds)
  |
  |
  
Album (1) -----< (N) Song (one-to-many preko albumId)

Artist (N) -----< (N) Song (many-to-many preko artistIds)
```

---

## Napomene

- **Users Service** koristi bcrypt za heširanje lozinki (hash & salt)
- **Content Service** čuva veze između entiteta preko ID-jeva (denormalizovano)
- **Notifications Service** za ocenu 6 može koristiti MongoDB, ali za više ocene se preporučuje Cassandra (wide-column)
- Svi ID-jevi su UUID stringovi
- Svi servisi imaju `createdAt` i `updatedAt` polja za audit logovanje
