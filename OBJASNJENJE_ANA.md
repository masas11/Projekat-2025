# 🎯 Objašnjenje za Anu - UKRATKO

## ✅ Nakon sinhronizacije

**Ana je pretplaćena na:**
- Pop
- Rock

**Rezultat:**
- ✅ **13 pesama** u sekciji "Na osnovu vaših pretplata" (12 Pop + 1 Rock)
- ✅ **Top pesma: Sao Paolo (R&B)** u sekciji "Otkrijte nešto novo"

## 📊 Zašto "Sao Paolo" a ne "Billie Jean"?

**Razlog:**
1. Ana je **pretplaćena na Pop** → Billie Jean (Pop) se prikazuje u "Na osnovu vaših pretplata"
2. Top pesma mora biti iz **nepretplaćenog žanra** (ne Pop, ne Rock)
3. Query traži pesmu sa najviše ocena 5 zvezdica iz nepretplaćenih žanrova
4. Pošto nema ocena 5, vraća se **prva pesma** po redosledu (alfabetski ili po ID-u)
5. **Sao Paolo (R&B)** je prva R&B pesma u bazi → vraća se Sao Paolo

## 🔍 Zašto je top pesma različita za različite korisnike?

**Ivana (pretplaćena na R&B):**
- Top pesma: **Billie Jean (Pop)**
- Razlog: Ivana NIJE pretplaćena na Pop → Billie Jean je iz nepretplaćenog žanra

**Ana (pretplaćena na Pop i Rock):**
- Top pesma: **Sao Paolo (R&B)**
- Razlog: Ana NIJE pretplaćena na R&B → Sao Paolo je iz nepretplaćenog žanra

**Zaključak:**
- Top pesma zavisi od **pretplaćenih žanrova** korisnika
- Svaki korisnik vidi top pesmu iz **različitog nepretplaćenog žanra**
- Ako nema ocena 5, vraća se prva pesma po redosledu u bazi

## 📋 Ukratko

1. **"Na osnovu vaših pretplata"** = pesme iz žanrova na koje si pretplaćen
   - Ana: 12 Pop pesama + 1 Rock pesma = 13 pesama

2. **"Otkrijte nešto novo"** = top pesma iz žanra na koji NISI pretplaćen
   - Ana: Sao Paolo (R&B) jer nije pretplaćena na R&B
   - Ivana: Billie Jean (Pop) jer nije pretplaćena na Pop

3. **Top pesma se bira:**
   - Po broju ocena 5 zvezdica
   - Ili prva pesma po redosledu ako nema ocena 5
