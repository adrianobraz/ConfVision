# Build e deploy - Portal do parceiro ConfService (Windows)
#
# Uso:
#   .\build.ps1                 # npm run build (usa .env.production se existir)
#   .\build.ps1 -Package        # gera deploy\confservice-portal\
#
# Antes: ajuste VITE_API_URL em .env.production
#   Mesmo dominio (Proxmox + Apache proxy):  https://confservice.com.br
#   Portal em hospedagem + API no Proxmox:   https://SEU-IP-OU-API:2020  (precisa HTTPS ou ambos HTTP)

param(
    [switch]$Package
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

if (-not (Test-Path "package.json")) {
    Write-Error "Execute na pasta confservice\web"
}

if (-not (Test-Path ".env.production")) {
    if (Test-Path ".env.example") {
        Copy-Item ".env.example" ".env.production"
        Write-Host "Criado .env.production a partir de .env.example - revise VITE_API_URL" -ForegroundColor Yellow
    } else {
        Write-Host "Aviso: sem .env.production - Vite pode usar localhost" -ForegroundColor Yellow
    }
}

if (-not (Test-Path "node_modules")) {
    Write-Host "npm install..." -ForegroundColor Cyan
    npm install
    if ($LASTEXITCODE -ne 0) {
        Write-Error "npm install falhou"
    }
}

Write-Host "npm run build..." -ForegroundColor Cyan
npm run build
if ($LASTEXITCODE -ne 0) {
    Write-Error "Build falhou"
}

$dist = Join-Path $PSScriptRoot "dist"
if (-not (Test-Path $dist)) {
    Write-Error "Pasta dist nao gerada"
}

# Garante .htaccess no dist (SPA em Apache / hospedagem)
$htaccessSrc = Join-Path $PSScriptRoot "public\.htaccess"
if (Test-Path $htaccessSrc) {
    Copy-Item $htaccessSrc -Destination (Join-Path $dist ".htaccess") -Force
}

$sizeMB = [math]::Round(((Get-ChildItem $dist -Recurse | Measure-Object Length -Sum).Sum / 1MB), 2)
Write-Host ""
Write-Host ('OK: dist/ (' + $sizeMB + ' MB)') -ForegroundColor Green

Write-Host ""
Write-Host "Arquivos essenciais:" -ForegroundColor Cyan
Write-Host "  dist\  -> site estatico (HTML/JS/CSS)"
Write-Host ""
Write-Host "Nao precisa subir: src\, node_modules\, package.json, .env (API ja embutida no build)" -ForegroundColor DarkGray

if ($Package) {
    $deployDir = Join-Path $PSScriptRoot "..\deploy\confservice-portal"
    if (Test-Path $deployDir) {
        Remove-Item $deployDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $deployDir -Force | Out-Null
    Copy-Item "$dist\*" -Destination $deployDir -Recurse -Force

    $apacheSrc = Join-Path $PSScriptRoot "..\deploy\apache-confservice.com.br.conf"
    if (Test-Path $apacheSrc) {
        Copy-Item $apacheSrc -Destination (Join-Path $deployDir "apache-confservice.com.br.conf") -Force
    }

    $leiaMe = @"
Portal parceiro ConfService - pacote pronto
==========================================

SUBA TODO o conteudo desta pasta (exceto este LEIA-ME e o .conf do Apache)
para a pasta publica da hospedagem (public_html) OU para o Proxmox:
  /home/confmonit/v4.0/confservice/web-dist/

Onde pode hospedar?
-------------------
A) Hospedagem normal (Apache/cPanel / KingHost) - SIM
   - Site estatico + cs-proxy.php + .htaccess (ja no pacote)
   - /parceiro/*, /marketplace/*, /health sao encaminhados para a API no Proxmox
   - Ajuste CS_API_BACKEND em cs-proxy.env se o IP mudar
   - VITE_API_URL=https://confservice.com.br (mesmo dominio; sem mixed content)
   - Sem o proxy, o SPA engole /parceiro/registrar e NAO grava no banco

B) Proxmox (recomendado se o dominio aponta para o core-4)
   - Suba web-dist + use apache-confservice.com.br.conf
   - Apache serve o portal e faz proxy /parceiro /marketplace /internal /health -> :2020
   - VITE_API_URL=https://confservice.com.br (mesmo dominio)

API (sempre no Proxmox):
  binario confservice + systemctl confmonit4confservice
"@
    Set-Content -Path (Join-Path $deployDir "LEIA-ME.txt") -Value $leiaMe -Encoding UTF8

    Write-Host ""
    Write-Host "Pacote: deploy\confservice-portal\" -ForegroundColor Green
    Write-Host "Suba o conteudo para public_html (hospedagem) ou web-dist/ (Proxmox)" -ForegroundColor DarkGray
}

Write-Host ""
Write-Host "Comando:" -ForegroundColor Cyan
Write-Host "  cd ...\confservice\web"
Write-Host "  powershell -ExecutionPolicy Bypass -File .\build.ps1 -Package"
