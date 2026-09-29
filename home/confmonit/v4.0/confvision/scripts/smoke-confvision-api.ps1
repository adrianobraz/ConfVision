# Smoke ConfVision API (health + vis postgres; requer rede até o host)
param(
    [string]$Base = "https://vision.confmonit2.com.br",
    [string]$WorkerKey = "",
    [string]$IdFranqueado = ""
)

$ErrorActionPreference = "Stop"

Write-Host "== vis_health =="
$h = Invoke-RestMethod -Uri "$Base/vis_health" -Method Get
$h | ConvertTo-Json -Compress

if ($WorkerKey) {
    Write-Host "== vis_camera_by_franqueado (worker) =="
    $hdr = @{ Authorization = "Bearer $WorkerKey" }
    $q = if ($IdFranqueado) { "?id_franqueado=$([uri]::EscapeDataString($IdFranqueado))" } else { "" }
    $c = Invoke-WebRequest -Uri "$Base/vis_camera_by_franqueado$q" -Headers $hdr -UseBasicParsing
    Write-Host "status $($c.StatusCode) len $($c.Content.Length)"
    if ($c.Content.Length -lt 1200) { Write-Host $c.Content }
}

Write-Host "OK smoke"
