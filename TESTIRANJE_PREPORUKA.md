# 🎯 Vodič za testiranje personalizovanih preporuka

## 📋 Kako rade preporuke (ukratko)

Sistem koristi **Neo4j graf bazu** za čuvanje veza između korisnika, pesama, žanrova i ocena.

### Algoritam preporuke:

1. **Pesme iz pretplaćenih žanrova** (`GetSubscribedGenreSongs`):
   - Pronalazi sve pesme koje pripadaju žanrovima na koje je korisnik pretplaćen
   - **Isključuje** pesme koje je korisnik ocenio sa **manje od 4 zvezdice**
   - **Uključuje** pesme koje je korisnik ocenio sa **4 ili 5 zvezdica** ILI **nije ocenio uopšte**
   - Maksimalno 50 pesama

2. **Top pesma iz nepretplaćenog žanra** (`GetTopRatedSongFromUnsubscribedGenre`):
   - Pronalazi pesmu koja pripada žanru na koji korisnik **NIJE** pretplaćen
   - Pesma mora imati **najviše ocena 5 zvezdica** od strane drugih korisnika
   - Ako korisnik nema pretplate, vraća najpopularniju pesmu uopšte

### Graf struktura:
- **User** → `SUBSCRIBED_TO` → **Genre**
- **User** → `RATED {rating: 1-5}` → **Song**
- **Song** → `BELONGS_TO` → **Genre**
- **Song** → `PERFORMED_BY` → **Artist**

---

## ✅ Da li radi prema specifikaciji?

**DA**, implementacija je u skladu sa specifikacijom:

✅ Prikazuje pesme iz pretplaćenih žanrova (ocena >= 4 ili nije ocenjena)  
✅ Prikazuje top pesmu iz nepretplaćenog žanra (najviše ocena 5)  
✅ Koristi Neo4j graf bazu  
✅ Ažurira se asinhrono preko eventova

---

## 🧪 Kako testirati

### **Test Scenario 1: Novi korisnik (bez pretplata i ocena)**

1. **Registruj se** kao novi korisnik (npr. `testuser`)
2. **Prijavi se** na sistem
3. **Idi na početnu stranicu** (`/`)
4. **Očekivani rezultat:**
   - Trebalo bi da vidiš **jednu pesmu** sa razlogom "Popular in genre you might like"
   - Ovo je najpopularnija pesma u sistemu (najviše ocena 5)

---

### **Test Scenario 2: Pretplata na žanr**

1. **Prijavi se** kao korisnik (npr. `uros`)
2. **Idi na stranicu Artists ili Albums**
3. **Pretplati se na žanr** (npr. "R&B")
4. **Idi na početnu stranicu** (`/`)
5. **Očekivani rezultat:**
   - Trebalo bi da vidiš **sekciju "Na osnovu vaših pretplata"** sa pesmama iz R&B žanra
   - Sve pesme koje pripadaju R&B žanru (osim onih koje si ocenio sa < 4)
   - Plus **sekciju "Otkrijte nešto novo"** sa top pesmom iz drugog žanra

---

### **Test Scenario 3: Ocenjivanje pesama**

1. **Prijavi se** kao korisnik
2. **Pretplati se na žanr** (npr. "Pop")
3. **Oceni neku Pop pesmu sa 3 zvezdice**
4. **Oceni drugu Pop pesmu sa 5 zvezdica**
5. **Idi na početnu stranicu** (`/`)
6. **Očekivani rezultat:**
   - **NE bi trebalo** da vidiš pesmu koju si ocenio sa 3 zvezdice
   - **TREBALO BI** da vidiš pesmu koju si ocenio sa 5 zvezdica
   - **TREBALO BI** da vidiš sve ostale Pop pesme koje nisi ocenio

---

### **Test Scenario 4: Kombinovano (pretplata + ocene)**

1. **Prijavi se** kao korisnik
2. **Pretplati se na "R&B"**
3. **Oceni 2 R&B pesme sa 5 zvezdica**
4. **Idi na početnu stranicu** (`/`)
5. **Očekivani rezultat:**
   - **Sekcija "Na osnovu vaših pretplata":**
     - 2 pesme koje si ocenio sa 5 zvezdica
     - Sve ostale R&B pesme koje nisi ocenio
   - **Sekcija "Otkrijte nešto novo":**
     - Top pesma iz drugog žanra (npr. Pop, Rock, itd.)

