# Exporta cadastro grade CVG + janelas arme/desarme do Xano para Postgres (005_import_ops.sql)
param(
    [string]$CvgBaseUrl = "https://xpcy-oyme-lno7.b2.xano.io/api:KiUjyOQR",
    [string]$RoboBaseUrl = "https://xpcy-oyme-lno7.b2.xano.io/api:Maqjhbl9",
    [string]$ConfVisionBaseUrl = "https://xpcy-oyme-lno7.b2.xano.io/api:AC7rgWwW",
    [string]$OutFile = (Join-Path $PSScriptRoot "005_import_ops.sql")
)

$ErrorActionPreference = "Stop"
$cvg = $CvgBaseUrl.TrimEnd("/")
$robo = $RoboBaseUrl.TrimEnd("/")
$vis = $ConfVisionBaseUrl.TrimEnd("/")

function Sql-Str([object]$v) {
    if ($null -eq $v) { return "NULL" }
    if ($v -is [bool]) { return $(if ($v) { "TRUE" } else { "FALSE" }) }
    if ($v -is [int] -or $v -is [long] -or $v -is [double]) { return [string]$v }
    $s = [string]$v
    if ($s -eq "") { return "NULL" }
    return "'" + ($s -replace "'", "''") + "'"
}

function Sql-Ts([object]$v) {
    if ($null -eq $v -or [string]$v -eq "") { return "NOW()" }
    if ($v -is [int64] -or $v -is [int] -or ($v -is [string] -and [string]$v -match '^\d+$')) {
        $ms = [int64]$v
        if ($ms -lt 100000000000) { $ms = $ms * 1000 }
        $dt = [DateTimeOffset]::FromUnixTimeMilliseconds($ms).UtcDateTime
        return "'" + $dt.ToString("yyyy-MM-dd HH:mm:ss+00") + "'"
    }
    return "'" + ([datetime]$v).ToUniversalTime().ToString("yyyy-MM-dd HH:mm:ss+00") + "'"
}

function Sql-Array([object]$arr) {
    if ($null -eq $arr -or $arr.Count -eq 0) { return "ARRAY[]::TEXT[]" }
    $items = @($arr | ForEach-Object { Sql-Str ([string]$_) })
    return "ARRAY[" + ($items -join ",") + "]"
}

function Fetch-Json($uri) {
    Write-Host "GET $uri"
    try {
        return Invoke-RestMethod -Uri $uri -Method Get -TimeoutSec 120
    } catch {
        Write-Warning ("Falha GET " + $uri + ": " + $_.Exception.Message)
        return $null
    }
}

function Normalize-List($resp) {
    if ($null -eq $resp) { return @() }
    if ($resp.PSObject.Properties.Name -contains "dados") { return @($resp.dados) }
    if ($resp -is [System.Array]) { return @($resp) }
    if ($resp.id) { return @($resp) }
    return @()
}

Write-Host "Exportando ops (grade + janelas)..." -ForegroundColor Cyan

# --- Grade: endpoint dedicado (requer push 2406_cvg_export_all) ---
$exportGrade = Fetch-Json ($cvg + '/cvg_export_all')
if ($exportGrade) {
    $gradeConfig = Normalize-List $exportGrade.grade_config
    $gradeSlots = Normalize-List $exportGrade.grade_slots
    $gradeEscopo = Normalize-List $exportGrade.grade_escopo
}

# Fallback endpoints de tabela / listagem
if ($gradeConfig.Count -eq 0 -and $gradeSlots.Count -eq 0 -and $gradeEscopo.Count -eq 0) {
    $gradeConfig = Normalize-List (Fetch-Json ($cvg + '/vis_cliente_grade_config'))
    $gradeSlots = Normalize-List (Fetch-Json ($cvg + '/vis_cliente_grade_slot'))
    $gradeEscopo = Normalize-List (Fetch-Json ($cvg + '/vis_cliente_grade_escopo'))
}

