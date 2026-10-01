# Túnel SSH: Postgres interno (10.2.2.120) via core-4
# Use quando o IP externo 191.96.156.116 bloquear seu IP no pg_hba.conf
#
# Uso (deixe esta janela aberta):
#   .\tunnel-postgres-via-core4.ps1
#
# Depois conecte em localhost:15432 (DBeaver, pgAdmin, go run -check):
#   POSTGRES_URL=postgres://confmonit:SENHA@127.0.0.1:15432/confmonit?sslmode=disable

param(
    [string]$Core4Host = "185.130.61.4",
    [string]$Core4User = "root",
    [string]$PostgresInternal = "10.2.2.120",
    [int]$PostgresPort = 5432,
    [int]$LocalPort = 15432
)

Write-Host "Abrindo tunel ${Core4User}@${Core4Host} -> ${PostgresInternal}:${PostgresPort} -> localhost:${LocalPort}" -ForegroundColor Cyan
Write-Host "POSTGRES_URL=postgres://confmonit:SENHA@127.0.0.1:${LocalPort}/confmonit?sslmode=disable" -ForegroundColor DarkGray
Write-Host "Ctrl+C para encerrar." -ForegroundColor DarkGray

ssh -N -L "${LocalPort}:${PostgresInternal}:${PostgresPort}" "${Core4User}@${Core4Host}"
