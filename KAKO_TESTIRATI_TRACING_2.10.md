# Kako Testirati Jaeger Tracing (2.10) - UKRATKO

## ✅ Implementacija Završena!

Tracing je implementiran u **SVIM servisima**:
- ✅ API Gateway
- ✅ Users Service
- ✅ Content Service
- ✅ Ratings Service
- ✅ Subscriptions Service
- ✅ Notifications Service
- ✅ Recommendation Service

## 🚀 Kako Testirati (3 koraka):

### 1. Rebuild-uj sve servise:
```powershell
docker-compose up -d --build
```

### 2. Generiši trace-ove (napravi HTTP pozive):
```powershell
# Opcija A: Koristi test skriptu
.\test-tracing-2.10.ps1

# Opcija B: Ili ručno napravi pozive
Invoke-WebRequest -Uri "http://localhost:8081/health" -UseBasicParsing
Invoke-WebRequest -Uri "http://localhost:8081/api/content/songs" -UseBasicParsing
```

### 3. Otvori Jaeger UI:
1. Otvori: **http://localhost:16686**
2. Izaberi servis: **`api-gateway`** (ne `jaeger-all-in-one`)
3. Klikni: **"Find Traces"**
4. Trebalo bi da vidiš trace-ove!

## 📊 Šta Proveriti:

- **Sinhronne operacije:** HTTP pozivi između servisa
- **Asinhrone operacije:** Event emisije (new_artist, new_album, new_song)
- **Span hijerarhija:** Parent-child odnos između servisa
- **Trace context propagacija:** Kroz HTTP headers i event-e

## ⚠️ Ako ne vidiš trace-ove:

1. Proveri da li su servisi rebuild-ovani: `docker-compose ps`
2. Proveri logove: `docker-compose logs api-gateway | Select-String "Tracing"`
3. Sačekaj 2-3 sekunde (trace-ovi se šalju asinhrono)

---

**Sve je spremno! Rebuild-uj servise i testiraj!** 🎉
