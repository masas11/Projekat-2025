# 🔐 HTTPS Test Endpoints - Kako da Pokažeš da HTTPS Radi

## ✅ DOBAR ZNAK: 404 Greška

**Šta znači 404:**
- ✅ HTTPS **RADI** - browser se uspešno povezao na HTTPS server
- ✅ Server **ODGOVARA** - nije connection error
- ⚠️ Samo nema root endpoint-a (`/`) - što je normalno

**"Not secure" upozorenje:**
- To je normalno za self-signed sertifikate
- Browser ne veruje sertifikat, ali HTTPS **RADI**
- Podaci su i dalje **ŠIFROVANI**

---

## 🎯 PRAVI ENDPOINT-I ZA TESTIRANJE

### 1. **API Gateway Health Check**

```
https://localhost:8081/health
```

**Očekivano:**
```json
{
  "status": "healthy",
  "service": "api-gateway"
}
```

### 2. **Users Service Health Check**

```
https://localhost:8081/api/users/health
```

**Očekivano:**
```
users-service is running
```

### 3. **Test Registracije**

```
https://localhost:8081/api/users/register
```

**Metoda:** POST
**Body:** JSON sa user podacima

---

## 📋 KAKO DA POKAŽEŠ NA ODBRANI

### Opcija 1: Browser (Najlakše)

1. **Otvorite u browser-u:**
   ```
   https://localhost:8081/health
   ```

2. **Kliknite "Advanced" → "Proceed to localhost (unsafe)"**

3. **Vidite JSON odgovor:**
   ```json
   {
     "status": "healthy",
     "service": "api-gateway"
   }
   ```

4. **Objasnite:**
   "Vidite da HTTPS radi - server odgovara sa JSON podacima. 'Not secure' upozorenje je normalno za self-signed sertifikate, ali HTTPS protokol radi i podaci su šifrovani."

### Opcija 2: PowerShell (Za demonstraciju)

```powershell
. .\https-helper.ps1
$result = Invoke-HTTPSRequest -Uri "https://localhost:8081/api/users/health"
Write-Host "Status: $($result.StatusCode)" -ForegroundColor Green
Write-Host "Content: $($result.Content)"
```

**Očekivano:**
```
Status: 200
Content: users-service is running
```

### Opcija 3: Browser Developer Tools (Najbolje za odbranu)

1. **Otvorite aplikaciju** (`localhost:3000`)
2. **Otvorite Developer Tools** (F12)
3. **Network tab**
4. **Napravite neki zahtev** (login, registracija)
5. **Pokažite da URL-ovi počinju sa `https://`**
6. **Pokažite da zahtevi prolaze** (Status 200, 201)

---

## 💬 ŠTA DA KAŽEŠ NA ODBRANI

### Ako te pitaju za 404:

"404 greška na root endpoint-u (`/`) je normalna jer API Gateway nema root endpoint. To je dobar znak - znači da HTTPS radi i server odgovara. Pravi endpoint-i su pod `/api/` putanjom, npr. `/api/users/health`."

### Ako te pitaju za "Not secure":

"'Not secure' upozorenje je normalno za self-signed sertifikate. Browser ne veruje sertifikat jer nije potpisan od strane Certificate Authority (CA). Međutim, HTTPS **RADI** - vidite da server odgovara i podaci su šifrovani u tranzitu. U production okruženju, sertifikati bi bili potpisani od strane CA (npr. Let's Encrypt), pa browser ne bi prikazivao upozorenje."

---

## ✅ CHECKLIST ZA DEMONSTRACIJU

- [ ] Otvori `https://localhost:8081/health` u browser-u
- [ ] Prihvati sertifikat ("Advanced" → "Proceed")
- [ ] Pokaži JSON odgovor
- [ ] Objasni da HTTPS radi (404 je normalno za root endpoint)
- [ ] Pokaži pravi endpoint (`/api/users/health`)
- [ ] Demonstriraj kroz Developer Tools → Network tab

---

## 🎯 KLJUČNE TAČKE

1. ✅ **404 je dobar znak** - HTTPS radi, samo nema root endpoint-a
2. ✅ **"Not secure" je normalno** - zbog self-signed sertifikata
3. ✅ **Koristi prave endpoint-e** - `/health` ili `/api/users/health`
4. ✅ **Najbolje pokazati kroz Network tab** - vidite sve zahteve sa `https://`

---

**Srećno na odbrani! 🚀**
