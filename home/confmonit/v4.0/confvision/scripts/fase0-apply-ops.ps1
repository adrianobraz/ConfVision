# Aplica fase0_pilot_ops.sql (Postgres). Requer POSTGRES_URL.
param(
    [string]$SqlPath = "c:\sistemaconfmonit\core4-rust-pilot\confvision-rust-processor\sql\fase0_pilot_ops.sql"
)
if (-not $env:POSTGRES_URL) {
    Write-Error "POSTGRES_URL nao definido"
    exit 2
}
$env:FASE0_SQL = $SqlPath
$root = Resolve-Path (Join-Path $PSScriptRoot "..")
Push-Location $root
go run ./scripts/fase0-apply-ops/
$code = $LASTEXITCODE
Pop-Location
exit $code
