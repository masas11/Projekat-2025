# 2.6 Asinhrona komunikacija između servisa

## 📋 Pregled

Sistem koristi **event-driven arhitekturu** za asinhronu komunikaciju između mikroservisa. Ovo omogućava:
- **Decoupling** - servisi su slabo povezani
- **Scalability** - servisi mogu nezavisno da se skaliraju
- **Resilience** - otkaz jednog servisa ne utiče na druge
- **Performance** - asinhroni pozivi ne blokiraju glavni tok

## 🔄 Event Flow

### 1. Content Service → Subscriptions Service
**Događaji:**
- `new_artist` - kada se kreira novi umetnik
- `new_album` - kada se kreira novi album
- `new_song` - kada se kreira nova pesma

**Svrha:** Subscriptions service proverava koje korisnike treba obavestiti o novom sadržaju.

**Implementacija:**
- `content-service/internal/events/emitter.go` - `EmitEvent()` funkcija
- Asinhrono slanje preko HTTP POST zahteva
- Timeout: 2 sekunde
- TLS podrška za HTTPS

### 2. Content Service → Recommendation Service
**Događaji:**
- `song_created` - kada se kreira nova pesma
- `artist_created` - kada se kreira novi umetnik
- `album_created` - kada se kreira novi album
- `song_deleted` - kada se obriše pesma
- `artist_deleted` - kada se obriše umetnik
- `album_deleted` - kada se obriše album

**Svrha:** Recommendation service ažurira Neo4j graf sa novim/obrisanim čvorovima.

**Implementacija:**
- Asinhrono slanje preko HTTP POST zahteva
- Recommendation service obrađuje događaje u goroutine-u
- Timeout: 30 sekundi za obradu događaja

### 3. Ratings Service → Recommendation Service
**Događaji:**
- `rating_created` - kada korisnik oceni pesmu
- `rating_updated` - kada korisnik promeni ocenu
- `rating_deleted` - kada korisnik obriše ocenu

**Svrha:** Recommendation service ažurira RATED relacije u Neo4j grafu.

**Implementacija:**
- `ratings-service/cmd/main.go` - `emitRatingEvent()` funkcija
- Asinhrono slanje preko HTTP POST zahteva
- Timeout: 5 sekundi

### 4. Subscriptions Service → Recommendation Service
**Događaji:**
- `subscription_created` - kada korisnik se pretplati na žanr
- `subscription_deleted` - kada korisnik otkaže pretplatu na žanr

**Svrha:** Recommendation service ažurira SUBSCRIBED_TO relacije u Neo4j grafu.

**Implementacija:**
- `subscriptions-service/cmd/main.go` - `emitSubscriptionEvent()` funkcija
- Asinhrono slanje preko HTTP POST zahteva
- Timeout: 5 sekundi

## 📡 Event Handlers

### Recommendation Service (`/events`)
**Endpoint:** `POST /events`

**Obrađuje događaje:**
- `rating_created`, `rating_updated` → `handleRatingEvent()`
- `rating_deleted` → `handleRatingDeleted()`
- `subscription_created` → `handleSubscriptionCreated()`
- `subscription_deleted` → `handleSubscriptionDeleted()`
- `song_created` → `handleSongCreated()`
- `song_deleted` → `handleSongDeleted()`
- `artist_created` → `handleArtistCreated()`
- `artist_deleted` → `handleArtistDeleted()`
- `album_deleted` → `handleAlbumDeleted()`

**Karakteristike:**
- Asinhrona obrada u goroutine-u
- Timeout: 30 sekundi za obradu
- Automatsko kreiranje korisnika ako ne postoji
- Logovanje svih događaja

### Subscriptions Service (`/events`)
**Endpoint:** `POST /events`

**Obrađuje događaje:**
- `new_artist` → `handleNewArtistEvent()`
- `new_album` → `handleNewAlbumEvent()`
- `new_song` → `handleNewSongEvent()`

**Karakteristike:**
- Proverava pretplate korisnika
- Kreira notifikacije za pretplaćene korisnike
- Timeout: 10 sekundi za obradu
- Retry mehanizam za kreiranje notifikacija

## 🔧 Implementacioni detalji

