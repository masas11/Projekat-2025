# Uputstvo za Demonstraciju Napada - 2.22

## 📋 Preduslovi

1. **Aplikacija mora biti pokrenuta:**
   ```powershell
   docker-compose up -d
   ```

2. **Proverite da li su servisi aktivni:**
   ```powershell
   docker-compose ps
   ```

3. **Proverite da li postoji `https-helper.ps1`:**
   - Ako ne postoji, kreirajte ga (već bi trebalo da postoji iz prethodnih testova)

## 🚀 Brzo Pokretanje

### Opcija 1: Pokrenite sve napade odjednom

```powershell
.\test-all-attacks.ps1
```

Ovo će pokrenuti sve 4 napada redom:
1. XSS napad
2. SQL Injection napad
3. Brute-force napad
4. DoS napad

### Opcija 2: Pokrenite napade pojedinačno

```powershell
# XSS napad
.\test-xss-attack.ps1

# SQL Injection napad
.\test-sql-injection-attack.ps1

# Brute-force napad
.\test-brute-force-attack.ps1

# DoS napad
.\test-dos-attack.ps1
```

## 📊 Očekivani Rezultati

### 1. XSS Napad
- **Status:** ✅ Blokiran
- **HTTP Status:** 400 Bad Request
- **Poruka:** "invalid input"
- **Log:** `VALIDATION_FAILURE` sa razlogom "XSS attempt detected"

### 2. SQL Injection Napad
- **Status:** ✅ Blokiran
- **HTTP Status:** 400 Bad Request
- **Poruka:** "invalid input"
- **Log:** `VALIDATION_FAILURE` sa razlogom "SQL injection attempt detected"

### 3. Brute-force Napad
- **Status:** ✅ Blokiran
- **HTTP Status:** 403 Forbidden (nakon 5 neuspešnih pokušaja)
- **Poruka:** "account locked"
- **Log:** `LOGIN_FAILURE` za svaki neuspešan pokušaj

### 4. DoS Napad
- **Status:** ✅ Blokiran
- **HTTP Status:** 429 Too Many Requests (nakon 100 zahteva/min)
- **Poruka:** "too many requests"
- **Log:** `ACCESS_CONTROL_FAILURE` sa razlogom "rate limit exceeded"

## 🔍 Provera Logova

### XSS i SQL Injection
```powershell
docker logs projekat-2025-2-users-service-1 | Select-String "VALIDATION_FAILURE"
```

### Brute-force
```powershell
docker logs projekat-2025-2-users-service-1 | Select-String "LOGIN_FAILURE"
```

### DoS
```powershell
docker logs projekat-2025-2-api-gateway-1 | Select-String "too many requests"
```

## 📝 Napomene

1. **Brute-force napad zahteva postojeći nalog:**
   - Pre pokretanja, kreirajte test korisnika:
     ```powershell
     .\kreiraj-test-korisnika.ps1
     ```
   - Ili registrujte korisnika "testuser" sa lozinkom "Test1234!" kroz frontend

2. **DoS napad može trajati nekoliko minuta:**
   - Test šalje 150 zahteva sa pauzom od 100ms između zahteva
   - Ukupno vreme: ~15 sekundi

3. **Ako testovi ne rade:**
   - Proverite da li su servisi pokrenuti
   - Proverite da li postoji `https-helper.ps1`
   - Proverite da li je HTTPS konfigurisan

## 📚 Detaljna Dokumentacija

Za detaljnu dokumentaciju o svakom napadu, pogledajte:
- `DEMONSTRACIJA_NAPADA_2.22.md`

## ✅ Checklist za Odbranu

- [ ] Svi servisi su pokrenuti
- [ ] XSS napad je testiran i blokiran
- [ ] SQL Injection napad je testiran i blokiran
- [ ] Brute-force napad je testiran i blokiran
- [ ] DoS napad je testiran i blokiran
- [ ] Logovi su provereni
- [ ] Dokumentacija je pročitana

---

**Srećna odbrana! 🎓**
