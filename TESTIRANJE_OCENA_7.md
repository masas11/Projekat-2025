# Testiranje zahteva za ocenu 7

## Preduslovi
1. Pokrenuti sve servise: `docker-compose up`
2. Pokrenuti frontend: `cd frontend && npm start`
3. Imati kreiranog korisnika (ne admin) i admin korisnika

---

## 1.7 Reprodukcija pesme (RK)

### Test koraci:
1. **Pristupite stranici sa pesmama**: `http://localhost:3000/songs`
2. **Kliknite na bilo koju pesmu** da otvorite SongDetail stranicu
3. **Proverite da li postoji AudioPlayer** sa kontrolama (play, pause, volume, progress bar)
4. **Kliknite na Play dugme** i proverite da li se pesma reprodukuje
5. **Testirajte kontrole**: volume slider, seek (klik na progress bar)

### Očekivani rezultat:
- AudioPlayer se prikazuje
- Pesma se može reprodukovati
- Sve kontrole rade

---

## 1.8 Filtriranje i pretraga (A, RK)

### Preduslov: Kreirajte test podatke
**VAŽNO**: Pre testiranja, kreirajte test podatke kao admin:
1. **Prijavite se kao admin**
2. **Kreirajte barem 3-5 umetnika** sa različitim žanrovima
3. **Kreirajte barem 5-10 pesama** sa različitim nazivima i žanrovima
4. **Kreirajte barem 3-5 albuma** sa različitim nazivima i žanrovima

### Test 1: Pretraga i filtriranje pesama
1. **Idite na**: `http://localhost:3000/songs`
2. **Unesite naziv pesme** u polje za pretragu (npr. deo naziva)
3. **Proverite da li se lista filtrira** po nazivu (broj pronađenih se menja)
4. **Očistite pretragu**
5. **Izaberite žanr** iz dropdown-a (npr. "Pop")
6. **Proverite da li se lista filtrira** po žanru
7. **Kombinujte pretragu i filter** - unesite naziv I izaberite žanr
8. **Proverite da li rade zajedno** (lista se filtrira po oba kriterijuma)

### Test 2: Pretraga i filtriranje umetnika
1. **Idite na**: `http://localhost:3000/artists`
2. **Unesite naziv umetnika** u polje za pretragu (deo naziva)
3. **Proverite da li se lista filtrira** po nazivu
4. **Očistite pretragu**
5. **Izaberite žanr** iz dropdown-a
6. **Proverite da li se lista filtrira** po žanru
7. **Kombinujte pretragu i filter** - proverite da li rade zajedno

### Test 3: Pretraga i filtriranje albuma
1. **Idite na**: `http://localhost:3000/albums`
2. **Unesite naziv albuma** u polje za pretragu
3. **Proverite da li se lista filtrira** po nazivu
4. **Očistite pretragu**
5. **Izaberite žanr** iz dropdown-a
6. **Proverite da li se lista filtrira** po žanru
7. **Kombinujte pretragu i filter** - proverite da li rade zajedno

### Očekivani rezultat:
- Pretraga radi po nazivu
- Filtriranje radi po žanru
- Kombinacija pretrage i filtera radi
- Prikazuje se broj pronađenih rezultata

---

## 1.9 Ocenjivanje pesama (RK)

### Preduslov: Kreirajte test podatke
1. **Kreirajte barem jednu pesmu** (kao admin)

### Test koraci:
1. **Prijavite se kao običan korisnik** (ne admin)
2. **Idite na stranicu sa pesmama**: `http://localhost:3000/songs`
3. **Kliknite na bilo koju pesmu** da otvorite SongDetail stranicu
4. **Proverite da li postoji sekcija "Oceni pesmu"** sa 5 zvezdica
5. **Kliknite na zvezdicu** (npr. 4. zvezdicu)
6. **Proverite poruku o uspešnom čuvanju**
7. **Proverite da li se ocena sačuvala** (zvezdice se popune do 4, prikaže se "Vaša ocena: 4/5")
8. **Osvežite stranicu** (F5)
9. **Proverite da li se ocena učitala** (zvezdice su i dalje popunjene do 4)
10. **Kliknite na drugu zvezdicu** (npr. 5. zvezdicu)
11. **Proverite da li se ocena izmenila** (sada je 5/5)
12. **Kliknite na "Obriši ocenu"** dugme
13. **Proverite da li se ocena obrisala** (zvezdice su prazne)
14. **Osvežite stranicu**
15. **Proverite da li je ocena i dalje obrisana**

### Test kao admin:
1. **Prijavite se kao admin**
2. **Idite na stranicu pesme**
3. **Proverite da li sekcija za ocenjivanje NIJE prikazana** (admin ne može ocenjivati)

