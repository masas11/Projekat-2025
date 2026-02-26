# Kako Radi HDFS Upload (2.11) - Detaljno Ukratko

## 📋 Ceo Flow (Korak po Korak)

### 1️⃣ **FRONTEND → API Gateway**
```
Korisnik (Admin) → Frontend → API Gateway (port 8081)
```
- Admin otvara formu za kreiranje/izmenu pesme
- Izabere audio fajl (mp3, wav, ogg, m4a, flac)
- Frontend šalje **multipart/form-data** request

### 2️⃣ **API Gateway → Content Service**
```
API Gateway → Content Service (port 8002)
POST /api/content/songs/{songId}/upload
```
- API Gateway prosleđuje request sa JWT tokenom
- Content Service proverava autentifikaciju (admin only)

### 3️⃣ **Content Service → HDFS**
```
Content Service → HDFS NameNode (port 9870)
```
- Content Service prima fajl (max 100MB)
- Validira ekstenziju (.mp3, .wav, .ogg, .m4a, .flac)
- Čita fajl u memoriju
- **Upload-uje na HDFS** putem WebHDFS API:
  - HDFS Path: `/audio/songs/{songId}.{ext}`
  - Primer: `/audio/songs/abc123.mp3`

### 4️⃣ **HDFS Čuva Fajl**
```
HDFS NameNode → HDFS DataNode
```
- NameNode upravlja metadata
- DataNode čuva stvarni fajl
- Fajl je sada na HDFS-u (ne lokalno!)

### 5️⃣ **Content Service → MongoDB**
```
Content Service → MongoDB
```
- Ažurira `song.audioFileURL` sa HDFS path-om
- Primer: `audioFileURL: "/audio/songs/abc123.mp3"`

### 6️⃣ **Stream (Kada Korisnik Sluša)**
```
Frontend → API Gateway → Content Service → HDFS → Frontend
GET /api/content/songs/{songId}/stream
```
- Frontend traži stream endpoint
- Content Service preuzima fajl sa HDFS-a
- Vraća audio fajl kao HTTP response
- Frontend reprodukuje audio

---

## 🔧 Tehnički Detalji

### **Upload Endpoint**
```
POST /api/content/songs/{songId}/upload
Headers:
  Authorization: Bearer {JWT_TOKEN}
Body (multipart/form-data):
  audio: [binary file]
  songId: {songId}
```

### **HDFS Client Operacije**
1. **Upload**: `HDFSClient.UploadData(fileData, hdfsPath)`
   - Kreira direktorijum `/audio/songs/` ako ne postoji
   - Upload-uje fajl na HDFS
   
2. **Download**: `HDFSClient.DownloadFile(hdfsPath)`
   - Preuzima fajl sa HDFS-a
   - Vraća `[]byte` sa audio podacima

### **HDFS Path Format**
```
/audio/songs/{songId}.{ext}
```
- Primer: `/audio/songs/507f1f77bcf86cd799439011.mp3`
- Uvek isti format za sve pesme

---

## 🎯 Kako Upload-ovati Kroz Frontend

### **Trenutno (Samo URL):**
1. Admin → Songs → Create/Edit Song
2. Unese `audioFileUrl` (URL ili lokalna putanja)
3. **Problem**: Ne upload-uje na HDFS, samo čuva URL

### **Trebalo bi (Sa Upload-om):**
1. Admin → Songs → Create/Edit Song
2. Klikne "Upload Audio File"
3. Izabere fajl sa računara
4. Frontend šalje na `/api/content/songs/{songId}/upload`
5. Content Service upload-uje na HDFS
6. Song se ažurira sa HDFS path-om

---

## 📝 Primer Request-a

### **Upload Request (curl)**
```bash
curl -X POST "http://localhost:8081/api/content/songs/abc123/upload" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -F "audio=@song.mp3" \
  -F "songId=abc123"
```

### **Response**
```json
{
  "message": "Audio file uploaded successfully",
  "songId": "abc123",
  "hdfsPath": "/audio/songs/abc123.mp3",
  "fileName": "song.mp3",
  "fileSize": 5242880
}
```

### **Stream Request**
```bash
curl http://localhost:8081/api/content/songs/abc123/stream -o downloaded.mp3
```

---

## ✅ Provera u HDFS Web UI

1. Otvori: `http://localhost:9870`
2. Klikni **Utilities** → **Browse the file system**
3. Navigiraj do `/audio/songs/`
4. Vidiš upload-ovane fajlove

---

## 🔍 Troubleshooting

### **Problem: Upload ne radi**
- Proveri da li je admin ulogovan (JWT token)
- Proveri da li je `content-service` pokrenut
- Proveri da li je HDFS pokrenut (`docker-compose logs hdfs-namenode`)

### **Problem: Stream ne radi**
- Proveri da li fajl postoji na HDFS-u (Web UI)
- Proveri `song.audioFileURL` u MongoDB-u
- Proveri logove `content-service`

### **Problem: "File not found" na HDFS**
- Proveri HDFS path format (`/audio/songs/{songId}.{ext}`)
- Proveri da li je upload bio uspešan
- Proveri HDFS Web UI

---

## 📌 Napomene

- **Maksimalna veličina**: 100MB po fajlu
- **Podržani formati**: mp3, wav, ogg, m4a, flac
- **HDFS podaci**: Čuvaju se u `./data/hdfs/` direktorijumu
- **Backup**: HDFS podaci se ne čuvaju u Git-u (dodaj u `.gitignore`)
