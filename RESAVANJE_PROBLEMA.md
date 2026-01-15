# 🔧 Rešavanje problema sa Docker-om

## Problem: "The system cannot find the file specified" / "dockerDesktopLinuxEngine"

**Uzrok:** Docker Desktop nije pokrenut!

---

## ✅ REŠENJE - Korak po Korak

### Korak 1: Otvorite Docker Desktop

1. Pritisnite `Windows` dugme
2. Ukucajte: `Docker Desktop`
3. Kliknite na "Docker Desktop" aplikaciju

**ILI**

Pronađite Docker Desktop ikonu u system tray-u (dole desno pored sata) i kliknite na nju.

### Korak 2: Sačekajte da se Docker Desktop pokrene

Kada otvorite Docker Desktop, videćete:
- Loading animaciju
- Poruku "Docker Desktop is starting..."
- Može potrajati 30 sekundi do 2 minuta

**Kako znati da je spreman:**
- Ikonica u system tray-u će biti zelena (bez animacije)
- U Docker Desktop prozoru će pisati "Docker Desktop is running"
- Status će biti "Running"

### Korak 3: Proverite da li radi

U CMD-u ukucajte:

```cmd
docker ps
```

Ako vidite praznu listu ili header (bez greške) - **Docker radi!** ✅

Ako i dalje dobijate grešku, proverite:

```cmd
docker --version
```

Ako ovo radi, Docker je instaliran, ali servis nije pokrenut.

---

## 🔄 ALTERNATIVNO REŠENJE: Restart Docker Desktop

Ako Docker Desktop ne želi da se pokrene:

1. Zatvorite Docker Desktop potpuno
2. Otvorite Task Manager (`Ctrl + Shift + Esc`)
3. Pronađite sve Docker procese i zatvorite ih:
   - `Docker Desktop`
   - `com.docker.backend`
   - `dockerd`
4. Sačekajte 10 sekundi
5. Otvorite Docker Desktop ponovo

---

## 🛠️ PROVERA: Da li je Docker Desktop instaliran?

Ako ne možete da pronađete Docker Desktop:

1. Proverite da li je instaliran:
   ```cmd
   where docker
   ```

2. Ako ne postoji, preuzmite sa: https://www.docker.com/products/docker-desktop/

3. Instalirajte i restartujte računar

---

## ✅ NAKON ŠTO SE DOCKER DESKTOP POKRENE

Vratite se u CMD i pokrenite:

```cmd
cd C:\Users\boris\OneDrive\Desktop\projekat\Projekat-2025
docker-compose up
```

Sada bi trebalo da radi! 🎉

---

## 📋 REDOSLED AKCIJA

1. ✅ Otvorite Docker Desktop
2. ✅ Sačekajte da se pokrene (zelena ikonica)
3. ✅ Proverite: `docker ps` (ne bi trebalo da da grešku)
4. ✅ Pokrenite: `docker-compose up`

---

## ❓ Još problema?

Ako i dalje imate problema, proverite:

```cmd
# Proverite Docker status
docker info

# Proverite da li Docker servis radi
sc query com.docker.service
```

Ako vidite greške, možda treba da restartujete računar nakon instalacije Docker Desktop-a.

