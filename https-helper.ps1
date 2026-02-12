# Helper funkcija za HTTPS zahteve sa ignorisanjem sertifikata
function Invoke-HTTPSRequest {
    param(
        [string]$Uri,
        [string]$Method = "GET",
        [hashtable]$Headers = @{},
        [object]$Body = $null,
        [string]$ContentType = "application/json"
    )
    
    # Ignoriši SSL sertifikat greške
    add-type @"
        using System.Net;
        using System.Security.Cryptography.X509Certificates;
        public class TrustAllCertsPolicy : ICertificatePolicy {
            public bool CheckValidationResult(
                ServicePoint srvPoint, X509Certificate certificate,
                WebRequest request, int certificateProblem) {
                return true;
            }
        }
"@
    [System.Net.ServicePointManager]::CertificatePolicy = New-Object TrustAllCertsPolicy
    [System.Net.ServicePointManager]::ServerCertificateValidationCallback = {$true}
    
    try {
        $request = [System.Net.HttpWebRequest]::Create($Uri)
        $request.Method = $Method
        $request.Timeout = 10000
        
        # Dodaj headers
        foreach ($key in $Headers.Keys) {
            if ($key -eq "Content-Type") {
                $request.ContentType = $Headers[$key]
            } elseif ($key -eq "Authorization") {
                $request.Headers.Add("Authorization", $Headers[$key])
            } else {
                $request.Headers.Add($key, $Headers[$key])
            }
        }
        
        # Dodaj body ako postoji
        if ($Body) {
            if (-not $request.ContentType) {
                $request.ContentType = $ContentType
            }
            $bodyBytes = [System.Text.Encoding]::UTF8.GetBytes($Body)
            $request.ContentLength = $bodyBytes.Length
            
            $requestStream = $request.GetRequestStream()
            $requestStream.Write($bodyBytes, 0, $bodyBytes.Length)
            $requestStream.Close()
        }
        
        $response = $request.GetResponse()
        $responseStream = $response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($responseStream)
        $responseBody = $reader.ReadToEnd()
        $reader.Close()
        $responseStream.Close()
        $response.Close()
        
        return @{
            StatusCode = [int]$response.StatusCode
            Content = $responseBody
            Error = $null
        }
    }
    catch {
        $statusCode = 0
        $errorMsg = $_.Exception.Message
        $errorContent = $null
        
        # Pokušaj da izvučeš status kod iz WebException
        if ($_.Exception -is [System.Net.WebException]) {
            $webException = $_.Exception
            if ($webException.Response) {
                $httpResponse = $webException.Response
                if ($httpResponse -is [System.Net.HttpWebResponse]) {
                    $statusCode = [int]$httpResponse.StatusCode
                    
                    # Pokušaj da pročitaš response body
                    try {
                        $responseStream = $httpResponse.GetResponseStream()
                        $reader = New-Object System.IO.StreamReader($responseStream)
                        $errorContent = $reader.ReadToEnd()
                        $reader.Close()
                        $responseStream.Close()
                    } catch {
                        # Ignoriši greške pri čitanju error response-a
                    }
                }
            }
        }
        
        # Ako nismo dobili status kod, pokušaj da ga izvučeš iz error poruke
        if ($statusCode -eq 0 -and $errorMsg -match '\((\d+)\)') {
            $statusCode = [int]$matches[1]
        }
        
        return @{
            StatusCode = $statusCode
            Error = $errorMsg
            Content = $errorContent
        }
    }
    finally {
        [System.Net.ServicePointManager]::CertificatePolicy = $null
        [System.Net.ServicePointManager]::ServerCertificateValidationCallback = $null
    }
}
