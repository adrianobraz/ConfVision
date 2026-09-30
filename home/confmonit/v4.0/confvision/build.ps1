# ConfVision — build cross-compile Windows -> Linux (Proxmox)
#
# Uso:
#   .\build.ps1                 # build endurecido (padrao)
#   .\build.ps1 -Obfuscate        # build com Garble (ofuscado)
#   .\build.ps1 -Local            # build para Windows (teste local)
#
# Depois: subir confvision via FileZilla + sudo systemctl restart confmonit4confvision

param(
    [switch]$Obfuscate,
    [switch]$Local
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

$OutputName = "confvision"
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
    $OutputName = "confvision.exe"
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

# Smoke: rotas worker API (D5) no binario; senao Dispatch retorna 404 em producao.
$binText = [System.IO.File]::ReadAllBytes($bin.FullName)
$ascii = [System.Text.Encoding]::ASCII.GetString($binText)
$d5Ok = $ascii.Contains("/vis_rust_processor_capacity")
$pingOk = $ascii.Contains("/vis_worker_ping")

Write-Host ""
Write-Host "OK: $($bin.FullName) ($sizeMB MB)" -ForegroundColor Green
Write-Host "Rotas no binario: vis_worker_ping=$pingOk vis_rust_processor_capacity=$d5Ok" -ForegroundColor $(if ($d5Ok -and $pingOk) { "Green" } else { "Yellow" })
if (-not $d5Ok) {
    Write-Host "AVISO: binario sem D5 - confira src/modulos/visdata/router.go antes do deploy." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Deploy API central vision.confmonit2.com.br via Proxmox (nao foxpro EasyPanel):" -ForegroundColor Cyan
Write-Host "  1. FileZilla: /home/confmonit/v4.0/confvision/confvision"
if (-not $Local) {
    Write-Host '  2. SSH: chmod +x confvision; sudo systemctl restart confmonit4confvision'
    Write-Host '  3. Teste D5: curl com Authorization Bearer em /vis_rust_processor_capacity (vision.confmonit2.com.br)'
}
