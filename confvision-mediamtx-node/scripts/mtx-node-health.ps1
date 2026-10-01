# Health check local ou remoto — Guard + MediaMTX (V2)
param(
    [string]$GuardBase = "http://127.0.0.1:8100",
    [string]$MtxApiBase = "",
    [string]$MtxUser = "",
    [string]$MtxPass = ""
)

$ErrorActionPreference = "Continue"
$ok = $true

Write-Host "=== ConfVision MediaMTX+Guard health ==="

try {
    $g = Invoke-RestMethod -Uri "$GuardBase/health" -TimeoutSec 10
    Write-Host "[OK] Guard /health status=$($g.status) secret=$($g.secret_configured) bans=$($g.bans)"
} catch {
    Write-Host "[FALHA] Guard $GuardBase/health — $($_.Exception.Message)"
    $ok = $false
}

try {
    $r = Invoke-WebRequest -Uri "$GuardBase/health/ready" -TimeoutSec 10 -UseBasicParsing
    $body = $r.Content | ConvertFrom-Json
    if ($r.StatusCode -eq 200 -and $body.status -eq "ready") {
        Write-Host "[OK] Guard /health/ready mediamtx_api=$($body.mediamtx_api)"
    } else {
        Write-Host "[WARN] Guard /health/ready HTTP $($r.StatusCode) status=$($body.status) mediamtx=$($body.mediamtx_api)"
        $ok = $false
    }
} catch {
    Write-Host "[FALHA] Guard /health/ready — $($_.Exception.Message)"
    $ok = $false
}

if ($MtxApiBase) {
    $pair = "${MtxUser}:${MtxPass}"
    $b64 = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes($pair))
    $h = @{ Authorization = "Basic $b64" }
    try {
        $null = Invoke-RestMethod -Uri "$MtxApiBase/v3/paths/list?itemsPerPage=1" -Headers $h -TimeoutSec 10
        Write-Host "[OK] MediaMTX API $MtxApiBase/v3/paths/list"
    } catch {
        Write-Host "[FALHA] MediaMTX API — $($_.Exception.Message)"
        $ok = $false
    }
} else {
    Write-Host "[SKIP] MediaMTX API (passe -MtxApiBase e credenciais)"
}

if ($ok) { exit 0 } else { exit 1 }
