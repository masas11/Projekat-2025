# 🧪 Kako Testirati Pretplatu na Žanrove

## 🎯 Brzi Vodič

### **1. Prijavite se kao korisnik**
- Idite na `/login`
- Unesite korisničko ime i lozinku
- Dobijte OTP i unesite ga

### **2. Idite na stranicu pesama**
- Kliknite na "Pesme" u navigaciji
- Ili idite direktno na `/songs`

### **3. Pretplata na žanr:**
1. **Izaberite žanr** iz dropdown-a "Filtriranje po žanru"
   - Primeri: Pop, Rock, Jazz, R&B, itd.
2. **Kliknite na ikonu 🔔** pored dropdown-a
3. **Poruka će se pojaviti:** "Uspešno ste se pretplatili na žanr: [naziv]!"
4. **Ikona se menja u ✓** (siva boja)

### **4. Otkazivanje pretplate:**
1. **Izaberite isti žanr** iz dropdown-a
2. **Kliknite na ikonu ✓** (sada je pretplaćen)
3. **Poruka će se pojaviti:** "Uspešno ste se odjavili sa pretplate na žanr: [naziv]!"
4. **Ikona se menja nazad u 🔔**

### **5. Pregled pretplata na profilu:**
1. **Idite na "Moj Profil"** u navigaciji
2. **Vidite sekciju "Pretplate na Žanrove"**
3. **Vidite sve žanrove na koje ste pretplaćeni**
4. **Kliknite "Otkaži pretplatu"** da otkažete bilo koju pretplatu

---

## 🔍 Detaljno Testiranje

### **Test 1: Pretplata na žanr**

```
1. Prijavite se kao korisnik
2. Idite na /songs
3. Izaberite "Pop" iz dropdown-a
4. Kliknite 🔔
5. Očekivani rezultat:
   - Poruka: "Uspešno ste se pretplatili na žanr: Pop!"
   - Ikona se menja u ✓
   - Na profilu vidite "Pop" u listi pretplata
```

### **Test 2: Otkazivanje pretplate**

```
1. Izaberite "Pop" iz dropdown-a (već ste pretplaćeni)
2. Kliknite ✓
3. Očekivani rezultat:
   - Poruka: "Uspešno ste se odjavili sa pretplate na žanr: Pop!"
   - Ikona se menja nazad u 🔔
   - Na profilu "Pop" više nije u listi
```

### **Test 3: Pregled na profilu**

```
1. Idite na /profile
2. Proverite sekciju "Pretplate na Žanrove"
3. Trebalo bi da vidite sve žanrove na koje ste pretplaćeni
4. Kliknite "Otkaži pretplatu" na bilo kom žanru
5. Očekivani rezultat:
   - Poruka: "Uspešno ste se odjavili sa pretplate"
   - Žanr se uklanja iz liste
```

### **Test 4: Zaštita od duplikata**

```
1. Pretplatite se na "Pop"
2. Pokušajte ponovo da se pretplatite na "Pop"
3. Očekivani rezultat:
   - Poruka: "Već ste pretplaćeni na ovaj žanr"
   - Ili: "Already subscribed to this genre"
```

---

## 🐛 Troubleshooting

### Problem: "Cannot read properties of null (reading 'filter')"
**Rešenje:** ✅ **Popravljeno!** Dodate provere da li je rezultat array pre pozivanja `.filter()`

### Problem: Ikona se ne menja
**Rešenje:** 
- Proverite da li ste prijavljeni
- Proverite browser konzolu za greške
- Osvežite stranicu

### Problem: Pretplata ne radi
**Rešenje:**
- Proverite da li je subscriptions-service pokrenut: `docker-compose ps`
- Proverite logove: `docker-compose logs subscriptions-service`
- Proverite da li je MongoDB pokrenut: `docker ps | findstr mongodb-subscriptions`

---

## ✅ Checklist za Testiranje

- [ ] Prijavljeni ste kao korisnik
- [ ] Idite na `/songs`
- [ ] Izaberite žanr iz dropdown-a
- [ ] Kliknite 🔔 za pretplatu
- [ ] Proverite da se ikona promenila u ✓
- [ ] Idite na `/profile`
- [ ] Proverite da je žanr u listi pretplata
- [ ] Otkažite pretplatu sa profila
- [ ] Proverite da se žanr uklonio iz liste

---

## 🎯 UI Elementi

### **Songs stranica:**
- Dropdown za filtriranje po žanru
- Ikona 🔔/✓ pored dropdown-a
- Poruka o uspehu/grešci

### **Profile stranica:**
- Sekcija "Pretplate na Žanrove"
- Lista svih pretplata sa datumom
- Dugme "Otkaži pretplatu" za svaki žanr

---

## 📝 Napomene

- Pretplata je samo na **žanrove**, ne na pesme
- Pretplata je samo na **umetnike**, ne na albume ili pesme
- Sve pretplate se čuvaju u MongoDB bazi
- Status pretplate se automatski ažurira u UI-u
