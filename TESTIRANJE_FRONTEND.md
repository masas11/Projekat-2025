# Vodič za testiranje bezbednosnih funkcionalnosti na frontendu

## 🚀 Pokretanje aplikacije

1. **Pokreni backend servise:**
   ```bash
   # U terminalu, pokreni sve servise (users-service, api-gateway, itd.)
   # Koristi docker-compose ili pokreni ručno
   ```

2. **Pokreni frontend:**
   ```bash
   cd frontend
   npm start
   ```

3. Aplikacija će se otvoriti na `http://localhost:3000`

---

## 📋 Test Scenariji

### 1. **Registracija naloga sa email verifikacijom**

**Koraci:**
1. Idi na `/register`
2. Popuni formu:
   - Ime: `Marko`
   - Prezime: `Marković`
   - Email: `marko@test.com`
   - Username: `marko123` (jedinstven!)
   - Lozinka: `Test1234` (mora imati min 8 karaktera, veliko slovo, broj)
   - Potvrdi lozinku: `Test1234`
3. Klikni "Registruj se"
4. **Rezultat:**
   - Treba da vidiš poruku: "Uspešna registracija! Email za verifikaciju je poslat..."
   - Proveri konzolu backend servera - videćeš log: `[MOCK EMAIL] Sending verification email...`
   - U logu će biti link: `http://localhost:8081/api/users/verify-email?token=...`
5. **Email verifikacija:**
   - Kopiraj token iz konzole
   - Idi na `http://localhost:3000/verify-email?token=TVOJ_TOKEN`
   - Treba da vidiš poruku o uspešnoj verifikaciji
   - Automatski te preusmerava na login stranicu

**Test slučajevi:**
- ✅ Registracija sa jakom lozinkom
- ❌ Registracija sa slabom lozinkom (bez velikog slova/broja)
- ❌ Registracija sa nepoklapanjem lozinki
- ❌ Registracija sa istim username-om (treba error)

---

### 2. **Prijava sa OTP (One-Time Password)**

**Prvo verifikuj email** (ako nisi u prethodnom koraku)!

**Koraci:**
1. Idi na `/login`
2. Unesi:
   - Username: `marko123` (ili bilo koji registrovan korisnik)
   - Lozinka: `Test1234`
3. Klikni "Zatraži OTP"
4. **Proveri konzolu backend servera:**
   - Videćeš: `[MOCK EMAIL] Sending OTP 123456 to marko@test.com`
   - Zapamti OTP kod (npr. `123456`)
5. Unesi OTP kod u formu
6. Klikni "Verifikuj OTP"
7. **Rezultat:** Treba da te prijavi i preusmeri na početnu stranicu

**Test slučajevi:**
- ✅ Prijava sa ispravnim korisničkim podacima
- ❌ Prijava sa neverifikovanim email-om (treba error: "email not verified")
- ❌ Prijava sa pogrešnim lozinkom (treba error)
- ❌ Prijava sa pogrešnim OTP kodom (treba error)
- ❌ Prijava sa isteklom lozinkom (>60 dana) - simuliraj promenom baze

---

### 3. **Logout**

**Koraci:**
1. Biti prijavljen
2. Klikni "Odjavi se" u navbar-u
3. **Rezultat:** Treba da te odjavi i preusmeri na login stranicu
4. **Proveri:**
   - Token je obrisan iz localStorage
   - User podaci su obrisani (encriptovani podaci)
   - Ne možeš pristupiti zaštićenim rutama

---

### 4. **Reset lozinke (email link)**

**Koraci:**
1. Idi na `/login`
2. Klikni "Zaboravljena lozinka?"
3. Unesi email adresu (npr. `marko@test.com`)
4. Klikni "Pošalji link za reset"
5. **Proveri konzolu backend servera:**
   - Videćeš: `[MOCK EMAIL] Sending password reset email to marko@test.com`
   - Kopiraj token iz linka: `http://localhost:8081/api/users/password/reset?token=...`
6. **Koristi token:**
   - Idi na `http://localhost:3000/reset-password?token=TVOJ_TOKEN`
   - Unesi novu lozinku: `NovaLozinka123`
   - Potvrdi lozinku: `NovaLozinka123`
   - Klikni "Resetuj lozinku"
7. **Rezultat:** Lozinka je promenjena, preusmerava te na login

**Test slučajevi:**
- ✅ Reset sa validnim tokenom
- ❌ Reset sa isteklim tokenom (>1 sat)
- ❌ Reset sa slabom novom lozinkom
- ❌ Reset sa nepoklapanjem lozinki

---

### 5. **Promena lozinke (mora biti stara najmanje 1 dan)**

**Koraci:**
1. Biti prijavljen
2. Idi na `/change-password`
3. Unesi:
   - Stara lozinka: `Test1234` (trenutna)
   - Nova lozinka: `NovaLozinka123`
   - Potvrdi novu lozinku: `NovaLozinka123`
