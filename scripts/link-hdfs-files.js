// Link HDFS audio files to songs in database
// Maps files from /audio/songs/ to songs by matching filenames

const songs = [
  { name: /mesecina/i, path: '/audio/songs/Mesecina.mp3' },
  { name: /lady.*gaga.*abracadabra/i, path: '/audio/songs/Lady-Gaga-Abracadabra.mp3' },
  { name: /abracadabra/i, path: '/audio/songs/Lady-Gaga-Abracadabra.mp3' }
];

songs.forEach(song => {
  const result = db.songs.updateMany(
    { name: song.name },
    { $set: { audioFileUrl: song.path } }
  );
  print(`Updated ${result.modifiedCount} songs for pattern ${song.name}`);
});

print('Done!');
