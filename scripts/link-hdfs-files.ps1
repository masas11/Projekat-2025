# Script to link HDFS audio files to songs in database
# Maps files from /audio/songs/ to songs by matching filenames

Write-Host "Linking HDFS audio files to songs..."

# Get all MP3 files from HDFS
$hdfsFiles = docker exec hdfs-namenode hdfs dfs -ls /audio/songs/*.mp3 2>&1 | Select-String -Pattern "\.mp3$"

foreach ($line in $hdfsFiles) {
    if ($line -match "/([^/]+\.mp3)$") {
        $fileName = $matches[1]
        $hdfsPath = "/audio/songs/$fileName"
        
        # Extract song name from filename (remove .mp3 and normalize)
        $songName = $fileName -replace "\.mp3$", "" -replace "-", " " -replace "_", " "
        
        Write-Host "Processing: $fileName -> Looking for song: $songName"
        
        # Try to find song by name (case insensitive, partial match)
        $updateCmd = "db.songs.updateMany({ name: { `$regex: '$songName', `$options: 'i' } }, { `$set: { audioFileUrl: '$hdfsPath' } })"
        
        $result = docker exec projekat-2025-2-mongodb-content-1 mongosh --eval $updateCmd 2>&1
        Write-Host "  Result: $result"
    }
}

Write-Host "Done!"