4. Klikni "Promeni lozinku"
5. **Rezultat:** 
   - Ako je lozinka stara više od 1 dana: uspešno
   - Ako je promenjena danas: error "password too new"

**Test slučajevi:**
- ✅ Promena sa starom lozinkom (stariju od 1 dana) - simuliraj u bazi ili sačkaj 1 dan
- ❌ Promena sa lozinkom promenjenom danas (treba error)
- ❌ Promena sa pogrešnom starom lozinkom
- ❌ Promena sa slabom novom lozinkom

---

### 6. **Magic Link - Povraćaj naloga**

**Koraci:**
1. Idi na `/login`
2. Klikni "Povraćaj naloga (Magic Link)"
3. Ili idi direktno na `/recover-account`
4. Unesi email adresu (npr. `marko@test.com`)
5. Klikni "Pošalji magic link"
6. **Proveri konzolu backend servera:**
   - Videćeš: `[MOCK EMAIL] Sending magic link to marko@test.com`
   - Kopiraj token iz linka: `http://localhost:8081/api/users/recover/verify?token=...`
7. **Koristi token:**
   - Otvori link direktno ili kopiraj token
   - Idi na `http://localhost:3000/verify-magic-link?token=TVOJ_TOKEN`
   - *Napomena: Ovo možda nije implementirano kao posebna stranica, ali možeš testirati direktno API endpoint*

**Test slučajevi:**
- ✅ Magic link sa validnim tokenom - automatski se prijavljuje
- ❌ Magic link sa isteklim tokenom (>15 minuta)

---

## 🔍 Provera bezbednosti

### Rate Limiting (DoS zaštita)
**Test:**
1. Otvori developer tools (F12)
2. Pokreni brzu petlju zahteva (npr. 20+ zahteva u sekundi)
3. Treba da vidiš error: "too many requests" nakon određenog broja zahteva

### Enkripcija state podataka
**Proveri:**
1. Nakon prijave, otvori DevTools → Application → Local Storage
2. Proveri `user` key:
   - Podaci treba da budu enkriptovani (ne čitljivi JSON)
   - Treba da postoji i `user_checksum` key
3. **Test integriteta:**
   - Promeni ručno vrednost u localStorage
   - Osvježi stranicu
   - Treba da detektuje promenu i obriše podatke

### Authorization middleware
**Test:**
1. Ne budi prijavljen
2. Pokušaj da pristupiš zaštićenim rutama (npr. `/notifications`)
3. Treba da te preusmeri na login ili pokaže error
4. **Nakon prijave:**
   - Pokušaj da kreiraš Artist (treba da budeš ADMIN)
   - Kao USER, treba da vidiš error "forbidden"

---

## 🐛 Debugging

### Gde naći logove:

**Backend (users-service):**
- Email poslate: Konzola gde radi `users-service`
- Traži: `[MOCK EMAIL] Sending...`
- OTP kodovi: Vidljivi u konzoli
- Tokeni: U logu linkova

**Frontend:**
- Browser DevTools → Console (greške)
- Browser DevTools → Network (HTTP zahtevi)
- Browser DevTools → Application → Local Storage (encriptovani podaci)

### Česti problemi:

1. **CORS error:**
   - Proveri da API Gateway radi na portu 8081
   - Proveri da frontend koristi pravu API URL

2. **Token expired:**
   - Tokeni imaju kratak vijek trajanja
   - Email verifikacija: 24 sata
   - Password reset: 1 sat
   - Magic link: 15 minuta
   - OTP: 5 minuta

3. **Email not verified:**
   - Proveri da si kliknuo na link za verifikaciju
   - Proveri u bazi da je `verified: true`

---

## ✅ Checklist za potpuno testiranje

- [ ] Registracija sa jakom lozinkom
- [ ] Email verifikacija (korišćenje tokena iz konzole)
- [ ] Prijava sa OTP (provera u konzoli)
- [ ] Logout funkcionalnost
- [ ] Reset lozinke (korišćenje email tokena)
- [ ] Promena lozinke (test "1 dan stara" ograničenja)
- [ ] Magic link za povraćaj naloga
- [ ] Rate limiting (DoS zaštita)
- [ ] Enkripcija state podataka (provera u localStorage)
- [ ] Authorization middleware (test zaštićenih ruta)
- [ ] Validacija na frontendu (slaba lozinka, nepoklapanje itd.)
- [ ] Validacija na backendu (SQL injection, XSS testovi)

---

## 📝 Napomene

- **Email servis je mock-ovano** - stvarni emailovi se ne šalju, ali su vidljivi u konzoli backend servera
- **Tokeni se prikazuju u konzoli** - koristi ih za testiranje
- **Sve lozinke moraju biti jake** - min 8 karaktera, veliko slovo, broj
- **Admin korisnik:** username: `admin`, password: `admin123` (kreira se automatski)

Srećno testiranje! 🎉
