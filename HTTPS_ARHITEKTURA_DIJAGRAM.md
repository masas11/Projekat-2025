# 🔐 HTTPS Arhitektura - Vizuelni Dijagram

## 📊 ARHITEKTURA KOMUNIKACIJE

```
┌─────────────────────────────────────────────────────────────────┐
│                        KLIJENT (Browser)                         │
│                    http://localhost:3000                         │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ HTTPS (šifrovano)
                             │ Port 8081
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      API GATEWAY                                 │
│              https://localhost:8081                              │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  TLS_CERT_FILE=/app/certs/server.crt                  │    │
│  │  TLS_KEY_FILE=/app/certs/server.key                   │    │
│  │                                                         │    │
│  │  ListenAndServeTLS(certFile, keyFile)                  │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                   │
│  Environment Variables:                                          │
│  • USERS_SERVICE_URL=https://users-service:8001                 │
│  • CONTENT_SERVICE_URL=https://content-service:8002              │
│  • RATINGS_SERVICE_URL=https://ratings-service:8003             │
│  • ...                                                           │
└───────┬───────────────┬───────────────┬───────────────┬────────┘
        │               │               │               │
        │ HTTPS         │ HTTPS         │ HTTPS         │ HTTPS
        │ (šifrovano)   │ (šifrovano)   │ (šifrovano)   │ (šifrovano)
        │               │               │               │
        ▼               ▼               ▼               ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│   USERS     │ │  CONTENT    │ │  RATINGS    │ │     ...      │
│  SERVICE    │ │  SERVICE    │ │  SERVICE    │ │              │
│             │ │             │ │             │ │              │
│ Port: 8001  │ │ Port: 8002  │ │ Port: 8003  │ │              │
│             │ │             │ │             │ │              │
│ TLS_CERT    │ │ TLS_CERT    │ │ TLS_CERT    │ │ TLS_CERT     │
│ TLS_KEY     │ │ TLS_KEY     │ │ TLS_KEY     │ │ TLS_KEY      │
│             │ │             │ │             │ │              │
│ ListenAnd   │ │ ListenAnd   │ │ ListenAnd   │ │ ListenAnd    │
│ ServeTLS    │ │ ServeTLS    │ │ ServeTLS    │ │ ServeTLS     │
└─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘
```

---

## 🔄 TLS HANDSHAKE PROCES

```
┌──────────┐                                    ┌──────────┐
│ KLIJENT  │                                    │ SERVER   │
└────┬─────┘                                    └────┬─────┘
     │                                               │
     │  1. ClientHello                               │
     │     - TLS verzija                            │
     │     - Cipher suite-ovi                        │
     │     - Random bytes                           │
     ├──────────────────────────────────────────────>│
     │                                               │
     │                         2. ServerHello       │
     │                            - Odabrana TLS    │
     │                            - Cipher suite   │
     │                            - Sertifikat      │
     │                            - Random bytes    │
     │<──────────────────────────────────────────────┤
     │                                               │
     │  3. Verifikacija sertifikata                 │
     │     (u production)                           │
     │                                               │
     │  4. Razmena ključeva                         │
     │     (Diffie-Hellman ili RSA)                 │
     ├──────────────────────────────────────────────>│
     │                                               │
     │                         5. Finished          │
     │<──────────────────────────────────────────────┤
     │                                               │
     │  6. Finished                                 │
     ├──────────────────────────────────────────────>│
     │                                               │
     │  ✅ Šifrovana komunikacija                   │
     │<══════════════════════════════════════════════>│
```

---

## 📦 STRUKTURA SERTIFIKATA

```
certs/
├── server.crt          (Javni sertifikat)
│   ├── Javni ključ
│   ├── Informacije o serveru
│   └── Potpis
│
└── server.key          (Privatni ključ)
    └── Privatni ključ (zaštićen permisijama 0600)
```

**Mount u Docker:**
```
Host: ./certs/  →  Container: /app/certs/:ro
```

---

## 🔐 ŠTA SE ŠIFRUJE

### Podaci koji se šalju preko HTTPS-a:

