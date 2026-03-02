# 📚 Vodič za Odbranu Projekta - Music Streaming Platform

**Tim:** 4 člana  
**Predmeti:** SOA, NoSQL, Informaciona bezbednost  
**Ocena:** 10+ (bez Kubernetes-a)

---

## 🎯 Kako koristiti ovaj vodič

Za svaki zahtev imate:
- **Šta je implementirano** - kratko objašnjenje
- **Kako radi** - tok implementacije
- **Detaljna implementacija** - konkretan kod sa objašnjenjima
- **Logika rada** - kako funkcioniše mehanizam korak po korak
- **Gde je kod** - lokacija u projektu
- **Kako pokazati** - demonstracija na odbrani
- **Šta profesor može da pita** - ključne tačke za objašnjenje

---

## ✅ OCENA 6 - Osnovne funkcionalnosti

### 1.1 Registracija naloga
**Šta:** Korisnik se registruje sa username, email, password  
**Kako:** 
- Frontend šalje POST `/api/users/register`
- Users-service validira podatke, hešira lozinku (bcrypt)
- Šalje verifikacioni email preko MailHog-a
- Korisnik klikne link → email verified

**Kod:**
- `services/users-service/internal/handler/auth_handler.go` - `Register()`
- `services/users-service/internal/service/password_service.go` - heširanje

