# Diagnóstico ponta a ponta — uma câmera (Control Plane + cadastro; não testa RTMP na VPS)
param(
    [Parameter(Mandatory = $true)]
    [int]$CameraId,
    [string]$ApiBase = $(if ($env:CONFVISION_API_URL) { $env:CONFVISION_API_URL } else { "https://vision.confmonit2.com.br" }).TrimEnd("/"),
    [string]$WorkerId = $(if ($env:WORKER_ID) { $env:WORKER_ID } else { "rust-processor-pilot-b-04" }),
    [string]$VisWorkerKey = $env:VIS_WORKER_API_KEY,
    [string]$RustHealthUrl = $env:RUST_HEALTH_URL
)

$ErrorActionPreference = "Stop"
if (-not $VisWorkerKey) {
    throw "Defina VIS_WORKER_API_KEY (mesma chave do EasyPanel confvision / Rust)."
}

$headers = @{ "X-Vis-Worker-Key" = $VisWorkerKey }

Write-Host "=== ConfVision diagnostico camera id=$CameraId ==="
Write-Host "API: $ApiBase | worker esperado: $WorkerId"

# 1 Go health
try {
    $h = Invoke-RestMethod -Uri "$ApiBase/vis_health" -TimeoutSec 15
    Write-Host "[OK] vis_health status=$($h.status) postgres=$($h.postgres)"
} catch {
    Write-Host "[FALHA] vis_health: $($_.Exception.Message)"
    exit 1
}

# 2 RTMP auth (Guard usa este endpoint)
try {
    $auth = Invoke-RestMethod -Uri "$ApiBase/vis_camera/rtmp_auth/$CameraId" -Headers $headers -TimeoutSec 20
    $okAuth = ($auth.ativo -eq $true -or [string]$auth.plano -eq "online") -and ($auth.bloqueado -ne $true)
    $motivo = $auth.stream_motivo_pausa
    if ($motivo -like "sistema_stream*") { $okAuth = $false }
    if ($okAuth) {
        Write-Host "[OK] rtmp_auth ativo=$($auth.ativo) bloqueado=$($auth.bloqueado) plano=$($auth.plano)"
    } else {
        Write-Host "[FALHA] rtmp_auth bloqueia publish: ativo=$($auth.ativo) bloqueado=$($auth.bloqueado) motivo=$motivo"
    }
    if ($auth.licenca_valido_ate) {
        $exp = [datetime]::Parse($auth.licenca_valido_ate)
        if ($exp -lt [datetime]::UtcNow) {
            Write-Host "[AVISO] licenca_valido_ate=$($auth.licenca_valido_ate) (expirada - renovar no painel)"
        }
    }
} catch {
    Write-Host "[FALHA] rtmp_auth: $($_.Exception.Message)"
}

# 3 Sync Rust
try {
    $q = [uri]::EscapeDataString($WorkerId)
    $sync = Invoke-RestMethod -Uri "$ApiBase/vis_camera_sync_ativas?worker_id=$q" -Headers $headers -TimeoutSec 30
    $cam = @($sync.cameras | Where-Object { [int]$_.id -eq $CameraId })
    if ($cam.Count -eq 0) {
        Write-Host "[FALHA] sync_ativas: camera $CameraId NAO lista para worker_id=$WorkerId"
        Write-Host "        Verifique vis_camera.worker_id, ativo, deteccao_humano, analitico_pausado"
    } else {
        $c = $cam[0]
        Write-Host "[OK] sync_ativas: worker_id=$($c.worker_id) rtsp_url_sec=$($c.rtsp_url_sec)"
        Write-Host "     mediamtx_rtsp_base=$($c.mediamtx_rtsp_base)"
        if ([string]$c.rtsp_url_sec -match "srv1\.dnsid\.com\.br") {
            Write-Host "[AVISO] rtsp_url_sec usa srv1:8554 (DNS ingest) — Rust no Docker deve usar foxpro_rust-mediamtx:8554 ou rtsp_url_sec vazio + nó"
        }
        if ([string]$c.rtsp_url_sec -match "foxpro_confvision") {
            Write-Host "[AVISO] rtsp_url_sec host legado foxpro_confvision — use foxpro_rust-mediamtx ou NULL"
        }
        if (-not [string]$c.rtsp_url_sec -and [string]$c.mediamtx_rtsp_base -match "foxpro_rust-mediamtx") {
            Write-Host "[OK] RTSP via nó MTX (rtsp_url_sec vazio + mediamtx_rtsp_base interno)"
        }
        if (-not [string]$c.rtsp_url_sec -and -not [string]$c.mediamtx_rtsp_base) {
            Write-Host "[AVISO] sem rtsp_url_sec e sem mediamtx_rtsp_base - Rust nao monta URL"
        }
    }
} catch {
    Write-Host "[FALHA] sync_ativas: $($_.Exception.Message)"
}

# 4 Rust health (opcional)
if ($RustHealthUrl) {
    try {
        $rh = Invoke-RestMethod -Uri $RustHealthUrl -TimeoutSec 15
        Write-Host "[OK] Rust health cameras_online=$($rh.cameras_online) fps_total=$($rh.fps_total)"
    } catch {
        Write-Host "[FALHA] Rust $RustHealthUrl - servico parado ou 502/503 (EasyPanel Start rust-pilot com WORKER_ID=$WorkerId)"
    }
} else {
    Write-Host "[SKIP] Rust health (defina RUST_HEALTH_URL, ex. https://foxpro-rust-pilot.../health)"
}

Write-Host ""
Write-Host "Proximos passos na VPS foxpro:"
Write-Host "  1) confvision (MediaMTX+Guard) Running"
Write-Host "  2) NVR publicando RTMP cam/HASH (hash = rtmp_token camera $CameraId)"
Write-Host "  3) rust-pilot Running WORKER_ID=$WorkerId MEDIAMTX_RTSP_BASE=rtsp://foxpro_rust-mediamtx:8554"
Write-Host "  4) Guard: unban IP se ip_banido nos logs"
