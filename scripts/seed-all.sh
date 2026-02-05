#!/bin/bash

# Script to seed all databases with initial data
# Make sure Docker containers are running first!

echo "🌱 Seeding databases with initial data..."

# Wait for MongoDB containers to be ready
echo "⏳ Waiting for MongoDB containers to be ready..."
sleep 5

# Get container names
USERS_CONTAINER=$(docker ps --filter "name=mongodb-users" --format "{{.Names}}" | head -1)
CONTENT_CONTAINER=$(docker ps --filter "name=mongodb-content" --format "{{.Names}}" | head -1)
RATINGS_CONTAINER=$(docker ps --filter "name=mongodb-ratings" --format "{{.Names}}" | head -1)

if [ -z "$USERS_CONTAINER" ] || [ -z "$CONTENT_CONTAINER" ] || [ -z "$RATINGS_CONTAINER" ]; then
    echo "❌ Error: MongoDB containers not found. Make sure they are running!"
    echo "   Run: docker-compose up -d"
    exit 1
fi

echo "📦 Seeding content database..."
docker exec -i "$CONTENT_CONTAINER" mongosh music_streaming < scripts/seed-content.js

echo "⭐ Seeding ratings database..."
docker exec -i "$RATINGS_CONTAINER" mongosh ratings_db < scripts/seed-ratings.js

echo "✅ All databases seeded successfully!"
echo ""
echo "📝 Note: Users database is automatically seeded by users-service on startup"
echo "   (Admin user: username='admin', password='admin123')"
