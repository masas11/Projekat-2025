# Skripta za ažuriranje postojećih pretplata sa artistName (CQRS migration)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  AZURIRANJE PRETPLATA SA ARTIST NAME" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$contentContainer = docker ps --filter "name=mongodb-content" --format "{{.Names}}" | Select-Object -First 1
$subscriptionsContainer = docker ps --filter "name=mongodb-subscriptions" --format "{{.Names}}" | Select-Object -First 1

if ([string]::IsNullOrEmpty($contentContainer) -or [string]::IsNullOrEmpty($subscriptionsContainer)) {
    Write-Host "ERROR: MongoDB kontejneri nisu pokrenuti!" -ForegroundColor Red
    exit 1
}

Write-Host "Ažuriram pretplate sa artistName..." -ForegroundColor Yellow

# JavaScript skripta koja ažurira sve pretplate
$updateScript = @"
use('subscriptions_db');

// Uzmi sve pretplate na umetnike bez artistName
const subscriptions = db.subscriptions.find({type: 'artist', artistName: {`$exists: false}}).toArray();

if (subscriptions.length === 0) {
    print('Nema pretplata za ažuriranje.');
} else {
    print('Pronađeno ' + subscriptions.length + ' pretplata za ažuriranje.');
    
    // Za svaku pretplatu, uzmi ime umetnika iz content-service
    subscriptions.forEach(function(sub) {
        // Pozovi content-service MongoDB direktno
        const artist = db.getSiblingDB('music_streaming').artists.findOne({_id: sub.artistId});
        
        if (artist && artist.name) {
            db.subscriptions.updateOne(
                { _id: sub._id },
                { `$set: { artistName: artist.name } }
            );
            print('Ažurirana pretplata za: ' + artist.name);
        } else {
            print('Umetnik ' + sub.artistId + ' nije pronađen');
        }
    });
    
    print('Ažuriranje završeno!');
}
"@

# Pokreni skriptu
docker exec $subscriptionsContainer mongosh subscriptions_db --quiet --eval $updateScript

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  AZURIRANJE ZAVRSENO!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Osveži frontend stranicu da vidiš izmene." -ForegroundColor Yellow
Write-Host ""
