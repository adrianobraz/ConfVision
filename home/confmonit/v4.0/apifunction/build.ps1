# apifunction - build Linux (VPS)

$ErrorActionPreference = "Stop"
$OutDir = Join-Path $PSScriptRoot "deploy\apifunction"
$BinName = "apifunction"

Push-Location $PSScriptRoot
try {
    go mod tidy
    if ($LASTEXITCODE -ne 0) { throw "go mod tidy falhou" }

    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    $env:CGO_ENABLED = "0"

    New-Item -ItemType Directory -Force -Path $OutDir | Out-Null
    go build -o (Join-Path $OutDir $BinName) .
    if ($LASTEXITCODE -ne 0) { throw "go build falhou" }

    if (Test-Path (Join-Path $OutDir ".env")) {
        Write-Host "Mantendo .env existente em deploy"
    } elseif (Test-Path ".env") {
        Copy-Item -Force ".env" (Join-Path $OutDir ".env")
    } elseif (Test-Path ".env.example") {
        Copy-Item -Force ".env.example" (Join-Path $OutDir ".env.example")
    }
    Copy-Item -Force -Recurse "sql" (Join-Path $OutDir "sql")
    if (Test-Path "deploy\confmonit4apifunction.service") {
        Copy-Item -Force "deploy\confmonit4apifunction.service" (Join-Path $OutDir "confmonit4apifunction.service")
    }

    @"
apifunction - deploy (governanca repasse)
=========================================

SUBIR para: /home/confmonit/v4.0/apifunction/

Arquivos:
  - apifunction   (binario Linux — migration Postgres fp_gov_* embutida no binario)
  - .env

MySQL confmonitV4 (uma vez, se ainda nao rodou):
  - sql/001_fp_contrato_schema.sql
  - sql/003_fp_fatura_estornada.sql

Postgres (fp_gov_*, fp_catalogo_*, fp_pacote_cota, fp_central_preco_config):
  migration automatica no start (POSTGRES_URL no .env).

Variaveis criticas .env:
  POSTGRES_URL, WORKER_ADMIN_TOKEN, WORKER_SECRET, XANO_API_FINANCEIRO
  XANO_META_ACCESS_TOKEN, XANO_META_WORKSPACE_ID=1

Xano fp_config_financeiro:
  apifunction_url = http://185.130.61.4:20001
  worker_secret   = fp-worker-change-me

Apos subir:
  chmod +x apifunction
  systemctl restart confmonit4apifunction
  curl http://127.0.0.1:20001/health

Governanca:
  POST /financeiro/governanca/worker/tick  (X-Worker-Key)
  GET  /financeiro/governanca/estado
  GET  /financeiro/governanca/franqueado

Catalogo PostgreSQL (admConfmonit):
  POST /financeiro/catalogo/listar|salvar|seed
  POST /financeiro/pacote-cota/listar|salvar|seed
  POST /financeiro/central-preco-config/listar|salvar

Faturas Central (admConfmonit):
  POST /financeiro/fatura/listar  (requer XANO_META_ACCESS_TOKEN no .env)
"@ | Set-Content (Join-Path $OutDir "LEIA-ME.txt") -Encoding UTF8

    Write-Host "OK: $OutDir\$BinName"
} finally {
    Pop-Location
}
