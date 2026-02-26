# Implementacija Jaeger Tracing (2.10) - Kratak Vodič

## ✅ Šta je urađeno:

1. **Jaeger servis dodat u docker-compose.yml**
2. **Shared tracing biblioteka kreirana** (`services/shared/tracing/`)
3. **Tracing inicijalizacija u API Gateway**
4. **Test skripta i dokumentacija kreirana**

## 📝 Šta treba dodati u ostale servise:

Za svaki servis (users, content, ratings, subscriptions, notifications, recommendation), dodati:

### 1. Import tracing biblioteke:

```go
import (
    // ... postojeći imports ...
    "shared/tracing"
)
```

### 2. Inicijalizacija u main() funkciji (na početku):

```go
func main() {
    cfg := config.Load()

    // Initialize tracing (2.10)
    cleanup, err := tracing.InitTracing("service-name") // npr. "users-service"
    if err != nil {
        log.Printf("Warning: Failed to initialize tracing: %v", err)
    } else {
        defer cleanup()
        log.Println("Tracing initialized for service-name")
    }

    // ... ostatak koda ...
}
```

### 3. Ažurirati go.mod:

Dodati u `require` sekciju:
```go
require (
    // ... postojeći ...
    shared v0.0.0
)

replace shared => ../shared
```

### 4. Za asinhrone operacije (event emisije):

U `services/content-service/internal/events/emitter.go`, dodati tracing context u event payload:

```go
import (
    "context"
    "shared/tracing"
)

func EmitEvent(ctx context.Context, subscriptionsServiceURL string, event interface{}) {
    // Start span for async event emission
    ctx, span := tracing.StartSpan(ctx, "emit.event")
    defer span.End()

    // ... postojeći kod za slanje event-a ...
    
    // Add trace context to event headers
    req.Header.Set("traceparent", getTraceParent(ctx))
}
```

## 🚀 Brza Implementacija:

Za svaki servis, dodati samo 3 linije koda u `main()` funkciju (kao što je urađeno za api-gateway).

## 📊 Testiranje:

1. Pokreni: `docker-compose up -d`
2. Pokreni: `.\test-tracing-2.10.ps1`
3. Otvori: http://localhost:16686

## ⚠️ Napomena:

Tracing je opcionalan - ako Jaeger nije dostupan, servisi će raditi normalno (no-op tracer).
