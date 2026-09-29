# Diagnostico Fase 0 — produção (ConfVision + Rust A/B + Postgres + sidecar)
param(
    [string]$VisionBase = "https://vision.confmonit2.com.br",
    [string]$RustA = "https://foxpro-rust-pilot.rkr351.easypanel.host",
    [string]$RustB = "https://foxpro-rust-pilot-b.rkr351.easypanel.host",
    [string]$YoloBase = "https://foxpro-rust-yolo-sidecar.rkr351.easypanel.host",
    [string]$WorkerKey = "",
    [string]$IdFranqueado = "2025110408303794600766334"
)

$ErrorActionPreference = "Continue"
$issues = [System.Collections.Generic.List[string]]::new()

function Section($t) { Write-Host ""; Write-Host "=== $t ===" -ForegroundColor Cyan }
function Ok($m) { Write-Host "[OK] $m" -ForegroundColor Green }
function Bad($m) { Write-Host "[PROBLEMA] $m" -ForegroundColor Red; $issues.Add($m) | Out-Null }
function Info($m) { Write-Host "[i] $m" -ForegroundColor Yellow }

Section "ConfVision"
try {
    $h = Invoke-RestMethod "$VisionBase/vis_health" -TimeoutSec 20
    if ($h.postgres -eq "ok") { Ok "vis_health postgres ok" } else { Bad "vis_health postgres=$($h.postgres)" }
} catch { Bad "vis_health: $_" }

if ($WorkerKey) {
    try {
        $hdr = @{ Authorization = "Bearer $WorkerKey" }
        $c = Invoke-RestMethod "$VisionBase/vis_camera_by_franqueado?id_franqueado=$([uri]::EscapeDataString($IdFranqueado))" -Headers $hdr -TimeoutSec 30
        $n = @($c.dados).Count
        if ($n -gt 0) { Ok "lista franqueado: $n cameras" } else { Bad "lista franqueado vazia" }
    } catch { Bad "vis_camera_by_franqueado: $_" }
} else {
    Info "WorkerKey omitido - pulei lista franqueado"
}

Section "Rust processors"
foreach ($pair in @(@{L='A';U=$RustA}, @{L='B';U=$RustB})) {
    $label = $pair.L
    $base = $pair.U
    try {
        $health = Invoke-RestMethod "$base/health" -TimeoutSec 25
        Info "$label health: online=$($health.cameras_online) total=$($health.cameras_total) pending=$($health.cameras_pending_admission) events=$($health.events_published) capacity=$($health.capacity_state) advisory=$($health.load_advisory) fps=$($health.fps_total) frames_rx=$($health.frames_received)"
        if ($health.cameras_online -lt 1) {
            Bad "Rust $label cameras_online=0 (analitico parado ou fila/shedding)"
        } else {
            Ok "Rust $label cameras_online=$($health.cameras_online)"
        }
        if ($health.events_published -eq 0 -and $health.cameras_online -eq 0) {
            Info "Rust $label events_published=0 coerente com zero workers online"
        }
        try {
            $cap = Invoke-RestMethod "$base/capacity-report" -TimeoutSec 25
            $mem = $cap.capacity.memory.percent
            $cpu = $cap.capacity.cpu.percent
            $allow = $cap.load.allow_new_camera
            Info "$label capacity: cpu=$([math]::Round($cpu,1))% mem=$([math]::Round($mem,1))% allow_new_camera=$allow admission=$($cap.load.admission_active)"
            if (-not $allow) { Bad "Rust $label allow_new_camera=false (admission bloqueando novas sessoes)" }
        } catch { Info "$label capacity-report: $_" }
    } catch { Bad "Rust $label /health inacessivel: $_" }
}

Section "YOLO sidecar"
try {
    $y = Invoke-RestMethod "$YoloBase/health" -TimeoutSec 15
    Ok "yolo sidecar status=$($y.status) busy=$($y.busy)"
} catch { Bad "yolo sidecar: $_" }

