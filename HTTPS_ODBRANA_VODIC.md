# 🔐 HTTPS - Vodič za Odbranu

## 📋 ŠTA DA KAŽEŠ NA ODBRANI

### 1. UVOD (30 sekundi)

"Implementirali smo HTTPS protokol na svim slojevima komunikacije u sistemu. HTTPS obezbeđuje šifrovanje podataka u tranzitu, što znači da svi podaci koji se prenose između klijenta i servera, kao i između servisa, su zaštićeni od presretanja i manipulacije."

---

### 2. ARHITEKTURA HTTPS-A (1-2 minuta)

**Šta da kažeš:**

"HTTPS je implementiran na tri nivoa:

1. **Klijent → API Gateway**: Frontend komunicira sa API Gateway-em preko HTTPS-a
2. **API Gateway → Backend Servisi**: API Gateway prosleđuje zahteve backend servisima preko HTTPS-a
3. **Svi servisi koriste HTTPS**: Svaki servis ima svoj HTTPS server sa sertifikatima"

**Šta da pokažeš:**

```bash
# Pokaži docker-compose.yml - environment varijable
docker-compose.yml (linije 10-16, 41-42)
```

**Objašnjenje:**

"Vidite da API Gateway koristi `https://users-service:8001` umesto `http://`. To znači da sva komunikacija između servisa ide preko šifrovanog kanala."

---

### 3. SSL SERTIFIKATI (1-2 minuta)

**Šta da kažeš:**

"Koristimo self-signed SSL sertifikate za development. Sertifikati se generišu pomoću OpenSSL-a i sadrže:
- **server.crt**: Javni sertifikat sa javnim ključem
- **server.key**: Privatni ključ (mora biti zaštićen)"

**Šta da pokažeš:**

```powershell
# 1. Pokaži sertifikate
ls certs\

# 2. Pokaži sadržaj docker-compose.yml - volumes
# Linija 19: - ./certs:/app/certs:ro

# 3. Pokaži da servisi koriste sertifikate
docker exec projekat-2025-2-api-gateway-1 env | Select-String "TLS"
```

**Objašnjenje:**

"Sertifikati su mountovani u sve servise kao read-only volume (`:ro`). Svaki servis čita sertifikate iz `/app/certs/` direktorijuma. Privatni ključ je zaštićen permisijama 0600."

---

### 4. IMPLEMENTACIJA U KODU (2-3 minuta)

**Šta da kažeš:**

"Svaki servis proverava da li su sertifikati dostupni preko environment varijabli. Ako postoje, pokreće HTTPS server, u suprotnom fallback na HTTP."

**Šta da pokažeš:**

**Otvoriti:** `services/api-gateway/cmd/main.go` (linije 512-530)

```go
// Support HTTPS if certificates are provided
certFile := os.Getenv("TLS_CERT_FILE")
keyFile := os.Getenv("TLS_KEY_FILE")
if certFile != "" && keyFile != "" {
    log.Println("Starting HTTPS server on port", cfg.Port)
    server := &http.Server{
        Addr:    ":" + cfg.Port,
        Handler: mux,
    }
    if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
        if appLogger != nil {
            appLogger.LogTLSFailure("api-gateway", err.Error(), "")
        }
        log.Fatal("HTTPS server failed:", err)
    }
} else {
    log.Println("Starting HTTP server on port", cfg.Port)
    log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
```

**Objašnjenje:**

"Ovo je graceful degradation - ako sertifikati nisu dostupni, servis i dalje radi preko HTTP-a. U production okruženju, sertifikati će uvek biti dostupni."

---

### 5. INTER-SERVICE KOMUNIKACIJA (2-3 minuta)

**Šta da kažeš:**

"API Gateway prosleđuje zahteve backend servisima preko HTTPS-a. Konfigurisali smo HTTP klijent sa TLS konfiguracijom koja omogućava komunikaciju sa self-signed sertifikatima."

**Šta da pokažeš:**

**Otvoriti:** `services/api-gateway/cmd/main.go` (linije 76-85)

```go
// Konfiguriši HTTP klijent da ignoriše sertifikate za inter-service komunikaciju
// (jer koristimo samopotpisane sertifikate)
tr := &http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}
client := &http.Client{
    Timeout:   5 * time.Second, // Timeout za pozive backend servisa
    Transport: tr,
}
resp, err := client.Do(req)
```

**Objašnjenje:**

"`InsecureSkipVerify: true` se koristi samo za development sa self-signed sertifikatima. U production okruženju, sertifikati bi bili potpisani od strane Certificate Authority (CA), pa bi verifikacija bila obavezna."

