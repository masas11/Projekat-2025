# Event Sourcing Implementation (2.14)

## Pregled

Event Sourcing pattern je implementiran za aktivnosti korisnika. Sve aktivnosti su predstavljene kao immutable događaji (events) koji se čuvaju u Event Store-u. Stanje se rekonstruiše replay-ovanjem svih događaja.

## Arhitektura

### Event Store
- **Baza**: MongoDB (`mongodb-analytics:27017`, database: `analytics_db`)
- **Collection**: `event_store`
- **Pattern**: Append-only event store sa event stream-om

### Event Model

Svaki događaj je immutable i sadrži:
- **EventID**: Jedinstveni identifikator događaja
- **EventType**: Tip događaja (SONG_PLAYED, RATING_GIVEN, itd.)
- **StreamID**: Identifikator stream-a (User ID)
- **Version**: Sekvencijalni broj događaja u stream-u
- **Timestamp**: Vreme kada se događaj desio
- **Payload**: Podaci specifični za događaj (immutable)
- **Metadata**: Dodatni metadata (opciono)

### Tipovi Događaja

1. **SONG_PLAYED** - Slušanje pesme
2. **RATING_GIVEN** - Davanje ocene pesmi
3. **GENRE_SUBSCRIBED** - Kreiranje pretplate na žanr
4. **GENRE_UNSUBSCRIBED** - Uklanjanje pretplate na žanr
5. **ARTIST_SUBSCRIBED** - Kreiranje pretplate na umetnika
6. **ARTIST_UNSUBSCRIBED** - Uklanjanje pretplate na umetnika

## Karakteristike Event Sourcing-a

### 1. Append-Only Store
- Svi događaji se dodaju na kraj stream-a
- Događaji su immutable (ne mogu se menjati ili brisati)
- Version kontrola osigurava redosled događaja

### 2. Event Stream
- Svaki korisnik ima svoj event stream
- Događaji su sekvencijalno numerisani (version)
- Mogućnost čitanja događaja od određenog version-a

### 3. State Reconstruction (Replay)
- Stanje se rekonstruiše replay-ovanjem svih događaja
- Mogućnost kreiranja različitih view-ova stanja
- Mogućnost vremenskog putovanja (time travel)

## API Endpoints

### 1. Log Activity (POST /activities)
```http
POST /api/analytics/activities
Authorization: Bearer <token>
Content-Type: application/json

{
  "userId": "user123",
  "type": "SONG_PLAYED",
  "songId": "song1",
  "songName": "Song Name"
}
```

**Funkcionalnost:**
- Čuva aktivnost u ActivityStore (backward compatibility)
- Dodaje događaj u Event Store (Event Sourcing)
- Generiše sledeći version za stream

### 2. Get Event Stream (GET /events/stream)
```http
GET /api/analytics/events/stream?userId=user123&fromVersion=0&limit=100
Authorization: Bearer <token>
```

**Query Parameters:**
- `userId` (required): User ID (stream ID)
- `fromVersion` (optional): Počni od određenog version-a (default: 0)
- `limit` (optional): Maksimalan broj događaja (default: svi)

**Response:**
```json
[
  {
    "id": "event-id-1",
    "eventId": "event-id-1",
    "eventType": "SONG_PLAYED",
    "streamId": "user123",
    "version": 1,
    "timestamp": "2026-02-27T19:00:00Z",
    "payload": {
      "songId": "song1",
      "songName": "Song Name"
    }
  },
  {
    "id": "event-id-2",
    "eventId": "event-id-2",
    "eventType": "RATING_GIVEN",
    "streamId": "user123",
    "version": 2,
    "timestamp": "2026-02-27T19:05:00Z",
    "payload": {
      "songId": "song1",
      "songName": "Song Name",
      "rating": 5
    }
  }
]
```

### 3. Replay Events (GET /events/replay)
```http
GET /api/analytics/events/replay?userId=user123
Authorization: Bearer <token>
```

**Query Parameters:**
- `userId` (required): User ID (stream ID)

