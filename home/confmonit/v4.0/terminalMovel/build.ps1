# Terminal Movel — build cross-compile Windows -> Linux (Proxmox)
#
# Uso:
#   powershell -ExecutionPolicy Bypass -File .\build.ps1
#   powershell -ExecutionPolicy Bypass -File .\build.ps1 -Local
#
# Nao use -Obfuscate: quebra templates HTML (mesmo problema do ConfVision).
#
# Depois: subir terminalMovel via FileZilla + sudo systemctl restart confmonit4terminalMovel

param(
    [switch]$Obfuscate,
    [switch]$Local
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

$OutputName = "terminalMovel"
$GoBin = Join-Path $env:USERPROFILE "go\bin"

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Error "Go nao encontrado. Instale: https://go.dev/dl/"
}

if ($Obfuscate) {
    Write-Host "AVISO: Garble quebra templates HTML neste projeto. Use build sem -Obfuscate." -ForegroundColor Yellow
}

if (-not $Local) {
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    $env:CGO_ENABLED = "0"
    Write-Host "Alvo: Linux amd64" -ForegroundColor Cyan
} else {
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    $OutputName = "terminalMovel.exe"
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
Write-Host "Deploy:" -ForegroundColor Cyan
Write-Host "  1. FileZilla -> /home/confmonit/v4.0/terminalMovel/terminalMovel"
Write-Host "     (+ .env e pasta assets/ se ainda nao estiverem no servidor)"
if (-not $Local) {
    Write-Host "  2. SSH: chmod +x terminalMovel && sudo systemctl restart confmonit4terminalMovel"
}