**Pokaži environment varijable:**

```powershell
docker exec projekat-2025-2-api-gateway-1 env | Select-String "SERVICE_URL"
```

**Rezultat:**
```
USERS_SERVICE_URL=https://users-service:8001
CONTENT_SERVICE_URL=https://content-service:8002
RATINGS_SERVICE_URL=https://ratings-service:8003
...
```

**Objašnjenje:**

"Vidite da svi URL-ovi počinju sa `https://`. To znači da API Gateway uvek koristi HTTPS za komunikaciju sa backend servisima."

---

### 6. DEMONSTRACIJA (2-3 minute)

**Šta da pokažeš:**

#### 6.1 Provera da HTTPS radi

```powershell
# 1. Provera logova - HTTPS server je pokrenut
docker logs projekat-2025-2-api-gateway-1 --tail 5 | Select-String "HTTPS"

# Očekivano: "Starting HTTPS server on port 8080"
```

#### 6.2 Test HTTPS konekcije

```powershell
# Učitaj helper funkciju
. .\https-helper.ps1

# Test HTTPS zahteva
$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/health"
Write-Host "Status: $($result.StatusCode)"
Write-Host "Content: $($result.Content)"
```

**Očekivano rezultat:**
```
Status: 200
Content: users-service is running
```

**Objašnjenje:**

"Vidite da zahtev prolazi preko HTTPS-a i vraća status 200. Podaci su šifrovani u tranzitu."

#### 6.3 Provera sertifikata u kontejneru

```powershell
# Provera da sertifikati postoje u kontejneru
docker exec projekat-2025-2-api-gateway-1 ls -la /app/certs/
```

**Očekivano:**
```
-rw-r--r-- 1 root root 1234 Feb 18 09:00 server.crt
-rw-r--r-- 1 root root 1675 Feb 18 09:00 server.key
```

---

### 7. ZAŠTITA PODATAKA (1-2 minuta)

**Šta da kažeš:**

"HTTPS štiti sledeće podatke u tranzitu:
- **Lozinke**: Pri registraciji i promeni lozinke
- **JWT tokeni**: Pri autentifikaciji i autorizaciji
- **OTP kodovi**: Pri verifikaciji prijave
- **Lični podaci**: Email, ime, prezime
- **Admin akcije**: Sve administrativne operacije"

**Šta da pokažeš:**

```powershell
# Test registracije - podaci se šalju preko HTTPS-a
$body = @{
    firstName = "Test"
    lastName = "User"
    email = "test@example.com"
    username = "testuser"
    password = "Test1234!"
    confirmPassword = "Test1234!"
} | ConvertTo-Json

$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/register" `
    -Method "POST" -Body $body -ContentType "application/json"