### Očekivani rezultat:
- Korisnik može oceniti pesmu (1-5)
- Može izmeniti postojeću ocenu
- Može obrisati ocenu
- Admin ne vidi opciju za ocenjivanje

---

## 1.10 Pretplata na sadržaj (RK)

### Test 1: Pretplata na umetnika
1. **Prijavite se kao običan korisnik**
2. **Idite na stranicu umetnika**: `http://localhost:3000/artists/{artistId}`
3. **Kliknite na "🔔 Pretplati se"** dugme
4. **Proverite poruku o uspešnoj pretplati**
5. **Proverite da li se dugme promenilo** u "✓ Pretplaćen"
6. **Kliknite ponovo** da prekinete pretplatu
7. **Proverite da li se pretplata prekinula**

### Test 2: Pretplata na žanr
1. **Idite na**: `http://localhost:3000/songs`
2. **U sekciji "Pretraga i Filtriranje"** izaberite žanr iz dropdown-a (npr. "Pop")
3. **Proverite da li se pojavilo dugme "🔔 Pretplati se"** pored dropdown-a
4. **Kliknite na "🔔 Pretplati se"** dugme
5. **Proverite poruku o uspešnoj pretplati**
6. **Proverite da li se dugme promenilo** u "✓ Pretplaćen"
7. **Osvežite stranicu** (F5)
8. **Izaberite isti žanr ponovo** - proverite da li je dugme i dalje "✓ Pretplaćen"
9. **Kliknite ponovo na dugme** da prekinete pretplatu
10. **Proverite da li se pretplata prekinula** (dugme je sada "🔔 Pretplati se")

### Test 3: Pregled pretplata na profilu
1. **Idite na profil**: `http://localhost:3000/profile`
2. **Proverite sekciju "Moje pretplate"**
3. **Proverite da li se prikazuju** pretplate na umetnike i žanrove
4. **Kliknite na "Odjavi se"** pored neke pretplate
5. **Proverite da li se pretplata uklonila** sa liste

### Očekivani rezultat:
- Može se pretplatiti na umetnika
- Može se pretplatiti na žanr
- Može se prekinuti pretplata
- Pretplate su vidljive na profilu

---

## 2.5 Sinhrona komunikacija između servisa

### Preduslov: Kreirajte test podatke
1. **Kreirajte barem jednu pesmu i jednog umetnika** (kao admin)

### Test 1: Ocenjivanje pesme - validacija postojanja (sa postojećom pesmom)
1. **Prijavite se kao korisnik**
2. **Idite na stranicu postojeće pesme** (koju ste kreirali)
3. **Pokušajte da ocenite pesmu** - trebalo bi da radi normalno
4. **Proverite da li se ocena sačuvala**

### Test 2: Pretplata na umetnika - validacija postojanja (sa postojećim umetnikom)
1. **Prijavite se kao korisnik**
2. **Idite na stranicu postojećeg umetnika** (kojeg ste kreirali)
3. **Pokušajte da se pretplatite** - trebalo bi da radi normalno
4. **Proverite da li se pretplata sačuvala**

### Test 3: Provera logova (bez zaustavljanja servisa)
1. **Otvori terminal** sa Docker logovima: `docker-compose logs -f ratings-service`
2. **U drugom terminalu**: `docker-compose logs -f subscriptions-service`
3. **Pokušajte da ocenite pesmu** (frontend)
4. **Proverite logove ratings-service** - trebalo bi da vidi poziv ka content-service
5. **Pokušajte da se pretplatite na umetnika** (frontend)
6. **Proverite logove subscriptions-service** - trebalo bi da vidi poziv ka content-service

### Očekivani rezultat:
- Servisi proveravaju postojanje preko content-service
- Vraćaju se odgovarajuće greške kada resurs ne postoji
- Logovi pokazuju sinhronu komunikaciju

---

## 2.7 Otpornost na parcijalne otkaze sistema

### ⚠️ NAPOMENA: Ovi testovi zahtevaju zaustavljanje servisa
**Pre testiranja, uradite sledeće:**
1. **Kreirajte test podatke** (pesme, umetnike) kao admin
2. **Prijavite se kao običan korisnik**
3. **Otvori terminal za logove**: `docker-compose logs -f ratings-service subscriptions-service`

### Test 2.7.1 i 2.7.2: HTTP klijent konfiguracija i timeout

#### Test 1: Timeout test
1. **U novom terminalu, zaustavite content-service**: `docker-compose stop content-service`
2. **Pokušajte da ocenite pesmu** (frontend)
3. **Proverite da li zahtev ne čeka beskonačno** - treba da se završi sa greškom nakon ~2 sekunde
4. **Proverite logove ratings-service** - trebalo bi da vidi timeout grešku ili "fallback activated"
5. **Pokrenite content-service**: `docker-compose start content-service`

