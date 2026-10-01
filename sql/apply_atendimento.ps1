# Aplica 008 + 009 no Postgres central (rodar NO SERVIDOR Core4 ou rede 10.2.2.x)
# Uso:
#   cd confvision\sql
#   .\apply_atendimento.ps1
#   .\apply_atendimento.ps1 -PostgresUrl "postgres://confmonit:SENHA@10.2.2.120:5432/confmonit?sslmode=disable"

param(
    [string]$PostgresUrl = $env:POSTGRES_URL
)

$ErrorActionPreference = "Stop"
$Root = $PSScriptRoot
$ApplyDir = Join-Path $Root "cmd\apply"

if (-not $PostgresUrl) {
    $PostgresUrl = "postgres://confmonit:20080328DriziN@191.96.156.116:5432/confmonit?sslmode=disable"
}

$env:POSTGRES_URL = $PostgresUrl
Write-Host "POSTGRES_URL=$PostgresUrl" -ForegroundColor Cyan

Push-Location $ApplyDir
try {
    Write-Host "`n=== 008 schema atendimento ===" -ForegroundColor Green
    go run . ..\..\008_atendimento_schema.sql
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

    Write-Host "`n=== 009 liberacao inicial ===" -ForegroundColor Green
    go run . ..\..\009_atendimento_liberacao_inicial.sql
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

    Write-Host "`n=== Conferencia ===" -ForegroundColor Green
    go run . -check 2>$null
} finally {
    Pop-Location
}

Write-Host "`nConcluido. Rode no Postgres:" -ForegroundColor Yellow
Write-Host @"
SELECT COUNT(*) FROM ops_stg_cliente_ativo;
SELECT COUNT(*) total,
       SUM(inteligencia_artificial::int) com_ia,
       SUM(finalizacao_automatica::int) com_autofim
FROM ops_cliente_atendimento_config;
"@