if ($gradeSlots.Count -eq 0) {
    Write-Host "Fallback grade: cvg_grade_list por franqueado (vis_camera)..." -ForegroundColor Yellow
    $cameras = Normalize-List (Fetch-Json ($vis + '/vis_camera'))
    $franqueados = $cameras | ForEach-Object { $_.id_franqueado } | Where-Object { $_ } | Sort-Object -Unique
    $seenClient = @{}
    foreach ($fra in $franqueados) {
        $list = Fetch-Json ($cvg + '/cvg_grade_list?id_franqueado=' + [uri]::EscapeDataString($fra))
        foreach ($item in @($list.itens)) {
            $key = [string]$item.id_cliente + '|' + [string]$item.id_dispositivo
            if ($seenClient[$key]) { continue }
            $seenClient[$key] = $true
            $gcUri = $cvg + '/cvg_grade_by_cliente?id_cliente=' + [uri]::EscapeDataString($item.id_cliente) + '&id_dispositivo=' + [uri]::EscapeDataString([string]$item.id_dispositivo)
            $gc = Fetch-Json $gcUri
            if ($gc.slots) { $gradeSlots += @($gc.slots) }
        }
    }
}

# --- Janelas arme: endpoint dedicado (requer push 2210_ops_export_janelas) ---
$janelas = @()
$exportJan = Fetch-Json ($robo + '/ops_export_janelas')
if ($exportJan) {
    $janelas = Normalize-List $exportJan
}

if ($janelas.Count -eq 0) {
    $janelas = Normalize-List (Fetch-Json ($robo + '/WhatsappEventCadJanela'))
}
if ($janelas.Count -eq 0) {
    Write-Host "Fallback janelas: Queryall por dia 0-7..." -ForegroundColor Yellow
    $seenJan = @{}
    0..7 | ForEach-Object {
        $d = $_
        $r = Fetch-Json ($robo + '/whatsappeventcadjanela/Queryall/WhatsEventCadJan?DiaSemana=' + $d)
        if ($null -eq $r -or $null -eq $r.dados) { return }
        foreach ($x in @($r.dados.desarme) + @($r.dados.arme)) {
            if ($null -eq $x -or -not $x.id) { continue }
            $seenJan[[string]$x.id] = $x
        }
    }
    $janelas = @($seenJan.Values)
}
if ($janelas.Count -eq 0) {
    Write-Host "Fallback janelas: WhatsappEventCadJanela (variante lowercase)..." -ForegroundColor Yellow
    $janelas = Normalize-List (Fetch-Json ($robo + '/whatsappeventcadjanela'))
}

Write-Host "  grade_config: $($gradeConfig.Count)  slots: $($gradeSlots.Count)  escopo: $($gradeEscopo.Count)  janelas: $($janelas.Count)"

$sb = New-Object System.Text.StringBuilder
[void]$sb.AppendLine("-- Gerado por export_ops_xano.ps1 em $(Get-Date -Format o)")
[void]$sb.AppendLine("BEGIN;")
[void]$sb.AppendLine("")

foreach ($c in $gradeConfig) {
    if (-not $c.id) { continue }
    $line = 'INSERT INTO vis_cliente_grade_config (id, created_at, id_franqueado, id_cliente, grade_ativa) VALUES (' +
        $c.id + ', ' + (Sql-Ts $c.created_at) + ', ' + (Sql-Str $c.id_franqueado) + ', ' + (Sql-Str $c.id_cliente) + ', ' + (Sql-Str $c.grade_ativa) +
        ') ON CONFLICT (id_cliente, id_franqueado) DO UPDATE SET grade_ativa = EXCLUDED.grade_ativa;'
    [void]$sb.AppendLine($line)
}

foreach ($s in $gradeSlots) {
    if (-not $s.id -and -not $s.dia_semana) { continue }
    $disp = if ($s.id_dispositivo) { $s.id_dispositivo } else { '' }
    $ativo = if ($null -ne $s.ativo) { $s.ativo } else { $true }
    $idCol = if ($s.id) { 'id, ' } else { '' }
    $idIns = if ($s.id) { ($s.id.ToString() + ', ') } else { '' }
    $line = 'INSERT INTO vis_cliente_grade_slot (' + $idCol + 'created_at, id_franqueado, id_cliente, id_dispositivo, dia_semana, hora, acao, ativo) VALUES (' +
        $idIns + (Sql-Ts $s.created_at) + ', ' + (Sql-Str $s.id_franqueado) + ', ' + (Sql-Str $s.id_cliente) + ', ' + (Sql-Str $disp) + ', ' +
        (Sql-Str $s.dia_semana) + ', ' + (Sql-Str $s.hora) + ', ' + (Sql-Str $s.acao) + ', ' + (Sql-Str $ativo) +
        ') ON CONFLICT DO NOTHING;'
    [void]$sb.AppendLine($line)
}

