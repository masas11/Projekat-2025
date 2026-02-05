// MongoDB seed script for content-service
// Run with: docker exec -i projekat-2025-mongodb-content-1 mongosh music_streaming < scripts/seed-content.js

// Artists
db.artists.insertMany([
  {
    _id: "artist1",
    name: "Michael Jackson",
    biography: "Michael Joseph Jackson was an American singer, songwriter, and dancer. Dubbed the 'King of Pop', he is regarded as one of the most significant cultural figures of the 20th century.",
    genres: ["Pop", "R&B", "Soul"],
    createdAt: new Date()
  },
  {
    _id: "artist2",
    name: "The Beatles",
    biography: "The Beatles were an English rock band, formed in Liverpool in 1960, that comprised John Lennon, Paul McCartney, George Harrison and Ringo Starr.",
    genres: ["Rock", "Pop"],
    createdAt: new Date()
  },
  {
    _id: "artist3",
    name: "Lady Gaga",
    biography: "Stefani Joanne Angelina Germanotta, known professionally as Lady Gaga, is an American singer, songwriter, and actress.",
    genres: ["Pop", "Dance", "Electronic"],
    createdAt: new Date()
  },
  {
    _id: "artist4",
    name: "The Weeknd",
    biography: "Abel Makkonen Tesfaye, known professionally as the Weeknd, is a Canadian singer, songwriter, and record producer.",
    genres: ["R&B", "Pop", "Hip-Hop"],
    createdAt: new Date()
  },
  {
    _id: "artist5",
    name: "Ed Sheeran",
    biography: "Edward Christopher Sheeran MBE is an English singer-songwriter.",
    genres: ["Pop", "Folk", "Acoustic"],
    createdAt: new Date()
  }
]);

// Albums
db.albums.insertMany([
  {
    _id: "album1",
    name: "Thriller",
    releaseDate: new Date("1982-11-30"),
    genre: "Pop",
    artistIds: ["artist1"],
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    _id: "album2",
    name: "Abbey Road",
    releaseDate: new Date("1969-09-26"),
    genre: "Rock",
    artistIds: ["artist2"],
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    _id: "album3",
    name: "The Fame",
    releaseDate: new Date("2008-08-19"),
    genre: "Pop",
    artistIds: ["artist3"],
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    _id: "album4",
    name: "After Hours",
    releaseDate: new Date("2020-03-20"),
    genre: "R&B",
    artistIds: ["artist4"],
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    _id: "album5",
    name: "÷ (Divide)",
    releaseDate: new Date("2017-03-03"),
    genre: "Pop",
    artistIds: ["artist5"],
    createdAt: new Date(),
    updatedAt: new Date()
  }
]);

// Songs
db.songs.insertMany([
  {
    _id: "song1",
    name: "Billie Jean",
    duration: 294, // seconds
    genre: "Pop",
    albumId: "album1",
    artistIds: ["artist1"],
    audioFileUrl: "/music/Lady-Gaga-Abracadabra.mp3", // placeholder
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    _id: "song2",
    name: "Beat It",
    duration: 258,
    genre: "Pop",
    albumId: "album1",
    artistIds: ["artist1"],
    audioFileUrl: "/music/Lady-Gaga-Abracadabra.mp3",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    _id: "song3",
    name: "Come Together",
    duration: 259,
    genre: "Rock",
    albumId: "album2",
    artistIds: ["artist2"],
    audioFileUrl: "/music/Lady-Gaga-Abracadabra.mp3",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    _id: "song4",
    name: "Poker Face",
    duration: 238,
    genre: "Pop",
    albumId: "album3",
    artistIds: ["artist3"],
    audioFileUrl: "/music/Lady-Gaga-Abracadabra.mp3",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    _id: "song5",
    name: "Blinding Lights",
    duration: 200,
    genre: "R&B",
    albumId: "album4",
    artistIds: ["artist4"],
    audioFileUrl: "/music/Lady-Gaga-Abracadabra.mp3",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    _id: "song6",
    name: "Shape of You",
    duration: 233,
    genre: "Pop",
    albumId: "album5",
    artistIds: ["artist5"],
    audioFileUrl: "/music/Lady-Gaga-Abracadabra.mp3",
    createdAt: new Date(),
    updatedAt: new Date()
  }
]);

print("Content seeded successfully!");