**Response:**
```json
{
  "userId": "user123",
  "totalSongsPlayed": 15,
  "totalRatingsGiven": 8,
  "subscribedGenres": ["Pop", "Rock"],
  "subscribedArtists": ["artist1", "artist2"],
  "lastActivityTime": "2026-02-27T19:30:00Z",
  "activityBreakdown": {
    "SONG_PLAYED": 15,
    "RATING_GIVEN": 8,
    "GENRE_SUBSCRIBED": 2,
    "ARTIST_SUBSCRIBED": 2
  },
  "recentActivities": [
    {
      "id": "activity-id",
      "userId": "user123",
      "type": "SONG_PLAYED",
      "timestamp": "2026-02-27T19:30:00Z",
      "songId": "song1",
      "songName": "Song Name"
    }
  ]
}
```

## Implementacija

### Event Store

```go
type EventStore struct {
    collection *mongo.Collection
}

// AppendEvent - dodaje novi događaj u store (append-only)
func (es *EventStore) AppendEvent(ctx context.Context, event *UserEvent) error

// GetEventStream - vraća sve događaje za stream
func (es *EventStore) GetEventStream(ctx context.Context, streamID string, fromVersion int64, limit int) ([]*UserEvent, error)

// ReplayEvents - rekonstruiše stanje replay-ovanjem događaja
func (es *EventStore) ReplayEvents(ctx context.Context, streamID string) (*UserActivityState, error)
```

### State Reconstruction

Stanje se rekonstruiše primenom svakog događaja u redosledu:

1. **SONG_PLAYED** → Povećava `totalSongsPlayed`
2. **RATING_GIVEN** → Povećava `totalRatingsGiven`
3. **GENRE_SUBSCRIBED** → Dodaje žanr u `subscribedGenres`
4. **GENRE_UNSUBSCRIBED** → Uklanja žanr iz `subscribedGenres`
5. **ARTIST_SUBSCRIBED** → Dodaje umetnika u `subscribedArtists`
6. **ARTIST_UNSUBSCRIBED** → Uklanja umetnika iz `subscribedArtists`

## Prednosti Event Sourcing-a

1. **Kompletan Audit Trail**: Svi događaji su sačuvani i nepromenljivi
2. **Time Travel**: Mogućnost rekonstrukcije stanja u bilo kom trenutku
3. **Debugging**: Mogućnost analize svih događaja koji su doveli do trenutnog stanja
4. **Scalability**: Append-only operacije su veoma efikasne
5. **Event Replay**: Mogućnost kreiranja novih view-ova stanja bez menjanja postojećih događaja

## Backward Compatibility

Postojeći `/activities` endpoint i dalje radi i koristi `ActivityStore` za backward compatibility. Svi novi događaji se automatski dodaju i u Event Store.

## MongoDB Indexes

Event Store koristi sledeće indexe za efikasne upite:

1. **Unique Index**: `(streamId, version)` - osigurava jedinstvenost version-a po stream-u
2. **Query Index**: `(streamId, timestamp)` - za efikasno čitanje događaja po vremenu
3. **Type Index**: `(eventType, timestamp)` - za filtriranje po tipu događaja

## Testiranje

### 1. Log Activity
```bash
curl -X POST http://localhost:8081/api/analytics/activities \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user123",
    "type": "SONG_PLAYED",
    "songId": "song1",
    "songName": "Test Song"
  }'
```

### 2. Get Event Stream
```bash
curl http://localhost:8081/api/analytics/events/stream?userId=user123 \
  -H "Authorization: Bearer <token>"
```

### 3. Replay Events
```bash
curl http://localhost:8081/api/analytics/events/replay?userId=user123 \
  -H "Authorization: Bearer <token>"
```

## Provera u MongoDB

```bash
# Pregled svih događaja
docker exec -it projekat-2025-2-mongodb-analytics-1 mongosh analytics_db --eval "db.event_store.find().pretty()"

# Događaji za određenog korisnika
docker exec -it projekat-2025-2-mongodb-analytics-1 mongosh analytics_db --eval "db.event_store.find({streamId: 'user123'}).sort({version: 1}).pretty()"

# Broj događaja po tipu
docker exec -it projekat-2025-2-mongodb-analytics-1 mongosh analytics_db --eval "db.event_store.aggregate([{$group: {_id: '$eventType', count: {$sum: 1}}}])"
```

## Napomene

- Event Store je append-only - događaji se ne mogu menjati ili brisati
- Version kontrola osigurava redosled događaja
- Replay može biti spor za korisnike sa velikim brojem događaja (razmotriti snapshot-ove)
- Backward compatibility je zadržana sa postojećim ActivityStore-om