foreach ($e in $gradeEscopo) {
    if (-not $e.id) { continue }
    $disp = if ($e.id_dispositivo) { $e.id_dispositivo } else { '' }
    $line = 'INSERT INTO vis_cliente_grade_escopo (id, created_at, id_franqueado, id_cliente, id_dispositivo, grade_ativa) VALUES (' +
        $e.id + ', ' + (Sql-Ts $e.created_at) + ', ' + (Sql-Str $e.id_franqueado) + ', ' + (Sql-Str $e.id_cliente) + ', ' + (Sql-Str $disp) + ', ' + (Sql-Str $e.grade_ativa) +
        ') ON CONFLICT (id_cliente, id_franqueado, id_dispositivo) DO UPDATE SET grade_ativa = EXCLUDED.grade_ativa;'
    [void]$sb.AppendLine($line)
}

foreach ($j in $janelas) {
    if (-not $j.id) { continue }
    $dias = $j.dias
    if ($null -eq $dias) { $dias = @() }
    $idCliente = if ($j.IdCliente) { $j.IdCliente } else { $j.id_cliente }
    $idFranqueado = if ($j.IdFranqueado) { $j.IdFranqueado } else { $j.id_franqueado }
    $nomeDisp = if ($j.NomeDispositivo) { $j.NomeDispositivo } else { $j.nome_dispositivo }
    $line = 'INSERT INTO ops_arme_janela (id, created_at, whatsappeventocad_id, dias, hora_inicio, hora_fim, id_cliente, id_franqueado, whatsapp, nome, id_dispositivo, nome_dispositivo) VALUES (' +
        $j.id + ', ' + (Sql-Ts $j.created_at) + ', ' + (Sql-Str $j.whatsappeventocad_id) + ', ' + (Sql-Array $dias) + ', ' +
        (Sql-Str $j.hora_inicio) + ', ' + (Sql-Str $j.hora_fim) + ', ' + (Sql-Str $idCliente) + ', ' + (Sql-Str $idFranqueado) + ', ' +
        (Sql-Str $j.whatsapp) + ', ' + (Sql-Str $j.nome) + ', ' + (Sql-Str $j.idDispositivo) + ', ' + (Sql-Str $nomeDisp) +
        ') ON CONFLICT (id) DO UPDATE SET dias = EXCLUDED.dias, hora_inicio = EXCLUDED.hora_inicio, hora_fim = EXCLUDED.hora_fim;'
    [void]$sb.AppendLine($line)
}

[void]$sb.AppendLine('')
[void]$sb.AppendLine('SELECT setval(pg_get_serial_sequence(''vis_cliente_grade_config'',''id''), COALESCE((SELECT MAX(id) FROM vis_cliente_grade_config), 1));')
[void]$sb.AppendLine('SELECT setval(pg_get_serial_sequence(''vis_cliente_grade_slot'',''id''), COALESCE((SELECT MAX(id) FROM vis_cliente_grade_slot), 1));')
[void]$sb.AppendLine('SELECT setval(pg_get_serial_sequence(''vis_cliente_grade_escopo'',''id''), COALESCE((SELECT MAX(id) FROM vis_cliente_grade_escopo), 1));')
[void]$sb.AppendLine('SELECT setval(pg_get_serial_sequence(''ops_arme_janela'',''id''), COALESCE((SELECT MAX(id) FROM ops_arme_janela), 1));')
[void]$sb.AppendLine('COMMIT;')

$utf8 = New-Object System.Text.UTF8Encoding $false
[System.IO.File]::WriteAllText($OutFile, $sb.ToString(), $utf8)
Write-Host "OK: $OutFile" -ForegroundColor Green
