# 🔒 Kako Testirati HTTPS

## 🎯 Brzi Test

### **1. Proveri da li servisi rade sa HTTPS:**

```powershell
# Test API Gateway preko HTTPS (ako je konfigurisan)
Invoke-WebRequest -Uri "https://localhost:8081/api/content/artists" -SkipCertificateCheck -UseBasicParsing
```

### **2. Proveri u browser-u:**

Otvorite u browser-u:
```
https://localhost:8081/api/content/artists
```

**Napomena:** Browser će pokazati upozorenje zbog samopotpisanog sertifikata - to je normalno za development.

---

## 📋 Detaljno Testiranje

### **Metoda 1: PowerShell (Windows)**

```powershell
# Test HTTPS endpoint
try {
    $response = Invoke-WebRequest -Uri "https://localhost:8081/api/content/artists" `
        -SkipCertificateCheck `
        -UseBasicParsing
    Write-Host "✅ HTTPS radi! Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Response: $($response.Content.Substring(0, [Math]::Min(100, $response.Content.Length)))" -ForegroundColor Gray
} catch {
    Write-Host "❌ HTTPS ne radi: $($_.Exception.Message)" -ForegroundColor Red
}
```

### **Metoda 2: curl (ako je instaliran)**

```powershell
# Test HTTPS sa curl
curl -k https://localhost:8081/api/content/artists
```

### **Metoda 3: Browser**

1. Otvorite browser
2. Idite na: `https://localhost:8081/api/content/artists`
3. Browser će pokazati upozorenje o sertifikatu
4. Kliknite "Advanced" → "Proceed to localhost"
5. Trebalo bi da vidite JSON sa umetnicima

---

## 🔍 Provera Konfiguracije

### **1. Proveri da li je HTTPS pokrenut:**

```powershell
# Proveri da li servisi slušaju na HTTPS portovima
netstat -ano | findstr :8081
```

### **2. Proveri docker-compose.https.yml:**

```powershell
# Proveri da li koristiš HTTPS compose fajl
docker-compose -f docker-compose.https.yml ps
```

### **3. Proveri logove:**

```powershell
# Proveri logove API Gateway-a
docker-compose logs api-gateway | Select-String -Pattern "HTTPS|TLS|SSL|listening"
```

---

## ⚠️ Važne Napomene

### **Trenutno Stanje:**

API Gateway trenutno koristi `http.ListenAndServe()` što znači da **ne podržava HTTPS direktno**.

### **Za HTTPS treba:**

1. **Ažurirati API Gateway** da koristi `http.ListenAndServeTLS()`
2. **Dodati sertifikate** u `certs/` folder
3. **Pokrenuti sa `docker-compose.https.yml`**

---

## 🚀 Kako Omogućiti HTTPS

### **1. Kreiraj sertifikate:**

```powershell
# Kreiraj certs folder ako ne postoji
New-Item -ItemType Directory -Force -Path certs

# Generiši sertifikat (zahteva OpenSSL)
openssl genrsa -out certs/server.key 2048
openssl req -new -key certs/server.key -out certs/server.csr -subj "/C=RS/ST=Serbia/L=Belgrade/O=MusicStreaming/OU=IT/CN=localhost"
openssl x509 -req -days 365 -in certs/server.csr -signkey certs/server.key -out certs/server.crt
Remove-Item certs/server.csr
```

### **2. Ažuriraj API Gateway:**

U `services/api-gateway/cmd/main.go` promeni:

```go
// Umesto:
log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))

// Koristi:
certFile := os.Getenv("TLS_CERT_FILE")
keyFile := os.Getenv("TLS_KEY_FILE")
if certFile != "" && keyFile != "" {
    log.Println("Starting HTTPS server on port", cfg.Port)
    log.Fatal(http.ListenAndServeTLS(":"+cfg.Port, certFile, keyFile, mux))
} else {
    log.Println("Starting HTTP server on port", cfg.Port)
    log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
```

### **3. Pokreni sa HTTPS:**

```powershell
docker-compose -f docker-compose.https.yml up -d --build
```

---

## ✅ Brzi Test Komande

### **Test HTTP (trenutno):**
```powershell
Invoke-WebRequest -Uri "http://localhost:8081/api/content/artists" -UseBasicParsing
```

### **Test HTTPS (ako je konfigurisan):**
```powershell
Invoke-WebRequest -Uri "https://localhost:8081/api/content/artists" -SkipCertificateCheck -UseBasicParsing
```

---

## 🔍 Provera da li HTTPS radi

### **Znak da HTTPS radi:**
- ✅ Status code 200 sa `https://` URL-om
- ✅ Browser pokazuje "Not Secure" ili upozorenje (zbog samopotpisanog sertifikata)
- ✅ Logovi pokazuju "Starting HTTPS server"

### **Znak da HTTPS NE radi:**
- ❌ Connection refused sa `https://` URL-om
- ❌ "This site can't be reached"
- ❌ Logovi pokazuju "Starting HTTP server"

---

## 💡 Savet

Za development, HTTP je dovoljan. HTTPS je potreban za production ili ako specifikacija eksplicitno zahteva HTTPS.
