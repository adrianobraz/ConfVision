# TaskXano — build cross-compile Windows -> Linux (Proxmox)
#
# Worker/daemon (WhatsApp, ligacao, autofim, arme/desarme). Sem frontend.
#
# Uso:
#   powershell -ExecutionPolicy Bypass -File .\build.ps1
#   powershell -ExecutionPolicy Bypass -File .\build.ps1 -Local
#
# Depois: subir taskxano + .env via FileZilla + sudo systemctl restart confmonit4taskxano

param(
    [switch]$Obfuscate,
    [switch]$Local
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

$OutputName = "taskxano"
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
    $OutputName = "taskxano.exe"
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
Write-Host "  1. FileZilla -> /home/confmonit/v4.0/taskxano/taskxano"
Write-Host "     (+ .env - unico outro arquivo necessario)"
if (-not $Local) {
    Write-Host "  2. SSH: chmod +x taskxano && sudo systemctl restart confmonit4taskxano"
}
