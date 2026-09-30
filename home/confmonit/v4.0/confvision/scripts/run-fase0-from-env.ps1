# Carrega POSTGRES_URL e VIS_WORKER_API_KEY de ../.env e roda diagnose + verify.
param(
    [int]$PilotBCameraId = 3,
    [switch]$SkipRustPilotA
)

$ErrorActionPreference = "Continue"
$root = Split-Path -Parent $PSScriptRoot
$envFile = Join-Path $root ".env"
if (-not (Test-Path $envFile)) {
    Write-Host "FAIL: $envFile nao encontrado"
    exit 2
}

Get-Content $envFile | ForEach-Object {
    $line = $_.Trim()
    if ($line -match '^\s*#') { return }
    if ($line -match '^POSTGRES_URL=(.+)$') {
        $env:POSTGRES_URL = $matches[1].Trim().Trim('"').Trim("'")
    }
    if ($line -match '^VIS_WORKER_API_KEY=(.+)$') {
        $script:WorkerKey = $matches[1].Trim().Trim('"').Trim("'")
    }
}

$diagArgs = @{ PilotBCameraId = $PilotBCameraId }
if ($SkipRustPilotA) { $diagArgs.SkipRustPilotA = $true }
if ($WorkerKey) { $diagArgs.WorkerKey = $WorkerKey }

Write-Host "=== fase0-diagnose ===" -ForegroundColor Cyan
& (Join-Path $PSScriptRoot "fase0-diagnose.ps1") @diagArgs
$diagExit = $LASTEXITCODE

Write-Host "`n=== fase0-verify ===" -ForegroundColor Cyan
$verArgs = @{ PilotCameraId = $PilotBCameraId }
if ($SkipRustPilotA) { $verArgs.SkipRustPilotA = $true }
if ($WorkerKey) { $verArgs.WorkerKey = $WorkerKey }
& (Join-Path $PSScriptRoot "fase0-verify.ps1") @verArgs
$verExit = $LASTEXITCODE

if (-not $env:POSTGRES_URL) {
    Write-Host "`n[i] POSTGRES_URL ausente no .env — pg-audit nao rodou" -ForegroundColor Yellow
} elseif ($diagExit -ne 0 -and $LASTEXITCODE -eq 0) {
    Write-Host "`n[i] Postgres inacessivel desta maquina (VPN/VPS?) — rode pg-audit no host com rede 10.2.2.x" -ForegroundColor Yellow
}

exit [Math]::Max($diagExit, $verExit)