Section "Pipeline eventos (Rust B / camera 15)"
try {
    $m = Invoke-RestMethod "$RustB/metrics" -TimeoutSec 25
    $gm = $m.metrics
    $c15 = $m.cameras | Where-Object { [int64]$_.camera_id -eq 15 } | Select-Object -First 1
    if ($c15) {
        Info "cam15 status=$($c15.status) fps=$([math]::Round($c15.fps,2)) decoded=$($c15.frames_decoded) motion_hits=$($c15.motion_detected) motion_score=$($c15.last_motion_score)"
        if ($c15.status -ne "online") { Bad "Camera 15 nao online no Rust B" }
        if ([int64]$c15.frames_decoded -gt 5 -and [int64]$c15.motion_detected -eq 0) {
            Info "YOLO pode rodar, mas motion gate=0 (cena estatica ou decode parcial)"
        }
    } else {
        Bad "Camera 15 ausente nas metricas do Rust B"
    }
    if ($gm.decode_errors -gt 0) {
        Info "decode_errors global=$($gm.decode_errors) (ver logs h264 NAL no container)"
    }
    Info "global: decoded=$($gm.frames_decoded) motion_detected=$($gm.motion_detected) enqueued=$($gm.frames_enqueued)"
    $hb = Invoke-RestMethod "$RustB/health" -TimeoutSec 20
    if ($hb.events_published -eq 0 -and $hb.cameras_online -ge 1) {
        Bad "Rust B online mas events_published=0 (YOLO sem match pessoa/area ou fila)"
    }
} catch { Bad "metrics pipeline: $_" }

Section "Postgres (pg-audit)"
if ($env:POSTGRES_URL) {
    $auditDir = Join-Path $PSScriptRoot "pg-audit"
    Push-Location $auditDir
    $out = go run . 2>&1 | Out-String
    Pop-Location
    if ($out -match "id\s+15\s+ESCRITORIO\s+offline\s+<nil>") {
        Info "vis_camera.status offline + ultimo_stream_ok_em null (status cadastro nao vem do Rust; stream_ok no ping)"
    }
    if ($out -match "rust-processor-pilot-b-02\s+(\d+)") {
        $n = [int]$Matches[1]
        if ($n -ge 1) { Ok "vis_worker pilot-b cameras_ativas=$n" } else { Bad "vis_worker pilot-b cameras_ativas=0" }
    } else {
        Bad "vis_worker pilot-b sem linha no pg-audit"
    }
    if ($out -match "Ultimos eventos c[^\r\n]+\r?\n\(sem linhas\)") { Bad "Camera 15 sem eventos em vis_evento" }
    if ($out -match "Sync Rust pilot-b[\s\S]*?\n3\t") { Ok "Camera 3 elegivel no sync pilot-b" }
    if ($out -match "id\s+15[\s\S]*?rust-processor-pilot-b-02" -and $out -notmatch "id\s+15[\s\S]*?analitico_pausado\s+true") {
        Ok "Camera 15 no worker B e analitico nao pausado (conferir bloco piloto)"
    } elseif ($out -match "id\s+15[\s\S]*?analitico_pausado\s+true") {
        Bad "Camera 15 ainda analitico_pausado=true (rodar fase0-apply-ops)"
    }
    $lines = ($out -split "`n" | Select-Object -First 35) -join "`n"
    Write-Host $lines
} else {
    Info "POSTGRES_URL nao definido - defina e rode de novo para audit completo"
}

Section "Hipoteses (causa raiz provavel)"
Info "1) LOAD_SHEDDING_ENABLED=1 nos logs derruba workers (15/21/20/22) com RAM~51% ou capacity critical"
Info "2) IDs em load_shed_ids nao reabrem no sync ate cooldown (ver logs 'elegivel para reabrir')"
Info "3) RTMP ok no MediaMTX != cameras_online Rust (RTSP decode no processor)"
Info "4) vis_camera.status offline = sem ping do worker, nao reflete RTMP guard"

Section "Resultado"
if ($issues.Count -eq 0) {
    Write-Host "Nenhum problema automatico detectado." -ForegroundColor Green
    exit 0
}
Write-Host "$($issues.Count) problema(s) detectado(s):" -ForegroundColor Red
$issues | ForEach-Object { Write-Host " - $_" }
exit [Math]::Min(99, $issues.Count)
