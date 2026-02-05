// MongoDB seed script for ratings-service
// Run with: docker exec -i projekat-2025-mongodb-ratings-1 mongosh ratings_db < scripts/seed-ratings.js

// Sample ratings (assuming user IDs from users_db)
db.ratings.insertMany([
  {
    userId: "f613665d-83bf-4c6c-bd3b-f712f9b04e84", // Ivana Markovic
    songId: "song1",
    rating: 5,
    createdAt: new Date()
  },
  {
    userId: "f613665d-83bf-4c6c-bd3b-f712f9b04e84",
    songId: "song2",
    rating: 4,
    createdAt: new Date()
  },
  {
    userId: "55def55d-fed3-466a-9d6a-ed2b15100411", // Ljubica
    songId: "song3",
    rating: 5,
    createdAt: new Date()
  },
  {
    userId: "55def55d-fed3-466a-9d6a-ed2b15100411",
    songId: "song4",
    rating: 4,
    createdAt: new Date()
  }
]);

print("Ratings seeded successfully!");
