# 🔍 Kako proveriti preporuke u browseru

## 📋 Korak po korak

### 1. Otvori Developer Tools
- Pritisni **F12** ili **Ctrl+Shift+I** (Windows/Linux)
- Ili **Cmd+Option+I** (Mac)
- Ili desni klik → "Inspect" / "Pregledaj element"

### 2. Otvori Network tab
- Klikni na tab **"Network"** (Mreža)
- Ili pritisni **Ctrl+Shift+E** (Windows/Linux)

### 3. Osveži stranicu
- Pritisni **F5** ili **Ctrl+R** da osvežiš stranicu
- Ovo će pokrenuti sve network zahteve

### 4. Filtriraj zahteve
- U polju za pretragu (Filter), unesi: `recommendations`
- Ili traži: `ratings/recommendations`
- Trebalo bi da vidiš zahtev sa imenom: `recommendations?userId=...`

### 5. Proveri zahtev
- Klikni na zahtev `recommendations?userId=...`
- Otvoriće se detalji zahteva

### 6. Proveri Response
- Klikni na tab **"Response"** (Odgovor)
- Trebalo bi da vidiš JSON sa:
  ```json
  {
    "subscribedGenreSongs": [
      {
        "songId": "...",
        "name": "Sao Paolo",
        "genre": "R&B",
        ...
      },
      {
        "songId": "...",
        "name": "Blinding Lights",
        "genre": "R&B",
        ...
      }
    ],
    "topRatedSong": {
      "songId": "...",
      "name": "Ubijas me usnama",
      "genre": "Pop",
      ...
    }
  }
  ```

### 7. Proveri Headers
- Klikni na tab **"Headers"** (Zaglavlja)
- Proveri:
  - **Request URL**: `http://localhost:8081/api/ratings/recommendations?userId=...`
  - **Status Code**: Trebalo bi da bude `200 OK`
  - **Authorization**: Trebalo bi da sadrži `Bearer TOKEN`

## ⚠️ Česti problemi

### Problem: Ne vidiš zahtev `recommendations`
**Rešenje:**
- Proveri da li si prijavljen kao korisnik (ne admin)
- Proveri da li si na Home stranici (`/`)
- Osveži stranicu ponovo (F5)

### Problem: Status Code je 401 (Unauthorized)
**Rešenje:**
- Token je istekao - prijavi se ponovo
- Proveri da li postoji `Authorization` header

### Problem: Status Code je 403 (Forbidden)
**Rešenje:**
- Admin korisnici ne mogu da vide preporuke
- Prijavi se kao običan korisnik

### Problem: Response je prazan ili `null`
**Rešenje:**
- Proveri da li imaš pretplate u Neo4j
- Proveri logove recommendation-service
- Proveri da li postoje pesme u Neo4j

### Problem: Vidiš samo `topRatedSong`, nema `subscribedGenreSongs`
**Rešenje:**
- Nemaš pretplate na žanrove
- Pretplati se na žanr (npr. R&B)
- Osveži stranicu

## 🧪 Testiranje direktno u Console tabu

Možeš i direktno testirati u **Console** tabu:

```javascript
// Zameni USER_ID sa tvojim ID-jem
const userId = "a0cfeaa6-6a37-44bb-a794-9ab0678a1214";
const token = localStorage.getItem('token');

fetch(`http://localhost:8081/api/ratings/recommendations?userId=${userId}`, {
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
})
  .then(res => res.json())
  .then(data => {
    console.log('Preporuke:', data);
    console.log('R&B pesme:', data.subscribedGenreSongs);
    console.log('Top pesma:', data.topRatedSong);
  })
  .catch(err => console.error('Greška:', err));
```

## 📊 Šta treba da vidiš

### Za korisnika pretplaćenog na R&B:
- `subscribedGenreSongs`: Array sa R&B pesmama (npr. "Sao Paolo", "Blinding Lights")
- `topRatedSong`: Jedna pesma iz drugog žanra (npr. Pop)

### Za korisnika bez pretplata:
- `subscribedGenreSongs`: Prazan array `[]` ili `null`
- `topRatedSong`: Jedna top pesma

### Za admin korisnika:
- Ne bi trebalo da vidiš zahtev (admin ne može da vidi preporuke)

## 🔧 React Router upozorenja

Upozorenja koja vidiš:
```
⚠️ React Router Future Flag Warning: React Router will begin wrapping state updates in React.startTransition in v7...
⚠️ React Router Future Flag Warning: Relative route resolution within Splat routes is changing in v7...
```

**Ovo nije greška!** To su samo upozorenja za buduće verzije React Routera. Ne utiču na funkcionalnost aplikacije. Možeš ih ignorisati ili dodati future flags u React Router konfiguraciju ako želiš.
