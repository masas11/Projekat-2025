// Update songs to use music files from frontend/public/music
// Set audioFileUrl to /music/filename.mp3 for songs that match

const updates = [
  { pattern: /mesecina/i, url: '/music/Mesecina.mp3' },
  { pattern: /lady.*gaga/i, url: '/music/Lady-Gaga-Abracadabra.mp3' },
  { pattern: /abracadabra/i, url: '/music/Lady-Gaga-Abracadabra.mp3' }
];

updates.forEach(update => {
  const result = db.songs.updateMany(
    { name: update.pattern },
    { $set: { audioFileUrl: update.url } }
  );
  print(`Updated ${result.modifiedCount} songs matching ${update.pattern} -> ${update.url}`);
});

// Also update songs that have no audioFileUrl but might match
const allSongs = db.songs.find({ audioFileUrl: { $exists: false } }).toArray();
print(`Found ${allSongs.length} songs without audioFileUrl`);

print('Done!');
