# Aplica coleta_operacional_migration.sql no Postgres (POSTGRES_URL).
$ErrorActionPreference = "Stop"
$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$envFile = Join-Path $root ".env"
if (Test-Path $envFile) {
    Get-Content $envFile | ForEach-Object {
        $line = $_.Trim()
        if ($line -match '^\s*#' -or $line -eq "") { return }
        if ($line -match '^POSTGRES_URL=(.+)$') {
            $env:POSTGRES_URL = $matches[1].Trim('"')
        }
    }
}
if (-not $env:POSTGRES_URL) {
    Write-Error "POSTGRES_URL nao encontrado em .env ou ambiente"
}
Push-Location $root
go run ./scripts/apply-coleta-migration/
$code = $LASTEXITCODE
Pop-Location
exit $code