**Kako pokazati:**
1. Otvori frontend → Register
2. Popuni formu → Register
3. Otvori MailHog (http://localhost:8025) → vidi email
4. Klikni na link → email verified

---

### 1.2 Prijava na sistem (OTP)
**Šta:** Kombinovana autentifikacija - password + OTP  
**Kako:**
- Korisnik unosi username/password
- Sistem šalje OTP na email
- Korisnik unosi OTP → dobija JWT token

**Kod:**
- `services/users-service/internal/handler/auth_handler.go` - `RequestOTP()`, `VerifyOTP()`

**Kako pokazati:**
1. Login → unesi username/password
2. Proveri MailHog → vidi OTP email
3. Unesi OTP → uspešna prijava

---

### 1.3 Kreiranje i izmena umetnika (Admin)
**Šta:** Admin kreira/izmenjuje umetnike (ime, biografija, žanrovi)  
**Kako:**
- Frontend šalje POST/PUT `/api/content/artists`
- Content-service čuva u MongoDB

**Kod:**
- `services/content-service/internal/handler/artist_handler.go` - `CreateArtist()`, `UpdateArtist()`

**Kako pokazati:**
1. Login kao admin
2. Artists → Add Artist
3. Popuni formu → Create
4. Umetnik se pojavljuje u listi

---

### 1.4 Kreiranje albuma i pesama (Admin)
**Šta:** Admin kreira albume i pesme  
**Kako:**
- Album se kreira prvo (vezan za umetnika)
- Pesma se kreira nakon albuma
- Sve u MongoDB (Content-service)

**Kod:**
- `services/content-service/internal/handler/album_handler.go`
- `services/content-service/internal/handler/song_handler.go`

**Kako pokazati:**
1. Artists → klikni umetnika → Add Album
2. Albums → klikni album → Add Song
3. Sve se prikazuje u hijerarhiji

---

### 1.5 Pregled umetnika, albuma i pesama
**Šta:** Korisnici pregledaju sadržaj  
**Kako:**
- Frontend poziva GET `/api/content/artists`, `/albums`, `/songs`
- Content-service vraća podatke iz MongoDB

**Kod:**
- `services/content-service/internal/handler/` - svi GET handleri

**Kako pokazati:**
1. Otvori frontend
2. Klikni Artists → vidi listu
3. Klikni umetnika → vidi albume
4. Klikni album → vidi pesme

---

### 1.11 Pregled notifikacija
**Šta:** Korisnik vidi sve notifikacije  
**Kako:**
- Notifications-service čuva u Cassandra
- Frontend poziva GET `/api/notifications`

**Kod:**
- `services/notifications-service/internal/handler/notification_handler.go`

**Kako pokazati:**
1. Login kao korisnik
2. Notifications → vidi listu notifikacija

---

### 2.1 Dizajn sistema
**Šta:** Model podataka i komunikacija između servisa  
**Kako:**
- Svaki servis ima svoje modele u `internal/model/`
- REST API za komunikaciju
- Event-based komunikacija za asinhrone operacije

**Kod:**
- `services/*/internal/model/` - modeli za svaki servis
- `services/*/internal/dto/` - DTO objekti

**Kako pokazati:**
- Pokaži strukturu projekta
- Pokaži modele u kodu

---

### 2.2 API Gateway
**Šta:** Jedinstvena ulazna tačka za sve zahteve  
**Kako:**
- API Gateway prima sve zahteve
- Prosleđuje ih odgovarajućim servisima
- Dodaje CORS, autentifikaciju, rate limiting

**Kod:**
- `services/api-gateway/cmd/main.go` - `proxyRequest()`

**Kako pokazati:**
1. Pokaži da svi zahtevi idu kroz `http://localhost:8081`
2. Pokaži routing logiku u kodu

---

### 2.3 Kontejnerizacija (Docker)
**Šta:** Svi servisi u Docker kontejnerima  
**Kako:**
- `docker-compose.yml` definiše sve servise
- Svaki servis ima `Dockerfile`

**Kod:**
- `docker-compose.yml`
- `services/*/Dockerfile`

**Kako pokazati:**
```powershell
docker-compose ps  # Pokaži sve kontejnere
docker-compose logs api-gateway  # Pokaži logove
```

---

### 2.4 Eksterna konfiguracija
**Šta:** Konfiguracija u env fajlovima, ne u kodu  
**Kako:**
- Svaki servis ima `config/config.go`
- Čita iz environment varijabli ili `.env` fajlova

**Kod:**
- `services/*/config/config.go`

**Kako pokazati:**
- Pokaži `docker-compose.yml` - environment varijable
- Pokaži kako servisi čitaju konfiguraciju

---

## ✅ OCENA 7 - Napredne funkcionalnosti

### 1.6 Reprodukcija pesme
**Šta:** Korisnik pušta pesme u browseru  
**Kako:**
- Frontend poziva GET `/api/content/songs/{id}/stream`
- Content-service streamuje audio iz HDFS-a ili vraća URL
- HTML5 audio player reprodukuje

**Kod:**
- `services/content-service/internal/handler/song_handler.go` - `StreamSong()`

**Kako pokazati:**
1. Songs → klikni pesmu
2. Klikni Play
3. Pesma se reprodukuje

---

### 1.7 Filtriranje i pretraga
**Šta:** Pretraga po nazivu, filtriranje po žanru  
**Kako:**
- Frontend šalje query parametre (`?search=...`, `?genre=...`)
- Content-service filtrira u MongoDB upitu

**Kod:**
- `services/content-service/internal/handler/artist_handler.go` - `GetArtists()`
- `services/content-service/internal/handler/song_handler.go` - `GetSongs()`

**Kako pokazati:**
1. Artists → unesi tekst u search box
2. Artists → izaberi žanr iz dropdown
3. Rezultati se filtriraju

---

### 1.8 Ocenjivanje pesama
**Šta:** Korisnik ocenjuje pesme (1-5)  
**Kako:**
- Frontend šalje POST `/api/ratings/rate-song`
- Ratings-service čuva u MongoDB
- Sinhrona komunikacija sa Content-service (provera da pesma postoji)

**Kod:**
- `services/ratings-service/cmd/main.go` - `/rate-song` handler
- Koristi circuit breaker i retry za Content-service

**Kako pokazati:**
1. Songs → klikni pesmu
2. Klikni zvezdu (1-5)
3. Ocena se čuva i prikazuje

---

### 1.9 Pretplata na umetnika i žanrove
**Šta:** Korisnik se pretplaćuje na umetnike/žanrove  
**Kako:**
- Frontend šalje POST `/api/subscriptions/subscribe-artist` ili `/subscribe-genre`
- Subscriptions-service čuva u MongoDB
- Sinhrona komunikacija sa Content-service (provera da umetnik/žanr postoje)

**Kod:**
- `services/subscriptions-service/cmd/main.go` - subscribe handleri

**Kako pokazati:**
1. Artists → klikni umetnika → Subscribe
2. Profile → Subscriptions → Subscribe to Genre
3. Pretplate se čuvaju

---

### 2.5 Sinhrona komunikacija
**Šta:** Servisi pozivaju druge servise direktno (HTTP)  
**Kako:**
- Ratings-service poziva Content-service da proveri da pesma postoji
- Subscriptions-service poziva Content-service da proveri da umetnik postoji
- Koristi HTTP klijent sa timeout-om

**Kod:**
- `services/ratings-service/cmd/main.go` - `checkSpecificSongExists()`
- `services/subscriptions-service/cmd/main.go` - `checkArtistExists()`

**Kako pokazati:**
- Pokaži kod gde se poziva drugi servis
- Pokaži logove kada se poziva

---

### 2.7.1-2.7.4 Otpornost na otkaze (osnovno)
**Šta:** HTTP klijent konfiguracija, timeout, fallback, circuit breaker  
**Kako:**
- **2.7.1:** HTTP Transport sa TLS, MaxIdleConns, IdleConnTimeout
- **2.7.2:** Timeout na HTTP klijentu (2s za ratings, 5s za subscriptions)
- **2.7.3:** Fallback logika - vraća default vrednost ako servis nije dostupan
- **2.7.4:** Circuit breaker - otvara se nakon 3 neuspeha

**Detaljna implementacija:**

#### 2.7.1 - Konfiguracija HTTP klijenta
**Kod:** `services/ratings-service/cmd/main.go` (linija 324-333)
```go
// HTTP Transport konfiguracija
tr := &http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},  // TLS podrška
    MaxIdleConns:    10,                                      // Maksimalno 10 idle konekcija
    IdleConnTimeout: 30 * time.Second,                        // Timeout za idle konekcije
}
clientHTTP := &http.Client{
    Timeout:   2 * time.Second,  // Request timeout (2.7.2)
    Transport: tr,
}
```
**Objašnjenje:**
- `TLSClientConfig` - omogućava HTTPS komunikaciju između servisa
- `MaxIdleConns` - maksimalan broj idle konekcija u pool-u (optimizacija performansi)
- `IdleConnTimeout` - nakon 30s neaktivnosti, konekcija se zatvara
- `Timeout` - maksimalno vreme čekanja na odgovor (2 sekunde)

#### 2.7.2 - Timeout na nivou zahteva
**Kod:** `services/ratings-service/cmd/main.go` (linija 223-224)
```go
reqCtx, reqCancel := context.WithTimeout(ctx, 2*time.Second)
defer reqCancel()
req, err := http.NewRequestWithContext(reqCtx, "GET", checkURL, nil)
```
**Objašnjenje:**
- `context.WithTimeout` - kreira context sa timeout-om od 2 sekunde
- Ako zahtev ne završi u roku, context se automatski otkazuje
- `reqCancel()` - oslobađa resurse nakon završetka

#### 2.7.3 - Fallback logika
**Kod:** `services/ratings-service/cmd/main.go` (linija 253-257)
```go
if err != nil {
    // Fallback logic (2.7.3): return false when service is unavailable
    log.Printf("Content-service unavailable for song %s after retries, fallback activated - assuming song does not exist. Last error: %v", songID, lastErr)
    return false  // Vraća false umesto da baci grešku
}
```
**Objašnjenje:**
- Ako svi retry pokušaji ne uspeju, umesto da baci grešku, vraća default vrednost (`false`)
- Sistem nastavlja da radi normalno, samo sa ograničenom funkcionalnošću
- Loguje se greška za debugging, ali korisnik dobija odgovor

#### 2.7.4 - Circuit Breaker
**Kod:** `services/shared/circuitbreaker/circuitbreaker.go`
```go
type CircuitBreaker struct {
    maxFailures    int           // 3 neuspeha
    resetTimeout   time.Duration // 5 sekundi
    state          State         // Closed, Open, HalfOpen
    failures       int           // Brojač neuspeha
    lastFailTime   time.Time     // Vreme poslednjeg neuspeha
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    // Proverava da li je circuit breaker otvoren
    if cb.state == Open && time.Since(cb.lastFailTime) > cb.resetTimeout {
        cb.setState(HalfOpen)  // Probni poziv
    }
    if cb.state == Open {
        return &CircuitBreakerError{Message: "circuit breaker is open"}
    }
    
    err := fn()
    if err != nil {
        cb.onFailure()  // Povećava brojač neuspeha
        return err
    }
    cb.onSuccess()  // Resetuje brojač
    return nil
}
```
**Objašnjenje:**
- **Closed** - normalan rad, svi pozivi prolaze
- **Open** - nakon 3 neuspeha, circuit breaker se otvara, svi pozivi se odbijaju
- **HalfOpen** - nakon 5 sekundi, probni poziv se šalje da proveri da li servis radi
- Ako probni poziv uspe → vraća se u **Closed**, ako ne → ostaje **Open**

**Kod:** `services/ratings-service/cmd/main.go` (linija 336)
```go
cb := NewCircuitBreaker(3, 5*time.Second) // 3 neuspeha, reset nakon 5s
```

**Kod:**
- `services/ratings-service/cmd/main.go` - HTTP klijent, circuit breaker
- `services/subscriptions-service/cmd/main.go` - HTTP klijent, circuit breaker
- `services/shared/circuitbreaker/circuitbreaker.go` - shared circuit breaker implementacija

**Kako pokazati:**
```powershell
.\test-resilience-kratko.ps1
```

**Šta profesor može da pita:**
- "Zašto koristite circuit breaker?" → **Odgovor:** "Sprečava preopterećenje servisa koji ne radi. Nakon 3 neuspeha, prestajemo da šaljemo zahteve na 5 sekundi, što daje servisu vreme da se oporavi."
- "Kako funkcioniše fallback?" → **Odgovor:** "Umesto da baci grešku korisniku, vraćamo default vrednost (npr. `false` za proveru postojanja pesme). Sistem nastavlja da radi, samo sa ograničenom funkcionalnošću."
- "Zašto različiti timeout-i?" → **Odgovor:** "Ratings-service ima 2s jer je brz poziv (samo provera postojanja), subscriptions-service ima 5s jer može da traje duže (dohvatanje imena umetnika)."

---

## ✅ OCENA 8 - Asinhrone operacije

### 1.10 Generisanje notifikacija
**Šta:** Notifikacije se generišu kada se doda novi sadržaj  
**Kako:**
- Content-service emituje event (`new_artist`, `new_album`, `new_song`)
- Subscriptions-service prima event i proverava pretplate
- Notifications-service kreira notifikaciju u Cassandra

**Kod:**
- `services/content-service/internal/events/emitter.go` - emisija eventova
- `services/subscriptions-service/cmd/main.go` - `handleNewSongEvent()`
- `services/notifications-service/` - kreiranje notifikacija

**Kako pokazati:**
1. Pretplati se na umetnika kao korisnik
2. Kao admin, kreiraj novu pesmu za tog umetnika
3. Sačekaj 5-10 sekundi
4. Notifikacija se pojavljuje u `/notifications`

---

### 1.12 Preporuke muzičkog sadržaja
**Šta:** Personalizovane preporuke na osnovu pretplata i ocena  
**Kako:**
- Recommendation-service koristi Neo4j graf bazu
- Algoritam: pesme iz pretplaćenih žanrova (ocena >= 4) + top pesma iz nepretplaćenog žanra
- Ažurira se asinhrono preko eventova

**Kod:**
- `services/recommendation-service/cmd/main.go` - recommendation logika
- `services/recommendation-service/internal/store/neo4j.go` - Neo4j upiti

**Kako pokazati:**
1. Pretplati se na žanr, oceni neke pesme
2. Home stranica → prikazuju se preporuke umesto svih umetnika

---

### 2.6 Asinhrona komunikacija
**Šta:** Servisi komuniciraju preko eventova (HTTP POST)  
**Kako:**
- Content-service emituje event kada se kreira novi sadržaj
- Subscriptions-service i Recommendation-service reaguju na evente
- Eventi se šalju asinhrono (goroutine)

**Kod:**
- `services/content-service/internal/events/emitter.go` - `EmitEvent()`
- `services/subscriptions-service/cmd/main.go` - event handleri
- `services/recommendation-service/cmd/main.go` - event handleri

**Kako pokazati:**
- Pokaži kod gde se emituje event
- Pokaži kod gde se prima event
- Pokaži logove kada se šalje/prima event

---

### 2.7.5-2.7.7 Otpornost na otkaze (napredno)
**Šta:** Retry mehanizam, timeout za korisnika, context cancellation  
**Kako:**
- **2.7.5:** Retry sa exponential backoff (3 pokušaja, 100ms → 200ms → 400ms)
- **2.7.6:** API Gateway vraća 408 Request Timeout nakon 5s
- **2.7.7:** Servisi prekidaju obradu ako context istekne

**Detaljna implementacija:**

#### 2.7.5 - Retry mehanizam sa exponential backoff
**Kod:** `services/ratings-service/cmd/main.go` (linija 163-202)
```go
func RetryWithExponentialBackoff(ctx context.Context, config RetryConfig, fn func() error) error {
    delay := config.InitialDelay  // 100ms
    
    for attempt := 0; attempt < config.MaxRetries; attempt++ {
        // Proverava da li je context otkazan (2.7.7)
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        
        err := fn()  // Izvršava funkciju
        if err == nil {
            return nil  // Uspeh!
        }
        
        // Ne retry-uj na poslednjem pokušaju
        if attempt < config.MaxRetries-1 {
            log.Printf("Retry attempt %d/%d failed: %v, retrying in %v...", attempt+1, config.MaxRetries, err, delay)
            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-time.After(delay):
                // Exponential backoff: 100ms → 200ms → 400ms
                delay = time.Duration(float64(delay) * config.BackoffMultiplier)  // 2.0
                if delay > config.MaxDelay {
                    delay = config.MaxDelay  // Maksimalno 2s
                }
            }
        }
    }
    return fmt.Errorf("all %d retry attempts failed", config.MaxRetries)
}
```
**Objašnjenje:**
- **Exponential backoff** - svaki sledeći pokušaj čeka duže (100ms → 200ms → 400ms)
- **Zašto?** - Daje servisu vreme da se oporavi, sprečava preopterećenje
- **Context cancellation** - ako context istekne, prekida se retry
- **3 pokušaja** - prvi + 2 retry pokušaja

**Kod:** `services/ratings-service/cmd/main.go` (linija 215-251)
```go
err := RetryWithExponentialBackoff(ctx, retryConfig, func() error {
    reqCtx, reqCancel := context.WithTimeout(ctx, 2*time.Second)
    defer reqCancel()
    
    req, err := http.NewRequestWithContext(reqCtx, "GET", checkURL, nil)
    resp, err := client.Do(req)
    // ... obrada odgovora
})
```

#### 2.7.6 - Eksplicitni timeout za korisnika
**Kod:** `services/api-gateway/cmd/main.go` (linija 60-151)
```go
func proxyRequest(w http.ResponseWriter, r *http.Request, targetURL string, timeout time.Duration, appLogger *logger.Logger) {
    // Kreira context sa timeout-om (5 sekundi)
    ctx, cancel := context.WithTimeout(r.Context(), timeout)
    defer cancel()
    
    // Pokreće zahtev u gorutini
    go func() {
        resp, err := client.Do(req)
        resultChan <- result{resp: resp, err: err}
    }()
    
    // Čeka na rezultat ili timeout
    select {
    case <-ctx.Done():
        // Timeout istekao - vraća 408 Request Timeout (2.7.6)
        log.Printf("Request timeout for %s: %v", targetURL, ctx.Err())
        enableCORS(w, r)
        w.WriteHeader(http.StatusRequestTimeout)  // 408
        w.Write([]byte("Request timeout - service did not respond in time"))
        return
    case res := <-resultChan:
        // Odgovor stigao na vreme
        resp = res.resp
        err = res.err
    }
}
```
**Objašnjenje:**
- API Gateway postavlja **maksimalno vreme čekanja** od 5 sekundi
- Ako servis ne odgovori u roku, korisnik dobija **408 Request Timeout**
- Korisnik ne čeka beskonačno, već dobija jasnu poruku o timeout-u

#### 2.7.7 - Upstream servis odustaje od obrade
**Kod:** `services/ratings-service/cmd/main.go` (linija 168-172, 217-220)
```go
// Proverava da li je context otkazan pre svakog retry pokušaja
select {
case <-ctx.Done():
    log.Printf("Context cancelled during retry attempt %d: %v", attempt+1, ctx.Err())
    return ctx.Err()  // Prekida obradu
default:
}

// Proverava i tokom čekanja na retry delay
select {
case <-ctx.Done():
    log.Printf("Context cancelled during retry delay: %v", ctx.Err())
    return ctx.Err()
case <-time.After(delay):
    // Nastavlja sa retry-om
}
```
**Objašnjenje:**
- Servis **kontinuirano proverava** da li je context otkazan
- Ako API Gateway pošalje timeout (2.7.6), context se otkazuje
- Servis **odmah prekida obradu** umesto da troši resurse na nepotrebne retry-eve
- **Zašto je važno?** - Oslobađa resurse (konekcije, memoriju) kada korisnik više ne čeka

**Kod:**
- `services/ratings-service/cmd/main.go` - `RetryWithExponentialBackoff()` (linija 163-202)
- `services/api-gateway/cmd/main.go` - `proxyRequest()` timeout logika (linija 60-151)
- `services/ratings-service/cmd/main.go` - context cancellation provere (linija 168-172, 217-220)

**Kako pokazati:**
```powershell
.\test-resilience-kratko.ps1
```

**Šta profesor može da pita:**
- "Zašto exponential backoff?" → **Odgovor:** "Sprečava preopterećenje servisa. Ako servis ne radi, čekamo duže između pokušaja (100ms → 200ms → 400ms), dajući mu vreme da se oporavi."
- "Kako API Gateway zna kada da vrati timeout?" → **Odgovor:** "Koristimo `context.WithTimeout` sa 5 sekundi. Ako servis ne odgovori u roku, `select` statement hvata `ctx.Done()` i vraća 408."
- "Zašto je važno da servis proverava context?" → **Odgovor:** "Ako korisnik više ne čeka (timeout), nema smisla da servis troši resurse na retry-eve. Context cancellation omogućava servisu da odmah prekine obradu."

---

## ✅ OCENA 9 - Napredni šabloni

### 2.8 API Composition
**Šta:** API Gateway kompozuje podatke iz više servisa  
**Kako:**
- Kada se traži pesma, API Gateway poziva Content-service i Ratings-service
- Kombinuje podatke i vraća korisniku
- Prikazuje broj ocena i prosečnu ocenu uz pesmu

**Detaljna implementacija:**

**Kod:** `services/api-gateway/cmd/main.go` (linija 337-437)
```go
func composeSongWithRatings(w http.ResponseWriter, r *http.Request, songID string, cfg *config.Config, appLogger *logger.Logger) {
    // KORAK 1: Dohvati pesmu iz Content-service
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    
    contentURL := cfg.ContentServiceURL + "/songs/" + songID
    contentReq, err := http.NewRequestWithContext(ctx, "GET", contentURL, nil)
    contentResp, err := client.Do(contentReq)
    // ... parsiranje odgovora
    var song map[string]interface{}
    json.NewDecoder(contentResp.Body).Decode(&song)
    
    // KORAK 2: Dohvati ocene iz Ratings-service
    ratingURL := cfg.RatingsServiceURL + "/average-rating?songId=" + songID
    ratingReq, err := http.NewRequestWithContext(ratingCtx, "GET", ratingURL, nil)
    ratingResp, err := ratingClient.Do(ratingReq)
    
    if ratingResp.StatusCode == http.StatusOK {
        var ratingData map[string]interface{}
        json.NewDecoder(ratingResp.Body).Decode(&ratingData)
        avg, _ := ratingData["averageRating"].(float64)
        count, _ := ratingData["ratingCount"].(float64)
        song["averageRating"] = avg      // Dodaje prosečnu ocenu
        song["ratingCount"] = int(count)  // Dodaje broj ocena
    } else {
        // Fallback: ako ratings-service ne radi, vraća pesmu bez ocena
        song["averageRating"] = 0.0
        song["ratingCount"] = 0
    }
    
    // KORAK 3: Vrati kombinovani odgovor
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(song)
}
```

**Za više pesama (paralelno):**
**Kod:** `services/api-gateway/cmd/main.go` (linija 209-335)
```go
func composeSongsWithRatings(w http.ResponseWriter, r *http.Request, cfg *config.Config, appLogger *logger.Logger) {
    // KORAK 1: Dohvati sve pesme iz Content-service
    contentResp, err := client.Do(contentReq)
    var songs []map[string]interface{}
    json.NewDecoder(contentResp.Body).Decode(&songs)
    
    // KORAK 2: Za svaku pesmu, paralelno dohvati ocene (goroutines)
    ratingChan := make(chan ratingResult, len(songs))
    
    for i, song := range songs {
        go func(idx int, s map[string]interface{}) {
            songID, _ := s["id"].(string)
            ratingURL := cfg.RatingsServiceURL + "/average-rating?songId=" + songID
            ratingResp, err := ratingClient.Do(ratingReq)
            
            if ratingResp.StatusCode == http.StatusOK {
                var ratingData map[string]interface{}
                json.NewDecoder(ratingResp.Body).Decode(&ratingData)
                ratingChan <- ratingResult{
                    index: idx,
                    averageRating: ratingData["averageRating"].(float64),
                    ratingCount: int(ratingData["ratingCount"].(float64)),
                }
            } else {
                ratingChan <- ratingResult{index: idx, averageRating: 0, ratingCount: 0}
            }
        }(i, song)
    }
    
    // KORAK 3: Prikupi sve rezultate
    ratings := make([]ratingResult, len(songs))
    for i := 0; i < len(songs); i++ {
        ratings[i] = <-ratingChan
    }
    
    // KORAK 4: Kombinuj pesme sa ocenama
    for _, rating := range ratings {
        songs[rating.index]["averageRating"] = rating.averageRating
        songs[rating.index]["ratingCount"] = rating.ratingCount
    }
    
    // KORAK 5: Vrati kombinovani odgovor
    json.NewEncoder(w).Encode(songs)
}
```

**Objašnjenje:**
- **Za jednu pesmu:** API Gateway poziva Content-service i Ratings-service **sekvencijalno**, kombinuje rezultate
- **Za više pesama:** Koristi **goroutines** za paralelne pozive ka Ratings-service (brže!)
- **Fallback:** Ako Ratings-service ne radi, vraća pesme sa `averageRating: 0` i `ratingCount: 0`
- **Korisnik dobija jedan API poziv** umesto da frontend pravi više poziva

**Kod:**
- `services/api-gateway/cmd/main.go` - `composeSongWithRatings()` (linija 337-437)
- `services/api-gateway/cmd/main.go` - `composeSongsWithRatings()` (linija 209-335)

**Kako pokazati:**
1. Songs → klikni pesmu
2. Prikazuje se: naziv, album, umetnici, **broj ocena, prosečna ocena**
3. Sve iz jednog API poziva
4. Proveri Network tab - samo jedan poziv ka `/api/content/songs/{id}`

**Šta profesor može da pita:**
- "Zašto API Composition umesto da frontend poziva oba servisa?" → **Odgovor:** "Smanjuje broj HTTP zahteva, poboljšava performanse, i omogućava fallback logiku ako jedan servis ne radi."
- "Kako funkcioniše paralelno dohvatanje ocena?" → **Odgovor:** "Koristimo goroutines - za svaku pesmu pokrećemo goroutine koja poziva Ratings-service. Sve se izvršava paralelno, što je brže nego sekvencijalno."
- "Šta se dešava ako Ratings-service ne radi?" → **Odgovor:** "Fallback logika vraća pesme sa `averageRating: 0` i `ratingCount: 0`. Korisnik i dalje vidi pesme, samo bez ocena."

---

### 2.9 CQRS
**Šta:** Read model za pretplate - imena umetnika se keširaju  
**Kako:**
- Subscriptions-service čuva ime umetnika uz pretplatu (denormalizacija)
- Ne poziva Content-service svaki put
- Ažurira se kada se umetnik promeni

**Detaljna implementacija:**

**Kod:** `services/subscriptions-service/internal/model/subscription.go`
```go
type Subscription struct {
    ID        string    `bson:"_id" json:"id"`
    UserID    string    `bson:"userId" json:"userId"`
    ArtistID  string    `bson:"artistId" json:"artistId"`
    ArtistName string   `bson:"artistName" json:"artistName"`  // DENORMALIZACIJA (CQRS)
    Genre     string    `bson:"genre,omitempty" json:"genre,omitempty"`
    CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
```
**Objašnjenje:**
- `ArtistName` se čuva **uz pretplatu** (denormalizacija)
- Ne moramo da pozivamo Content-service svaki put kada prikazujemo pretplate
- **Read model** - optimizovan za čitanje

**Kod:** `services/subscriptions-service/cmd/main.go` (linija 151-216)
```go
func getArtistName(client *http.Client, contentURL, artistID string, cb *CircuitBreaker, ctx context.Context) (string, bool) {
    // Poziva Content-service samo ako ime nije u read modelu
    checkURL := contentURL + "/artists/" + url.QueryEscape(artistID)
    
    // Koristi circuit breaker i retry
    err := cb.Call(func() error {
        return RetryWithExponentialBackoff(ctx, retryConfig, func() error {
            req, err := http.NewRequestWithContext(reqCtx, "GET", checkURL, nil)
            resp, err := client.Do(req)
            
            if resp.StatusCode == http.StatusOK {
                var artist map[string]interface{}
                json.NewDecoder(resp.Body).Decode(&artist)
                if name, ok := artist["name"].(string); ok {
                    artistName = name  // Čuva ime za read model
                    exists = true
                    return nil
                }
            }
            return err
        })
    })
    
    return artistName, exists
}
```

**Kod:** `services/subscriptions-service/cmd/main.go` - kreiranje pretplate
```go
// Kada se kreira pretplata, dohvata se ime umetnika i čuva u read modelu
artistName, exists := getArtistName(client, contentURL, artistID, cb, ctx)
if !exists {
    return fmt.Errorf("artist not found")
}

subscription := &Subscription{
    UserID:     userID,
    ArtistID:    artistID,
    ArtistName: artistName,  // Čuva se u read modelu
    CreatedAt:  time.Now(),
}
// Čuva se u MongoDB sa imenom umetnika
```

**Kod:** `services/subscriptions-service/cmd/main.go` - prikaz pretplata
```go
// Kada se prikazuju pretplate, čita se direktno iz read modela
subscriptions, err := subscriptionStore.GetUserSubscriptions(userID)
// subscriptions[i].ArtistName je već tu - nema poziva ka Content-service!
```

**Objašnjenje:**
- **Write Side:** Kada se kreira pretplata, dohvata se ime umetnika iz Content-service i čuva u read modelu
- **Read Side:** Kada se prikazuju pretplate, čita se direktno iz read modela (brže!)
- **Ažuriranje:** Ako se umetnik promeni, trebalo bi da se ažurira read model (u našem slučaju, pretplate se ne ažuriraju automatski, ali bi trebalo)

**Kod:**
- `services/subscriptions-service/cmd/main.go` - `getArtistName()` (linija 151-216)
- `services/subscriptions-service/internal/model/subscription.go` - `Subscription` model sa `ArtistName`
- `services/subscriptions-service/cmd/main.go` - kreiranje pretplate sa čuvanjem imena

**Kako pokazati:**
1. Profile → Subscriptions
2. Prikazuju se imena umetnika (ne samo ID-ovi)
3. Proveri Network tab - nema ponovnih poziva ka Content-service
4. Proveri MongoDB: `db.subscriptions.find()` - vidiš da `artistName` postoji u dokumentu

**Šta profesor može da pita:**
- "Zašto CQRS umesto da svaki put pozivate Content-service?" → **Odgovor:** "Poboljšava performanse - ne moramo da čekamo HTTP poziv svaki put. Read model je optimizovan za čitanje, čuva podatke koji su često potrebni."
- "Šta je denormalizacija?" → **Odgovor:** "Čuvanje podataka na više mesta. Umesto da čuvamo samo `artistId`, čuvamo i `artistName` uz pretplatu. To je trade-off između performansi i konzistentnosti."
- "Kako ažurirate read model kada se umetnik promeni?" → **Odgovor:** "Trenutno se ne ažurira automatski, ali bi trebalo da Content-service emituje event kada se umetnik promeni, a Subscriptions-service ažurira read model."

---

### 2.10 Tracing (Jaeger)
**Šta:** Distributed tracing kroz sve servise  
**Kako:**
- OpenTelemetry + Jaeger
- Svaki HTTP zahtev ima trace ID
- Trace se propagira kroz sve servise

**Kod:**
- `services/shared/tracing/tracing.go` - inicijalizacija
- `services/shared/tracing/http.go` - HTTP middleware

**Kako pokazati:**
1. Pokreni: `.\test-tracing-2.10.ps1`
2. Otvori Jaeger UI: http://localhost:16686
3. Klikni Find Traces → vidiš trace kroz sve servise

---

### 2.11 HDFS
**Šta:** Audio fajlovi se čuvaju na HDFS-u  
**Kako:**
- Content-service uploaduje audio fajlove na HDFS preko WebHDFS API-ja
- Stream endpoint čita iz HDFS-a
- Redis keš za najslušanije pesme

**Detaljna implementacija:**

**Kod:** `services/content-service/internal/storage/hdfs.go` (linija 49-137)
```go
func (c *HDFSClient) UploadFile(localPath, hdfsPath string) error {
    // KORAK 1: Otvara lokalni fajl
    file, err := os.Open(localPath)
    defer file.Close()
    
    // KORAK 2: Kreira HDFS direktorijum ako ne postoji
    dir := filepath.Dir(hdfsPath)
    if dir != "." && dir != "/" {
        c.Mkdir(dir, true)
    }
    
    // KORAK 3: WebHDFS CREATE operacija (redirect pattern)
    createURL := fmt.Sprintf("%s/webhdfs/v1%s?op=CREATE&overwrite=true&user.name=root", c.baseURL, hdfsPath)
    req, err := http.NewRequest("PUT", createURL, nil)
    resp, err := c.httpClient.Do(req)
    
    // KORAK 4: HDFS vraća 307 redirect sa Location header-om
    if resp.StatusCode == http.StatusTemporaryRedirect {
        redirectURL := resp.Header.Get("Location")
        
        // KORAK 5: Upload fajla na redirect URL (DataNode)
        fileData, err := io.ReadAll(file)
        req, err = http.NewRequest("PUT", redirectURL, bytes.NewReader(fileData))
        resp, err = c.httpClient.Do(req)
        
        if resp.StatusCode == http.StatusCreated {
            return nil  // Uspešno!
        }
    }
    return fmt.Errorf("upload failed")
}
```

**Kod:** `services/content-service/internal/storage/hdfs.go` (linija 262-321)
```go
func (c *HDFSClient) DownloadFile(hdfsPath string) ([]byte, error) {
    // KORAK 1: WebHDFS OPEN operacija
    openURL := fmt.Sprintf("%s/webhdfs/v1%s?op=OPEN&user.name=root", c.baseURL, hdfsPath)
    req, err := http.NewRequest("GET", openURL, nil)
    resp, err := c.httpClient.Do(req)
    
    // KORAK 2: HDFS vraća 307 redirect
    if resp.StatusCode == http.StatusTemporaryRedirect {
        redirectURL := resp.Header.Get("Location")
        
        // KORAK 3: Download fajla sa DataNode-a
        req, err = http.NewRequest("GET", redirectURL, nil)
        resp, err = c.httpClient.Do(req)
        
        // KORAK 4: Čita fajl u memoriju
        data, err := io.ReadAll(resp.Body)
        return data, nil
    }
    return nil, fmt.Errorf("download failed")
}
```

**Kod:** `services/content-service/internal/handler/song_handler.go` - Stream
```go
func (h *SongHandler) StreamSong(w http.ResponseWriter, r *http.Request) {
    songID := mux.Vars(r)["id"]
    
    // Dohvata pesmu iz MongoDB
    song, err := h.songStore.GetSongByID(songID)
    if err != nil {
        http.Error(w, "Song not found", http.StatusNotFound)
        return
    }
    
    // HDFS path za audio fajl
    hdfsPath := fmt.Sprintf("/audio/songs/%s.mp3", songID)
    
    // Download fajla iz HDFS-a
    audioData, err := h.hdfsClient.DownloadFile(hdfsPath)
    if err != nil {
        http.Error(w, "Failed to download audio", http.StatusInternalServerError)
        return
    }
    
    // Stream-uje audio korisniku
    w.Header().Set("Content-Type", "audio/mpeg")
    w.Header().Set("Content-Length", strconv.Itoa(len(audioData)))
    w.Write(audioData)
}
```

**Objašnjenje:**
- **WebHDFS REST API** - HDFS nudi REST API za upload/download
- **Redirect pattern** - NameNode vraća 307 redirect sa Location header-om koji pokazuje na DataNode
- **Two-step upload** - prvo CREATE zahtev (dobija redirect), pa upload na DataNode
- **Two-step download** - prvo OPEN zahtev (dobija redirect), pa download sa DataNode-a
- **Timeout konfiguracija** - 600 sekundi za velike audio fajlove

**Kod:**
- `services/content-service/internal/storage/hdfs.go` - `UploadFile()` (linija 49-137)
- `services/content-service/internal/storage/hdfs.go` - `DownloadFile()` (linija 262-321)
- `services/content-service/internal/handler/song_handler.go` - `StreamSong()`
- `services/content-service/internal/handler/song_handler.go` - `UploadAudio()`

**Kako pokazati:**
1. Pokreni: `.\test-hdfs-2.11.ps1`
2. Admin → Songs → Edit → Upload audio fajl
3. Otvori HDFS Web UI: http://localhost:9870
4. Vidi fajlove u `/audio/songs/`
5. Proveri stream: `curl http://localhost:8081/api/content/songs/{id}/stream`

**Šta profesor može da pita:**
- "Zašto HDFS umesto običnog fajl sistema?" → **Odgovor:** "HDFS je distribuirani fajl sistem - fajlovi se repliciraju na više DataNode-a, što omogućava skalabilnost i otpornost na otkaze."
- "Kako funkcioniše WebHDFS redirect pattern?" → **Odgovor:** "NameNode vraća 307 redirect sa Location header-om koji pokazuje na DataNode gde se fajl nalazi. Klijent direktno komunicira sa DataNode-om, što smanjuje opterećenje NameNode-a."
- "Zašto timeout od 600 sekundi?" → **Odgovor:** "Audio fajlovi mogu biti veliki (nekoliko MB). Timeout od 10 minuta omogućava upload velikih fajlova bez prekida."

---

## ✅ OCENA 10 - Kompleksni šabloni

### 1.13 Brisanje pesama (Saga)
**Šta:** Brisanje pesme koordiniše više servisa  
**Kako:**
- Saga orchestrator koordiniše: Backup → Delete Ratings → Delete Neo4j → Delete HDFS → Delete MongoDB
- Ako neki korak ne uspe, izvršava se kompenzacija u obrnutom redosledu

**Kod:**
- `services/saga-service/internal/orchestrator/song_deletion_saga.go` - saga logika
- `services/content-service/internal/handler/song_handler.go` - poziva saga-service

**Kako pokazati:**
```powershell
.\test-saga-complete.ps1
```
Ili kroz frontend: Admin → Songs → Delete Song

---

### 1.14 Istorija aktivnosti (Event Sourcing)
**Šta:** Sve aktivnosti su događaji u Event Store-u  
**Kako:**
- Svaka aktivnost (slušanje, ocenjivanje, pretplata) se čuva kao događaj
- Event Store u MongoDB (`event_store` kolekcija)
- Svaki događaj ima version (sequence number)

**Kod:**
- `services/analytics-service/internal/store/event_store.go` - Event Store
- `services/analytics-service/internal/handler/activity_handler.go` - `LogActivity()`

**Kako pokazati:**
1. Pokreni: `.\test-event-sourcing-provera.ps1`
2. Profile → Activity History → vidiš sve aktivnosti iz Event Store-a

---

### 2.12 Keširanje (Redis)
**Šta:** Najslušanije pesme se keširaju u Redis-u  
**Kako:**
- Play count se čuva u Redis (`song_play_count:*`)
- Most played songs se kešira (`most_played_songs`)
- Cache se invalidira kada se play count poveća

**Detaljna implementacija:**

**Kod:** `services/content-service/internal/cache/redis_cache.go` (linija 46-61)
```go
func (c *RedisCache) IncrementPlayCount(ctx context.Context, songID string) error {
    key := fmt.Sprintf(SongPlayCountKey, songID)  // "song_play_count:songID"
    
    // INCREMENT counter u Redis-u (atomic operacija)
    _, err := c.client.Incr(ctx, key).Result()
    if err != nil {
        return fmt.Errorf("failed to increment play count: %w", err)
    }
    
    // Postavlja expiration (24 sata)
    c.client.Expire(ctx, key, 24*time.Hour)
    
    // INVALIDIRA most played cache (write-through invalidation)
    c.client.Del(ctx, MostPlayedCacheKey)  // "most_played_songs"
    
    return nil
}
```

**Kod:** `services/content-service/internal/cache/redis_cache.go` (linija 63-82)
```go
func (c *RedisCache) GetMostPlayedSongs(ctx context.Context, limit int) ([]*MostPlayedSong, error) {
    // KORAK 1: Proverava cache (read-through)
    cached, err := c.client.Get(ctx, MostPlayedCacheKey).Result()
    if err == nil && cached != "" {
        // Cache HIT - vraća iz cache-a
        var songs []*MostPlayedSong
        if err := json.Unmarshal([]byte(cached), &songs); err == nil {
            log.Printf("Retrieved %d most played songs from Redis cache", len(songs))
            if len(songs) > limit {
                return songs[:limit], nil
            }
            return songs, nil
        }
    }
    
    // KORAK 2: Cache MISS - računa iz play count-ova
    log.Printf("Cache miss for most played songs, computing from play counts...")
    return c.computeMostPlayedSongs(ctx, limit)
}
```

**Kod:** `services/content-service/internal/cache/redis_cache.go` (linija 84-149)
```go
func (c *RedisCache) computeMostPlayedSongs(ctx context.Context, limit int) ([]*MostPlayedSong, error) {
    // KORAK 1: Dohvata sve play count ključeve
    keys, err := c.client.Keys(ctx, "song_play_count:*").Result()
    if err != nil {
        return nil, fmt.Errorf("failed to get play count keys: %w", err)
    }
    
    if len(keys) == 0 {
        return []*MostPlayedSong{}, nil
    }
    
    // KORAK 2: Dohvata play count za svaku pesmu
    type songCount struct {
        SongID string
        Count  int64
    }
    var songsWithCounts []songCount
    
    for _, key := range keys {
        count, err := c.client.Get(ctx, key).Int64()
        if err != nil {
            continue
        }
        songID := key[len("song_play_count:"):]  // Ekstraktuje songID
        songsWithCounts = append(songsWithCounts, songCount{
            SongID: songID,
            Count:  count,
        })
    }
    
    // KORAK 3: Sortira po count-u (descending)
    for i := 0; i < len(songsWithCounts)-1; i++ {
        for j := i + 1; j < len(songsWithCounts); j++ {
            if songsWithCounts[i].Count < songsWithCounts[j].Count {
                songsWithCounts[i], songsWithCounts[j] = songsWithCounts[j], songsWithCounts[i]
            }
        }
    }
    
    // KORAK 4: Uzima top N pesama
    result := make([]*MostPlayedSong, 0, limit)
    for i, sc := range songsWithCounts {
        if i >= limit {
            break
        }
        result = append(result, &MostPlayedSong{
            SongID: sc.SongID,
            Count:  int(sc.Count),
        })
    }
    
    // KORAK 5: Kešira rezultat (1 sat TTL)
    if len(result) > 0 {
        cacheData, err := json.Marshal(result)
        if err == nil {
            c.client.Set(ctx, MostPlayedCacheKey, cacheData, CacheTTL)  // 1 hour
            log.Printf("Cached %d most played songs in Redis", len(result))
        }
    }
    
    return result, nil
}
```

**Kod:** `services/content-service/internal/handler/song_handler.go` - IncrementPlayCount
```go
func (h *SongHandler) StreamSong(w http.ResponseWriter, r *http.Request) {
    songID := mux.Vars(r)["id"]
    
    // Stream-uje audio...
    
    // INCREMENT play count u Redis-u (asinhrono)
    go func() {
        if err := h.cache.IncrementPlayCount(context.Background(), songID); err != nil {
            log.Printf("Failed to increment play count: %v", err)
        }
    }()
}
```

**Objašnjenje:**
- **Play Count Storage:**
  - Ključ: `song_play_count:{songID}`
  - Vrednost: broj (integer)
  - TTL: 24 sata
  - Atomic increment: `INCR` operacija
- **Most Played Cache:**
  - Ključ: `most_played_songs`
  - Vrednost: JSON array top N pesama
  - TTL: 1 sat
  - Write-through invalidation: kada se play count poveća, cache se invalidira
- **Cache Strategy:**
  - **Read-through:** Ako cache postoji, vraća iz cache-a. Ako ne, računa i kešira.
  - **Write-through invalidation:** Kada se play count poveća, most played cache se briše (invalidira).

**Kod:**
- `services/content-service/internal/cache/redis_cache.go` - `IncrementPlayCount()` (linija 46-61)
- `services/content-service/internal/cache/redis_cache.go` - `GetMostPlayedSongs()` (linija 63-82)
- `services/content-service/internal/cache/redis_cache.go` - `computeMostPlayedSongs()` (linija 84-149)
- `services/content-service/internal/handler/song_handler.go` - `StreamSong()` sa increment play count

**Kako pokazati:**
```powershell
.\test-cache-saga-kratko.ps1
```
Ili ručno:
```powershell
# Proveri play count-ove
docker exec projekat-2025-2-redis-1 redis-cli KEYS "song_play_count:*"

# Proveri play count za određenu pesmu
docker exec projekat-2025-2-redis-1 redis-cli GET "song_play_count:songID"

# Proveri most played cache
docker exec projekat-2025-2-redis-1 redis-cli GET "most_played_songs"
```

**Šta profesor može da pita:**
- "Zašto Redis umesto MongoDB za play count?" → **Odgovor:** "Redis je in-memory baza - brža je za čitanje i pisanje. Atomic `INCR` operacija omogućava thread-safe increment bez lock-a."
- "Kako funkcioniše cache invalidation?" → **Odgovor:** "Write-through invalidation - kada se play count poveća, most played cache se briše (`DEL`). Sledeći zahtev će računati iz play count-ova i keširati novi rezultat."
- "Zašto TTL od 24 sata za play count?" → **Odgovor:** "Play count-ovi se resetuju svakih 24 sata, što omogućava dnevne statistike. Most played cache ima TTL od 1 sata jer se češće menja."
- "Zašto asinhrono increment play count?" → **Odgovor:** "Ne blokira stream odgovor. Korisnik dobija audio odmah, a play count se ažurira u pozadini."

---

### 2.13 Saga (detaljno)
**Šta:** Distributed transaction pattern za brisanje pesme  
**Kako:**
- **Uspešan tok:** Sve koraci se izvršavaju redom
- **Neuspešan tok:** Ako neki korak ne uspe, kompenzacija se izvršava u obrnutom redosledu
- Saga transakcije se čuvaju u MongoDB

**Detaljna implementacija:**

**Kod:** `services/saga-service/internal/orchestrator/song_deletion_saga.go` (linija 30-93)
```go
func (s *SongDeletionSaga) Execute(ctx context.Context, songID string) (*model.SagaTransaction, error) {
    // KORAK 1: Kreira saga transakciju sa svim koracima
    saga := &model.SagaTransaction{
        ID:     fmt.Sprintf("saga_%s_%d", songID, time.Now().Unix()),
        Type:   "DELETE_SONG",
        Status: model.SagaStatusPending,
        SongID: songID,
        Steps: []model.SagaStep{
            {Name: model.StepBackupSong, Status: model.StepStatusPending, Order: 1},
            {Name: model.StepDeleteRatings, Status: model.StepStatusPending, Order: 2},
            {Name: model.StepDeleteFromNeo4j, Status: model.StepStatusPending, Order: 3},
            {Name: model.StepDeleteFromHDFS, Status: model.StepStatusPending, Order: 4},
            {Name: model.StepDeleteFromMongo, Status: model.StepStatusPending, Order: 5},
        },
    }
    s.store.CreateTransaction(ctx, saga)  // Čuva u MongoDB
    
    // KORAK 2: Izvršava korake redom
    for i := range saga.Steps {
        step := &saga.Steps[i]
        log.Printf("Executing step %d: %s for song %s", step.Order, step.Name, songID)
        
        step.Status = model.StepStatusPending
        s.store.UpdateStepStatus(ctx, saga.ID, step.Name, step.Status, "")
        
        // KORAK 3: Izvršava korak
        err := s.executeStep(ctx, saga, step)
        if err != nil {
            log.Printf("Step %s failed: %v", step.Name, err)
            step.Status = model.StepStatusFailed
            step.Error = err.Error()
            s.store.UpdateStepStatus(ctx, saga.ID, step.Name, step.Status, err.Error())
            
            // KORAK 4: Kompenzacija - rollback svih prethodnih koraka
            saga.Status = model.SagaStatusCompensating
            s.store.UpdateTransaction(ctx, saga)
            s.compensate(ctx, saga, i)  // Kompenzuje sve korake do i-1
            saga.Status = model.SagaStatusCompensated
            saga.Error = fmt.Sprintf("Step %s failed: %v", step.Name, err)
            s.store.UpdateTransaction(ctx, saga)
            return saga, fmt.Errorf("saga failed at step %s: %w", step.Name, err)
        }
        
        // KORAK 5: Korak uspeo - označi kao completed
        step.Status = model.StepStatusCompleted
        s.store.UpdateStepStatus(ctx, saga.ID, step.Name, step.Status, "")
    }
    
    // KORAK 6: Svi koraci uspeli
    saga.Status = model.SagaStatusCompleted
    s.store.UpdateTransaction(ctx, saga)
    return saga, nil
}
```

**Kod:** `services/saga-service/internal/orchestrator/song_deletion_saga.go` (linija 95-111)
```go
func (s *SongDeletionSaga) executeStep(ctx context.Context, saga *model.SagaTransaction, step *model.SagaStep) error {
    switch step.Name {
    case model.StepBackupSong:
        return s.backupSong(ctx, saga)  // Backup pesme pre brisanja
    case model.StepDeleteRatings:
        return s.deleteRatings(ctx, saga.SongID)  // Briše ocene
    case model.StepDeleteFromNeo4j:
        return s.deleteFromNeo4j(ctx, saga.SongID)  // Briše iz Neo4j grafa
    case model.StepDeleteFromHDFS:
        return s.deleteFromHDFS(ctx, saga)  // Briše audio fajl
    case model.StepDeleteFromMongo:
        return s.deleteFromMongo(ctx, saga.SongID)  // Briše iz MongoDB
    default:
        return fmt.Errorf("unknown step: %s", step.Name)
    }
}
```

**Kod:** `services/saga-service/internal/orchestrator/song_deletion_saga.go` (linija 248-270)
```go
func (s *SongDeletionSaga) compensate(ctx context.Context, saga *model.SagaTransaction, failedStepIndex int) {
    log.Printf("Starting compensation for saga %s (failed at step %d)", saga.ID, failedStepIndex)
    
    // Kompenzacija u OBRNUTOM redosledu (od failedStepIndex-1 do 0)
    for i := failedStepIndex - 1; i >= 0; i-- {
        step := &saga.Steps[i]
        if step.Status != model.StepStatusCompleted {
            continue  // Preskače korake koji nisu bili izvršeni
        }
        
        log.Printf("Compensating step: %s", step.Name)
        err := s.compensateStep(ctx, saga, step)
        if err != nil {
            log.Printf("Compensation failed for step %s: %v", step.Name, err)
            // Nastavlja sa drugim kompenzacijama čak i ako jedna ne uspe
        } else {
            step.Status = model.StepStatusCompensated
            s.store.UpdateStepStatus(ctx, saga.ID, step.Name, step.Status, "")
        }
    }
}
```

**Kod:** `services/saga-service/internal/orchestrator/song_deletion_saga.go` (linija 272-293)
```go
func (s *SongDeletionSaga) compensateStep(ctx context.Context, saga *model.SagaTransaction, step *model.SagaStep) error {
    switch step.Name {
    case model.StepDeleteFromNeo4j:
        // Restore song to Neo4j (rekreira čvor)
        log.Printf("Compensating: Would restore song %s to Neo4j", saga.SongID)
        return nil
    case model.StepDeleteFromHDFS:
        // Restore audio file to HDFS (iz backup-a)
        log.Printf("Compensating: Would restore audio file for song %s to HDFS", saga.SongID)
        return nil
    case model.StepDeleteFromMongo:
        // Restore song to MongoDB (iz backup-a)
        return s.restoreToMongo(ctx, saga)
    case model.StepBackupSong, model.StepDeleteRatings:
        // Nema kompenzacije - backup je samo čitanje, brisanje ocena je ireverzibilno
        return nil
    default:
        return fmt.Errorf("unknown step for compensation: %s", step.Name)
    }
}
```

**Objašnjenje:**
- **Orchestrator pattern** - Saga-service koordiniše sve korake
- **Sekvencijalno izvršavanje** - koraci se izvršavaju redom (1 → 2 → 3 → 4 → 5)
- **Kompenzacija u obrnutom redosledu** - ako korak 3 ne uspe, kompenzuje se korak 2, pa korak 1
- **Persistence** - svaki korak se čuva u MongoDB, omogućava retry i monitoring
- **Idempotentnost** - svaki korak može da se izvrši više puta bez problema

**Primer uspešnog toka:**
1. Backup Song ✅
2. Delete Ratings ✅
3. Delete from Neo4j ✅
4. Delete from HDFS ✅
5. Delete from MongoDB ✅
→ **Saga Status: COMPLETED**

**Primer neuspešnog toka (korak 3 ne uspe):**
1. Backup Song ✅
2. Delete Ratings ✅
3. Delete from Neo4j ❌ (neuspeh)
4. **Kompenzacija:**
   - Compensate Delete Ratings (nema kompenzacije - ocene su već obrisane)
   - Compensate Backup Song (nema kompenzacije - backup je samo čitanje)
→ **Saga Status: COMPENSATED**

**Kod:**
- `services/saga-service/internal/orchestrator/song_deletion_saga.go` - `Execute()` (linija 30-93)
- `services/saga-service/internal/orchestrator/song_deletion_saga.go` - `compensate()` (linija 248-270)
- `services/saga-service/internal/orchestrator/song_deletion_saga.go` - `executeStep()` (linija 95-111)
- `services/saga-service/internal/store/saga_store.go` - čuvanje saga transakcija u MongoDB

**Kako pokazati:**
1. **Uspešan tok:** Admin → Delete Song → sve uspešno
   - Proveri MongoDB: `db.saga_transactions.find()` - vidiš `status: "COMPLETED"`
2. **Neuspešan tok:** Zaustavi ratings-service → Delete Song → saga se rollback-uje
   - Proveri MongoDB: `db.saga_transactions.find()` - vidiš `status: "COMPENSATED"` i `error` polje

**Šta profesor može da pita:**
- "Zašto Saga umesto distributed transaction?" → **Odgovor:** "Distributed transactions (2PC) su spore i blokiraju resurse. Saga koristi kompenzaciju - svaki korak ima svoju kompenzaciju, što je brže i fleksibilnije."
- "Kako funkcioniše kompenzacija?" → **Odgovor:** "Ako korak N ne uspe, kompenzujemo sve prethodne korake (N-1, N-2, ...) u obrnutom redosledu. Svaki korak ima svoju kompenzaciju (npr. restore iz backup-a)."
- "Šta se dešava ako kompenzacija ne uspe?" → **Odgovor:** "Logujemo grešku, ali nastavljamo sa drugim kompenzacijama. U produkciji bi trebalo da imamo retry mehanizam za kompenzacije."
- "Zašto čuvate saga transakcije u MongoDB?" → **Odgovor:** "Omogućava retry mehanizam, monitoring, i debugging. Možemo da vidimo tačno gde je saga stala i zašto."

---

### 2.14 Event Sourcing (detaljno)
**Šta:** Sve aktivnosti su immutable događaji  
**Kako:**
- Event Store čuva sve događaje (`event_store` kolekcija)
- Event stream endpoint vraća sve događaje za korisnika
- Replay endpoint rekonstruiše stanje iz događaja

**Detaljna implementacija:**

**Kod:** `services/analytics-service/internal/model/event.go`
```go
type UserEvent struct {
    ID        primitive.ObjectID `bson:"_id" json:"id"`
    EventID   string             `bson:"eventId" json:"eventId"`
    StreamID  string             `bson:"streamId" json:"streamId"`  // userID
    EventType EventType           `bson:"eventType" json:"eventType"`
    Version   int64              `bson:"version" json:"version"`      // Sequence number
    Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
    Data      map[string]interface{} `bson:"data" json:"data"`
}

type EventType string
const (
    EventTypeSongPlayed        EventType = "SONG_PLAYED"
    EventTypeRatingGiven       EventType = "RATING_GIVEN"
    EventTypeSubscribedToGenre EventType = "SUBSCRIBED_TO_GENRE"
    EventTypeUnsubscribedFromGenre EventType = "UNSUBSCRIBED_FROM_GENRE"
    EventTypeSubscribedToArtist EventType = "SUBSCRIBED_TO_ARTIST"
    EventTypeUnsubscribedFromArtist EventType = "UNSUBSCRIBED_FROM_ARTIST"
)
```

**Kod:** `services/analytics-service/internal/store/event_store.go` (linija 61-100)
```go
func (es *EventStore) AppendEvent(ctx context.Context, event *model.UserEvent) error {
    // Generiše event ID ako nije dat
    if event.EventID == "" {
        event.EventID = primitive.NewObjectID().Hex()
    }
    
    // Generiše version (sequence number) za stream
    if event.Version == 0 {
        nextVersion, err := es.getNextVersion(ctx, event.StreamID)
        if err != nil {
            return fmt.Errorf("failed to get next version: %w", err)
        }
        event.Version = nextVersion  // 1, 2, 3, ...
    }
    
    // Setuje timestamp
    if event.Timestamp.IsZero() {
        event.Timestamp = time.Now()
    }
    
    // APPEND-ONLY - samo dodavanje, nikada brisanje ili izmena
    _, err := es.collection.InsertOne(ctx, event)
    if err != nil {
        if mongo.IsDuplicateKeyError(err) {
            return fmt.Errorf("event version conflict: version %d already exists for stream %s", event.Version, event.StreamID)
        }
        return fmt.Errorf("failed to append event: %w", err)
    }
    
    log.Printf("Event appended successfully: streamId=%s, eventType=%s, version=%d", event.StreamID, event.EventType, event.Version)
    return nil
}
```

**Kod:** `services/analytics-service/internal/store/event_store.go` (linija 102-116)
```go
func (es *EventStore) getNextVersion(ctx context.Context, streamID string) (int64, error) {
    // Pronalazi poslednji event za stream
    opts := options.FindOne().SetSort(bson.D{{Key: "version", Value: -1}})
    var lastEvent model.UserEvent
    err := es.collection.FindOne(ctx, bson.M{"streamId": streamID}, opts).Decode(&lastEvent)
    
    if err == mongo.ErrNoDocuments {
        return 1, nil  // Prvi event u stream-u
    }
    if err != nil {
        return 0, fmt.Errorf("failed to get last version: %w", err)
    }
    
    return lastEvent.Version + 1, nil  // Sledeći version
}
```

**Kod:** `services/analytics-service/internal/store/event_store.go` (linija 118-144)
```go
func (es *EventStore) GetEventStream(ctx context.Context, streamID string, fromVersion int64, limit int) ([]*model.UserEvent, error) {
    filter := bson.M{"streamId": streamID}
    if fromVersion > 0 {
        filter["version"] = bson.M{"$gte": fromVersion}  // Od određenog version-a
    }
    
    // Sortira po version-u (rastuće)
    opts := options.Find().SetSort(bson.D{{Key: "version", Value: 1}})
    if limit > 0 {
        opts.SetLimit(int64(limit))
    }
    
    cursor, err := es.collection.Find(ctx, filter, opts)
    if err != nil {
        return nil, fmt.Errorf("failed to get event stream: %w", err)
    }
    defer cursor.Close(ctx)
    
    var events []*model.UserEvent
    if err = cursor.All(ctx, &events); err != nil {
        return nil, fmt.Errorf("failed to decode events: %w", err)
    }
    
    return events, nil
}
```

**Kod:** `services/analytics-service/internal/store/event_store.go` (linija 171-192)
```go
func (es *EventStore) ReplayEvents(ctx context.Context, streamID string) (*model.UserActivityState, error) {
    // Dohvata SVE događaje za korisnika
    events, err := es.GetEventStream(ctx, streamID, 0, 0)
    if err != nil {
        return nil, fmt.Errorf("failed to get events for replay: %w", err)
    }
    
    // Kreira početno stanje
    state := &model.UserActivityState{
        UserID:            streamID,
        SubscribedGenres:  make([]string, 0),
        SubscribedArtists: make([]string, 0),
        ActivityBreakdown: make(map[string]int),
        RecentActivities:  make([]*model.UserActivity, 0),
    }
    
    // REKONSTRUISANJE STANJA - primenjuje sve događaje redom
    for _, event := range events {
        es.applyEvent(state, event)  // Ažurira stanje na osnovu događaja
    }
    
    return state, nil
}
```

**Kod:** `services/analytics-service/internal/store/event_store.go` (linija 194-220)
```go
func (es *EventStore) applyEvent(state *model.UserActivityState, event *model.UserEvent) {
    // Ažurira activity breakdown
    state.ActivityBreakdown[string(event.EventType)]++
    
    // Ažurira stanje na osnovu tipa događaja
    switch event.EventType {
    case model.EventTypeSubscribedToGenre:
        genre := event.Data["genre"].(string)
        state.SubscribedGenres = append(state.SubscribedGenres, genre)
    case model.EventTypeUnsubscribedFromGenre:
        genre := event.Data["genre"].(string)
        // Uklanja iz liste
        for i, g := range state.SubscribedGenres {
            if g == genre {
                state.SubscribedGenres = append(state.SubscribedGenres[:i], state.SubscribedGenres[i+1:]...)
                break
            }
        }
    case model.EventTypeSongPlayed:
        songID := event.Data["songId"].(string)
        state.RecentActivities = append(state.RecentActivities, &model.UserActivity{
            Type:      "SONG_PLAYED",
            SongID:    songID,
            Timestamp: event.Timestamp,
        })
    // ... ostali tipovi događaja
    }
}
```

**Objašnjenje:**
- **Append-only** - događaji se samo dodaju, nikada ne brišu ili menjaju (immutable)
- **Version (sequence number)** - svaki događaj ima version koji se povećava (1, 2, 3, ...)
- **Stream ID** - svi događaji za korisnika su u jednom stream-u (`userID`)
- **Replay** - stanje se rekonstruiše primenom svih događaja redom
- **Time travel** - možemo da vidimo stanje u bilo kom trenutku replay-ovanjem događaja do tog trenutka

**Kod:**
- `services/analytics-service/internal/store/event_store.go` - `AppendEvent()` (linija 61-100)
- `services/analytics-service/internal/store/event_store.go` - `GetEventStream()` (linija 118-144)
- `services/analytics-service/internal/store/event_store.go` - `ReplayEvents()` (linija 171-192)
- `services/analytics-service/internal/store/event_store.go` - `applyEvent()` (linija 194-220)
- `services/analytics-service/internal/model/event.go` - `UserEvent` struktura

**Kako pokazati:**
```powershell
.\test-event-sourcing-provera.ps1
```
Ili ručno:
```powershell
# Proveri Event Store
docker exec projekat-2025-2-mongodb-analytics-1 mongosh analytics_db --eval 'db.event_store.find().sort({version: 1}).limit(5).pretty()'

# Proveri replay
curl http://localhost:8081/api/analytics/users/{userId}/replay
```

**Šta profesor može da pita:**
- "Zašto Event Sourcing umesto da čuvate trenutno stanje?" → **Odgovor:** "Omogućava time travel, audit trail, i rekonstrukciju stanja. Možemo da vidimo tačnu istoriju svih aktivnosti korisnika."
- "Šta je version (sequence number)?" → **Odgovor:** "Redni broj događaja u stream-u. Osigurava redosled i sprečava duplicate događaje (unique constraint na streamId + version)."
- "Kako funkcioniše replay?" → **Odgovor:** "Dohvatamo sve događaje za korisnika, sortirane po version-u, i primenjujemo ih redom na početno stanje. Na kraju dobijamo trenutno stanje."
- "Zašto append-only?" → **Odgovor:** "Događaji su immutable - jednom kreirani, ne menjaju se. Ako treba da se ispravi greška, kreiramo novi događaj koji ispravlja prethodni."

---

## ✅ OCENA 10+ - Event Sourcing + CQRS

### 1.15 Prikaz i računanje analitika
**Šta:** Analitike se računaju iz Event Store-a i čuvaju u Read Model-u  
**Kako:**
- **CQRS Command Side:** Aktivnost → Command → Event → Event Store
- **CQRS Query Side:** Query Handler čita iz Projection Store (read model)
- Projection Store čuva izračunate analitike (totalSongsPlayed, totalRatings, itd.)

**Kod:**
- `services/analytics-service/internal/cqrs/command_handler.go` - Command Handler
- `services/analytics-service/internal/cqrs/query_handler.go` - Query Handler
- `services/analytics-service/internal/store/projection_store.go` - Projection Store

**Kako pokazati:**
1. Profile → Analytics → prikazuju se analitike
2. Pokreni: `.\test-event-sourcing-provera.ps1` → vidiš Projection Store

---

### 2.15 Event Sourcing + CQRS (kombinovano)
**Šta:** Event Sourcing za čuvanje događaja + CQRS za čitanje analitika  
**Kako:**
- **Write Side:** Command Handler kreira događaje u Event Store
- **Read Side:** Query Handler čita iz Projection Store (brže)
- Event Handler ažurira Projection Store kada se kreira novi događaj

**Detaljna implementacija:**

#### CQRS Command Side (Write)
**Kod:** `services/analytics-service/internal/cqrs/command_handler.go` (linija 23-66)
```go
func (ch *CommandHandler) HandleCommand(ctx context.Context, cmd Command) *CommandResult {
    var event *model.UserEvent
    
    // Konvertuje Command u Event
    switch c := cmd.(type) {
    case *PlaySongCommand:
        event = ch.handlePlaySongCommand(ctx, c)
    case *RateSongCommand:
        event = ch.handleRateSongCommand(ctx, c)
    case *SubscribeToArtistCommand:
        event = ch.handleSubscribeToArtistCommand(ctx, c)
    // ... ostali command tipovi
    }
    
    if event == nil {
        return &CommandResult{Success: false, Error: fmt.Errorf("failed to create event")}
    }
    
    // APPEND EVENT TO EVENT STORE (Event Sourcing)
    if err := ch.eventStore.AppendEvent(ctx, event); err != nil {
        return &CommandResult{Event: event, Success: false, Error: err}
    }
    
    log.Printf("Command handled successfully: userID=%s, eventType=%s", cmd.GetUserID(), event.EventType)
    return &CommandResult{Event: event, Success: true}
}
```

**Kod:** `services/analytics-service/internal/handler/activity_handler.go` (linija 82-105)
```go
func (h *ActivityHandler) LogActivity(w http.ResponseWriter, r *http.Request) {
    // Parsira aktivnost iz zahteva
    var activity model.UserActivity
    json.NewDecoder(r.Body).Decode(&activity)
    
    // KONVERTUJE AKTIVNOST U COMMAND
    cmd := h.activityToCommand(&activity)
    
    // COMMAND HANDLER KREIRA EVENT I APPEND-UJE U EVENT STORE
    result := h.CommandHandler.HandleCommand(ctx, cmd)
    if !result.Success {
        log.Printf("Error handling command: %v", result.Error)
        return
    }
    
    log.Printf("Command handled successfully via CQRS: streamId=%s, eventType=%s", result.Event.StreamID, result.Event.EventType)
    
    // EVENT HANDLER AŽURIRA PROJECTION STORE (Read Model)
    if h.ProjectionStore != nil {
        eventHandler := cqrs.NewEventHandler(h.ProjectionStore, h.Config)
        if err := eventHandler.HandleEvent(ctx, result.Event); err != nil {
            log.Printf("Error processing event for projection: %v", err)
        }
    }
}
```

#### CQRS Event Handler (Projection Update)
**Kod:** `services/analytics-service/internal/cqrs/event_handler.go` (linija 30-120)
```go
func (eh *EventHandler) HandleEvent(ctx context.Context, event *model.UserEvent) error {
    // Dohvata trenutnu projekciju za korisnika
    projection, err := eh.projectionStore.GetProjection(ctx, event.StreamID)
    if err != nil {
        // Ako ne postoji, kreira novu
        projection = &model.AnalyticsProjection{
            UserID: event.StreamID,
        }
    }
    
    // AŽURIRA PROJECTION NA OSNOVU DOGAĐAJA
    switch event.EventType {
    case model.EventTypeSongPlayed:
        projection.TotalSongsPlayed++
        songID := event.Data["songId"].(string)
        // Ažurira top songs
        if projection.TopSongs == nil {
            projection.TopSongs = make(map[string]int)
        }
        projection.TopSongs[songID]++
        
    case model.EventTypeRatingGiven:
        projection.TotalRatings++
        rating := int(event.Data["rating"].(float64))
        projection.TotalRatingSum += rating
        projection.AverageRating = float64(projection.TotalRatingSum) / float64(projection.TotalRatings)
        
    case model.EventTypeSubscribedToGenre:
        genre := event.Data["genre"].(string)
        if projection.GenresPlayed == nil {
            projection.GenresPlayed = make(map[string]int)
        }
        projection.GenresPlayed[genre]++
        
    // ... ostali tipovi događaja
    }
    
    // ČUVA AŽURIRANU PROJECTION U READ MODEL
    return eh.projectionStore.UpdateProjection(ctx, projection)
}
```

#### CQRS Query Side (Read)
**Kod:** `services/analytics-service/internal/cqrs/query_handler.go` (linija 32-100)
```go
func (qh *QueryHandler) HandleQuery(ctx context.Context, query Query) *QueryResult {
    switch q := query.(type) {
    case *GetUserAnalyticsQuery:
        return qh.handleGetUserAnalyticsQuery(ctx, q)
    default:
        return &QueryResult{Error: fmt.Errorf("unknown query type")}
    }
}

func (qh *QueryHandler) handleGetUserAnalyticsQuery(ctx context.Context, query *GetUserAnalyticsQuery) *QueryResult {
    // ČITA IZ PROJECTION STORE (READ MODEL) - BRZO!
    projection, err := qh.projectionStore.GetProjection(ctx, query.UserID)
    if err != nil {
        return &QueryResult{Error: err}
    }
    
    // Vraća analitike iz read modela
    return &QueryResult{
        Data: map[string]interface{}{
            "totalSongsPlayed": projection.TotalSongsPlayed,
            "averageRating":   projection.AverageRating,
            "genresPlayed":     projection.GenresPlayed,
            "topArtists":       projection.TopArtists,
            "subscribedArtistsCount": len(projection.SubscribedArtists),
        },
    }
}
```

**Kod:** `services/analytics-service/internal/store/projection_store.go`
```go
type AnalyticsProjection struct {
    UserID            string            `bson:"userId" json:"userId"`
    TotalSongsPlayed   int               `bson:"totalSongsPlayed" json:"totalSongsPlayed"`
    TotalRatings       int               `bson:"totalRatings" json:"totalRatings"`
    TotalRatingSum    int               `bson:"totalRatingSum" json:"totalRatingSum"`
    AverageRating      float64           `bson:"averageRating" json:"averageRating"`
    GenresPlayed       map[string]int    `bson:"genresPlayed" json:"genresPlayed"`
    TopSongs          map[string]int    `bson:"topSongs" json:"topSongs"`
    TopArtists         []string          `bson:"topArtists" json:"topArtists"`
    SubscribedArtists  []string          `bson:"subscribedArtists" json:"subscribedArtists"`
    LastUpdated        time.Time         `bson:"lastUpdated" json:"lastUpdated"`
}
```

**Objašnjenje:**
- **Command Side (Write):**
  1. Aktivnost → Command → Event → Event Store (append-only)
  2. Event Handler automatski ažurira Projection Store
- **Query Side (Read):**
  1. Query → Query Handler → Projection Store (brzo čitanje)
  2. Nema potrebe da se replay-uju svi događaji svaki put
- **Projection Store (Read Model):**
  - Optimizovan za čitanje (denormalizovani podaci)
  - Ažurira se automatski kada se kreira novi događaj
  - Brže od replay-ovanja svih događaja

**Tok podataka:**
```
1. Korisnik sluša pesmu
   ↓
2. LogActivity() prima aktivnost
   ↓
3. CommandHandler.HandleCommand() kreira Event
   ↓
4. EventStore.AppendEvent() čuva događaj (Event Sourcing)
   ↓
5. EventHandler.HandleEvent() ažurira Projection Store (CQRS)
   ↓
6. QueryHandler.HandleQuery() čita iz Projection Store (brzo!)
```

**Kod:**
- `services/analytics-service/internal/cqrs/command_handler.go` - Command Handler (linija 23-66)
- `services/analytics-service/internal/cqrs/event_handler.go` - Event Handler (linija 30-120)
- `services/analytics-service/internal/cqrs/query_handler.go` - Query Handler (linija 32-100)
- `services/analytics-service/internal/store/projection_store.go` - Projection Store
- `services/analytics-service/internal/handler/activity_handler.go` - `LogActivity()` integracija (linija 82-105)

**Kako pokazati:**
```powershell
.\test-event-sourcing-provera.ps1
```
Pokaži:
- Event Store: 136 događaja (`db.event_store.count()`)
- Projection Store: 4 projekcije (`db.analytics_projections.find()`)
- CQRS Command Handler radi (logovi: "Command handled successfully")
- CQRS Query Handler radi (brzo vraća analitike iz Projection Store)

**Šta profesor može da pita:**
- "Zašto kombinujete Event Sourcing i CQRS?" → **Odgovor:** "Event Sourcing omogućava audit trail i time travel, ali replay-ovanje svih događaja je sporo. CQRS Projection Store omogućava brzo čitanje analitika, a ažurira se automatski kada se kreira novi događaj."
- "Kako se Projection Store ažurira?" → **Odgovor:** "Event Handler se poziva nakon što se događaj append-uje u Event Store. Event Handler ažurira Projection Store na osnovu tipa događaja (npr. `SONG_PLAYED` → `TotalSongsPlayed++`)."
- "Šta ako Projection Store nije ažuriran?" → **Odgovor:** "Možemo da replay-ujemo sve događaje iz Event Store-a i rekonstruišemo Projection Store. Event Store je source of truth."
- "Zašto ne čitate direktno iz Event Store-a?" → **Odgovor:** "Za analitike bi trebalo da replay-ujemo sve događaje svaki put, što je sporo. Projection Store je optimizovan za čitanje - sve analitike su već izračunate."

---

## 📝 Napomene za odbranu

### Opšte napomene:
1. **Spremite test skripte** - pokrenite ih pre odbrane
2. **Proverite logove** - `docker-compose logs [service-name]`
3. **Spremite kod lokacije** - znajte gde je svaki deo implementacije
4. **Prikažite arhitekturu** - kako servisi komuniciraju

### Ako pitaju detalje:
- **Event Sourcing:** "Sve aktivnosti su immutable događaji sa version number"
- **CQRS:** "Command Side piše u Event Store, Query Side čita iz Projection Store"
- **Saga:** "Orchestrator koordiniše više koraka, kompenzacija u obrnutom redosledu"
- **Circuit Breaker:** "Otvara se nakon 3 neuspeha, resetuje se nakon 5 sekundi"

---

## 📖 Kako koristiti ovaj dokument na odbrani

### Pre odbrane:
1. **Pročitajte detaljna objašnjenja** za sve zahteve koje ste uradili
2. **Pogledajte kod** na lokacijama navedenim u dokumentu
3. **Pokrenite test skripte** da proverite da sve radi
4. **Spremite MongoDB/Redis komande** za demonstraciju

### Tokom odbrane:
1. **Kada profesor pita "Kako ste implementirali X?"**
   - Otvorite ovaj dokument
   - Pronađite zahtev X
   - Pročitajte sekciju "Detaljna implementacija"
   - Objasnite korak po korak kako funkcioniše
   - Pokažite kod ako je moguće

2. **Kada profesor pita "Zašto ste tako uradili?"**
   - Pročitajte sekciju "Objašnjenje" u detaljnoj implementaciji
   - Koristite odgovore iz sekcije "Šta profesor može da pita"

3. **Kada treba da pokažete funkcionalnost:**
   - Koristite sekciju "Kako pokazati"
   - Pokrenite test skripte
   - Pokažite frontend ili logove

### Ključne stvari za zapamtiti:
- **Sve zahteve imate detaljno objašnjene** - ne morate da pogađate
- **Kod lokacije su navedene** - možete brzo da pronađete implementaciju
- **Objašnjenja su student-friendly** - jednostavnim rečima, sa primerima
- **Svaki mehanizam ima logiku rada** - korak po korak objašnjenje

---

**Srećno na odbrani! 🎉**
