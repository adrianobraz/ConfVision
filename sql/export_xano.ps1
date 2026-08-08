# Exporta vis_mediamtx_node, vis_camera e vis_camera_area do Xano para SQL Postgres.
# Uso:
#   .\export_xano.ps1
#   .\export_xano.ps1 -XanoBaseUrl "https://....xano.io/api:AC7rgWwW" -OutFile "003_import_xano.sql"
param(
    [string]$XanoBaseUrl = "https://xpcy-oyme-lno7.b2.xano.io/api:AC7rgWwW",
    [string]$OutFile = (Join-Path $PSScriptRoot "003_import_xano.sql")
)

$ErrorActionPreference = "Stop"
$base = $XanoBaseUrl.TrimEnd("/")

function Sql-Str([object]$v) {
    if ($null -eq $v) { return "NULL" }
    if ($v -is [bool]) { return $(if ($v) { "TRUE" } else { "FALSE" }) }
    if ($v -is [int] -or $v -is [long] -or $v -is [double] -or $v -is [decimal]) { return [string]$v }
    $s = [string]$v
    if ($s -eq "") { return "NULL" }
    return "'" + ($s -replace "'", "''") + "'"
}

function Sql-Ts([object]$v) {
    if ($null -eq $v -or [string]$v -eq "") { return "NULL" }
    if ($v -is [int64] -or $v -is [int] -or $v -is [double] -or ($v -is [string] -and [string]$v -match '^\d+$')) {
        $ms = [int64]$v
        if ($ms -lt 100000000000) { $ms = $ms * 1000 }
        $dt = [DateTimeOffset]::FromUnixTimeMilliseconds($ms).UtcDateTime
        return "'" + $dt.ToString("yyyy-MM-dd HH:mm:ss+00") + "'"
    }
    return "'" + ([datetime]$v).ToUniversalTime().ToString("yyyy-MM-dd HH:mm:ss+00") + "'"
}

function Fetch-Json($path) {
    $uri = "$base/$path"
    Write-Host "GET $uri"
    return Invoke-RestMethod -Uri $uri -Method Get -TimeoutSec 120
}

function Normalize-List($resp) {
    if ($null -eq $resp) { return @() }
    if ($resp.PSObject.Properties.Name -contains "dados") {
        return @($resp.dados)
    }
    if ($resp -is [System.Array]) { return @($resp) }
    if ($resp.id) { return @($resp) }
    return @()
}

Write-Host "Exportando do Xano confVision..." -ForegroundColor Cyan

$nodes = Normalize-List (Fetch-Json "vis_mediamtx_node")
$cameras = Normalize-List (Fetch-Json "vis_camera")
$areas = Normalize-List (Fetch-Json "vis_camera_area_query_ativas")

$franqueados = New-Object System.Collections.Generic.HashSet[string]
foreach ($c in $cameras) {
    if ($c.id_franqueado) { [void]$franqueados.Add([string]$c.id_franqueado) }
}
$licencas = @()
foreach ($fra in $franqueados) {
    $licResp = Fetch-Json ("vis_licenca_by_franqueado?id_franqueado=" + [uri]::EscapeDataString($fra))
    if ($licResp.dados) { $licencas += @($licResp.dados) }
}

Write-Host "  nos: $($nodes.Count)  cameras: $($cameras.Count)  areas: $($areas.Count)  licencas: $($licencas.Count)"

$sb = New-Object System.Text.StringBuilder
[void]$sb.AppendLine("-- Gerado por export_xano.ps1 em $(Get-Date -Format o)")
[void]$sb.AppendLine("BEGIN;")
[void]$sb.AppendLine("")

