# Opção A — preflight + (opcional) aplicar CSV rtsp_url_sec + verificar Rust health
param(
    [string]$CsvPath = (Join-Path $PSScriptRoot "..\sql\opcao_a_foxpro_6_cameras.csv"),
    [Parameter(Mandatory = $true)]
    [string]$VisWorkerKey,
    [switch]$Apply,
    [switch]$SkipPreflight
)

$ErrorActionPreference = "Stop"
$env:VIS_WORKER_API_KEY = $VisWorkerKey

if (-not $SkipPreflight) {
    & (Join-Path $PSScriptRoot "opcao-a-preflight.ps1") -VisWorkerKey $VisWorkerKey
}

if (-not $Apply) {
    Write-Host "`nDry-run (sem -Apply). Para gravar API:"
    & (Join-Path $PSScriptRoot "apply-rtsp-option-a.ps1") -CsvPath $CsvPath -DryRun
    exit 0
}

$content = Get-Content -LiteralPath $CsvPath -Raw
if ($content -match "USUARIO|SENHA|IP_DVR") {
    throw "CSV ainda tem placeholders (USUARIO/SENHA/IP_DVR). Preencha RTSP reais em: $CsvPath"
}

& (Join-Path $PSScriptRoot "apply-rtsp-option-a.ps1") -CsvPath $CsvPath

Write-Host "`n== Verificar sync =="
if (Get-Command bash -ErrorAction SilentlyContinue) {
    $env:VIS_WORKER_API_KEY = $VisWorkerKey
    bash (Join-Path $PSScriptRoot "opcao-a-verify.sh") 2>$null
    if ($LASTEXITCODE -ne 0) { Write-Host "opcao-a-verify: exit $LASTEXITCODE (instale jq no Git Bash ou rode no CT111)" }
} else {
    Write-Host "Instale Git Bash ou rode: bash confvision-rust-processor/scripts/opcao-a-verify.sh"
}

Write-Host "`n== Rust health (30s) =="
Start-Sleep -Seconds 15
foreach ($u in @(
    "https://foxpro-rust-pilot.rkr351.easypanel.host",
    "https://foxpro-rust-pilot-b.rkr351.easypanel.host"
)) {
    $h = Invoke-RestMethod -Uri "$u/health" -TimeoutSec 20
    Write-Host "$u -> online=$($h.cameras_online) total=$($h.cameras_total)"
}
