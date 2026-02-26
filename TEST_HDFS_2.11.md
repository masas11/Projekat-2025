# Testiranje HDFS Implementacije (2.11)

## Opis
HDFS (Hadoop Distributed File System) je implementiran za čuvanje audio zapisa pesama. Audio fajlovi se čuvaju na HDFS-u umesto lokalno.

## Komponente

### 1. HDFS Servisi
- **hdfs-namenode**: NameNode servis (port 9870 - Web UI, port 9000 - RPC)
- **hdfs-datanode**: DataNode servis (port 9864 - Web UI)

### 2. HDFS Client
- **Lokacija**: `services/content-service/internal/storage/hdfs.go`
- **Funkcionalnosti**:
  - Upload fajlova na HDFS
  - Download fajlova sa HDFS-a
  - Provera postojanja fajlova
  - Brisanje fajlova
  - Kreiranje direktorijuma

### 3. Content Service Modifikacije
- **Upload endpoint**: `POST /api/content/songs/{songId}/upload`
- **Stream endpoint**: `GET /api/content/songs/{songId}/stream` (sada koristi HDFS)
- **HDFS path format**: `/audio/songs/{songId}.{ext}`

## Kako Testirati

### 1. Pokretanje Servisa

```powershell
# Pokreni sve servise
docker-compose up -d

# Sačekaj 1-2 minuta da se HDFS potpuno pokrene
# Proveri status:
docker-compose ps
docker-compose logs hdfs-namenode
```

### 2. Provera HDFS Statusa

```powershell
# Proveri HDFS Web UI
# Otvori: http://localhost:9870

# Ili kroz API:
Invoke-WebRequest -Uri "http://localhost:9870/webhdfs/v1/?op=LISTSTATUS" -UseBasicParsing
```

### 3. Upload Audio Fajla

#### Opcija A: Kroz API (curl)

```powershell
# 1. Uloguj se kao admin i dobij JWT token
$loginResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/users/login" `
    -Method POST `
    -ContentType "application/json" `
    -Body '{"username":"admin","password":"admin123"}' `
    -UseBasicParsing

$token = ($loginResponse.Content | ConvertFrom-Json).token

# 2. Kreiraj pesmu (ako već ne postoji)
$songResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/content/songs" `
    -Method POST `
    -Headers @{Authorization="Bearer $token"} `
    -ContentType "application/json" `
    -Body '{
        "name": "Test Song",
        "duration": 180,
        "genre": "Pop",
        "albumId": "{albumId}",
        "artistIds": ["{artistId}"]
    }' `
    -UseBasicParsing

$songId = ($songResponse.Content | ConvertFrom-Json).id

# 3. Upload audio fajl
curl -X POST "http://localhost:8081/api/content/songs/$songId/upload" `
    -H "Authorization: Bearer $token" `
    -F "audio=@C:\path\to\audio.mp3" `
    -F "songId=$songId"
```

#### Opcija B: Kroz Frontend

1. Uloguj se kao admin
2. Idi na **Songs** -> **Edit Song**
3. Klikni **Upload Audio**
4. Izaberi audio fajl (mp3, wav, ogg, m4a, flac)
5. Klikni **Upload**

### 4. Testiranje Stream-a

```powershell
# Download audio fajl sa HDFS-a kroz stream endpoint
Invoke-WebRequest -Uri "http://localhost:8081/api/content/songs/{songId}/stream" `
    -OutFile "downloaded-audio.mp3" `
    -UseBasicParsing

# Ili kroz frontend:
# - Otvori pesmu
# - Klikni Play
# - Audio bi trebalo da se reprodukuje
```

### 5. Provera HDFS Web UI

1. Otvori: **http://localhost:9870**
2. Klikni **Utilities** -> **Browse the file system**
3. Navigiraj do `/audio/songs/`
4. Proveri da li postoje upload-ovani fajlovi

### 6. Automatski Test

```powershell
# Pokreni test skriptu
.\test-hdfs-2.11.ps1
```

## Očekivani Rezultati

### ✅ Uspešan Upload
- HTTP 200 OK
- Response sa `hdfsPath`, `fileName`, `fileSize`
- Fajl vidljiv u HDFS Web UI na `/audio/songs/{songId}.{ext}`

### ✅ Uspešan Stream
- HTTP 200 OK
- Audio fajl se preuzima i reprodukuje
- Content-Type: `audio/mpeg`, `audio/wav`, itd.

### ✅ Provera u HDFS Web UI
- Fajlovi su vidljivi na `/audio/songs/`
- File status prikazuje veličinu, vreme kreiranja, itd.

## Troubleshooting

### Problem: HDFS nije dostupan
**Rešenje:**
```powershell
# Proveri logove
docker-compose logs hdfs-namenode
docker-compose logs hdfs-datanode

# Restart HDFS servisa
docker-compose restart hdfs-namenode hdfs-datanode

# Sačekaj 1-2 minuta
```

### Problem: "Connection refused" pri upload-u
**Rešenje:**
- Proveri da li je `HDFS_NAMENODE_URL` postavljen u `docker-compose.yml`
- Proveri da li je `content-service` zavisan od `hdfs-namenode` (depends_on)

### Problem: "File not found" pri stream-u
**Rešenje:**
- Proveri da li je fajl upload-ovan na HDFS
- Proveri HDFS path u bazi podataka (`song.audioFileURL`)
- Proveri HDFS Web UI da li fajl postoji

### Problem: Upload radi, ali stream ne radi
**Rešenje:**
- Proveri da li je `HDFSClient` inicijalizovan u `song_handler.go`
- Proveri logove `content-service` za greške
- Proveri da li je HDFS path ispravno formatiran (`/audio/songs/{songId}.{ext}`)

## Napomene

- HDFS može da traje 1-2 minuta da se potpuno pokrene
- Audio fajlovi se čuvaju na HDFS-u, ne lokalno
- Maksimalna veličina upload-a: 100MB
- Podržani formati: mp3, wav, ogg, m4a, flac
- HDFS podaci se čuvaju u `./data/hdfs/` direktorijumu
