# Opção A — estado atual sync/query (sem vazar rtsp_url_sec completa)
param(
    [string]$ApiBase = $(if ($env:CONFVISION_API_URL) { $env:CONFVISION_API_URL } else { "https://vision.confmonit2.com.br" }).TrimEnd("/"),
    [string]$VisWorkerKey = $env:VIS_WORKER_API_KEY
)

$ErrorActionPreference = "Stop"
if (-not $VisWorkerKey) {
    throw "Defina VIS_WORKER_API_KEY ou -VisWorkerKey"
}

$headers = @{ "X-Vis-Worker-Key" = $VisWorkerKey }

function Get-VisJson {
    param([string]$Path)
    return Invoke-RestMethod -Uri "$ApiBase$Path" -Headers $headers -TimeoutSec 30
}

function Show-Cameras {
    param([string]$Label, [object]$Data)
    $list = @($Data.cameras)
    if ($list.Count -eq 0 -and $Data.dados) { $list = @($Data.dados) }
    if ($list.Count -eq 0 -and $Data -is [array]) { $list = @($Data) }
    Write-Host "`n== $Label (count=$($list.Count)) =="
    foreach ($c in $list) {
        $rtsp = [string]$c.rtsp_url_sec
        $has = if ($rtsp.Trim().Length -gt 0) { "sim" } else { "NAO" }
        $preview = if ($rtsp.Length -gt 45) { $rtsp.Substring(0, 45) + "..." } elseif ($rtsp) { $rtsp } else { "-" }
        $onvif = [string]$c.onvif_host
        if ($onvif.Length -gt 30) { $onvif = $onvif.Substring(0, 30) + "..." }
        Write-Host ("  id={0} worker={1} rtsp_sec={2} pausado={3} nome={4}" -f `
            $c.id, ($(if ($null -ne $c.worker_id) { $c.worker_id } else { "?" })), $has, ($(if ($null -ne $c.analitico_pausado) { $c.analitico_pausado } else { $false })), ($(if ($null -ne $c.nome) { $c.nome } else { "" })))
        if ($has -eq "sim") { Write-Host "    preview: $preview" }
        elseif ($onvif) { Write-Host "    onvif_host: $onvif (preencher rtsp_url_sec Opção A)" }
    }
}

Write-Host "Opção A preflight — $ApiBase"

foreach ($wid in @("rust-processor-pilot-a-01", "rust-processor-pilot-b-02")) {
    $q = [uri]::EscapeDataString($wid)
    Show-Cameras "sync_ativas worker=$wid" (Get-VisJson "/vis_camera_sync_ativas?worker_id=$q")
    Show-Cameras "query_ativas worker=$wid" (Get-VisJson "/vis_camera_query_ativas?worker_id=$q")
}

Write-Host "`nProximo: editar sql/opcao_a_foxpro_6_cameras.csv com RTSP reais (acessivel da VPS foxpro)."
Write-Host "Aplicar: .\opcao-a-run.ps1 -VisWorkerKey `$env:VIS_WORKER_API_KEY -Apply"
