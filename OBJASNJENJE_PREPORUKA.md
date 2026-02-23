# 🎯 Objašnjenje sistema preporuka

## 📋 Kako rade preporuke

### 1. **Pesme iz pretplaćenih žanrova** (`subscribedGenreSongs`)

**Algoritam:**
- Pronalazi sve pesme koje pripadaju žanrovima na koje je korisnik **pretplaćen**
- **Isključuje** pesme koje je korisnik ocenio sa **< 4 zvezdice**
- **Uključuje** pesme koje je korisnik ocenio sa **>= 4 zvezdice** ILI **nije ocenio uopšte**
- Maksimalno 50 pesama

**Primer:**
- Uros je pretplaćen na **R&B** → vidi R&B pesme (Sao Paolo, Blinding Lights)
- Ivana je pretplaćena na **R&B** → vidi R&B pesme (Sao Paolo, Blinding Lights)
- Ana **NEMA** pretplata na žanrove → vidi **prazan array** `[]`

### 2. **Top pesma iz nepretplaćenog žanra** (`topRatedSong`)

**Algoritam:**
- Pronalazi pesmu koja pripada žanru na koji korisnik **NIJE** pretplaćen
- Pesma mora imati **najviše ocena 5 zvezdica** od strane drugih korisnika
- Ako nema ocena 5, vraća se prva pesma iz nepretplaćenog žanra (po redosledu u bazi)

**Primer:**
- Uros je pretplaćen na **R&B** → top pesma je iz **Pop** žanra (Ubijas me usnama)
- Ivana je pretplaćena na **R&B** → top pesma je iz **Pop** žanra (Ubijas me usnama)
- Ana **NEMA** pretplata → top pesma je iz **bilo kog žanra** (Ubijas me usnama)

## ⚠️ Zašto je top pesma ista za sve?

**Problem:** Svi korisnici vide istu top pesmu ("Ubijas me usnama" - Pop)

**Razlog:**
1. **Nema ocena 5 zvezdica** u sistemu
2. Kada nema ocena 5, query vraća **prvu pesmu** iz nepretplaćenog žanra
3. Pošto je query isti za sve korisnike (osim pretplaćenih žanrova), vraća istu pesmu

**Rešenje:**
- Dodaj ocene 5 zvezdica različitim pesmama
- Ili proširi algoritam da uzima u obzir i druge faktore (npr. broj svih ocena, prosek ocena, itd.)

## 📊 Očekivani rezultati

### **Uros** (pretplaćen na R&B):
```json
{
  "subscribedGenreSongs": [
    {"name": "Sao Paolo", "genre": "R&B"},
    {"name": "Blinding Lights", "genre": "R&B"}
  ],
  "topRatedSong": {
    "name": "Ubijas me usnama",
    "genre": "Pop"
  }
}
```

### **Ivana** (pretplaćena na R&B):
```json
{
  "subscribedGenreSongs": [
    {"name": "Sao Paolo", "genre": "R&B"},
    {"name": "Blinding Lights", "genre": "R&B"}
  ],
  "topRatedSong": {
    "name": "Ubijas me usnama",
    "genre": "Pop"
  }
}
```

### **Ana** (pretplaćena samo na The Weeknd - umetnika, ne žanr):
```json
{
  "subscribedGenreSongs": [],  // Nema pretplata na žanrove
  "topRatedSong": {
    "name": "Ubijas me usnama",
    "genre": "Pop"
  }
}
```

## 🔍 Važno

**Pretplate na umetnike NE utiču na preporuke!**

Preporuke se baziraju **samo na pretplatama na žanrove**, ne na pretplatama na umetnike.

Ako želiš da Ana vidi preporuke, mora se **pretplatiti na žanr** (npr. Pop, Rock, itd.), ne samo na umetnika.

## 💡 Kako poboljšati algoritam

### Opcija 1: Dodaj ocene 5 zvezdica
- Oceni različite pesme sa 5 zvezdica
- Top pesma će se menjati na osnovu broja ocena 5

### Opcija 2: Proširi algoritam
- Uzmi u obzir sve ocene (ne samo 5)
- Uzmi u obzir prosek ocena
- Uzmi u obzir broj svih ocena
- Dodaj randomizaciju za raznovrsnost

### Opcija 3: Dodaj pretplate na umetnike u algoritam
- Modifikuj query da uzima u obzir i pretplate na umetnike
- Prikaži pesme umetnika na koje je korisnik pretplaćen
