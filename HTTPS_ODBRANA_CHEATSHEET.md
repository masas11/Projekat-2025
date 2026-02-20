# 🔐 HTTPS - Cheat Sheet za Odbranu

## ⚡ BRZE KOMANDE ZA DEMONSTRACIJU

### 1. Provera da HTTPS radi

```powershell
# Provera logova - HTTPS server je pokrenut
docker logs projekat-2025-2-api-gateway-1 --tail 5 | Select-String "HTTPS"

# Očekivano: "Starting HTTPS server on port 8080"
```

### 2. Test HTTPS konekcije

```powershell
# Učitaj helper funkciju
. .\https-helper.ps1

# Test HTTPS zahteva
$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/health"
Write-Host "Status: $($result.StatusCode)" -ForegroundColor Green
Write-Host "Content: $($result.Content)"
```

### 3. Provera sertifikata

```powershell
# Lokalni sertifikati
ls certs\ | Select-Object Name, Length, LastWriteTime

# Sertifikati u kontejneru
docker exec projekat-2025-2-api-gateway-1 ls -la /app/certs/
```

### 4. Provera environment varijabli

```powershell
# API Gateway - HTTPS URL-ovi ka servisima
docker exec projekat-2025-2-api-gateway-1 env | Select-String "SERVICE_URL"

# Očekivano:
# USERS_SERVICE_URL=https://users-service:8001
# CONTENT_SERVICE_URL=https://content-service:8002
# ...

# TLS sertifikati
docker exec projekat-2025-2-api-gateway-1 env | Select-String "TLS"
```

### 5. Test registracije preko HTTPS-a

```powershell
. .\https-helper.ps1

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

Write-Host "Status: $($result.StatusCode)" -ForegroundColor $(if ($result.StatusCode -eq 201) { "Green" } else { "Yellow" })
```

### 6. Provera statusa servisa

```powershell
# Status svih servisa
docker-compose ps

# Provera da li servisi koriste HTTPS
docker logs projekat-2025-2-users-service-1 --tail 3 | Select-String "HTTPS|HTTP"
docker logs projekat-2025-2-content-service-1 --tail 3 | Select-String "HTTPS|HTTP"
```

---

## 📋 KLJUČNE TAČKE ZA ODBRANU

### ✅ Šta je implementirano:

1. ✅ HTTPS na svim slojevima komunikacije
2. ✅ SSL sertifikati (self-signed za development)
3. ✅ Inter-service HTTPS komunikacija
4. ✅ Graceful degradation (fallback na HTTP)
5. ✅ TLS error logging

### 🎯 Šta da naglasiš:

- **Defense in depth**: HTTPS na svim slojevima
- **Inter-service security**: Čak i komunikacija između servisa je šifrovana
- **Production ready**: Kod je spreman za production sertifikate
- **Monitoring**: TLS greške se loguju

---

## 🔍 KLJUČNI FAJLOVI ZA POKAZIVANJE

1. **docker-compose.yml**
   - Linije 10-16: API Gateway environment varijable
   - Linije 41-42: Users Service TLS varijable
   - Linija 19: Volume mount za sertifikate

2. **services/api-gateway/cmd/main.go**
   - Linije 512-530: HTTPS server pokretanje
   - Linije 76-85: Inter-service HTTPS komunikacija

3. **services/users-service/cmd/main.go**
   - Linije 131-149: HTTPS server pokretanje

4. **certs/**
   - `server.crt`: Javni sertifikat
   - `server.key`: Privatni ključ

---

## 💬 ODGOVORI NA PITANJA

### P: Zašto self-signed sertifikati?
**O:** "Za development. U production bi koristili CA-signed sertifikate (Let's Encrypt)."

### P: Zašto InsecureSkipVerify?
**O:** "Samo za development sa self-signed sertifikatima. U production bi bilo false."

### P: Kako se generišu sertifikati?
**O:** "OpenSSL - genrsa → req → x509. Sertifikati se čuvaju u `./certs/` i mountuju u servise."

### P: Šta ako sertifikat nije dostupan?
**O:** "Graceful degradation - servis se pokreće preko HTTP-a. U production bi uvek bio dostupan."

---

## ⏱️ TIMING ZA ODBRANU

- **Uvod**: 30s
- **Arhitektura**: 1-2min
- **Sertifikati**: 1-2min
- **Kod**: 2-3min
- **Demonstracija**: 2-3min
- **Zaštita podataka**: 1-2min
- **Logovanje**: 1min

**Ukupno: ~10-15 minuta**

---

**Srećno! 🚀**
