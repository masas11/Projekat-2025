# Test saga transaction with failure scenarios (2.13)
# This script demonstrates different failure scenarios

$baseURL = "http://localhost:8081"
$token = Read-Host "Enter admin JWT token"

Write-Host "Saga Failure Test Scenarios" -ForegroundColor Cyan
Write-Host "===========================`n"

# Scenario 1: Delete non-existent song (should fail at BACKUP_SONG step)
Write-Host "Scenario 1: Deleting non-existent song (should fail at BACKUP_SONG)" -ForegroundColor Yellow
$deleteData1 = @{
    songId = "non-existent-song-id-12345"
} | ConvertTo-Json

try {
    $sagaResponse = Invoke-WebRequest -Uri "$baseURL/api/sagas/delete-song" `
        -Method POST `
        -Headers @{
            "Authorization" = "Bearer $token"
            "Content-Type" = "application/json"
        } `
        -Body $deleteData1 `
        -UseBasicParsing

    $saga = $sagaResponse.Content | ConvertFrom-Json
    Write-Host "  Saga Status: $($saga.status)" -ForegroundColor $(if ($saga.status -eq "COMPENSATED") { "Yellow" } else { "Red" })
    Write-Host "  Failed Step: $($saga.steps | Where-Object { $_.status -eq 'FAILED' } | Select-Object -First 1 -ExpandProperty name)"
    Write-Host "  Error: $($saga.error)`n"
} catch {
    Write-Host "  Error: $($_.Exception.Message)`n" -ForegroundColor Red
}

# Scenario 2: Create a song, then simulate failure by stopping ratings-service
Write-Host "Scenario 2: Create song, then test with ratings-service down" -ForegroundColor Yellow
Write-Host "  (Stop ratings-service container to simulate failure)"
Write-Host "  Command: docker compose stop ratings-service`n"

# Scenario 3: Test compensation by checking saga status
Write-Host "Scenario 3: Check saga transaction status" -ForegroundColor Yellow
$sagaId = Read-Host "Enter saga ID to check"

if ($sagaId) {
    try {
        $statusResponse = Invoke-WebRequest -Uri "$baseURL/api/sagas/$sagaId" `
            -Method GET `
            -Headers @{
                "Authorization" = "Bearer $token"
            } `
            -UseBasicParsing

        $saga = $statusResponse.Content | ConvertFrom-Json
        Write-Host "`nSaga Transaction Details:"
        Write-Host "  ID: $($saga.id)"
        Write-Host "  Status: $($saga.status)"
        Write-Host "  Song ID: $($saga.songId)"
        Write-Host "  Steps:"
        foreach ($step in $saga.steps) {
            $statusColor = switch ($step.status) {
                "COMPLETED" { "Green" }
                "FAILED" { "Red" }
                "COMPENSATED" { "Yellow" }
                default { "White" }
            }
            Write-Host "    - $($step.name): $($step.status)" -ForegroundColor $statusColor
            if ($step.error) {
                Write-Host "      Error: $($step.error)" -ForegroundColor Red
            }
        }
        if ($saga.error) {
            Write-Host "  Transaction Error: $($saga.error)" -ForegroundColor Red
        }
    } catch {
        Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Red
    }
}
