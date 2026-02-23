# 🎯 Objašnjenje preporuka - UKRATKO

## 📊 Za Ivanu (pretplaćena na R&B)

**Rezultat:**
- ✅ **2 R&B pesme** u sekciji "Na osnovu vaših pretplata" (Sao Paolo, Blinding Lights)
- ✅ **1 Pop pesma** (Billie Jean) u sekciji "Otkrijte nešto novo"

**Zašto Billie Jean, a ne Come Together?**
- Query traži pesmu iz **nepretplaćenog žanra** (ne R&B)
- Pošto nema ocena 5 zvezdica, vraća se **prva pesma** po redosledu u bazi
- Billie Jean je prva Pop pesma u bazi → vraća se Billie Jean
- Come Together je Rock pesma → ne vraća se jer je možda kasnije u redosledu

## 📊 Za Anu (pretplaćena na Rock i Pop)

**Rezultat:**
- ✅ **Come Together (Rock)** u sekciji **"Na osnovu vaših pretplata"** ← OVO JE PERSONALIZOVANA PREPORUKA
- ✅ **Billie Jean (Pop)** u sekciji **"Otkrijte nešto novo"** ← OVO JE TOP PESMA

**Objašnjenje:**
- **"Na osnovu vaših pretplata"** = pesme iz žanrova na koje si pretplaćen (Rock, Pop)
  - Come Together je Rock pesma → prikazuje se jer si pretplaćena na Rock
- **"Otkrijte nešto novo"** = top pesma iz žanra na koji NISI pretplaćen
  - Ana je pretplaćena na Rock i Pop
  - Top pesma je iz drugog žanra (npr. R&B, Jazz, itd.) ili prva pesma ako nema ocena
  - Billie Jean je Pop → **GREŠKA!** Ana je pretplaćena na Pop, ne bi trebalo da vidi Pop pesmu u "Otkrijte nešto novo"

## ⚠️ Problem sa Aninom top pesmom

**Problem:** Ana vidi Billie Jean (Pop) u "Otkrijte nešto novo", ali je pretplaćena na Pop!

**Razlog:** Query možda ne radi pravilno ili je Billie Jean prva pesma u bazi pre nego što se filtriraju pretplate.

**Rešenje:** Treba proveriti da li query pravilno isključuje pretplaćene žanrove.

## 📋 Ukratko

1. **"Na osnovu vaših pretplata"** = pesme iz žanrova na koje si pretplaćen
2. **"Otkrijte nešto novo"** = top pesma iz žanra na koji NISI pretplaćen
3. **Top pesma se bira:** po broju ocena 5 zvezdica, ili prva pesma ako nema ocena
4. **Zašto je ista svima:** nema ocena 5 → vraća se prva pesma po redosledu u bazi