foreach ($n in $nodes) {
    [void]$sb.AppendLine(@"
INSERT INTO vis_mediamtx_node (
    id, created_at, nome, rtmp_public, hls_public, rtsp_internal,
    max_cameras, ordem, status, observacao
) VALUES (
    $($n.id), $(Sql-Ts $n.created_at), $(Sql-Str $n.nome), $(Sql-Str $n.rtmp_public),
    $(Sql-Str $n.hls_public), $(Sql-Str $n.rtsp_internal), $(Sql-Str $(if ($n.max_cameras) { $n.max_cameras } else { 200 })),
    $(Sql-Str $(if ($n.ordem) { $n.ordem } else { 1 })), $(Sql-Str $(if ($n.status) { $n.status } else { "ativo" })),
    $(Sql-Str $n.observacao)
) ON CONFLICT (id) DO UPDATE SET
    nome = EXCLUDED.nome,
    rtmp_public = EXCLUDED.rtmp_public,
    hls_public = EXCLUDED.hls_public,
    rtsp_internal = EXCLUDED.rtsp_internal,
    max_cameras = EXCLUDED.max_cameras,
    ordem = EXCLUDED.ordem,
    status = EXCLUDED.status,
    observacao = EXCLUDED.observacao;
"@)
}

[void]$sb.AppendLine("")
foreach ($c in $cameras) {
    if (-not $c.id) { continue }
    [void]$sb.AppendLine(@"
INSERT INTO vis_camera (
    id, created_at, ativo, bloqueado, nome, id_franqueado, id_cliente, id_dispositivo,
    conta, particao, canal, setor, protocolo, rtsp_url_sec, onvif_host, onvif_porta,
    onvif_usuario, onvif_senha, confianca_min, cooldown_seg, somente_armado,
    deteccao_humano, deteccao_veiculo, status, ultimo_evento_em, worker_id,
    ultimo_ping_em, zonauser, captura_sensor, captura_analitico, analitico_pausado,
    id_setor, snapshot_url, vis_licenca_id, plano, ativado_em, evento_grava_foto,
    evento_grava_video, vis_licenca_gravacao_id, grava_continua, retencao_dias,
    gravacao_ativada_em, gravacao_status, grava_movimento, grava_timelapse,
    gravacao_flush_pedido, modo_deteccao, vis_mediamtx_node_id
) VALUES (
    $($c.id), $(Sql-Ts $c.created_at), $(Sql-Str $c.ativo), $(Sql-Str $c.bloqueado),
    $(Sql-Str $c.nome), $(Sql-Str $c.id_franqueado), $(Sql-Str $c.id_cliente), $(Sql-Str $c.id_dispositivo),
    $(Sql-Str $c.conta), $(Sql-Str $c.particao), $(Sql-Str $c.canal), $(Sql-Str $c.setor),
    $(Sql-Str $c.protocolo), $(Sql-Str $c.rtsp_url_sec), $(Sql-Str $c.onvif_host), $(Sql-Str $c.onvif_porta),
    $(Sql-Str $c.onvif_usuario), $(Sql-Str $c.onvif_senha), $(Sql-Str $c.confianca_min), $(Sql-Str $c.cooldown_seg),
    $(Sql-Str $c.somente_armado), $(Sql-Str $c.deteccao_humano), $(Sql-Str $c.deteccao_veiculo),
    $(Sql-Str $c.status), $(Sql-Ts $c.ultimo_evento_em), $(Sql-Str $c.worker_id),
    $(Sql-Ts $c.ultimo_ping_em), $(Sql-Str $c.zonauser), $(Sql-Str $c.captura_sensor),
    $(Sql-Str $c.captura_analitico), $(Sql-Str $c.analitico_pausado), $(Sql-Str $c.id_setor),
    $(Sql-Str $c.snapshot_url), $(Sql-Str $c.vis_licenca_id), $(Sql-Str $c.plano), $(Sql-Ts $c.ativado_em),
    $(Sql-Str $c.evento_grava_foto), $(Sql-Str $c.evento_grava_video), $(Sql-Str $c.vis_licenca_gravacao_id),
    $(Sql-Str $c.grava_continua), $(Sql-Str $c.retencao_dias), $(Sql-Ts $c.gravacao_ativada_em),
    $(Sql-Str $c.gravacao_status), $(Sql-Str $c.grava_movimento), $(Sql-Str $c.grava_timelapse),
    $(Sql-Str $c.gravacao_flush_pedido), $(Sql-Str $c.modo_deteccao), $(Sql-Str $c.vis_mediamtx_node_id)
) ON CONFLICT (id) DO NOTHING;
"@)
}

[void]$sb.AppendLine("")
foreach ($a in $areas) {
    if (-not $a.id) { continue }
    [void]$sb.AppendLine(@"
INSERT INTO vis_camera_area (
    id, created_at, vis_camera_id, nome, ativo, poligono_json, cor
) VALUES (
    $($a.id), $(Sql-Ts $a.created_at), $(Sql-Str $a.vis_camera_id), $(Sql-Str $a.nome),
    $(Sql-Str $a.ativo), $(Sql-Str $a.poligono_json), $(Sql-Str $a.cor)
) ON CONFLICT (id) DO NOTHING;
"@)
}

[void]$sb.AppendLine("")
foreach ($l in $licencas) {
    if (-not $l.id) { continue }
    [void]$sb.AppendLine(@"
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES (
    $($l.id), $(Sql-Ts $l.created_at), $(Sql-Str $l.id_franqueado), $(Sql-Str $l.plano),
    $(Sql-Str $(if ($l.unidade) { $l.unidade } else { 'camera' })), $(Sql-Str $l.valor),
    $(Sql-Ts $l.pago_em), $(Sql-Ts $l.valido_ate), $(Sql-Str $(if ($l.status) { $l.status } else { 'disponivel' })),
    $(Sql-Str $l.id_dispositivo), $(Sql-Str $l.id_fatura), $(Sql-Str $l.id_pagamento),
    $(Sql-Str $l.observacao), $(Sql-Str $l.vis_camera_id)
) ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status, vis_camera_id = EXCLUDED.vis_camera_id, plano = EXCLUDED.plano;
"@)
}

[void]$sb.AppendLine("")
[void]$sb.AppendLine("SELECT setval(pg_get_serial_sequence('vis_mediamtx_node','id'), COALESCE((SELECT MAX(id) FROM vis_mediamtx_node), 1));")
[void]$sb.AppendLine("SELECT setval(pg_get_serial_sequence('vis_camera','id'), COALESCE((SELECT MAX(id) FROM vis_camera), 1));")
[void]$sb.AppendLine("SELECT setval(pg_get_serial_sequence('vis_camera_area','id'), COALESCE((SELECT MAX(id) FROM vis_camera_area), 1));")
[void]$sb.AppendLine("SELECT setval(pg_get_serial_sequence('vis_licenca','id'), COALESCE((SELECT MAX(id) FROM vis_licenca), 1));")
[void]$sb.AppendLine("SELECT vis_mediamtx_node_recalc_pontos(id) FROM vis_mediamtx_node;")
[void]$sb.AppendLine("COMMIT;")

$utf8NoBom = New-Object System.Text.UTF8Encoding $false
[System.IO.File]::WriteAllText($OutFile, $sb.ToString(), $utf8NoBom)
Write-Host "OK: $OutFile" -ForegroundColor Green
