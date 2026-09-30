# Probes stack + HLS manifest (foxpro). Read-only.
$ErrorActionPreference = "Continue"
$fail = 0

function Get-Http($Url, $TimeoutSec = 25) {
    try {
        $r = Invoke-WebRequest -Uri $Url -TimeoutSec $TimeoutSec -UseBasicParsing
        return @{ Ok = $true; Code = [int]$r.StatusCode; Body = $r.Content }
    } catch {
        if ($_.Exception.Response) {
            return @{ Ok = $false; Code = [int]$_.Exception.Response.StatusCode.value__; Body = "" }
        }
        return @{ Ok = $false; Code = 0; Body = $_.Exception.Message }
    }
}

Write-Host "==> Go vis_health"
$h = Get-Http "https://vision.confmonit2.com.br/vis_health"
if ($h.Code -eq 200) { Write-Host "    OK" } else { Write-Host "    FAIL $($h.Code)"; $fail++ }

$rustBases = @(
    "https://foxpro-rust-pilot.rkr351.easypanel.host",
    "https://foxpro-rust-pilot-b.rkr351.easypanel.host"
)
foreach ($base in $rustBases) {
    Write-Host "==> Rust $base/health"
    $r = Get-Http "$base/health"
    if ($r.Code -eq 200) {
        $snippet = $r.Body
        if ($snippet.Length -gt 120) { $snippet = $snippet.Substring(0, 120) }
        Write-Host "    OK $snippet"
    } else {
        Write-Host "    FAIL $($r.Code)"
        $fail++
    }
}

Write-Host "==> confvision root"
$cv = Get-Http "https://foxpro-confvision.rkr351.easypanel.host/"
if ($cv.Code -eq 200) { Write-Host "    OK" } else { Write-Host "    FAIL $($cv.Code)"; $fail++ }

Write-Host "==> live-player.js recovery markers (repo)"
$lpPath = Join-Path $PSScriptRoot "..\recursos\modulos\confvision\live-player.js"
$lp = Get-Content -Raw -Path $lpPath
if ($lp -match "RECOVERY_AFTER_EXHAUST_MS" -and $lp -match "agendarRecuperacaoLenta") {
    Write-Host "    OK"
} else {
    Write-Host "    FAIL"
    $fail++
}

$hlsUrl = "https://foxpro-confvision.rkr351.easypanel.host:8888/cam/dkn59r69jvqr/index.m3u8"
Write-Host "==> HLS manifest $hlsUrl"
$m = Get-Http $hlsUrl 15
if ($m.Code -eq 200 -and $m.Body -match "#EXTM3U") {
    Write-Host "    OK manifest"
} else {
    Write-Host "    WARN code=$($m.Code) (offline or 8888 not public)"
}

if ($fail -gt 0) {
    Write-Host ""
    Write-Host "RESULT: FAIL ($fail critical)"
    exit 1
}
Write-Host ""
Write-Host "RESULT: OK critical probes"
