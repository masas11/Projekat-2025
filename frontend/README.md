# Music Streaming Frontend

React frontend aplikacija za Music Streaming mikroservisnu aplikaciju.

## Instalacija

```bash
cd frontend
npm install
```

## Pokretanje

```bash
npm start
```

Aplikacija će se pokrenuti na `http://localhost:3000`.

## Konfiguracija

Aplikacija koristi API Gateway na `http://localhost:8081` po defaultu. Možete promeniti URL kroz environment varijablu:

```bash
REACT_APP_API_URL=http://localhost:8081 npm start
```

## Funkcionalnosti

- **Autentifikacija**: Registracija i prijava sa OTP verifikacijom
- **Izvođači**: Pregled, kreiranje i ažuriranje izvođača (admin)
- **Albumi**: Pregled i kreiranje albuma (admin)
- **Pesme**: Pregled i kreiranje pesama (admin)
- **Notifikacije**: Pregled notifikacija za korisnika

## Struktura

- `src/components/` - React komponente
- `src/context/` - React Context za autentifikaciju
- `src/services/` - API servis za komunikaciju sa backend-om
