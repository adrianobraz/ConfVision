# Fase 0 - checklist automatico (producao). Nao substitui login no painel.
param(
    [string]$VisionBase = "https://vision.confmonit2.com.br",
    [string]$CamBase = "https://cam.rvasecurity.online",
    [string]$RustB = "https://foxpro-rust-pilot-b.rkr351.easypanel.host",
    [string]$RustA = "https://foxpro-rust-pilot.rkr351.easypanel.host",
    [string]$WorkerKey = "",
    [string]$IdFranqueado = "2025110408303794600766334",
    [int]$PilotCameraId = 15,
    [int]$MinPilotBCamerasAtivas = 1
)

function Pass($msg) { Write-Host "[OK] $msg" -ForegroundColor Green }
function Fail($msg) { Write-Host "[FALTA] $msg" -ForegroundColor Red }
function Warn($msg) { Write-Host "[?] $msg" -ForegroundColor Yellow }

$script:fail = 0

# 1 ConfVision + recursos
try {
    $h = Invoke-RestMethod -Uri "$VisionBase/vis_health" -TimeoutSec 20
    if ($h.postgres -eq "ok") { Pass "vis_health postgres ok" } else { Fail "vis_health postgres=$($h.postgres)"; $script:fail++ }
} catch { Fail "vis_health inacessivel: $_"; $script:fail++ }

try {
    $r = Invoke-WebRequest -Uri "$CamBase/recursos/modulos/confvision/dashboard.js" -UseBasicParsing -TimeoutSec 20
    if ($r.StatusCode -eq 200) { Pass "dashboard.js no whitelabel (HTTP 200)" } else { Fail "dashboard.js HTTP $($r.StatusCode)"; $script:fail++ }
} catch { Fail "dashboard.js: $_"; $script:fail++ }

if ($WorkerKey) {
    try {
        $hdr = @{ Authorization = "Bearer $WorkerKey" }
        $c = Invoke-RestMethod -Uri "$VisionBase/vis_camera_by_franqueado?id_franqueado=$([uri]::EscapeDataString($IdFranqueado))" -Headers $hdr -TimeoutSec 30
        $n = @($c.dados).Count
        if ($n -gt 0) { Pass "Postgres lista $n cameras (worker API)" } else { Fail "Postgres lista vazia"; $script:fail++ }
    } catch { Fail "vis_camera_by_franqueado: $_"; $script:fail++ }
} else {
    Warn "WorkerKey omitido - pulei lista Postgres"
}

function Test-RustHealth($label, $base) {
    try {
        $j = Invoke-RestMethod -Uri "$base/health" -TimeoutSec 20
        if ($null -ne $j.PSObject.Properties['cameras_pending_admission']) {
            Pass "Rust $label build admission (pending=$($j.cameras_pending_admission))"
        } else {
            Warn "Rust $label health sem cameras_pending_admission (imagem antiga?)"
        }
        if ([int64]$j.events_published -gt 0) {
            Pass "Rust $label events_published=$($j.events_published)"
        } else {
            Fail "Rust $label events_published=0"; $script:fail++
        }
        if ([int64]$j.cameras_online -ge 1) {
            Pass "Rust $label cameras_online=$($j.cameras_online)"
        } else {
            Fail "Rust $label cameras_online=$($j.cameras_online) (meta Fase0: >=1)"; $script:fail++
        }
    } catch {
        Fail "Rust $label /health: $_"; $script:fail++
    }
}

Test-RustHealth "B" $RustB
Test-RustHealth "A" $RustA

# 3 Postgres piloto (requer POSTGRES_URL no ambiente)
if ($env:POSTGRES_URL) {
    $auditDir = Join-Path $PSScriptRoot "pg-audit"
    Push-Location $auditDir
    $out = go run . 2>&1 | Out-String
    Pop-Location
    if ($out -match "sem linhas" -or $out -match "\(sem linhas\)") {
        Fail "vis_evento camera $PilotCameraId sem eventos recentes"
        $script:fail++
    } elseif ($out -match "Ultimos eventos") {
        Warn "Conferir manualmente bloco eventos camera $PilotCameraId no pg-audit"
    }
    if ($out -match "id\s+15\s+ESCRITORIO\s+offline\s+<nil>") {
        Warn "vis_camera.status=offline e ultimo_stream_ok_em null (campo status e legado; ver vis_worker ping)"
    }
    if ($out -match "rust-processor-pilot-b-02\s+(\d+)\s+(\d{4}-\d{2}-\d{2})") {
        $camAtivas = [int]$Matches[1]
        if ($camAtivas -ge $MinPilotBCamerasAtivas) {
            Pass "vis_worker pilot-b cameras_ativas=$camAtivas (min=$MinPilotBCamerasAtivas)"
        } else {
            Fail "vis_worker pilot-b cameras_ativas=$camAtivas (min=$MinPilotBCamerasAtivas)"
            $script:fail++
        }
    } else {
        Fail "vis_worker rust-processor-pilot-b-02 sem ping recente no pg-audit"
        $script:fail++
    }
} else {
    Warn "POSTGRES_URL nao definido - rode: `$env:POSTGRES_URL='...'; .\fase0-verify.ps1 -WorkerKey '...'"
}

Write-Host ""
if ($script:fail -eq 0) { Write-Host "FASE 0: todos os checks automaticos passaram." -ForegroundColor Green }
else { Write-Host "FASE 0: $($script:fail) item(ns) ainda pendente(s)." -ForegroundColor Red }
exit $script:fail
