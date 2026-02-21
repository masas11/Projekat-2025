# Testiranje Zahteva 2.5 i 2.7

## Pregled

Ovaj dokument opisuje kako da testirate zahteve:
- **2.5**: Sinhrona komunikacija između servisa
- **2.7**: Otpornost na parcijalne otkaze sistema
  - **2.7.1**: Konfiguracija HTTP klijenta
  - **2.7.2**: Postavljanje timeout-a na nivou zahteva
  - **2.7.3**: Fallback logika kada servis ne vrati odgovor
  - **2.7.4**: Circuit Breaker

## Preduslovi

1. Svi servisi su pokrenuti: `docker-compose up -d`
2. Postoje test podaci (umetnici, pesme, albumi) - kreirajte ih kao admin
3. Imate korisnički nalog sa RK (Regular User) rolom

## Test 2.5: Sinhrona komunikacija

### Cilj
Proveriti da servisi sinhrono komuniciraju pre nego što izvrše akciju.

### Test koraci

#### Test 2.5.1: Pretplata na umetnika (sinhrona komunikacija)

1. **Pokrenite sve servise** (uključujući `content-service`):
   ```powershell
   docker-compose up -d
   ```

2. **Ulogujte se kao korisnik (RK role)** preko frontenda

3. **Otvorite logove za subscriptions-service**:
   ```powershell
   docker-compose logs -f subscriptions-service
   ```

4. **Pretplatite se na umetnika** preko frontenda (Artists stranica -> kliknite na zvonce kod umetnika)

5. **Proverite logove** - trebalo bi da vidite:
   ```
   Retrying call to content-service for artist <artistID>... (attempt 1)
   ```
   ili
   ```
   checkArtistExists pozvan za artist <artistID>
   ```

#### Test 2.5.2: Ocenjivanje pesme (sinhrona komunikacija)

1. **Otvorite logove za ratings-service**:
   ```powershell
   docker-compose logs -f ratings-service
   ```

2. **Ocenite pesmu** preko frontenda (Songs stranica -> otvorite pesmu -> ocenite)

3. **Proverite logove** - trebalo bi da vidite:
   ```
   Retrying call to content-service for song <songID>... (attempt 1)
   ```
   ili
   ```
   checkSpecificSongExists pozvan za song <songID>
   ```

## Test 2.7: Otpornost na parcijalne otkaze sistema

### Test 2.7.1: HTTP Client konfiguracija

**Cilj**: Proveriti da je HTTP klijent pravilno konfigurisan.

**Provera**: Otvorite kod:
- `services/subscriptions-service/cmd/main.go` (oko linije 150-170)
- `services/ratings-service/cmd/main.go` (oko linije 150-170)

**Trebalo bi da vidite**:
```go
transport := &http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    MaxIdleConns:    100,
    IdleConnTimeout: 90 * time.Second,
}
client := &http.Client{
    Transport: transport,
    Timeout:   2 * time.Second, // 2.7.2 - Timeout
}
```

**Status**: ✅ Implementirano u kodu

### Test 2.7.2: Timeout na nivou zahteva

**Cilj**: Proveriti da zahtevi imaju timeout od 2 sekunde.

#### Test koraci

1. **Zaustavite content-service**:
   ```powershell
   docker-compose stop content-service
   ```

2. **Otvorite logove za subscriptions-service**:
   ```powershell
   docker-compose logs -f subscriptions-service
   ```

3. **Pokušajte da se pretplatite na umetnika** preko frontenda

4. **Proverite logove** - trebalo bi da vidite:
   ```
   Retrying call to content-service for artist <artistID>... (attempt 1)
   ```
   Nakon ~2 sekunde, zahtev će timeout-ovati i aktivirati fallback.

5. **Ponovite za ratings-service**:
   ```powershell
   docker-compose logs -f ratings-service
   ```
   Pokušajte da ocenite pesmu i proverite logove.

**Status**: ✅ Implementirano - timeout od 2 sekunde u `context.WithTimeout`

### Test 2.7.3: Fallback logika

**Cilj**: Proveriti da servisi vraćaju bezbedan odgovor kada drugi servis nije dostupan.

#### Test koraci

1. **Zaustavite content-service**:
   ```powershell
   docker-compose stop content-service
   ```

2. **Otvorite logove za subscriptions-service**:
   ```powershell
   docker-compose logs -f subscriptions-service
   ```

3. **Pokušajte da se pretplatite na umetnika** preko frontenda

4. **Proverite logove** - trebalo bi da vidite:
   ```
   Error checking artist <artistID>: <error>, using fallback
   Circuit breaker open for artist <artistID>, using fallback
   Content-service unavailable for artist <artistID>, fallback activated - assuming artist does not exist
   ```

5. **Ponovite za ratings-service**:
   ```powershell
   docker-compose logs -f ratings-service
   ```
   Pokušajte da ocenite pesmu i proverite logove:
   ```
   Content-service unavailable for song <songID>, fallback activated - assuming song does not exist
   ```

**Fallback ponašanje**:
- `subscriptions-service`: Vraća `false` (umetnik ne postoji) -> pretplata neće biti kreirana
- `ratings-service`: Vraća `false` (pesma ne postoji) -> ocena neće biti sačuvana

**Status**: ✅ Implementirano - fallback vraća `false` kada servis nije dostupan

### Test 2.7.4: Circuit Breaker

**Cilj**: Proveriti da Circuit Breaker sprečava preopterećenje nefunkcionalnog servisa.

#### Test koraci

1. **Zaustavite content-service**:
   ```powershell
   docker-compose stop content-service
   ```

