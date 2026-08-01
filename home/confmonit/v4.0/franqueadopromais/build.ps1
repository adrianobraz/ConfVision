# FranqueadoPro — build cross-compile Windows -> Linux (Proxmox)
#
# Uso:
#   .\build.ps1                 # build endurecido (padrao)
#   .\build.ps1 -Obfuscate      # build com Garble (ofuscado)
#   .\build.ps1 -Local          # build para Windows (teste local)
#   .\build.ps1 -Package        # alem do build, separa os arquivos essenciais em .\deploy\
#
# Depois: subir franqueadopro via FileZilla + sudo systemctl restart confmonit4franqueadopro

param(
    [switch]$Obfuscate,
    [switch]$Local,
    [switch]$Package
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

$OutputName = "franqueadopro"
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
    $OutputName = "franqueadopro.exe"
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
        Write-Host "  (adicione $GoBin ao PATH se necessario)" -ForegroundColor DarkGray
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

# ------------------------------------------------------------------
# Arquivos essenciais para subir e funcionar
# ------------------------------------------------------------------
Write-Host ""
Write-Host "Arquivos essenciais para o servidor:" -ForegroundColor Cyan
Write-Host "  1. $OutputName        -> binario compilado"
Write-Host "  2. .env               -> configuracao (porta, API, certificados, titulo)"
Write-Host "  3. recursos\          -> templates, css, js e imagens"
Write-Host ""
Write-Host "Nao precisa subir: codigo-fonte (*.go), go.mod, go.sum, build.ps1, *.map, deploy\" -ForegroundColor DarkGray

# ------------------------------------------------------------------
# -Package: separa os essenciais em .\deploy\franqueadopro\
# ------------------------------------------------------------------
if ($Package) {
    $deployDir = Join-Path $PSScriptRoot "deploy\franqueadopro"

    Write-Host ""
    Write-Host "Empacotando em: $deployDir" -ForegroundColor Cyan

    if (Test-Path $deployDir) {
        Remove-Item $deployDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $deployDir -Force | Out-Null

    Copy-Item $OutputName -Destination $deployDir -Force
    if (Test-Path ".env") {
        Copy-Item ".env" -Destination $deployDir -Force
    }

    # recursos\ sem source maps para manter o pacote enxuto
    $recursosDest = Join-Path $deployDir "recursos"
    Copy-Item "recursos" -Destination $recursosDest -Recurse -Force
    Get-ChildItem $recursosDest -Recurse -Include *.map | Remove-Item -Force -ErrorAction SilentlyContinue

    $leiaMe = @"
FranqueadoPro — pacote pronto para subir
========================================

SUBA TODO o conteudo desta pasta para:
  /home/confmonit/v4.0/franqueadopro/

Arquivos:
  - franqueadopro   (binario)
  - .env            (ja configurado — nao editar)
  - recursos\       (pasta inteira)

No servidor (uma vez apos subir):
  chmod +x franqueadopro
  systemctl restart confmonit4franqueadopro

Teste: login -> menu Meu Plano
"@
    Set-Content -Path (Join-Path $deployDir "LEIA-ME.txt") -Value $leiaMe -Encoding UTF8

    $totalMB = [math]::Round(((Get-ChildItem $deployDir -Recurse | Measure-Object Length -Sum).Sum / 1MB), 2)
    Write-Host "OK: pacote pronto ($totalMB MB)" -ForegroundColor Green
    Write-Host "Suba todo o conteudo de deploy\franqueadopro\ para /home/confmonit/v4.0/franqueadopro/" -ForegroundColor DarkGray
}

Write-Host ""
Write-Host "Deploy:" -ForegroundColor Cyan
Write-Host "  1. FileZilla -> /home/confmonit/v4.0/franqueadopro/franqueadopro"
if (-not $Local) {
    Write-Host "  2. SSH: chmod +x franqueadopro && sudo systemctl restart confmonit4franqueadopro"
    Write-Host "  3. SSH: sudo systemctl status confmonit4franqueadopro"
}
