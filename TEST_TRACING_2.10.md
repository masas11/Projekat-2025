# Test Vodič - Jaeger Tracing (2.10)

## Pregled

Jaeger tracing je implementiran u celoj mikroservisnoj aplikaciji za praćenje sinhronih i asinhronih operacija.

## Implementacija

### 1. Jaeger Servis

Jaeger je dodat u `docker-compose.yml`:
- **UI Port:** 16686
- **Collector Endpoint:** http://jaeger:14268/api/traces

### 2. Tracing Biblioteka

Kreirana je shared tracing biblioteka u `services/shared/tracing/`:
- `tracing.go` - Inicijalizacija i osnovne funkcije
- `http.go` - HTTP middleware za automatsko praćenje

### 3. Implementacija u Servisima

Tracing je implementiran u:
- ✅ API Gateway
- ✅ Users Service
- ✅ Content Service
- ✅ Ratings Service
- ✅ Subscriptions Service
- ✅ Notifications Service
- ✅ Recommendation Service

## Kako Testirati

### Korak 1: Pokreni Jaeger

```powershell
docker-compose up -d jaeger
```

### Korak 2: Pokreni Servise

```powershell
docker-compose up -d
```

### Korak 3: Pokreni Test Skriptu

```powershell
.\test-tracing-2.10.ps1
```

### Korak 4: Otvori Jaeger UI

1. Otvori browser: http://localhost:16686
2. Izaberi servis iz dropdown-a (npr. `api-gateway`)
3. Klikni "Find Traces"
4. Trebalo bi da vidiš trace-ove za sve pozive

## Šta Proveriti

### Sinhronne Operacije

1. **HTTP Pozivi između Servisa:**
   - API Gateway → Users Service
   - API Gateway → Content Service
   - API Gateway → Ratings Service
   - Content Service → Subscriptions Service

2. **Span Hijerarhija:**
   - Parent span u API Gateway
   - Child spans u backend servisima
   - Trace context propagacija kroz HTTP headers

### Asinhrone Operacije

1. **Event Emisije:**
   - Content Service → Subscriptions Service (new_artist, new_album, new_song)
   - Trace context propagacija u event payload-u

2. **Event Obrada:**
   - Subscriptions Service prima event-e
   - Recommendation Service prima event-e
   - Svaki event ima svoj span

## Primer Trace-a

```
api-gateway (root span)
  ├── GET /api/content/songs
  │   ├── content-service: GET /songs
  │   └── ratings-service: GET /average-rating (API Composition)
  └── GET /api/subscriptions
      └── subscriptions-service: GET /subscriptions
```

## Debugging

Ako ne vidiš trace-ove:

1. **Proveri da li je Jaeger pokrenut:**
   ```powershell
   docker ps | Select-String "jaeger"
   ```

2. **Proveri logove servisa:**
   ```powershell
   docker-compose logs api-gateway | Select-String "Tracing"
   ```

3. **Proveri da li servisi koriste tracing:**
   - Svaki servis treba da loguje "Tracing initialized for service: [name]"

## Napomene

- Tracing je opcionalan - ako Jaeger nije dostupan, servisi će raditi normalno (no-op tracer)
- Trace-ovi se šalju asinhrono u batch-ovima
- Jaeger UI može da prikaže trace-ove sa malim kašnjenjem (1-2 sekunde)