2. **Otvorite logove za subscriptions-service**:
   ```powershell
   docker-compose logs -f subscriptions-service
   ```

3. **Pokušajte da se pretplatite na umetnika 3+ puta** preko frontenda (ili curl komandama)

4. **Proverite logove** - trebalo bi da vidite sledeće stanja:

   **Prvi zahtev** (Circuit Breaker: `closed`):
   ```
   Circuit breaker state: closed, failures: 0
   Retrying call to content-service for artist <artistID>... (attempt 1)
   Function failed, failures: 1/3
   ```

   **Drugi zahtev** (Circuit Breaker: još uvek `closed`):
   ```
   Circuit breaker state: closed, failures: 1
   Retrying call to content-service for artist <artistID>... (attempt 1)
   Function failed, failures: 2/3
   ```

   **Treći zahtev** (Circuit Breaker: `open`):
   ```
   Circuit breaker state: closed, failures: 2
   Retrying call to content-service for artist <artistID>... (attempt 1)
   Function failed, failures: 3/3
   Circuit breaker opened after 3 failures
   ```

   **Četvrti zahtev** (Circuit Breaker: `open` - odmah vraća grešku):
   ```
   Circuit breaker is open
   Circuit breaker open for artist <artistID>, using fallback
   ```

5. **Ponovite za ratings-service**:
   ```powershell
   docker-compose logs -f ratings-service
   ```
   Pokušajte da ocenite pesmu 3+ puta i proverite logove.

6. **Ponovo pokrenite content-service**:
   ```powershell
   docker-compose start content-service
   ```

7. **Sačekajte 5 sekundi** (reset timeout za Circuit Breaker)

8. **Pokušajte ponovo da se pretplatite/ocenite** - Circuit Breaker bi trebalo da pređe u `half-open` stanje:
   ```
   Circuit breaker transitioning to half-open
   ```

   Ako zahtev uspe, Circuit Breaker se vraća u `closed`:
   ```
   Circuit breaker closed again
   ```

**Circuit Breaker konfiguracija**:
- **Max failures**: 3
- **Reset timeout**: 5 sekundi
- **Stanja**: `closed` -> `open` -> `half-open` -> `closed`

**Status**: ✅ Implementirano - Circuit Breaker sa 3 failure threshold i 5 sekundi reset timeout

## Kompletan test scenario

### Korak 1: Priprema

```powershell
# Pokrenite sve servise
docker-compose up -d

# Proverite da su svi servisi pokrenuti
docker-compose ps
```

### Korak 2: Test sinhronog komunikacije (2.5)

```powershell
# Otvorite logove u posebnom terminalu
docker-compose logs -f subscriptions-service ratings-service

# U drugom terminalu, testirajte pretplatu i ocenjivanje preko frontenda
# Proverite logove za pozive checkArtistExists i checkSpecificSongExists
```

### Korak 3: Test otpornosti (2.7)

```powershell
# Zaustavite content-service
docker-compose stop content-service

# Testirajte pretplatu i ocenjivanje preko frontenda
# Proverite logove za:
# - Retry logiku (2.7.2)
# - Fallback logiku (2.7.3)
# - Circuit Breaker (2.7.4)

# Ponovo pokrenite content-service
docker-compose start content-service
```

## Očekivani rezultati

### ✅ Test 2.5: Sinhrona komunikacija
- Logovi pokazuju pozive `checkArtistExists` i `checkSpecificSongExists`
- Servisi proveravaju postojanje pre nego što izvrše akciju

### ✅ Test 2.7.1: HTTP Client konfiguracija
- Kod sadrži `TLSClientConfig`, `MaxIdleConns`, `IdleConnTimeout`
- HTTP klijent ima `Timeout` od 2 sekunde

### ✅ Test 2.7.2: Timeout
- Logovi pokazuju retry pokušaje
- Zahtevi timeout-uju nakon ~2 sekunde

### ✅ Test 2.7.3: Fallback
- Logovi pokazuju "fallback activated" poruke
- Servisi vraćaju bezbedan odgovor (`false`) kada servis nije dostupan

### ✅ Test 2.7.4: Circuit Breaker
- Logovi pokazuju promene stanja: `closed` -> `open` -> `half-open` -> `closed`
- Nakon 3 neuspešna zahteva, Circuit Breaker se otvara
- Nakon 5 sekundi, Circuit Breaker prelazi u `half-open` stanje

## Troubleshooting

### Problem: Ne vidim logove
**Rešenje**: Proverite da su servisi pokrenuti: `docker-compose ps`

### Problem: Circuit Breaker se ne otvara
**Rešenje**: Pokušajte 3+ puta da se pretplatite/ocenite dok je `content-service` zaustavljen

### Problem: Ne vidim retry logiku
**Rešenje**: Proverite da je `content-service` zaista zaustavljen: `docker-compose ps`

### Problem: Fallback ne radi
**Rešenje**: Proverite kod u `checkArtistExists` i `checkSpecificSongExists` funkcijama - trebalo bi da vraćaju `false` kada servis nije dostupan

## Dodatne napomene

- **Sinhrona komunikacija (2.5)**: Implementirana u `subscriptions-service` (pretplata na umetnika) i `ratings-service` (ocenjivanje pesme)
- **Resilience mehanizmi (2.7)**: Implementirani u oba servisa sa istim konfiguracijama:
  - HTTP Client sa timeout-om od 2 sekunde
  - Retry logika (2 pokušaja sa 100ms delay)
  - Fallback logika (vraća `false`)
  - Circuit Breaker (3 failures, 5 sekundi reset timeout)
