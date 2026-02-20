# 🔐 HTTPS - Kratak Summary za Odbranu

## 🎯 GLAVNA PORUKA

**"Implementirali smo HTTPS protokol na svim slojevima komunikacije - od klijenta do API Gateway-a, i od API Gateway-a do backend servisa. Svi osetljivi podaci se šalju preko šifrovanog kanala."**

---

## 📋 ŠTA DA KAŽEŠ (3 minuta)

### 1. UVOD (30s)
"HTTPS obezbeđuje šifrovanje podataka u tranzitu. Implementirali smo ga na tri nivoa:
- Klijent → API Gateway
- API Gateway → Backend Servisi  
- Svi servisi imaju HTTPS servere"

### 2. IMPLEMENTACIJA (1min)
"Svaki servis proverava da li su sertifikati dostupni preko environment varijabli. Ako postoje, pokreće HTTPS server, u suprotnom fallback na HTTP. API Gateway prosleđuje zahteve backend servisima preko HTTPS-a koristeći konfigurisani HTTP klijent sa TLS podrškom."

### 3. DEMONSTRACIJA (1min)
"Pokazujem da HTTPS radi - testiram zahtev i proveravam sertifikate. Vidite da svi URL-ovi počinju sa `https://` i da komunikacija prolazi preko šifrovanog kanala."

---

## 🔑 KLJUČNE TAČKE

1. ✅ **HTTPS na svim slojevima** - ne samo na jednom
2. ✅ **Self-signed sertifikati** - za development
3. ✅ **Inter-service HTTPS** - čak i komunikacija između servisa
4. ✅ **Graceful degradation** - fallback na HTTP ako sertifikati nisu dostupni
5. ✅ **Production ready** - samo zameniti sertifikate

---

## 📁 FAJLOVI ZA POKAZIVANJE

1. **docker-compose.yml** (linije 10-16, 41-42, 19)
   - Environment varijable sa HTTPS URL-ovima
   - TLS sertifikati
   - Volume mount

2. **services/api-gateway/cmd/main.go** (linije 512-530, 76-85)
   - HTTPS server pokretanje
   - Inter-service HTTPS komunikacija

3. **certs/** direktorijum
   - `server.crt` i `server.key`

---

## ⚡ BRZE KOMANDE

```powershell
# 1. Provera HTTPS logova
docker logs projekat-2025-2-api-gateway-1 --tail 5 | Select-String "HTTPS"

# 2. Test HTTPS zahteva
. .\https-helper.ps1
Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/health"

# 3. Provera sertifikata
docker exec projekat-2025-2-api-gateway-1 ls -la /app/certs/

# 4. Provera environment varijabli
docker exec projekat-2025-2-api-gateway-1 env | Select-String "SERVICE_URL"
```

---

## 💬 ODGOVORI NA PITANJA

**P: Zašto self-signed sertifikati?**
O: "Za development. U production bi koristili CA-signed sertifikate."

**P: Zašto InsecureSkipVerify?**
O: "Samo za development sa self-signed sertifikatima. U production bi bilo false."

**P: Kako funkcioniše?**
O: "Svaki servis proverava TLS_CERT_FILE i TLS_KEY_FILE. Ako postoje, pokreće HTTPS server. API Gateway koristi HTTPS URL-ove za komunikaciju sa backend servisima."

---

## 📚 DETALJNI VODIČI

- **HTTPS_ODBRANA_VODIC.md** - Kompletan vodič sa objašnjenjima
- **HTTPS_ODBRANA_CHEATSHEET.md** - Brze komande i checklist
- **HTTPS_ARHITEKTURA_DIJAGRAM.md** - Vizuelni dijagrami

---

**Srećno na odbrani! 🚀**
