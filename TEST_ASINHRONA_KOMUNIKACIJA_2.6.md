# 🧪 Testiranje asinhrone komunikacije (2.6)

## 📋 Brzi testovi

### 1. Test kreiranja pesme → Notifikacije
**Koraci:**
1. Prijavi se kao korisnik (npr. "ana")
2. Pretplati se na žanr (npr. "Pop")
3. Odjavi se i prijavi se kao admin
4. Kreiraj novu Pop pesmu
5. Odjavi se i prijavi se kao korisnik
6. Proveri notifikacije - trebalo bi da vidiš notifikaciju o novoj pesmi

**Provera logova:**
```bash
docker-compose logs subscriptions-service --tail 20 | grep -i "event\|notification"
```

### 2. Test kreiranja pesme → Neo4j graf
**Koraci:**
1. Prijavi se kao admin
2. Kreiraj novu pesmu
3. Proveri Neo4j graf

**Provera:**
```bash
docker-compose exec neo4j cypher-shell -u neo4j -p password "MATCH (s:Song) RETURN s.name AS song ORDER BY s.name DESC LIMIT 5"
```

**Provera logova:**
```bash
docker-compose logs recommendation-service --tail 20 | grep -i "song_created\|event"
```

### 3. Test ocenjivanja pesme → Neo4j graf
**Koraci:**
1. Prijavi se kao korisnik
2. Oceni pesmu (npr. 5 zvezdica)
3. Proveri Neo4j graf

**Provera:**
```bash
docker-compose exec neo4j cypher-shell -u neo4j -p password "MATCH (u:User)-[r:RATED]->(s:Song) RETURN u.id AS user, s.name AS song, r.rating AS rating ORDER BY r.rating DESC LIMIT 5"
```

**Provera logova:**
```bash
docker-compose logs recommendation-service --tail 20 | grep -i "rating\|event"
```

### 4. Test pretplate na žanr → Neo4j graf
**Koraci:**
1. Prijavi se kao korisnik
2. Pretplati se na žanr (npr. "Rock")
3. Proveri Neo4j graf

**Provera:**
```bash
docker-compose exec neo4j cypher-shell -u neo4j -p password "MATCH (u:User)-[:SUBSCRIBED_TO]->(g:Genre) RETURN u.id AS user, g.name AS genre"
```

**Provera logova:**
```bash
docker-compose logs recommendation-service --tail 20 | grep -i "subscription\|event"
```

### 5. Test brisanja pesme → Neo4j graf
**Koraci:**
1. Zabeleži ID pesme (npr. "song123")
2. Prijavi se kao admin
3. Obriši pesmu
4. Proveri Neo4j graf - pesma bi trebalo da ne postoji

**Provera:**
```bash
docker-compose exec neo4j cypher-shell -u neo4j -p password "MATCH (s:Song {id: 'SONG_ID'}) RETURN s"
# Trebalo bi da vrati prazan rezultat
```

**Provera logova:**
```bash
docker-compose logs recommendation-service --tail 20 | grep -i "song_deleted\|deleted"
```

## ✅ Checklist

- [ ] Kreiranje pesme emituje događaj ka subscriptions-service
- [ ] Kreiranje pesme emituje događaj ka recommendation-service
- [ ] Ocenjivanje pesme emituje događaj ka recommendation-service
- [ ] Pretplata na žanr emituje događaj ka recommendation-service
- [ ] Brisanje pesme emituje događaj ka recommendation-service
- [ ] Neo4j graf se ažurira na osnovu događaja
- [ ] Notifikacije se kreiraju na osnovu događaja

## 🔍 Provera logova - Svi servisi

```bash
# Content service - emituje događaje
docker-compose logs content-service --tail 30 | grep -i "event\|emit"

# Subscriptions service - prima i obrađuje događaje
docker-compose logs subscriptions-service --tail 30 | grep -i "event\|received\|notification"

# Recommendation service - prima i obrađuje događaje
docker-compose logs recommendation-service --tail 30 | grep -i "event\|received\|processed"

# Ratings service - emituje događaje
docker-compose logs ratings-service --tail 30 | grep -i "event\|emit"
```

## 🎯 Ključni indikatori uspeha

1. **Event emitted** u logovima content-service/ratings-service/subscriptions-service
2. **Event received** u logovima recommendation-service/subscriptions-service
3. **Event processed** u logovima recommendation-service
4. **Neo4j graf se ažurira** - proveri direktno u Neo4j
5. **Notifikacije se kreiraju** - proveri u frontendu

## ⚠️ Česti problemi

- **Event se ne emituje** → Proveri da li servis radi (`docker-compose ps`)
- **Event se ne prima** → Proveri URL-ove servisa u konfiguraciji
- **Neo4j se ne ažurira** → Proveri konekciju sa Neo4j (`docker-compose logs neo4j`)
- **Notifikacije se ne kreiraju** → Proveri logove subscriptions-service
