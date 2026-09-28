# Gera CSV local (gitignored) com RTSP MediaMTX publico: rtsp://host:8554/cam/{hash12}
param(
    [int[]]$CameraIds = @(),
    [switch]$AllCameras,
    [string]$ApiBase = $(if ($env:CONFVISION_API_URL) { $env:CONFVISION_API_URL } else { "https://vision.confmonit2.com.br" }).TrimEnd("/"),
    [string]$VisWorkerKey = $env:VIS_WORKER_API_KEY,
    [string]$ConfvisionEnv = "C:\sistemaconfmonit\core4\home\confmonit\v4.0\confvision\.env",
    [string]$OutCsv = (Join-Path $PSScriptRoot "..\sql\opcao_a_mediamtx.local.csv"),
    [switch]$DespausarAnalitico
)

$ErrorActionPreference = "Stop"
if (-not (Test-Path -LiteralPath $ConfvisionEnv)) {
    throw "Env ConfVision nao encontrado: $ConfvisionEnv"
}

function Get-EnvVal([string]$name) {
    $m = Select-String -Path $ConfvisionEnv -Pattern "^$name=(.+)$" | Select-Object -First 1
    if (-not $m) { return "" }
    return $m.Matches.Groups[1].Value.Trim()
}

$secret = Get-EnvVal "RTMP_PUBLISH_SECRET"
if (-not $secret) { throw "RTMP_PUBLISH_SECRET vazio em .env" }

$rtspBase = Get-EnvVal "MEDIAMTX_RTSP_BASE"
if (-not $rtspBase) { $rtspBase = "rtsp://srv1.dnsid.com.br:8554" }
$rtspBase = $rtspBase.Trim().TrimEnd("/")

if ($AllCameras) {
    if (-not $VisWorkerKey) {
        $VisWorkerKey = Get-EnvVal "VIS_WORKER_API_KEY"
    }
    if (-not $VisWorkerKey) { throw "VIS_WORKER_API_KEY necessario para -AllCameras" }
    $headers = @{ "X-Vis-Worker-Key" = $VisWorkerKey }
    $list = Invoke-RestMethod -Uri "$ApiBase/vis_camera?all=1" -Headers $headers -TimeoutSec 120
    if ($list -is [array]) {
        $CameraIds = @($list | ForEach-Object { [int]$_.id } | Where-Object { $_ -gt 0 } | Sort-Object -Unique)
    } elseif ($list.dados) {
        $CameraIds = @($list.dados | ForEach-Object { [int]$_.id } | Where-Object { $_ -gt 0 } | Sort-Object -Unique)
    } else {
        throw "Resposta inesperada de GET /vis_camera?all=1"
    }
    Write-Host "AllCameras: $($CameraIds.Count) ids da tabela vis_camera"
} elseif ($CameraIds.Count -eq 0) {
    $CameraIds = @(5, 18, 19, 21, 22, 26, 27)
}

$despausarCol = if ($DespausarAnalitico) { "true" } else { "false" }

$confRoot = "C:\sistemaconfmonit\core4\home\confmonit\v4.0\confvision"
$env:RTMP_PUBLISH_SECRET = $secret
Push-Location $confRoot
try {
    $lines = @("id,rtsp_url_sec,despausar_analitico")
    foreach ($id in $CameraIds) {
        $out = & go run ./scripts/probe_stream_path $id 2>&1
        $path = ($out | Select-Object -Last 1).ToString().Trim()
        if (-not $path -or $path -notmatch "^cam/") {
            Write-Warning "camera id=$id path invalido: $path"
            continue
        }
        $url = "$rtspBase/$path"
        $lines += "$id,$url,$despausarCol"
        Write-Host "  id=$id -> $url"
    }
}
finally {
    Pop-Location
}

if ($lines.Count -le 1) { throw "Nenhuma linha gerada" }
$lines | Set-Content -LiteralPath $OutCsv -Encoding utf8
Write-Host "`nCSV: $OutCsv ($($lines.Count - 1) cameras)"
