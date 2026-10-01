# Verificacao pre-teste Config Atendimento (Postgres + APIs publicas)
# Uso:
#   cd confvision\sql
#   .\verify_teste_atendimento.ps1
#   .\verify_teste_atendimento.ps1 -IdFranqueado "UUID" -IdCliente "UUID"

param(
    [string]$PostgresUrl = $env:POSTGRES_URL,
    [string]$IdFranqueado = "",
    [string]$IdCliente = "",
    [string]$GatewayHost = "185.130.61.4",
    [int]$ConfVisionPort = 8086
)

$ErrorActionPreference = "Continue"

if (-not $PostgresUrl) {
    $PostgresUrl = "postgres://confmonit:20080328DriziN@191.96.156.116:5432/confmonit?sslmode=disable"
}

Write-Host "=== VERIFY TESTE CONFIG ATENDIMENTO ===" -ForegroundColor Cyan

function Test-HttpJson($url) {
    try {
        $raw = curl.exe -s --max-time 8 $url
        if ($LASTEXITCODE -ne 0) { return @{ ok = $false; raw = [string]$raw } }
        return @{ ok = $true; raw = [string]$raw }
    } catch {
        return @{ ok = $false; raw = $_.Exception.Message }
    }
}

function Format-HttpPreview($raw) {
    if (-not $raw) { return "(vazio)" }
    $oneLine = ($raw -replace '\s+', ' ').Trim()
    if ($oneLine.Length -le 60) { return $oneLine }
    return $oneLine.Substring(0, 60) + "..."
}

Write-Host ""
Write-Host "--- Servicos ---" -ForegroundColor Green
$fail = 0
foreach ($pair in @(
    @{ name = "eventgateway"; url = "http://${GatewayHost}:8082/healthz"; expectOk = $true; optional = $false },
    @{ name = "taskxano"; url = "http://${GatewayHost}:8081/healthz"; expectOk = $true; optional = $false },
    @{ name = "confvision"; url = "http://${GatewayHost}:${ConfVisionPort}/vis_health"; expectOk = $true; optional = $true },
    @{ name = "franqueadopro"; url = "http://${GatewayHost}:2005/"; expectOk = $false; optional = $false }
)) {
    $r = Test-HttpJson $pair.url
    $connected = [bool]$r.raw
    $ok = ($r.raw -match '"ok"\s*:\s*true') -or (-not $pair.expectOk -and $r.ok)
    if (-not $ok -and $pair.expectOk) {
        if ($pair.optional -and -not $connected) {
            Write-Host ("  {0,-14} AVISO  porta {1} inacessivel daqui - teste no core-4" -f $pair.name, $ConfVisionPort) -ForegroundColor Yellow
            continue
        }
        $fail++
    }
    $status = if ($ok) { "OK" } else { "FALHA" }
    Write-Host ("  {0,-14} {1}  {2}" -f $pair.name, $status, (Format-HttpPreview $r.raw))
}

Write-Host ""
Write-Host "--- Postgres ---" -ForegroundColor Green
$ApplyDir = Join-Path $PSScriptRoot "cmd\apply"
$env:POSTGRES_URL = $PostgresUrl
$env:ID_FRA = $IdFranqueado
Push-Location $ApplyDir
try {
    go run . -check-atendimento
    if ($LASTEXITCODE -ne 0) { $fail++ }
} finally {
    Pop-Location
}

if ($IdFranqueado -and $IdCliente) {
    Write-Host ""
    Write-Host "--- Resolve ConfVision ---" -ForegroundColor Green
    $base = "http://${GatewayHost}:${ConfVisionPort}/ops/atendimento/resolve?id_franqueado=$IdFranqueado&id_cliente=$IdCliente"
    foreach ($rec in @("inteligencia_artificial", "finalizacao_automatica")) {
        $r = Test-HttpJson "${base}&recurso=$rec"
        Write-Host "  $rec : $($r.raw)"
    }
} else {
    Write-Host ""
    Write-Host "Dica: passe -IdFranqueado e -IdCliente para testar /ops/atendimento/resolve" -ForegroundColor DarkGray
}

Write-Host ""
if ($fail -gt 0) {
    Write-Host "RESULTADO: $fail falha(s) - corrija antes dos testes manuais" -ForegroundColor Red
    exit 1
}
Write-Host "RESULTADO: OK - pode iniciar TESTE_CONFIG_ATENDIMENTO.md" -ForegroundColor Green
