# D5 — assign worker_id Rust para analíticas sem processor (POST /ops/d5/auto_assign_analiticas)
param(
    [string]$ApiBase = "https://vision.confmonit2.com.br",
    [string]$VisWorkerKey = $env:VIS_WORKER_API_KEY,
    [switch]$DryRun,
    [int]$Limit = 500
)

$ErrorActionPreference = "Stop"
if (-not $VisWorkerKey) {
    $envFile = "C:\sistemaconfmonit\core4\home\confmonit\v4.0\confvision\.env"
    if (Test-Path $envFile) {
        $VisWorkerKey = (Select-String -Path $envFile -Pattern '^VIS_WORKER_API_KEY=(.+)$').Matches.Groups[1].Value.Trim()
    }
}
if (-not $VisWorkerKey) { throw "Defina VIS_WORKER_API_KEY" }

$uri = "$($ApiBase.TrimEnd('/'))/ops/d5/auto_assign_analiticas"
$body = @{ dry_run = [bool]$DryRun; limit = $Limit } | ConvertTo-Json
$headers = @{
    Authorization      = "Bearer $VisWorkerKey"
    "X-Vis-Worker-Key" = $VisWorkerKey
    "Content-Type"     = "application/json"
}

Write-Host "==> D5 auto_assign dry_run=$DryRun limit=$Limit"
$r = Invoke-RestMethod -Uri $uri -Method Post -Headers $headers -Body $body -TimeoutSec 120
$r | ConvertTo-Json -Depth 8
$errCount = @($r.errors).Count
if ($errCount -gt 0 -and -not $DryRun) {
    Write-Host "WARN: $errCount erros no assign"
    exit 2
}
Write-Host "OK total_unassigned=$($r.total_unassigned) assigned=$($r.assigned.Count)"
exit 0
