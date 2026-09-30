# U2 — Atualiza stack do host (Windows → SSH ou WSL; local lab com Docker Desktop)
param(
    [string]$EnvFile = ".env.host",
    [string]$ComposeFile = "docker-compose.host.example.yml"
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot\..

if (-not (Test-Path $EnvFile)) {
    Write-Error "Faltando $EnvFile — copie env/host.env.example"
}

Write-Host "docker compose -f $ComposeFile --env-file $EnvFile up -d --build" -ForegroundColor Cyan
docker compose -f $ComposeFile --env-file $EnvFile up -d --build
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Start-Sleep -Seconds 5
$smokeSh = Join-Path $PSScriptRoot "tenant-stack-smoke.sh"
if (Test-Path $smokeSh) {
    bash $smokeSh 2>$null
}
Write-Host "Update host concluido." -ForegroundColor Green
