# TLS Handshake Greške - Objašnjenje

## ⚠️ Greške koje vidite:

```
TLS handshake error from 172.20.0.1:48226: EOF
TLS handshake error from 172.20.0.1:48212: EOF
TLS handshake error from 172.20.0.1:48230: EOF
```

## 🔍 Šta to znači?

**TLS Handshake Error:**
- Klijent (browser) pokušava da se konektuje na HTTPS server
- TLS handshake proces se pokreće (razmena sertifikata, šifrovanje)
- Klijent **prekida konekciju** pre završetka handshake-a (EOF = End Of File)

**IP adresa `172.20.0.1`:**
- Ovo je Docker host IP adresa
- Browser komunicira sa API Gateway preko ove adrese

## ✅ Da li je to problem?

**NE - ovo je normalno za development sa self-signed sertifikatima!**

**Razlog:**
1. Browser pokušava HTTPS konekciju
2. Vidi self-signed sertifikat
3. Prekida konekciju jer ne veruje sertifikat
4. Kada korisnik prihvati sertifikat → handshake uspeva

## 🔧 Kako da rešite?

### 1. Prihvatite sertifikat u browser-u:

1. Otvorite: `https://localhost:8081/api/users/health`
2. Browser će pokazati upozorenje
3. Kliknite **"Advanced"** ili **"Napredno"**
4. Kliknite **"Proceed to localhost (unsafe)"** ili **"Nastavi na localhost"**
5. Browser će zapamtiti sertifikat

### 2. Proverite da li HTTPS radi:

```powershell
# Proverite logove - trebalo bi da vidite i uspešne zahteve
docker logs projekat-2025-1-api-gateway-1 --tail 20
```

**Nakon prihvatanja sertifikata:**
- TLS handshake greške će nestati
- Videćete samo uspešne zahteve

## 📊 Status Sertifikata

### Provera sertifikata:

```powershell
ls certs\
```

**Trebalo bi da vidite:**
- `server.crt` - SSL sertifikat ✅
- `server.key` - Privatni ključ ✅

### Provera da li HTTPS server radi:

```powershell
docker logs projekat-2025-1-api-gateway-1 --tail 3
```

**Trebalo bi da vidite:**
- `Starting HTTPS server on port 8080` ✅

## ✅ Zaključak

**Sertifikati su OK:**
- ✅ Sertifikati postoje u `certs/` direktorijumu
- ✅ HTTPS server je pokrenut
- ✅ TLS handshake greške su normalne za self-signed sertifikate

**Šta uraditi:**
1. Prihvatite sertifikat u browser-u
2. Ove greške će nestati
3. HTTPS će raditi normalno

**Za odbranu:**
- Ove greške su dokaz da HTTPS radi (handshake se pokušava)
- Nakon prihvatanja sertifikata, greške nestaju
- To je normalno ponašanje za development sa self-signed sertifikatima