### Event Emitter Pattern
```go
// content-service/internal/events/emitter.go
func EmitEvent(serviceURL string, event interface{}) {
    go func() {
        // Asinhrono slanje HTTP POST zahteva
        // TLS podrška
        // Error handling
    }()
}
```

### Event Handler Pattern
```go
// recommendation-service/cmd/main.go
mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
    // Prihvata događaj
    go func() {
        // Asinhrona obrada u goroutine-u
        // Dugačak timeout za Neo4j operacije
    }()
    w.WriteHeader(http.StatusAccepted)
})
```

## 📊 Event Types

### Content Events
| Event Type | Payload | Emitted By | Handled By |
|------------|---------|------------|------------|
| `new_artist` | `artistId`, `name`, `genres` | Content Service | Subscriptions Service |
| `new_album` | `albumId`, `name`, `genre`, `artistIds`, `artistNames` | Content Service | Subscriptions Service |
| `new_song` | `songId`, `name`, `genre`, `artistIds`, `artistNames`, `albumId` | Content Service | Subscriptions Service |
| `song_created` | `songId`, `name`, `genre`, `artistIds`, `albumId`, `duration` | Content Service | Recommendation Service |
| `artist_created` | `artistId`, `name`, `genres` | Content Service | Recommendation Service |
| `album_created` | `albumId`, `name`, `genre`, `artistIds` | Content Service | Recommendation Service |
| `song_deleted` | `songId` | Content Service | Recommendation Service |
| `artist_deleted` | `artistId` | Content Service | Recommendation Service |
| `album_deleted` | `albumId` | Content Service | Recommendation Service |

### Rating Events
| Event Type | Payload | Emitted By | Handled By |
|------------|---------|------------|------------|
| `rating_created` | `userId`, `songId`, `rating` | Ratings Service | Recommendation Service |
| `rating_updated` | `userId`, `songId`, `rating` | Ratings Service | Recommendation Service |
| `rating_deleted` | `userId`, `songId` | Ratings Service | Recommendation Service |

### Subscription Events
| Event Type | Payload | Emitted By | Handled By |
|------------|---------|------------|------------|
| `subscription_created` | `userId`, `genre` | Subscriptions Service | Recommendation Service |
| `subscription_deleted` | `userId`, `genre` | Subscriptions Service | Recommendation Service |

## ✅ Prednosti asinhrone komunikacije

1. **Decoupling** - Servisi su nezavisni, promene u jednom servisu ne utiču na druge
2. **Performance** - Glavni tok se ne blokira čekanjem odgovora
3. **Resilience** - Otkaz jednog servisa ne sprečava rad drugih
4. **Scalability** - Servisi mogu nezavisno da se skaliraju
5. **Eventual Consistency** - Podaci će se eventualno sinhronizovati

## 🔍 Primeri korišćenja

### Emitovanje događaja
```go
// content-service/internal/handler/song_handler.go
events.EmitEvent(h.SubscriptionsServiceURL, event)
events.EmitEvent(h.RecommendationServiceURL, map[string]interface{}{
    "type": "song_created",
    "songId": song.ID,
    // ...
})
```

### Obrađivanje događaja
```go
// recommendation-service/cmd/main.go
case "song_created":
    handleSongCreated(eventCtx, event, neo4jStore)
```

## 📝 Napomene

- Svi događaji se šalju **asinhrono** (u goroutine-u)
- Event handlers vraćaju `202 Accepted` odmah, obrađuju događaje u pozadini
- Postoji **retry mehanizam** za kritične operacije (npr. kreiranje notifikacija)
- **TLS podrška** za sigurnu komunikaciju između servisa
- **Timeout-ovi** su postavljeni da spreče beskonačno čekanje

## 🚀 Testiranje

Za testiranje asinhrone komunikacije:

1. **Kreiraj novu pesmu** → Proveri logove subscriptions-service i recommendation-service
2. **Oceni pesmu** → Proveri logove recommendation-service
3. **Pretplati se na žanr** → Proveri logove recommendation-service
4. **Obriši pesmu** → Proveri logove recommendation-service (pesma bi trebalo da se obriše iz Neo4j grafa)
