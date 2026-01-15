# 📦 Kako Instalirati Node.js i npm

## 🎯 Šta je npm?

**npm** (Node Package Manager) je alat za upravljanje JavaScript paketima. Dolazi automatski sa **Node.js**.

---

## ✅ KORAK 1: Preuzmite Node.js

### Opcija A: Preuzimanje sa zvaničnog sajta (PREPORUČENO)

1. Idite na: **https://nodejs.org/**
2. Kliknite na veliko zeleno dugme **"Download Node.js (LTS)"**
   - LTS = Long Term Support (stabilna verzija)
   - Trenutno je to verovatno **v20.x** ili **v22.x**

3. Fajl će se automatski preuzeti (npr. `node-v20.11.0-x64.msi`)

### Opcija B: Direktan link

- **Windows 64-bit:** https://nodejs.org/dist/v20.11.0/node-v20.11.0-x64.msi
- **Windows 32-bit:** https://nodejs.org/dist/v20.11.0/node-v20.11.0-x86.msi

---

## ✅ KORAK 2: Instalirajte Node.js

1. **Otvorite preuzeti fajl** (npr. `node-v20.11.0-x64.msi`)

2. **Kliknite "Next"** kroz instalaciju:
   - Prihvatite licence
   - Izaberite folder za instalaciju (ostavite podrazumevani)
   - **VAŽNO:** Obavezno označite opciju **"Add to PATH"** (obično je već označena)

3. **Kliknite "Install"**
   - Može trajati 1-2 minuta

4. **Kliknite "Finish"**

---

## ✅ KORAK 3: Restartujte CMD

**VAŽNO:** Zatvorite sve CMD prozore i otvorite NOVI CMD prozor!

Node.js se neće učitati u postojeće CMD prozore.

---

## ✅ KORAK 4: Proverite da li je instalirano

U NOVOM CMD prozoru ukucajte:

```cmd
node --version
```

Trebalo bi da vidite nešto kao:
```
v20.11.0
```

Zatim:

```cmd
npm --version
```

Trebalo bi da vidite nešto kao:
```
10.2.4
```

**Ako vidite verzije - INSTALACIJA JE USPEŠNA!** ✅

---

## ❓ Problemi i Rešenja

### Problem 1: "node is not recognized"

**Rešenje:**
1. Zatvorite SVE CMD prozore
2. Otvorite NOVI CMD prozor
3. Pokušajte ponovo: `node --version`

Ako i dalje ne radi:
1. Restartujte računar
2. Proverite da li je Node.js instaliran:
   - Otvorite "Add or Remove Programs" u Windows Settings
   - Tražite "Node.js"
   - Ako ne postoji, instalirajte ponovo

### Problem 2: "npm is not recognized"

**Rešenje:**
- npm dolazi sa Node.js, tako da ako Node.js radi, npm bi trebalo da radi
- Proverite: `npm --version`
- Ako ne radi, restartujte računar

### Problem 3: Instalacija ne želi da se završi

**Rešenje:**
1. Zatvorite sve programe
2. Pokrenite instalaciju kao Administrator:
   - Desni klik na `.msi` fajl
   - Izaberite "Run as administrator"

---

## ✅ KORAK 5: Instalirajte Frontend Dependencies

Nakon što je Node.js instaliran, idite u frontend folder:

```cmd
cd D:\projekat\Projekat-2025\frontend
npm install
```

**Šta se dešava:**
- npm preuzima sve potrebne pakete (React, itd.)
- Može potrajati 2-5 minuta (prvi put)
- Kreiram se folder `node_modules` sa svim paketima

**Kako znati da je gotovo:**
- Videćete: `added XXX packages`
- Folder `node_modules` će biti kreiran

---

## ✅ KORAK 6: Pokrenite Frontend

```cmd
npm start
```

**Šta se dešava:**
- React development server se pokreće
- Browser će se automatski otvoriti na `http://localhost:3000`
- Može potrajati 30 sekundi - 1 minut

**Kako znati da radi:**
- Videćete: `webpack compiled successfully`
- Browser će se otvoriti sa frontend aplikacijom

---

## 📊 Rezime - Brzi Start

1. ✅ Preuzmite Node.js sa: https://nodejs.org/
2. ✅ Instalirajte (kliknite Next, Next, Install)
3. ✅ **RESTARTUJTE CMD** (zatvorite i otvorite novi)
4. ✅ Proverite: `node --version` i `npm --version`
5. ✅ Idite u frontend: `cd D:\projekat\Projekat-2025\frontend`
6. ✅ Instalirajte: `npm install`
7. ✅ Pokrenite: `npm start`

---

## 🎯 Alternativa: Koristite Chocolatey (Naprednije)

Ako imate **Chocolatey** instaliran:

```cmd
choco install nodejs
```

Ali ovo je opciono - standardna instalacija je lakša!

---

## ✅ Provera - Sve što treba da znate

**Node.js = JavaScript runtime** (pokreće JavaScript kod)
**npm = Package manager** (preuzima JavaScript pakete)

Oba dolaze zajedno u jednom instalacijskom paketu!

---

## 🆘 Ako i dalje imate problema

1. **Proverite da li je Node.js instaliran:**
   - Windows Settings → Apps → Tražite "Node.js"

2. **Proverite PATH:**
   - Otvorite System Properties → Environment Variables
   - U "Path" trebalo bi da vidite: `C:\Program Files\nodejs\`

3. **Restartujte računar** (ponekad je potrebno)

4. **Reinstalirajte Node.js** ako ništa ne pomaže

---

**Nakon instalacije, javite mi i pokrenimo frontend!** 🚀

