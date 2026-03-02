// Auto-link HDFS files to songs by ID
// If song exists, update audioFileUrl
// If song doesn't exist, we can't create it without more info

const hdfsFiles = [
  'fe628ff3-c877-4f6d-b44e-4d32bc9d66da.mp3',
  'song1.mp3'
];

hdfsFiles.forEach(file => {
  const songId = file.replace('.mp3', '');
  const hdfsPath = `/audio/songs/${file}`;
  
  // Update song if it exists
  const result = db.songs.updateOne(
    { _id: songId },
    { $set: { audioFileUrl: hdfsPath } }
  );
  
  if (result.matchedCount > 0) {
    print(`✓ Updated song ${songId} -> ${hdfsPath}`);
  } else {
    print(`✗ Song ${songId} not found - file exists on HDFS but song missing from DB`);
  }
});

print('\nDone! If songs are missing, create them first, then run this script.');
