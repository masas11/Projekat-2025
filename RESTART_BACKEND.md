# Restart Backend Servisa za CORS Fix

Dodata je CORS podrška u API Gateway. Da bi promene stupile na snagu, moraš restartovati backend servise.

## Koraci:

1. **Zaustavi postojeće kontejnere:**
```powershell
docker-compose down
```

2. **Ponovo pokreni kontejnere (sa rebuild-om):**
```powershell
docker-compose up -d --build
```

Ovo će:
- Rebuild-ovati API Gateway sa novim CORS kodom
- Pokrenuti sve servise ponovo

3. **Proveri da li su servisi pokrenuti:**
```powershell
docker-compose ps
```

4. **Proveri logove API Gateway-a da vidiš da li radi:**
```powershell
docker-compose logs api-gateway
```

Trebalo bi da vidiš: "API Gateway running on port 8080"

5. **Testiraj u browseru:**
- Osveži stranicu na `http://localhost:3000`
- CORS greške bi trebalo da nestanu

## Ako i dalje ima problema:

Proveri da li API Gateway radi:
```powershell
# Testiraj direktno API Gateway
curl http://localhost:8081/api/content/health
```

Ili otvori u browseru: `http://localhost:8081/api/content/health`

Ako dobiješ odgovor, API Gateway radi. Ako ne, proveri logove:
```powershell
docker-compose logs api-gateway
```
