param(
    [Parameter(Mandatory = $true)]
    [string]$DatabaseUrl,
    [string]$MigrationFile = ""
)

$ErrorActionPreference = "Stop"
$root = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
if (-not $MigrationFile) {
    $MigrationFile = Join-Path $root "home\confmonit\v4.0\confvision\sql\migrations\20260928_vis_coleta_operacional.sql"
}
if (-not (Test-Path $MigrationFile)) {
    throw "Migration nao encontrada: $MigrationFile"
}

$psql = Get-Command psql -ErrorAction SilentlyContinue
if (-not $psql) {
    Write-Host "psql nao esta no PATH. Aplique manualmente:" -ForegroundColor Yellow
    Write-Host "  psql `"$DatabaseUrl`" -f `"$MigrationFile`""
    exit 2
}

Write-Host "Aplicando: $MigrationFile"
& psql $DatabaseUrl -v ON_ERROR_STOP=1 -f $MigrationFile
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Write-Host "OK — migration aplicada." -ForegroundColor Green
