# Test successful saga transaction (2.13)
# This script tests a successful song deletion via saga pattern

$baseURL = "http://localhost:8081"
$token = Read-Host "Enter admin JWT token"

# First, create a test song
Write-Host "Creating a test song..."
$songData = @{
    name = "Test Song for Saga"
    duration = 180
    genre = "Pop"
    albumId = "album1"
    artistIds = @("artist1")
} | ConvertTo-Json

$createResponse = Invoke-WebRequest -Uri "$baseURL/api/content/songs" `
    -Method POST `
    -Headers @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    } `
    -Body $songData `
    -UseBasicParsing

$song = $createResponse.Content | ConvertFrom-Json
$songId = $song.id
Write-Host "Created test song with ID: $songId"

# Play the song a few times to generate play counts
Write-Host "Playing song to generate activity..."
for ($i = 1; $i -le 3; $i++) {
    try {
        Invoke-WebRequest -Uri "$baseURL/api/content/songs/$songId/stream" `
            -Method GET `
            -Headers @{
                "Authorization" = "Bearer $token"
            } `
            -UseBasicParsing `
            -TimeoutSec 2 | Out-Null
    } catch {
        # Ignore errors (song might not have audio file)
    }
}

# Now delete the song via saga
Write-Host "`nDeleting song via saga transaction..."
$deleteData = @{
    songId = $songId
} | ConvertTo-Json

try {
    $sagaResponse = Invoke-WebRequest -Uri "$baseURL/api/sagas/delete-song" `
        -Method POST `
        -Headers @{
            "Authorization" = "Bearer $token"
            "Content-Type" = "application/json"
        } `
        -Body $deleteData `
        -UseBasicParsing

    $saga = $sagaResponse.Content | ConvertFrom-Json
    Write-Host "`nSaga Transaction Result:"
    Write-Host "  ID: $($saga.id)"
    Write-Host "  Status: $($saga.status)"
    Write-Host "  Steps:"
    foreach ($step in $saga.steps) {
        Write-Host "    - $($step.name): $($step.status)"
    }

    if ($saga.status -eq "COMPLETED") {
        Write-Host "`n✅ SUCCESS: Saga transaction completed successfully!" -ForegroundColor Green
    } else {
        Write-Host "`n❌ FAILED: Saga transaction failed with status: $($saga.status)" -ForegroundColor Red
        if ($saga.error) {
            Write-Host "  Error: $($saga.error)" -ForegroundColor Red
        }
    }
} catch {
    Write-Host "`n❌ ERROR: Failed to execute saga transaction" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
}
