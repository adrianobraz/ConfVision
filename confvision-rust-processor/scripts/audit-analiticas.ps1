# Lista câmeras analíticas ativas: RTSP sec, worker_id, pausa — todas (sem limite artificial)
param(
    [string]$ApiBase = "https://vision.confmonit2.com.br",
    [string]$VisWorkerKey = $env:VIS_WORKER_API_KEY
)

$ErrorActionPreference = "Stop"
if (-not $VisWorkerKey) {
    $envFile = "C:\sistemaconfmonit\core4\home\confmonit\v4.0\confvision\.env"
    if (Test-Path $envFile) {
        $VisWorkerKey = (Select-String -Path $envFile -Pattern '^VIS_WORKER_API_KEY=(.+)$').Matches.Groups[1].Value.Trim()
    }
}
if (-not $VisWorkerKey) { throw "Defina VIS_WORKER_API_KEY" }

$headers = @{
    Authorization      = "Bearer $VisWorkerKey"
    "X-Vis-Worker-Key" = $VisWorkerKey
}

function Test-VisBool($v) {
    if ($null -eq $v) { return $false }
    if ($v -is [bool]) { return $v }
    $s = "$v".Trim().ToLowerInvariant()
    return $s -in @("true", "1", "t", "yes", "sim")
}

$uri = "$($ApiBase.TrimEnd('/'))/vis_camera?all=1"
$list = Invoke-RestMethod -Uri $uri -Headers $headers -TimeoutSec 120
if ($null -ne $list.PSObject.Properties["dados"]) {
    $rows = @($list.dados)
} else {
    $rows = @($list)
}

$anal = $rows | Where-Object {
    $_ -and (Test-VisBool $_.ativo) -and (Test-VisBool $_.deteccao_humano) -and -not (Test-VisBool $_.analitico_pausado)
}

$noRtsp = @($anal | Where-Object { -not $_.rtsp_url_sec -or [string]::IsNullOrWhiteSpace($_.rtsp_url_sec) })
$noWorker = @($anal | Where-Object { -not $_.worker_id -or [string]::IsNullOrWhiteSpace($_.worker_id) })

Write-Host "==> Analiticas ativas: $($anal.Count) / total cameras $($rows.Count)"
Write-Host "    sem rtsp_url_sec: $($noRtsp.Count)"
Write-Host "    sem worker_id:    $($noWorker.Count)"
if ($noRtsp.Count -gt 0) {
    Write-Host "    IDs sem RTSP:" ($noRtsp.id -join ", ")
}
if ($noWorker.Count -gt 0) {
    Write-Host "    IDs sem worker:" ($noWorker.id -join ", ")
}

$byWorker = $anal | Group-Object worker_id | Sort-Object Name
foreach ($g in $byWorker) {
    $name = if ($g.Name) { $g.Name } else { "(vazio)" }
    Write-Host "    worker_id $name : $($g.Count)"
}

if ($noRtsp.Count -gt 0 -or $noWorker.Count -gt 0) { exit 2 }
exit 0