```
┌─────────────────────────────────────────┐
│  KLIJENT → API GATEWAY                   │
├─────────────────────────────────────────┤
│  ✅ Lozinke (registracija, promena)     │
│  ✅ JWT tokeni (autentifikacija)         │
│  ✅ OTP kodovi (verifikacija)            │
│  ✅ Lični podaci (email, ime, prezime)  │
│  ✅ Admin akcije                         │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│  API GATEWAY → BACKEND SERVISI          │
├─────────────────────────────────────────┤
│  ✅ Svi zahtevi sa podacima             │
│  ✅ JWT tokeni u headers                │
│  ✅ User ID i drugi identifikatori     │
│  ✅ Query parametri sa osetljivim       │
│     podacima                            │
└─────────────────────────────────────────┘
```

---

## 🛡️ ZAŠTITA NA RAZLIČITIM NIVOIMA

```
┌─────────────────────────────────────────────────────────┐
│  NIVO 1: Klijent → API Gateway                          │
│  ✅ HTTPS šifrovanje                                    │
│  ✅ Zaštita od presretanja                              │
│  ✅ Zaštita od man-in-the-middle napada                 │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────┐
│  NIVO 2: API Gateway → Backend Servisi                  │
│  ✅ HTTPS šifrovanje                                    │
│  ✅ Zaštita inter-service komunikacije                  │
│  ✅ Defense in depth                                    │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────┐
│  NIVO 3: Backend Servisi                                │
│  ✅ HTTPS serveri na svim servisima                    │
│  ✅ Zaštita od presretanja u Docker mreži               │
│  ✅ Zaštita podataka u tranzitu                         │
└─────────────────────────────────────────────────────────┘
```

---

## 🔍 IMPLEMENTACIJA U KODU

### API Gateway - HTTPS Server

```go
certFile := os.Getenv("TLS_CERT_FILE")  // /app/certs/server.crt
keyFile := os.Getenv("TLS_KEY_FILE")    // /app/certs/server.key

if certFile != "" && keyFile != "" {
    // HTTPS server
    server := &http.Server{
        Addr:    ":" + cfg.Port,
        Handler: mux,
    }
    server.ListenAndServeTLS(certFile, keyFile)
} else {
    // Fallback na HTTP
    http.ListenAndServe(":"+cfg.Port, mux)
}
```

### API Gateway - Inter-Service HTTPS

```go
// Konfiguracija HTTP klijenta za HTTPS
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

---

## 📊 FLOW DIJAGRAM ZAHTEVA

```
1. Klijent šalje zahtev
   ┌──────────┐
   │ Browser  │
   └────┬─────┘
        │ HTTPS POST /api/users/register
        │ { email, password, ... }
        ▼
   ┌─────────────┐
   │ API Gateway │
   │ (HTTPS)     │
   └────┬────────┘
        │ HTTPS POST https://users-service:8001/register
        │ { email, password, ... }
        ▼
   ┌─────────────┐
   │ Users       │
   │ Service     │
   │ (HTTPS)     │
   └─────────────┘
        │
        │ Procesiranje
        │ Validacija
        │ Hash lozinke
        │
        │ HTTPS Response
        │ { status: 201, ... }
        ▼
   ┌─────────────┐
   │ API Gateway │
   └────┬────────┘
        │ HTTPS Response
        │ { status: 201, ... }
        ▼
   ┌──────────┐
   │ Browser  │
   └──────────┘
```

---

## ✅ CHECKLIST ZA ODBRANU

- [ ] Pokaži docker-compose.yml - environment varijable
- [ ] Pokaži kod - HTTPS server implementacija
- [ ] Pokaži kod - Inter-service HTTPS komunikacija
- [ ] Demonstriraj - Test HTTPS zahteva
- [ ] Demonstriraj - Provera sertifikata
- [ ] Objasni - TLS handshake proces
- [ ] Objasni - Zašto HTTPS na svim slojevima
- [ ] Objasni - Production readiness

---

**Koristi ovaj dijagram za vizuelno objašnjenje na odbrani! 📊**
