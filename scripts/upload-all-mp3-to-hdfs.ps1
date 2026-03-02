# Upload all MP3 files from a folder to HDFS
# Usage: .\scripts\upload-all-mp3-to-hdfs.ps1 -FolderPath "C:\path\to\mp3\files"

param(
    [string]$FolderPath = ".\songs"
)

Write-Host "Uploading MP3 files from $FolderPath to HDFS..."

if (-not (Test-Path $FolderPath)) {
    Write-Host "Folder not found: $FolderPath"
    Write-Host "Please create the folder and add MP3 files, or specify a different path."
    exit 1
}

$mp3Files = Get-ChildItem -Path $FolderPath -Filter "*.mp3" -File

if ($mp3Files.Count -eq 0) {
    Write-Host "No MP3 files found in $FolderPath"
    exit 1
}

Write-Host "Found $($mp3Files.Count) MP3 file(s)"

foreach ($file in $mp3Files) {
    Write-Host "Uploading $($file.Name)..."
    
    # Copy file to HDFS namenode container
    docker cp $file.FullName hdfs-namenode:/tmp/$($file.Name)
    
    # Upload to HDFS
    $hdfsPath = "/audio/songs/$($file.Name)"
    $result = docker exec hdfs-namenode hdfs dfs -put /tmp/$($file.Name) $hdfsPath 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ✓ Uploaded to $hdfsPath"
    } else {
        Write-Host "  ✗ Failed: $result"
    }
    
    # Clean up temp file
    docker exec hdfs-namenode rm -f /tmp/$($file.Name)
}

Write-Host "`nDone! Files are now on HDFS at /audio/songs/"
Write-Host "Note: You still need to create songs in the database and link them to these files."
