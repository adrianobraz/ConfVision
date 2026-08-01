# =============================================================================
# FranqueadoPro SaaS — build e pacote de deploy (dominio automatico)
# =============================================================================
#
# Compila para Linux (servidor core-4) e gera pasta deploy\fp-dominio-saas\
# pronta para subir via FileZilla.
#
# Uso (PowerShell, na pasta v4.0):
#   .\build-fp-dominio-saas.ps1
#   .\build-fp-dominio-saas.ps1 -ProvisionerKey "sua-chave-secreta-aqui"
#
# Depois: leia DEPLOY-DOMINIO-SAAS.txt e suba as pastas indicadas.

param(
    [string]$ProvisionerKey = "ALTERE-ESTA-CHAVE-NO-SERVIDOR"
)

$ErrorActionPreference = "Stop"
$Root = $PSScriptRoot
$DeployRoot = Join-Path $Root "deploy\fp-dominio-saas"

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Error "Go nao encontrado. Instale: https://go.dev/dl/"
}

$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

Write-Host "=== Build FranqueadoPro SaaS (Linux amd64) ===" -ForegroundColor Cyan
Write-Host "Go: $(go version)" -ForegroundColor DarkGray

# ---------------------------------------------------------------------------
# 1) FranqueadoPro
# ---------------------------------------------------------------------------
$FpDir = Join-Path $Root "franqueadopro"
$FpOut = Join-Path $FpDir "franqueadopro"
Write-Host ""
Write-Host "[1/2] franqueadopro..." -ForegroundColor Yellow
Push-Location $FpDir
go build -ldflags="-s -w" -trimpath -o franqueadopro .
if ($LASTEXITCODE -ne 0) { Pop-Location; Write-Error "Build franqueadopro falhou" }
Pop-Location
Write-Host "  OK: $FpOut" -ForegroundColor Green

# ---------------------------------------------------------------------------
# 2) fp-dominio-provision
# ---------------------------------------------------------------------------
$ProvDir = Join-Path $Root "fp-dominio-provision"
$ProvOut = Join-Path $ProvDir "fp-dominio-provision"
Write-Host ""
Write-Host "[2/2] fp-dominio-provision..." -ForegroundColor Yellow
Push-Location $ProvDir
go mod tidy
go build -ldflags="-s -w" -trimpath -o fp-dominio-provision .
if ($LASTEXITCODE -ne 0) { Pop-Location; Write-Error "Build fp-dominio-provision falhou" }
Pop-Location
Write-Host "  OK: $ProvOut" -ForegroundColor Green

# ---------------------------------------------------------------------------
# 3) Empacotar deploy
# ---------------------------------------------------------------------------
Write-Host ""
Write-Host "Empacotando: $DeployRoot" -ForegroundColor Cyan

if (Test-Path $DeployRoot) {
    Remove-Item $DeployRoot -Recurse -Force
}

$FpDeploy = Join-Path $DeployRoot "franqueadopro"
$ProvDeploy = Join-Path $DeployRoot "fp-dominio-provision"
$SvcDeploy = Join-Path $DeployRoot "systemd"
$StartDeploy = Join-Path $DeployRoot "start"

New-Item -ItemType Directory -Path $FpDeploy -Force | Out-Null
New-Item -ItemType Directory -Path $ProvDeploy -Force | Out-Null
New-Item -ItemType Directory -Path $SvcDeploy -Force | Out-Null
New-Item -ItemType Directory -Path $StartDeploy -Force | Out-Null

# FranqueadoPro
Copy-Item $FpOut -Destination $FpDeploy -Force
Copy-Item (Join-Path $FpDir "recursos") -Destination (Join-Path $FpDeploy "recursos") -Recurse -Force
Get-ChildItem (Join-Path $FpDeploy "recursos") -Recurse -Include *.map -ErrorAction SilentlyContinue | Remove-Item -Force

# .env franqueadopro (gera com PROVISIONER_KEY)
$fpEnvSrc = Join-Path $FpDir ".env"
$fpEnvDest = Join-Path $FpDeploy ".env"
if (Test-Path $fpEnvSrc) {
    $envContent = Get-Content $fpEnvSrc -Raw
    if ($envContent -match 'PROVISIONER_KEY=') {
        $envContent = $envContent -replace 'PROVISIONER_KEY="[^"]*"', "PROVISIONER_KEY=`"$ProvisionerKey`""
    } else {
        $envContent += "`nPROVISIONER_URL=`"http://127.0.0.1:2015`"`nPROVISIONER_KEY=`"$ProvisionerKey`"`n"
    }
    if ($envContent -notmatch 'PROVISIONER_URL=') {
        $envContent += "`nPROVISIONER_URL=`"http://127.0.0.1:2015`"`n"
    }
    Set-Content -Path $fpEnvDest -Value $envContent.TrimEnd() -Encoding UTF8 -NoNewline
    Add-Content -Path $fpEnvDest -Value "`n" -Encoding UTF8
} else {
    @"
TITULO_SITE=FranqueadoPro
URL_API=http://localhost:2000
PORTA=2005
HTTPS=NAO
PROVISIONER_URL=http://127.0.0.1:2015
PROVISIONER_KEY=$ProvisionerKey
"@ | Set-Content -Path $fpEnvDest -Encoding UTF8
}

# Provisionador
Copy-Item $ProvOut -Destination $ProvDeploy -Force
@"
PORTA=2015
API_KEY=$ProvisionerKey
PROXY_TARGET=http://127.0.0.1:2005/
APACHE_SITES_DIR=/etc/apache2/sites-available
CERTBOT_EMAIL=suporte@kitsite.com.br
"@ | Set-Content -Path (Join-Path $ProvDeploy ".env") -Encoding UTF8

# systemd + start (copiar do repo home/confmonit)
$HomeConfmonit = Join-Path $Root ".."
if (Test-Path (Join-Path $HomeConfmonit "service\confmonit4fp-dominio-provision.service")) {
    Copy-Item (Join-Path $HomeConfmonit "service\confmonit4fp-dominio-provision.service") $SvcDeploy -Force
}
if (Test-Path (Join-Path $HomeConfmonit "start\fp-dominio-provision.sh")) {
    Copy-Item (Join-Path $HomeConfmonit "start\fp-dominio-provision.sh") $StartDeploy -Force
}

# Guia rapido no pacote
Copy-Item (Join-Path $Root "DEPLOY-DOMINIO-SAAS.txt") $DeployRoot -Force -ErrorAction SilentlyContinue

$totalMB = [math]::Round(((Get-ChildItem $DeployRoot -Recurse -File | Measure-Object Length -Sum).Sum / 1MB), 2)

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host " PACOTE PRONTO: $DeployRoot ($totalMB MB)" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Subir para o servidor:" -ForegroundColor Cyan
Write-Host "  deploy\fp-dominio-saas\franqueadopro\*     -> /home/confmonit/v4.0/franqueadopro/" -ForegroundColor White
Write-Host "  deploy\fp-dominio-saas\fp-dominio-provision\* -> /home/confmonit/v4.0/fp-dominio-provision/" -ForegroundColor White
Write-Host "  deploy\fp-dominio-saas\start\fp-dominio-provision.sh -> /home/confmonit/start/" -ForegroundColor White
Write-Host "  deploy\fp-dominio-saas\systemd\*.service   -> /etc/systemd/system/ (uma vez)" -ForegroundColor White
Write-Host ""
Write-Host "Xano: push dos arquivos listados em DEPLOY-DOMINIO-SAAS.txt" -ForegroundColor Yellow
Write-Host "Chave provisionador: $ProvisionerKey" -ForegroundColor DarkGray
