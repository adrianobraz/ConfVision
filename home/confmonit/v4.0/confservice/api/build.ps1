# confservice API - build cross-compile Windows -> Linux (core-4)
#
# Uso:
#   .\build.ps1                 # build endurecido Linux amd64 (padrao)
#   .\build.ps1 -Obfuscate      # build com Garble (ofuscado)
#   .\build.ps1 -Local          # build para Windows (teste local)
#   .\build.ps1 -Package        # alem do build, separa arquivos em ..\deploy\confservice\
#
# Depois: FileZilla + systemctl restart confmonit4confservice

param(
    [switch]$Obfuscate,
    [switch]$Local,
    [switch]$Package
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

$OutputName = "confservice"
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
    Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
    $OutputName = "confservice.exe"
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
Write-Host "  1. $OutputName   -> /home/confmonit/v4.0/confservice/"
Write-Host "  2. .env          -> mesma pasta (MYSQL_DSN, API_KEY, JWT_SECRET)"
Write-Host ""
Write-Host "Uma vez no servidor:" -ForegroundColor Cyan
Write-Host "  /home/confmonit/start/confservice.sh"
Write-Host "  /etc/systemd/system/confmonit4confservice.service"
Write-Host ""
Write-Host "Nao precisa subir: *.go, go.mod, go.sum, build.ps1, web\" -ForegroundColor DarkGray

if ($Package) {
    $deployDir = Join-Path $PSScriptRoot "..\deploy\confservice"
    $homeConfmonit = Join-Path $PSScriptRoot "..\..\.."

    Write-Host ""
    Write-Host "Empacotando em: $deployDir" -ForegroundColor Cyan

    if (Test-Path $deployDir) {
        Remove-Item $deployDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $deployDir -Force | Out-Null
    New-Item -ItemType Directory -Path (Join-Path $deployDir "sql") -Force | Out-Null

    Copy-Item $OutputName -Destination $deployDir -Force

    if (Test-Path ".env") {
        Copy-Item ".env" -Destination $deployDir -Force
    } elseif (Test-Path ".env.example") {
        Copy-Item ".env.example" -Destination (Join-Path $deployDir ".env") -Force
        Write-Host "  Aviso: .env nao existe - copiado .env.example (edite no servidor)" -ForegroundColor Yellow
    }

    if (Test-Path "sql\schema.sql") {
        Copy-Item "sql\schema.sql" -Destination (Join-Path $deployDir "sql\schema.sql") -Force
    }
    if (Test-Path "sql\schema-tables.sql") {
        Copy-Item "sql\schema-tables.sql" -Destination (Join-Path $deployDir "sql\schema-tables.sql") -Force
    }

    $svcSrc = Join-Path $homeConfmonit "service\confmonit4confservice.service"
    $startSrc = Join-Path $homeConfmonit "start\confservice.sh"
    if (Test-Path $svcSrc) {
        New-Item -ItemType Directory -Path (Join-Path $deployDir "systemd") -Force | Out-Null
        Copy-Item $svcSrc (Join-Path $deployDir "systemd") -Force
    }
    if (Test-Path $startSrc) {
        New-Item -ItemType Directory -Path (Join-Path $deployDir "start") -Force | Out-Null
        Copy-Item $startSrc (Join-Path $deployDir "start") -Force
    }

    $faturaSrc = Join-Path $homeConfmonit "start\confservice-fatura-fechar.sh"
    $installCronSrc = Join-Path $homeConfmonit "start\install-confservice-fatura-cron.sh"
    if (Test-Path $faturaSrc) {
        New-Item -ItemType Directory -Path (Join-Path $deployDir "start") -Force | Out-Null
        Copy-Item $faturaSrc (Join-Path $deployDir "start") -Force
    }
    if (Test-Path $installCronSrc) {
        New-Item -ItemType Directory -Path (Join-Path $deployDir "start") -Force | Out-Null
        Copy-Item $installCronSrc (Join-Path $deployDir "start") -Force
    }

    $cronSrc = Join-Path $homeConfmonit "cron\confservice-fatura"
    $cronDir = Join-Path $deployDir "cron"
    New-Item -ItemType Directory -Path $cronDir -Force | Out-Null
    if (Test-Path $cronSrc) {
        Copy-Item $cronSrc (Join-Path $cronDir "confservice-fatura") -Force
    } else {
        @(
            "# ConfService - fechamento quinzenal de faturas"
            "SHELL=/bin/bash"
            "PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin"
            ""
            "5 0 1,16 * * root /home/confmonit/start/confservice-fatura-fechar.sh"
        ) | Set-Content -Path (Join-Path $cronDir "confservice-fatura") -Encoding ascii
    }

    $leiaMePath = Join-Path $deployDir "LEIA-ME.txt"
    @(
        "confservice - pacote pronto para subir"
        "======================================"
        ""
        "SUBA o conteudo principal para:"
        "  /home/confmonit/v4.0/confservice/"
        ""
        "Arquivos desta pasta (deploy\confservice\):"
        "  - confservice            (binario Linux)"
        "  - .env                   (MYSQL_DSN com user confmonit + DB confservice)"
        "  - sql\schema-tables.sql  (colar no phpMyAdmin se tabelas ainda nao existem)"
        ""
        "Uma vez no servidor (se ainda nao fez):"
        "  start\confservice.sh  -> /home/confmonit/start/"
        "  systemd\confmonit4confservice.service -> /etc/systemd/system/"
        ""
        "Cron fatura (dias 1 e 16) - FileZilla NAO grava em /etc/cron.d:"
        "  Suba para /home/confmonit/v4.0/confservice/:"
        "    cron\confservice-fatura"
        "    start\confservice-fatura-fechar.sh"
        "    start\install-confservice-fatura-cron.sh"
        "  Depois no SSH:"
        "    chmod +x /home/confmonit/v4.0/confservice/start/install-confservice-fatura-cron.sh"
        "    sudo /home/confmonit/v4.0/confservice/start/install-confservice-fatura-cron.sh"
        "    /home/confmonit/start/confservice-fatura-fechar.sh"
        ""
        "MySQL (ja feito se voce criou o DB):"
        "  GRANT ALL ON confservice.* TO 'confmonit'@'%';"
        "  Importar sql/schema-tables.sql no banco confservice"
        ""
        "Comandos SSH:"
        "  chmod +x /home/confmonit/v4.0/confservice/confservice"
        "  chmod +x /home/confmonit/start/confservice.sh"
        "  systemctl daemon-reload"
        "  systemctl enable confmonit4confservice"
        "  systemctl restart confmonit4confservice"
        "  curl http://127.0.0.1:2020/health"
        ""
        "Franqueado Pro .env (mesma API_KEY):"
        "  CONFSERVICE_URL=http://127.0.0.1:2020"
        "  CONFSERVICE_API_KEY=<igual API_KEY do .env deste pacote>"
    ) | Set-Content -Path $leiaMePath -Encoding UTF8

    $totalMB = [math]::Round(((Get-ChildItem $deployDir -Recurse -File | Measure-Object Length -Sum).Sum / 1MB), 2)
    Write-Host "OK: pacote pronto ($totalMB MB)" -ForegroundColor Green
    Write-Host "Suba deploy\confservice\ para /home/confmonit/v4.0/confservice/" -ForegroundColor DarkGray
}

Write-Host ""
Write-Host "Deploy:" -ForegroundColor Cyan
Write-Host "  1. FileZilla -> /home/confmonit/v4.0/confservice/confservice"
Write-Host "  2. FileZilla -> /home/confmonit/v4.0/confservice/.env"
if (-not $Local) {
    Write-Host "  3. SSH: chmod +x confservice && sudo systemctl restart confmonit4confservice"
    Write-Host "  4. SSH: curl http://127.0.0.1:2020/health"
}
