# fp-dominio-provision - build cross-compile Windows -> Linux (core-4)
#
# Uso:
#   .\build.ps1                 # build endurecido (padrao)
#   .\build.ps1 -Obfuscate      # build com Garble (ofuscado)
#   .\build.ps1 -Local          # build para Windows (teste local)
#   .\build.ps1 -Package        # alem do build, separa os arquivos essenciais em .\deploy\
#
# Depois: subir via FileZilla + systemctl restart confmonit4fp-dominio-provision

param(
    [switch]$Obfuscate,
    [switch]$Local,
    [switch]$Package
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

$OutputName = "fp-dominio-provision"
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
    $OutputName = "fp-dominio-provision.exe"
    Write-Host "Alvo: Windows (teste local)" -ForegroundColor Cyan
}

Write-Host "Go: $(go version)" -ForegroundColor DarkGray

go mod tidy

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

Write-Host ""
Write-Host "Arquivos essenciais para o servidor:" -ForegroundColor Cyan
Write-Host "  1. $OutputName        -> binario compilado"
Write-Host "  2. .env               -> configuracao (PORTA, API_KEY, PROXY_TARGET)"
Write-Host ""
Write-Host "Tambem necessario no servidor (uma vez):" -ForegroundColor Cyan
Write-Host "  /home/confmonit/start/fp-dominio-provision.sh"
Write-Host "  /etc/systemd/system/confmonit4fp-dominio-provision.service"
Write-Host ""
Write-Host "Nao precisa subir: codigo-fonte (*.go), go.mod, go.sum, build.ps1, deploy\" -ForegroundColor DarkGray

if ($Package) {
    $deployDir = Join-Path $PSScriptRoot "deploy\fp-dominio-provision"
    $homeConfmonit = Join-Path $PSScriptRoot "..\.."

    Write-Host ""
    Write-Host "Empacotando em: $deployDir" -ForegroundColor Cyan

    if (Test-Path $deployDir) {
        Remove-Item $deployDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $deployDir -Force | Out-Null

    Copy-Item $OutputName -Destination $deployDir -Force

    if (Test-Path ".env") {
        Copy-Item ".env" -Destination $deployDir -Force
    } elseif (Test-Path ".env.example") {
        Copy-Item ".env.example" -Destination (Join-Path $deployDir ".env") -Force
        Write-Host "  Aviso: .env nao existe - copiado .env.example (edite API_KEY no servidor)" -ForegroundColor Yellow
    }

    $svcSrc = Join-Path $homeConfmonit "service\confmonit4fp-dominio-provision.service"
    $startSrc = Join-Path $homeConfmonit "start\fp-dominio-provision.sh"
    if (Test-Path $svcSrc) {
        New-Item -ItemType Directory -Path (Join-Path $deployDir "systemd") -Force | Out-Null
        Copy-Item $svcSrc (Join-Path $deployDir "systemd") -Force
    }
    if (Test-Path $startSrc) {
        New-Item -ItemType Directory -Path (Join-Path $deployDir "start") -Force | Out-Null
        Copy-Item $startSrc (Join-Path $deployDir "start") -Force
    }

    $leiaMePath = Join-Path $deployDir "LEIA-ME.txt"
    @(
        "fp-dominio-provision - pacote pronto para subir"
        "================================================"
        ""
        "SUBA o conteudo principal para:"
        "  /home/confmonit/v4.0/fp-dominio-provision/"
        ""
        "Arquivos desta pasta (deploy\fp-dominio-provision\):"
        "  - fp-dominio-provision   (binario)"
        "  - .env                   (API_KEY = PROVISIONER_KEY do franqueadopro)"
        ""
        "Uma vez no servidor (se ainda nao fez):"
        "  start\fp-dominio-provision.sh  -> /home/confmonit/start/"
        "  systemd\confmonit4fp-dominio-provision.service -> /etc/systemd/system/"
        ""
        "Comandos SSH:"
        "  chmod +x /home/confmonit/v4.0/fp-dominio-provision/fp-dominio-provision"
        "  chmod +x /home/confmonit/start/fp-dominio-provision.sh"
        "  systemctl daemon-reload"
        "  systemctl enable confmonit4fp-dominio-provision"
        "  systemctl restart confmonit4fp-dominio-provision"
        "  curl http://127.0.0.1:2015/health"
    ) | Set-Content -Path $leiaMePath -Encoding UTF8

    $totalMB = [math]::Round(((Get-ChildItem $deployDir -Recurse -File | Measure-Object Length -Sum).Sum / 1MB), 2)
    Write-Host "OK: pacote pronto ($totalMB MB)" -ForegroundColor Green
    Write-Host "Suba deploy\fp-dominio-provision\ para /home/confmonit/v4.0/fp-dominio-provision/" -ForegroundColor DarkGray
}

Write-Host ""
Write-Host "Deploy:" -ForegroundColor Cyan
Write-Host "  1. FileZilla -> /home/confmonit/v4.0/fp-dominio-provision/fp-dominio-provision"
Write-Host "  2. FileZilla -> /home/confmonit/v4.0/fp-dominio-provision/.env"
if (-not $Local) {
    Write-Host "  3. SSH: chmod +x fp-dominio-provision && sudo systemctl restart confmonit4fp-dominio-provision"
    Write-Host "  4. SSH: curl http://127.0.0.1:2015/health"
}
