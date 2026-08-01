# fp-billing-worker — build cross-compile Windows -> Linux (Proxmox)
#
# Worker de faturamento (FranqueadoPro + ConfVision). Sem frontend.
#
# Uso:
#   powershell -ExecutionPolicy Bypass -File .\build.ps1
#   powershell -ExecutionPolicy Bypass -File .\build.ps1 -Local
#   powershell -ExecutionPolicy Bypass -File .\build.ps1 -Package
#
# Depois: subir fp-billing-worker + .env via FileZilla + systemctl restart

param(
    [switch]$Obfuscate,
    [switch]$Local,
    [switch]$Package
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

$OutputName = "fp-billing-worker"
$GoBin = Join-Path $env:USERPROFILE "go\bin"

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Error "Go nao encontrado. Instale: https://go.dev/dl/"
}

if (-not $Local) {
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    $env:CGO_ENABLED = "0"
    Write-Host "Alvo: Linux amd64" -ForegroundColor Cyan
} else {
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    $OutputName = "fp-billing-worker.exe"
    Write-Host "Alvo: Windows (teste local)" -ForegroundColor Cyan
}

Write-Host "Go: $(go version)" -ForegroundColor DarkGray

if ($Obfuscate) {
    $garble = Get-Command garble -ErrorAction SilentlyContinue
    if (-not $garble) {
        $garblePath = Join-Path $GoBin "garble.exe"
        if (Test-Path $garblePath) {
            $garble = $garblePath
        }
    }

    if (-not $garble) {
        Write-Host ""
        Write-Host "Garble nao instalado. Execute:" -ForegroundColor Yellow
        Write-Host "  go install mvdan.cc/garble@latest" -ForegroundColor White
        exit 1
    }

    Write-Host "Modo: Garble (-literals -tiny)" -ForegroundColor Magenta
    & $garble -literals -tiny build -o $OutputName .
} else {
    Write-Host "Modo: endurecido (-ldflags -s -w -trimpath)" -ForegroundColor Green
    go build -ldflags="-s -w" -trimpath -o $OutputName .
}

if ($LASTEXITCODE -ne 0) {
    Write-Error "Build falhou (exit $LASTEXITCODE)"
}

$bin = Get-Item $OutputName
$sizeMB = [math]::Round($bin.Length / 1MB, 2)

Write-Host ""
Write-Host "OK: $($bin.FullName) ($sizeMB MB)" -ForegroundColor Green

Write-Host ""
Write-Host "Arquivos essenciais para o servidor:" -ForegroundColor Cyan
Write-Host "  1. $OutputName        -> binario compilado"
Write-Host "  2. .env               -> XANO_API_FINANCEIRO, WORKER_SECRET, TZ, etc."
Write-Host ""
Write-Host "Nao precisa subir: *.go, go.mod, build.ps1, deploy\" -ForegroundColor DarkGray

if ($Package) {
    $deployDir = Join-Path $PSScriptRoot "deploy\fp-billing-worker"

    Write-Host ""
    Write-Host "Empacotando em: $deployDir" -ForegroundColor Cyan

    if (Test-Path $deployDir) {
        Remove-Item $deployDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $deployDir -Force | Out-Null

    Copy-Item $OutputName -Destination $deployDir -Force
    if (-not (Test-Path ".env") -and (Test-Path ".env.example")) {
        Copy-Item ".env.example" -Destination ".env" -Force
        Write-Host 'Criado .env a partir de .env.example' -ForegroundColor Yellow
    }
    if (Test-Path ".env") {
        Copy-Item ".env" -Destination $deployDir -Force
    }

    $leiaMe = @"
fp-billing-worker — pacote pronto para subir
============================================

SUBA os 2 arquivos desta pasta para:
  /home/confmonit/v4.0/fp-billing-worker/

Arquivos:
  - fp-billing-worker   (binario)
  - .env                (ja configurado — nao editar)

No servidor (uma vez apos subir):
  chmod +x fp-billing-worker
  systemctl restart confmonit4fpbillingworker

IMPORTANTE — antes do worker funcionar:
  Abra o admConfmonit -> Financeiro -> "Seed catalogo"
  (cria worker_secret no Xano igual ao .env desta pasta)

"@
    Set-Content -Path (Join-Path $deployDir "LEIA-ME.txt") -Value $leiaMe -Encoding UTF8

    $totalMB = [math]::Round(((Get-ChildItem $deployDir -Recurse | Measure-Object Length -Sum).Sum / 1MB), 2)
    Write-Host ('OK: pacote pronto (' + $totalMB + ' MB)') -ForegroundColor Green
    Write-Host 'Suba deploy\fp-billing-worker\ para /home/confmonit/v4.0/fp-billing-worker/' -ForegroundColor DarkGray
}

Write-Host ""
Write-Host "Deploy:" -ForegroundColor Cyan
Write-Host "  1. FileZilla -> /home/confmonit/v4.0/fp-billing-worker/fp-billing-worker"
Write-Host '     (e .env na mesma pasta)'
if (-not $Local) {
    Write-Host "  2. SSH: chmod +x fp-billing-worker"
    Write-Host "  3. SSH: sudo systemctl restart confmonit4fpbillingworker"
    Write-Host "  4. SSH: sudo systemctl status confmonit4fpbillingworker"
    Write-Host ""
    Write-Host "Teste manual (uma execucao):" -ForegroundColor DarkGray
    Write-Host '  ./fp-billing-worker --run-once'
}
