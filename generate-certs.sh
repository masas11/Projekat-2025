#!/bin/bash

# Kreiranje samopotpisanih SSL sertifikata za HTTPS
echo "🔐 Kreiranje SSL sertifikata..."

# Kreiraj privatni ključ
openssl genrsa -out certs/server.key 2048

# Kreiraj CSR (Certificate Signing Request)
openssl req -new -key certs/server.key -out certs/server.csr -subj "/C=RS/ST=Serbia/L=Belgrade/O=MusicStreaming/OU=IT/CN=localhost"

# Kreiraj samopotpisani sertifikat
openssl x509 -req -days 365 -in certs/server.csr -signkey certs/server.key -out certs/server.crt

# Očisti CSR
rm certs/server.csr

echo "✅ SSL sertifikati kreirani!"
echo "📁 Fajlovi:"
echo "   - certs/server.crt (sertifikat)"
echo "   - certs/server.key (privatni ključ)"
echo ""
echo "🚀 Pokreni sa: docker-compose -f docker-compose.https.yml up -d"