---

### **Test Scenario 5: Brisanje pretplate**

1. **Prijavi se** kao korisnik
2. **Pretplati se na "Pop"**
3. **Idi na početnu stranicu** - vidiš Pop pesme
4. **Odpretplati se od "Pop"** (na Profile stranici)
5. **Osveži početnu stranicu**
6. **Očekivani rezultat:**
   - **NE bi trebalo** da vidiš Pop pesme u sekciji "Na osnovu vaših pretplata"
   - **TREBALO BI** da vidiš samo top pesmu iz nepretplaćenog žanra

---

## 🔍 Debugging - Kako proveriti da li radi

### 1. Proveri Neo4j graf bazu:

```bash
# Pristupi Neo4j browseru
# URL: http://localhost:7474
# Username: neo4j
# Password: password
```

**Query za proveru pretplata:**
```cypher
MATCH (u:User {id: "uros"})-[:SUBSCRIBED_TO]->(g:Genre)
RETURN u.id, g.name
```

**Query za proveru ocena:**
```cypher
MATCH (u:User {id: "uros"})-[r:RATED]->(s:Song)
RETURN s.name, r.rating
ORDER BY r.rating DESC
```

**Query za proveru preporuka:**
```cypher
MATCH (u:User {id: "uros"})-[:SUBSCRIBED_TO]->(g:Genre)<-[:BELONGS_TO]-(s:Song)
OPTIONAL MATCH (u)-[r:RATED]->(s)
WHERE r IS NULL OR r.rating >= 4
RETURN s.name, g.name, r.rating
LIMIT 10
```

### 2. Proveri API endpoint direktno:

```bash
# Zameni TOKEN sa tvojim JWT tokenom
curl -X GET "http://localhost:8081/api/ratings/recommendations?userId=uros" \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json"
```

### 3. Proveri logove servisa:

```bash
# Recommendation service logovi
docker-compose logs recommendation-service | tail -50

# Subscriptions service logovi (za pretplate)
docker-compose logs subscriptions-service | tail -50

# Ratings service logovi (za ocene)
docker-compose logs ratings-service | tail -50
```

---

## ⚠️ Česti problemi i rešenja

### Problem: Ne vidim preporuke nakon pretplate

**Rešenje:**
1. Proveri da li je event poslat: `docker-compose logs subscriptions-service | grep "subscription_created"`
2. Proveri da li je Neo4j ažuriran: Query za pretplate (gore)
3. Osveži stranicu u browseru

### Problem: Vidim pesme koje sam ocenio sa < 4

**Rešenje:**
1. Proveri da li je rating event poslat: `docker-compose logs ratings-service | grep "rating_created"`
2. Proveri Neo4j: Query za ocene (gore)
3. Proveri da li je rating >= 4 u Neo4j

### Problem: Ne vidim top pesmu iz nepretplaćenog žanra

**Rešenje:**
1. Proveri da li postoje pesme sa ocenama 5: Query u Neo4j
2. Proveri da li korisnik ima pretplate na sve žanrove
3. Ako korisnik nema pretplate, trebalo bi da vidi top pesmu

---

## 📊 Primer očekivanog JSON odgovora

```json
{
  "subscribedGenreSongs": [
    {
      "songId": "song123",
      "name": "Song Name",
      "genre": "R&B",
      "artistIDs": ["artist1"],
      "albumID": "album1",
      "duration": 240,
      "reason": "Based on your genre subscriptions"
    }
  ],
  "topRatedSong": {
    "songId": "song456",
    "name": "Popular Song",
    "genre": "Pop",
    "artistIDs": ["artist2"],
    "albumID": "album2",
    "duration": 200,
    "reason": "Popular in genre you might like"
  }
}
```

---

## ✅ Checklist za testiranje

- [ ] Novi korisnik vidi top pesmu
- [ ] Pretplata na žanr prikazuje pesme iz tog žanra
- [ ] Pesme ocenjene sa < 4 se ne prikazuju
- [ ] Pesme ocenjene sa >= 4 se prikazuju
- [ ] Nepretplaćeni žanrovi prikazuju top pesmu
- [ ] Brisanje pretplate uklanja pesme iz preporuka
- [ ] Ocenjivanje pesme ažurira preporuke
- [ ] Frontend prikazuje obe sekcije (pretplate + top pesma)