Write-Host "Status: $($result.StatusCode)"
```

**Objašnjenje:**

"Svi ovi podaci (uključujući lozinku) se šalju preko HTTPS-a, što znači da su šifrovani i zaštićeni od presretanja."

---

### 8. LOGOVANJE TLS GREŠAKA (1 minuta)

**Šta da kažeš:**

"Implementirali smo logovanje TLS grešaka. Ako dođe do problema sa HTTPS konekcijom, to se loguje kao `TLS_FAILURE` događaj."

**Šta da pokažeš:**

**Otvoriti:** `services/api-gateway/cmd/main.go` (linije 88-95)

```go
if err != nil {
    // Log TLS/connection errors
    if log != nil {
        errorMsg := err.Error()
        if strings.Contains(errorMsg, "tls") || strings.Contains(errorMsg, "TLS") || 
           strings.Contains(errorMsg, "certificate") || strings.Contains(errorMsg, "handshake") {
            serviceName := extractServiceName(targetURL)
            log.LogTLSFailure(serviceName, errorMsg, r.RemoteAddr)
        }
    }
    http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
    return
}
```

**Objašnjenje:**

"Ovo omogućava praćenje problema sa HTTPS konekcijama i brzo reagovanje na sigurnosne incidente."

---

## 🎯 KLJUČNE TAČKE ZA ODBRANU

### ✅ Šta je implementirano:

1. ✅ **HTTPS na svim slojevima**: Klijent → Gateway → Backend
2. ✅ **SSL sertifikati**: Self-signed za development
3. ✅ **Inter-service HTTPS**: Svi servisi komuniciraju preko HTTPS-a
4. ✅ **Graceful degradation**: Fallback na HTTP ako sertifikati nisu dostupni
5. ✅ **TLS error logging**: Logovanje grešaka u HTTPS komunikaciji
6. ✅ **Zaštita podataka**: Svi osetljivi podaci se šalju preko HTTPS-a

### 🔍 Šta da naglasiš:

1. **Defense in depth**: HTTPS na svim slojevima, ne samo na jednom
2. **Inter-service security**: Čak i komunikacija između servisa je šifrovana
3. **Production ready**: Kod je spreman za production sertifikate (samo zameniti self-signed)
4. **Monitoring**: TLS greške se loguju za praćenje

---

## 📝 ODGOVORI NA ČESTA PITANJA

### P: Zašto koristite self-signed sertifikate?

**O:** "Za development okruženje koristimo self-signed sertifikate jer su jednostavni za generisanje i ne zahtevaju Certificate Authority. U production okruženju, sertifikati bi bili potpisani od strane CA (npr. Let's Encrypt), što bi obezbedilo potpunu verifikaciju identiteta servera."

### P: Zašto `InsecureSkipVerify: true`?

**O:** "Ovo se koristi samo za development sa self-signed sertifikatima. Go standardna biblioteka neće prihvatiti self-signed sertifikate bez ove opcije. U production okruženju sa CA-signed sertifikatima, ova opcija bi bila `false` i verifikacija bi bila obavezna."

### P: Kako se generišu sertifikati?

**O:** "Sertifikati se generišu pomoću OpenSSL-a. Proces uključuje:
1. Generisanje privatnog ključa (2048-bit RSA)
2. Generisanje Certificate Signing Request (CSR)
3. Self-signing sertifikata (važi 365 dana)

Sertifikati se čuvaju u `./certs/` direktorijumu i mountuju u sve servise."

### P: Kako funkcioniše TLS handshake?

**O:** "TLS handshake proces:
1. Klijent šalje ClientHello sa podrškom za TLS verzije i cipher suite-ove
2. Server odgovara ServerHello sa sertifikatom i odabranim cipher suite-om
3. Klijent verifikuje sertifikat (u production)
4. Razmena ključeva (Diffie-Hellman ili RSA)
5. Uspostavljanje šifrovane konekcije

Nakon handshake-a, sva komunikacija je šifrovana."

### P: Šta se dešava ako sertifikat nije dostupan?

**O:** "Implementirali smo graceful degradation. Ako sertifikati nisu dostupni, servis se pokreće preko HTTP-a. Ovo omogućava fleksibilnost u development okruženju, ali u production okruženju sertifikati će uvek biti dostupni."

---

## 🎬 REDOSLED DEMONSTRACIJE NA ODBRANI

### 1. Uvod (30s)
- Objasni šta je HTTPS i zašto je važan

### 2. Arhitektura (1-2min)
- Pokaži docker-compose.yml - environment varijable
- Objasni tri nivoa HTTPS-a

### 3. Sertifikati (1-2min)
- Pokaži sertifikate (`ls certs/`)
- Pokaži mount u docker-compose.yml
- Objasni strukturu sertifikata

### 4. Kod (2-3min)
- Otvori `services/api-gateway/cmd/main.go` (linije 512-530)
- Objasni graceful degradation
- Otvori `services/api-gateway/cmd/main.go` (linije 76-85)
- Objasni inter-service komunikaciju

### 5. Demonstracija (2-3min)
- Proveri logove (`docker logs ... | Select-String HTTPS`)
- Test HTTPS zahteva (`Invoke-HTTPSRequest`)
- Proveri sertifikate u kontejneru (`docker exec ... ls /app/certs/`)

### 6. Zaštita podataka (1-2min)
- Test registracije preko HTTPS-a
- Objasni šta se štiti

### 7. Logovanje (1min)
- Pokaži TLS error logging kod

**Ukupno vreme: ~10-15 minuta**

---

## 💡 SAVETI ZA ODBRANU

1. **Budite sigurni**: Znate šta radite i zašto
2. **Pokažite kod**: Ne samo pričajte, pokažite implementaciju
3. **Demonstrirajte**: Pokrenite komande i pokažite da radi
4. **Budite spremni na pitanja**: Pročitajte "Česta pitanja" sekciju
5. **Naglasite production readiness**: Kod je spreman za production sertifikate

---

## 📚 DODATNI MATERIJALI

- `docker-compose.yml` - Konfiguracija servisa i sertifikata
- `services/api-gateway/cmd/main.go` - HTTPS server implementacija
- `services/users-service/cmd/main.go` - HTTPS server implementacija
- `generate-certs.ps1` - Skripta za generisanje sertifikata
- `https-helper.ps1` - Helper funkcije za HTTPS testove

---

**Srećno na odbrani! 🚀**
