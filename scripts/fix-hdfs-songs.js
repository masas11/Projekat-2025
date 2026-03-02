// Fix songs in database to point to HDFS files
// Match songs by ID from filename

const hdfsFiles = [
  'fe628ff3-c877-4f6d-b44e-4d32bc9d66da.mp3',
  'song1.mp3'
];

hdfsFiles.forEach(file => {
  // Extract song ID from filename (remove .mp3)
  const songId = file.replace('.mp3', '');
  const hdfsPath = `/audio/songs/${file}`;
  
  // Update song with HDFS path
  const result = db.songs.updateOne(
    { _id: songId },
    { $set: { audioFileUrl: hdfsPath } }
  );
  
  if (result.matchedCount > 0) {
    print(`Updated song ${songId} -> ${hdfsPath}`);
  } else {
    print(`Song ${songId} not found in database`);
  }
});

print('Done!');
