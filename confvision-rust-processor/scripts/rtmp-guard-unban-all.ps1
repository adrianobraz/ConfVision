# Desbanir todos os IPs listados no RTMP Guard (foxpro confvision :8100 ou URL publica)
param(
    [string]$GuardBase = $env:RTMP_GUARD_URL,
    [string]$AdminKey = $env:RTMP_GUARD_ADMIN_KEY,
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"
if (-not $GuardBase) { $GuardBase = "https://foxpro-confvision.rkr351.easypanel.host:8100" }
$GuardBase = $GuardBase.TrimEnd("/")

$headers = @{ "Content-Type" = "application/json" }
if ($AdminKey) { $headers["X-RTMP-Guard-Key"] = $AdminKey }

$bansUri = "$GuardBase/bans"
if ($AdminKey) { $bansUri = "$bansUri?key=$AdminKey" }

Write-Host "==> GET $bansUri"
$list = Invoke-RestMethod -Uri $bansUri -Headers $headers -TimeoutSec 30
$rows = @($list.dados)
if (-not $rows) {
    Write-Host "Nenhum ban ativo"
    exit 0
}

foreach ($b in $rows) {
    $ip = $b.ip
    if (-not $ip) { continue }
    if ($DryRun) {
        Write-Host "DRY unban $ip ($($b.motivo))"
        continue
    }
    $body = @{ ip = $ip } | ConvertTo-Json
    $r = Invoke-RestMethod -Uri "$GuardBase/unban" -Method Post -Headers $headers -Body $body -TimeoutSec 15
    Write-Host "unban $ip -> $($r.desbanido)"
}
Write-Host "OK"