#### Test 2: Retry logika
1. **Zaustavite content-service**: `docker-compose stop content-service`
2. **Pokušajte da se pretplatite na umetnika** (frontend)
3. **Proverite logove subscriptions-service** - trebalo bi da vidi 2 retry pokušaja (attempt 1, attempt 2)
4. **Proverite da li se vraća greška** nakon retry pokušaja
5. **Pokrenite content-service**: `docker-compose start content-service`
6. **Ponovo pokušajte pretplatu** - trebalo bi da radi

### Test 2.7.3: Fallback logika

#### Test 1: Fallback kada servis nije dostupan
1. **Zaustavite content-service**: `docker-compose stop content-service`
2. **Pokušajte da ocenite pesmu** (frontend)
3. **Proverite da li se vraća greška** "Song not found" (fallback vraća false)
4. **Proverite logove ratings-service** - trebalo bi da vidi "fallback activated" ili "Content-service unavailable"
5. **Pokrenite content-service**: `docker-compose start content-service`

#### Test 2: Fallback za pretplatu
1. **Zaustavite content-service**: `docker-compose stop content-service`
2. **Pokušajte da se pretplatite na umetnika** (frontend)
3. **Proverite da li se vraća greška** "Artist not found" (fallback)
4. **Proverite logove subscriptions-service** - trebalo bi da vidi "fallback activated" ili "Content-service unavailable"
5. **Pokrenite content-service**: `docker-compose start content-service`

### Test 2.7.4: Circuit breaker

#### Test 1: Circuit breaker otvaranje
1. **Zaustavite content-service**: `docker-compose stop content-service`
2. **Pokušajte da ocenite pesmu 3 puta** (frontend, 3 puta zaredom)
3. **Proverite logove ratings-service** - trebalo bi da vidi:
   - Prva 3 neuspeha (failures: 1, 2, 3)
   - "Circuit breaker opened after 3 failures"
4. **Pokušajte da ocenite ponovo** (4. put)
5. **Proverite logove** - trebalo bi da vidi "Circuit breaker is open"
6. **Proverite da li se vraća greška** "Service temporarily unavailable - circuit breaker open"
7. **Sačekajte 5 sekundi**
8. **Pokušajte ponovo** (5. put)
9. **Proverite logove** - trebalo bi da vidi "Circuit breaker transitioning to half-open"

#### Test 2: Circuit breaker reset
1. **Nakon što je circuit breaker otvoren, sačekajte 5 sekundi**
2. **Pokrenite content-service**: `docker-compose start content-service`
3. **Pokušajte da ocenite pesmu** (frontend)
4. **Proverite logove ratings-service** - trebalo bi da vidi "Circuit breaker closed again" (ako je uspešno)
5. **Proverite da li se ocena sačuvala**

#### Test 3: Circuit breaker u subscriptions-service
1. **Zaustavite content-service**: `docker-compose stop content-service`
2. **Pokušajte da se pretplatite na umetnika 3 puta** (frontend, 3 puta zaredom)
3. **Proverite logove subscriptions-service** - trebalo bi da vidi circuit breaker otvaranje
4. **Proverite da li se vraća greška** nakon 3. pokušaja
5. **Pokrenite content-service**: `docker-compose start content-service`

### Očekivani rezultat:
- Timeout-ovi rade (zahtevi se ne blokiraju beskonačno)
- Retry logika radi (2 pokušaja)
- Fallback logika radi (vraća bezbednu vrednost)
- Circuit breaker se otvara nakon 3 neuspeha
- Circuit breaker se resetuje nakon 5 sekundi
- Ostali servisi nastavljaju da rade normalno kada jedan servis nije dostupan

---

## Brzi test checklist

- [ ] 1.7 - AudioPlayer se prikazuje i reprodukuje pesme
- [ ] 1.8 - Pretraga i filtriranje rade na Songs, Artists, Albums
- [ ] 1.9 - Ocenjivanje pesama radi (dodavanje, izmena, brisanje)
- [ ] 1.10 - Pretplata na umetnike i žanrove radi
- [ ] 2.5 - Sinhrona validacija radi (provera postojanja preko content-service)
- [ ] 2.7.1 - HTTP klijent je pravilno konfigurisan
- [ ] 2.7.2 - Timeout-ovi rade (zahtevi se ne blokiraju)
- [ ] 2.7.3 - Fallback logika radi kada servis nije dostupan
- [ ] 2.7.4 - Circuit breaker se otvara i resetuje pravilno

---

## Napomene

- Za testiranje otpornosti, koristite `docker-compose stop/start` za zaustavljanje/pokretanje servisa
- Proverite logove servisa u Docker kontejnerima: `docker-compose logs -f [service-name]`
- Za testiranje sa nepostojećim ID-jevima, koristite UUID format (npr. `00000000-0000-0000-0000-000000000000`)
