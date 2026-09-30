param(
    [Parameter(Mandatory = $true)]
    [string]$CsvPath,
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"

$base = if ($env:CONFVISION_API_URL) { $env:CONFVISION_API_URL } else { "https://vision.confmonit2.com.br" }
$base = $base.TrimEnd("/")
$key = $env:VIS_WORKER_API_KEY
if (-not $key) {
    throw "Defina VIS_WORKER_API_KEY (mesma chave dos processors Rust / visapi)."
}
if (-not (Test-Path -LiteralPath $CsvPath)) {
    throw "CSV nao encontrado: $CsvPath"
}

$headers = @{
    "X-Vis-Worker-Key" = $key
    "Content-Type"     = "application/json"
}

$rows = Import-Csv -LiteralPath $CsvPath
if ($rows.Count -eq 0) {
    Write-Host "CSV vazio."
    exit 0
}

foreach ($row in $rows) {
    $rawId = if ($row.id) { $row.id } else { $row.vis_camera_id }
    $id = $rawId.ToString().Trim()
    if (-not $id) { continue }

    $rtsp = [string]$row.rtsp_url_sec
    $rtsp = $rtsp.Trim()
    if (-not $rtsp) {
        Write-Warning "camera id=$id sem rtsp_url_sec - ignorando"
        continue
    }
    $lower = $rtsp.ToLowerInvariant()
    if (-not ($lower.StartsWith("rtsp://") -or $lower.StartsWith("rtsps://"))) {
        Write-Warning "camera id=$id URL invalida (precisa rtsp:// ou rtsps://) — ignorando"
        continue
    }
    if ($lower.Contains("/live/")) {
        Write-Warning "camera id=$id URL com /live/ (legado) — ignorando"
        continue
    }

    $body = @{ rtsp_url_sec = $rtsp }
    $despausar = [string]$row.despausar_analitico
    $despausar = $despausar.Trim().ToLowerInvariant()
    if ($despausar -in @("1", "true", "yes", "sim")) {
        $body.analitico_pausado = $false
    }

    $uri = "$base/vis_camera/$id"
    if ($DryRun) {
        Write-Host "[dry-run] PUT $uri rtsp_url_sec=<set> analitico_pausado=$($body.analitico_pausado)"
        continue
    }

    $json = $body | ConvertTo-Json -Compress
    try {
        $resp = Invoke-RestMethod -Method Put -Uri $uri -Headers $headers -Body $json
        Write-Host "OK camera id=$id nome=$($resp.nome)"
    }
    catch {
        Write-Error "Falha camera id=$id : $($_.Exception.Message)"
    }
}
