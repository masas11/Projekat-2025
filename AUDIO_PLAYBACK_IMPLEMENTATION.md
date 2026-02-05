# Implementacija Reprodukcije Pesama (1.6)

## Opis
Implementirana je funkcionalnost za reprodukciju pesama kroz browser sa podrškom za lokalne fajlove i stream sa backend-a.

## Komponente

### 1. AudioPlayer Component
**Lokacija:** `frontend/src/components/AudioPlayer.js`
**Funkcionalnosti:**
- Play/Pause kontrola
- Volume kontrola
- Seek (premotavanje)
- Prikaz trenutnog vremena i ukupnog trajanja
- Error handling
- Loading states

### 2. Backend Streaming Endpoint
**Lokacija:** `services/content-service/internal/handler/song_handler.go`
**Endpoint:** `GET /api/content/songs/{id}/stream`
**Funkcionalnosti:**
- Stream audio fajlova sa backend-a
- Podrška za lokalne fajlove
- Podrška za eksterne URL-ove
- Redirekcija na eksterne audio source-ove

## Model Podataka

### Song Model Update
Dodato polje `AudioFileURL` u Song model:
```go
type Song struct {
    ID           string   `json:"id" bson:"_id"`
    Name         string   `json:"name" bson:"name"`
    Duration     int      `json:"duration" bson:"duration"`
    Genre        string   `json:"genre" bson:"genre"`
    AlbumID      string   `json:"albumId" bson:"albumId"`
    ArtistIDs    []string `json:"artistIds" bson:"artistIds"`
    AudioFileURL string   `json:"audioFileUrl" bson:"audioFileUrl,omitempty"`
    CreatedAt    string   `json:"createdAt" bson:"createdAt"`
    UpdatedAt    string   `json:"updatedAt" bson:"updatedAt"`
}
```

## Frontend Integracija

### 1. SongDetail Page
- Dodat AudioPlayer komponenta
- Prikazuje se za svaku pesmu na detaljnoj strani

### 2. Songs Management Form
- Dodato polje za unos AudioFileURL
- Opciono polje za unos URL-a do audio fajla

### 3. API Service
- Dodata `getStreamUrl(songId)` metoda
- Konstrukcija URL-a za streaming

## Način Korišćenja

### 1. Kreiranje Pesme sa Audio Fajlom
```json
{
    "name": "Moja Pesma",
    "duration": 180,
    "genre": "Pop",
    "albumId": "album-id",
    "artistIds": ["artist-id"],
    "audioFileUrl": "https://example.com/song.mp3"
}
```

### 2. Lokalni Fajlovi
```json
{
    "audioFileUrl": "/path/to/local/song.mp3"
}
```

### 3. Stream sa Backend-a
Frontend automatski koristi `/api/content/songs/{id}/stream` endpoint za reprodukciju.

## Testiranje

### Test Route
- URL: `/test-audio`
- Komponenta: `TestAudioPlayer.js`
- Omogućava testiranje AudioPlayer komponente

## Backend Logika

### StreamSong Handler
1. Proverava da li pesma postoji
2. Ako postoji `AudioFileURL`:
   - Za HTTP/HTTPS URL-ove: redirect
   - Za lokalne fajlove: serve file
3. Ako ne postoji: vrati error

## CORS i Sigurnost
- Podržani su samo audio fajlovi
- Content-Type se postavlja automatski
- Cache-Control podešen za streaming

## Buduća Unapređenja
1. File upload funkcionalnost
2. Podrška za više audio formata
3. Audio preprocessing
4. Playlist funkcionalnost
5. Audio vizualizacija

## Tehnologije
- **Frontend:** React, HTML5 Audio API
- **Backend:** Go, net/http
- **Audio Format:** MP3 (podržani i drugi formati)
