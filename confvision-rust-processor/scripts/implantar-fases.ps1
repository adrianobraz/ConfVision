# Implantação operacional Fases 0–4 (+ Fase 5 rampa opcional) - foxpro / todas as câmeras analíticas
# Não altera EasyPanel remotamente: aplica cadastro (Opção A), D5 assign, verifica stack.
# Uso: .\implantar-fases.ps1 [-ApplyOpcaoA] [-AssignD5] [-VerifyOnly] [-SkipNetworkChecks]
param(
    [string]$ApiBase = "https://vision.confmonit2.com.br",
    [string]$RustA = "https://foxpro-rust-pilot.rkr351.easypanel.host",
    [string]$RustB = "https://foxpro-rust-pilot-b.rkr351.easypanel.host",
    [switch]$ApplyOpcaoA,
    [switch]$DespausarAnalitico,
    [switch]$AssignD5,
    [switch]$VerifyOnly,
    [switch]$SkipNetworkChecks,
    [switch]$FullD3Test
)

$ErrorActionPreference = "Continue"
$ScriptDir = $PSScriptRoot
$fail = 0

function Step($title) {
    Write-Host ""
    Write-Host "######## $title ########" -ForegroundColor Cyan
}

if (-not $VerifyOnly) {
    if (-not $ApplyOpcaoA -and -not $AssignD5) {
        $ApplyOpcaoA = $true
        $AssignD5 = $true
    }
}

Step "Fase 0 - cadastro RTSP Opção A (todas as câmeras)"
if ($VerifyOnly) {
    Write-Host "SKIP apply (VerifyOnly)"
} elseif ($ApplyOpcaoA) {
    $oaArgs = @{ AllCameras = $true }
    if ($DespausarAnalitico) { $oaArgs.DespausarAnalitico = $true }
    try {
        & (Join-Path $ScriptDir "opcao-a-apply-all.ps1") @oaArgs
        if ($LASTEXITCODE -ne 0) {
            Write-Host "WARN opcao-a exit $LASTEXITCODE (verify bash/jq no Windows e OK se PUT 30/30)"
        }
    } catch {
        Write-Host "FAIL opcao-a: $_"
        $fail = 1
    }
} else {
    Write-Host "SKIP (-ApplyOpcaoA não passado)"
}

Step "Fase 0 - auditoria analíticas (todas)"
try {
    & (Join-Path $ScriptDir "audit-analiticas.ps1") -ApiBase $ApiBase
    if ($LASTEXITCODE -ne 0) { $fail = 1 }
} catch {
    Write-Host "FAIL audit: $_"
    $fail = 1
}

Step "Fase 0 / D5 - assign processors (só câmeras sem worker_id)"
if ($VerifyOnly) {
    Write-Host "SKIP assign (VerifyOnly)"
} elseif ($AssignD5) {
    try {
        & (Join-Path $ScriptDir "d5-auto-assign-analiticas.ps1") -ApiBase $ApiBase -DryRun
        & (Join-Path $ScriptDir "d5-auto-assign-analiticas.ps1") -ApiBase $ApiBase
        if ($LASTEXITCODE -ne 0) { Write-Host "WARN assign parcial" }
    } catch {
        Write-Host "FAIL d5: $_"
        $fail = 1
    }
}

if ($SkipNetworkChecks) {
    Write-Host ""
    Write-Host "SkipNetworkChecks - fases 2-4 verify remoto omitido"
} else {
    Step "Fase 0 + D6 - stack foxpro"
    try {
        & (Join-Path $ScriptDir "foxpro-stack-verify.ps1") -RustBase $RustA
        if ($LASTEXITCODE -ne 0) { $fail = 1 }
    } catch { $fail = 1 }

    Write-Host "==> Rust B health"
    try {
        $hb = Invoke-RestMethod -Uri "$RustB/health" -TimeoutSec 25
        Write-Host "    status=$($hb.status) cams=$($hb.cameras_online)/$($hb.cameras_total) capacity=$($hb.capacity_state)"
    } catch {
        Write-Host "FAIL Rust B"
        $fail = 1
    }

    Step "Fase 3 - D2 Redis (processors)"
    $bash = Get-Command bash -ErrorAction SilentlyContinue
    if ($bash) {
        bash (Join-Path $ScriptDir "d2-redis-verify.sh") $RustA $RustB
        if ($LASTEXITCODE -ne 0) { Write-Host "WARN D2 (redis pode estar none em VPS)" }
    } else {
        Write-Host "INFO: bash ausente - pule d2-redis-verify.sh no CT111"
    }

    Step "Fase 4 - D6 verify"
    if ($bash) {
        try {
            bash (Join-Path $ScriptDir "phase-d6-verify.sh")
            if ($LASTEXITCODE -ne 0) { Write-Host "WARN phase-d6-verify (rode no CT111 se WSL indisponivel)" }
        } catch {
            Write-Host "WARN bash/WSL indisponivel - pule D6 local"
        }
    } else {
        Write-Host "INFO: rode no CT111: bash confvision-rust-processor/scripts/phase-d6-verify.sh"
    }

    Step "Fase 0 / D3 - teste online"
    $node = Get-Command node -ErrorAction SilentlyContinue
    if ($node) {
        $d3Args = @((Join-Path $ScriptDir "d3-online-test.mjs"), "--quick")
        if ($FullD3Test) { $d3Args = @((Join-Path $ScriptDir "d3-online-test.mjs"), "--full") }
        & node @d3Args
        if ($LASTEXITCODE -ne 0) { $fail = 1 }
    }
}

Step "Fase 1 - multi-tenant (manual por novo cliente)"
Write-Host @"
  Novo cliente: deploy/tenant-stack/scripts/tenant-init-env.sh + docker-compose.tenant.example.yml
  Doc: deploy/tenant-stack/docs/FASE1_PROVISIONAMENTO_TENANT.md
"@

Step "Fase 2 - GPU / NVDEC"
Write-Host @"
  Host com GPU: VIDEO_ACCELERATION=auto + benchmark phase-4.1-verify.sh
  Doc: docs/FASE_4_1_FECHAMENTO.md - foxpro atual = CPU (easypanel.env.fase-c.vps.example)
"@

Step "Fase 5 - capacidade (todas câmeras, rampa)"
Write-Host "  CT111: F5_STAGES=`"5 10 20 30`" bash scripts/phase-f5-ramp-run.sh  (doc FASE_5.md)"

Write-Host ""
if ($fail -eq 0) {
    Write-Host "RESULT: implantar-fases OK (revise WARN D2/redis e env EasyPanel - docs/IMPLANTAR_FASES_0_4.md)" -ForegroundColor Green
    exit 0
}
Write-Host "RESULT: implantar-fases FAIL - corrija itens acima" -ForegroundColor Red
exit 1
