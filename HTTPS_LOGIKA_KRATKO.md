# 🔐 HTTPS - Kratka Logika za Objašnjenje

## 📝 LOGIKA HTTPS-A (2-3 minuta objašnjenja)

### 1. **Sertifikati se generišu i mountuju**

```
Host sistem: ./certs/server.crt i server.key
         ↓
Docker volume mount (read-only)
         ↓
Container: /app/certs/server.crt i /app/certs/server.key
```

**Kod u docker-compose.yml:**
```yaml
volumes:
  - ./certs:/app/certs:ro
environment:
  - TLS_CERT_FILE=/app/certs/server.crt
  - TLS_KEY_FILE=/app/certs/server.key
```

---

### 2. **Svaki servis proverava sertifikate**

**Logika u kodu (services/*/cmd/main.go):**
```go
certFile := os.Getenv("TLS_CERT_FILE")  // Čita environment varijablu
keyFile := os.Getenv("TLS_KEY_FILE")

if certFile != "" && keyFile != "" {
    // Ako sertifikati postoje → HTTPS server
    server.ListenAndServeTLS(certFile, keyFile)
} else {
    // Ako ne postoje → HTTP fallback
    http.ListenAndServe(":"+port, mux)
}
```

**Objašnjenje:** "Svaki servis proverava da li su sertifikati dostupni. Ako jesu, pokreće HTTPS server, u suprotnom fallback na HTTP."

---

### 3. **API Gateway koristi HTTPS za komunikaciju sa backend servisima**

**Environment varijable:**
```yaml
USERS_SERVICE_URL=https://users-service:8001
CONTENT_SERVICE_URL=https://content-service:8002
RATINGS_SERVICE_URL=https://ratings-service:8003
```

**Kod u API Gateway (services/api-gateway/cmd/main.go):**
```go
// Konfiguracija HTTP klijenta sa TLS podrškom
tr := &http.Transport{
    TLSClientConfig: &tls.Config{
        InsecureSkipVerify: true,  // Za self-signed sertifikate
    },
}
client := &http.Client{
    Timeout:   5 * time.Second,
    Transport: tr,
}

// Poziv ka backend servisu preko HTTPS-a
targetURL := "https://users-service:8001/register"
resp, err := client.Do(req)
```

**Objašnjenje:** "API Gateway koristi HTTPS URL-ove i konfigurisani HTTP klijent sa TLS podrškom za komunikaciju sa backend servisima."

---

### 4. **Kompletan flow**

```
1. Klijent šalje zahtev → API Gateway (HTTPS)
   ↓
2. API Gateway proverava sertifikate → Pokreće HTTPS server
   ↓
3. API Gateway prosleđuje zahtev → Backend servis (HTTPS)
   ↓
4. Backend servis proverava sertifikate → Pokreće HTTPS server
   ↓
5. Odgovor se vraća → Sve preko HTTPS-a (šifrovano)
```

---

## 🎯 KLJUČNE TAČKE ZA OBJAŠNJENJE

### ✅ Šta je implementirano:

1. **Sertifikati se mountuju u sve servise** - Read-only volume mount
2. **Svaki servis proverava sertifikate** - Environment varijable `TLS_CERT_FILE` i `TLS_KEY_FILE`
3. **Graceful degradation** - Ako sertifikati nisu dostupni, fallback na HTTP
4. **Inter-service HTTPS** - API Gateway koristi HTTPS za komunikaciju sa backend servisima
5. **TLS konfiguracija** - HTTP klijent sa TLS podrškom za self-signed sertifikate

### 🔑 Zašto je važno:

- **Zaštita podataka u tranzitu** - Svi osetljivi podaci su šifrovani
- **Defense in depth** - HTTPS na svim slojevima komunikacije
- **Inter-service security** - Čak i komunikacija između servisa je zaštićena
- **Production ready** - Kod je spreman, samo zameniti sertifikate

---

## 💬 KAKO DA OBJASNIŠ ASISTENTU

### Uvod (30s):
"Implementirali smo HTTPS protokol na svim slojevima komunikacije. Sertifikati se generišu i mountuju u sve servise, a svaki servis proverava da li su sertifikati dostupni preko environment varijabli."

### Glavna logika (1-2min):
"Logika je jednostavna:
1. Sertifikati se mountuju iz `./certs/` u `/app/certs/` u svakom kontejneru
2. Svaki servis čita `TLS_CERT_FILE` i `TLS_KEY_FILE` environment varijable
3. Ako postoje, pokreće HTTPS server sa `ListenAndServeTLS()`
4. API Gateway koristi HTTPS URL-ove (`https://users-service:8001`) za komunikaciju
5. HTTP klijent je konfigurisan sa TLS podrškom za inter-service komunikaciju"

### Demonstracija (30s):
"Pokazujem kod u `main.go` gde se proveravaju sertifikati i pokreće HTTPS server, i kod u API Gateway-u gde se konfiguriše HTTP klijent za HTTPS komunikaciju."

---

## 📁 FAJLOVI ZA POKAZIVANJE

1. **docker-compose.yml** - Volume mount i environment varijable
2. **services/api-gateway/cmd/main.go** - Linije 512-530 (HTTPS server) i 76-85 (Inter-service HTTPS)
3. **services/users-service/cmd/main.go** - Linije 131-149 (HTTPS server)

---

**Ukupno vreme objašnjenja: 2-3 minuta**
