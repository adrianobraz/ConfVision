# Deploy Capacidade de Processamento — ConfVision
# Uso: .\scripts\deploy-capacidade.ps1
# Requer: Go, FileZilla manual ou SCP para core-4

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

Write-Host "== 1/4 Normalizar .env (LF, UTF-8) ==" -ForegroundColor Cyan
& "$PSScriptRoot\prepare-env-for-linux.ps1"

Write-Host "== 2/4 Build Linux amd64 ==" -ForegroundColor Cyan
& "$root\build.ps1"

Write-Host "== 3/4 Verificar SQL capacidade (opcional) ==" -ForegroundColor Cyan
$sqlApply = Join-Path $root "..\..\..\confvision\sql\cmd\apply"
$sqlFile = Join-Path $root "..\..\..\confvision\sql\014_capacidade_processamento.sql"
if (Test-Path $sqlApply) {
    $line = Get-Content "$root\.env" | Where-Object { $_ -match '^POSTGRES_URL=' } | Select-Object -First 1
    if ($line) {
        $env:POSTGRES_URL = ($line -replace '^POSTGRES_URL=', '').Trim()
        Push-Location $sqlApply
        go run . $sqlFile
        Pop-Location
    }
}

Write-Host ""
Write-Host "== 4/4 Subir para VPS (manual) ==" -ForegroundColor Yellow
Write-Host "Destino: /home/confmonit/v4.0/confvision/"
Write-Host ""
Write-Host "Arquivos:" -ForegroundColor Green
Write-Host "  confvision                          (binario)"
Write-Host "  .env"
Write-Host "  recursos/modulos/minha-empresa/minhas-licencas/minhas-licencas.html"
Write-Host "  recursos/modulos/minha-empresa/minhas-licencas/minhas-licencas.js"
Write-Host "  recursos/modulos/minha-empresa/minhas-licencas/minhas-licencas.css"
Write-Host ""
Write-Host "No servidor:" -ForegroundColor Green
Write-Host "  chmod +x /home/confmonit/v4.0/confvision/confvision"
Write-Host "  sudo systemctl restart confmonit4confvision"
Write-Host ""
Write-Host "Xano (push se ainda nao fez):" -ForegroundColor Magenta
Write-Host "  548_fn_fp_confvision_capacidade_faturar"
Write-Host "  549_fn_fp_confvision_ativar_capacidade_pagamento"
Write-Host "  2398_fp_confvision_capacidade_faturar (fp_franqueadoPro)"
Write-Host "  2272_fp_fatura_registrar_pagamento (item vis_capacidade_contrato)"
Write-Host "  452_fn_fp_confvision_fatura_venda (ref_id Postgres obrigatorio nos itens)"
Write-Host "  453_fn_fp_confvision_reconciliar_licencas_pagas (451 Postgres-only)"
Write-Host "  546_fn_fp_confvision_licenca_postgres_obter / 547 listar"
Write-Host "  448 / 2274 / 2300 resumo Postgres"
Write-Host "  2276 expirar via ops"
Write-Host "  2434_fp_confvision_reconciliar_licencas_pagas (adm recuperacao)"
Write-Host "  2433/2435 portal renovacao (fp_franqueadoPro — nao usar 2383/2384, IDs do financeiro orfaos)"
Write-Host "  543_fn_fp_confvision_renovacao_listar"
Write-Host "  544_fn_fp_confvision_renovacao_faturar"
Write-Host "  545_fn_fp_confvision_postgres_sync_licenca (URL confvision_api_url alinhada a 549)"
Write-Host "  2286_fp_fatura_minhas (incluir_itens para reconciliacao ConfCam)"
Write-Host "  (remover no Xano functions obsoletas 520-522 se existirem duplicadas)"
Write-Host "  451_fn_fp_confvision_ativar_licenca_pagamento (reativa em_uso)"
Write-Host "  2289_fp_confvision_fatura_renovacao (adm)"
Write-Host "  2383_fp_confvision_listar_renovacao / 2384_fp_confvision_fatura_renovacao (portal)"
Write-Host "  531_fn_fp_fatura_estornar_pagamento"
Write-Host "  2272_fp_fatura_registrar_pagamento"
Write-Host "  2277_fp_vis_licenca_listar_pendentes_renovacao"
Write-Host "  2278_fp_fatura_gerar_automatica_confvision"
Write-Host ""
Write-Host "Billing worker: rebuild e restart se usar renovacao automatica." -ForegroundColor DarkGray
