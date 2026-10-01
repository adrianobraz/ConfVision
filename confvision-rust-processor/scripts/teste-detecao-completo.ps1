# Teste detecção ponta a ponta — câmeras 32 e 6 (API + Rust health + áreas + fila)
param(
    [int[]]$CameraIds = @(32, 6),
    [string]$ApiBase = $(if ($env:CONFVISION_API_URL) { $env:CONFVISION_API_URL } else { "https://vision.confmonit2.com.br" }).TrimEnd("/"),
    [string]$VisWorkerKey = $env:VIS_WORKER_API_KEY,
    [hashtable]$WorkerHealth = @{
        "rust-processor-pilot-b-02" = "https://foxpro-rust-pilot-b.rkr351.easypanel.host/health"
        "rust-processor-pilot-b-04" = "https://foxpro-rust-pilot-b-04.rkr351.easypanel.host/health"
        "rust-processor-pilot-a-01" = "https://foxpro-rust-pilot.rkr351.easypanel.host/health"
    }
)

$ErrorActionPreference = "Continue"
if (-not $VisWorkerKey) { throw "Defina VIS_WORKER_API_KEY" }
$headers = @{ "X-Vis-Worker-Key" = $VisWorkerKey }

Write-Host "=== ConfVision teste detecao completo ==="
Write-Host "API: $ApiBase`n"

try {
    $h = Invoke-RestMethod -Uri "$ApiBase/vis_health" -TimeoutSec 15
    Write-Host "[OK] vis_health postgres=$($h.postgres)"
} catch {
    Write-Host "[FALHA] vis_health"; exit 1
}

$fail = $false
foreach ($cid in $CameraIds) {
    Write-Host "`n--- Camera $cid ---"
    $auth = Invoke-RestMethod -Uri "$ApiBase/vis_camera/rtmp_auth/$cid" -Headers $headers -TimeoutSec 20
    $wid = [string]$auth.worker_id
    if (-not $wid) {
        $q = [uri]::EscapeDataString("rust-processor-pilot-b-02")
        $s = Invoke-RestMethod -Uri "$ApiBase/vis_camera_sync_ativas?worker_id=$q" -Headers $headers
        foreach ($w in @("rust-processor-pilot-b-02", "rust-processor-pilot-b-04", "rust-processor-pilot-a-01")) {
            $q2 = [uri]::EscapeDataString($w)
            $s2 = Invoke-RestMethod -Uri "$ApiBase/vis_camera_sync_ativas?worker_id=$q2" -Headers $headers
            if (@($s2.cameras | Where-Object { $_.id -eq $cid }).Count -gt 0) { $wid = $w; break }
        }
    }
    Write-Host "worker_id=$wid plano=$($auth.plano) pausado=$($auth.analitico_pausado)"

    $areas = Invoke-RestMethod -Uri "$ApiBase/vis_camera_area_by_camera?vis_camera_id=$cid" -Headers $headers
    $ac = @($areas.dados); if ($null -eq $ac -and $areas -is [array]) { $ac = @($areas) }
    Write-Host "areas cadastradas=$($ac.Count) (Rust precisa receber no sync; ver deploy areas merge)"

    if ($wid -and $WorkerHealth.ContainsKey($wid)) {
        $url = $WorkerHealth[$wid]
        try {
            $rh = Invoke-RestMethod -Uri $url -TimeoutSec 20
            Write-Host "[Rust $wid] online=$($rh.cameras_online) fps=$([math]::Round($rh.fps_total,1))"
            Write-Host "  yolo_inf=$($rh.yolo_inferences) matches=$($rh.yolo_matches) published=$($rh.events_published) captured=$($rh.events_captured)"
            Write-Host "  queue_depth=$($rh.event_queue_depth) yolo_skipped_busy=$($rh.yolo_skipped_busy) cooldown_suppressed=$($rh.events_suppressed_cooldown)"
            if (($rh.yolo_inferences -gt 20) -and ($rh.yolo_matches -eq 0)) {
                Write-Host "[AVISO] YOLO roda mas zero match - area/motion gate/timeout sidecar ou bug areas no sync (corrigir deploy Rust)"
            }
            if (($rh.events_published -eq 0) -and ($rh.yolo_matches -gt 0)) {
                Write-Host "[AVISO] match sem publish - cooldown ou fila cheia"
            }
            $listed = @($rh.cameras_online) -ge 1
            $q = [uri]::EscapeDataString($wid)
            $sync = Invoke-RestMethod -Uri "$ApiBase/vis_camera_sync_ativas?worker_id=$q" -Headers $headers
            $inSync = @($sync.cameras | Where-Object { $_.id -eq $cid }).Count -gt 0
            if (-not $inSync) { Write-Host "[FALHA] camera nao esta no sync do worker $wid"; $fail = $true }
        } catch {
            Write-Host "[FALHA] Rust $url - processor OFF ou URL errada"; $fail = $true
        }
    } else {
        Write-Host "[FALHA] worker desconhecido ou sem health URL"; $fail = $true
    }
}

Write-Host "`n=== Checklist eventos no painel ==="
Write-Host "1) Pessoa DENTRO do poligono + movimento (se ANALYSIS_ONLY_ON_MOTION=1)"
Write-Host "2) Apos deploy Rust com merge de areas: yolo_matches e events_published sobem"
Write-Host "3) events_captured maior que 0 - fila consumida - vis_evento no Postgres"
if ($fail) { exit 1 }
Write-Host "`nRESULT: probes OK (ver avisos acima)"
