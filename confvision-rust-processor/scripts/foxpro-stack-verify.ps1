# Sonda foxpro: rust-pilot, confvision, confvision-worker, Go API
# Uso: .\foxpro-stack-verify.ps1 [-RustBase URL]

param(
    [string]$RustBase = "https://foxpro-rust-pilot.rkr351.easypanel.host"
)

$ErrorActionPreference = "Continue"
$fail = $false

function Get-Http {
    param([string]$Url, [int]$TimeoutSec = 20)
    try {
        $r = Invoke-WebRequest -Uri $Url -TimeoutSec $TimeoutSec -UseBasicParsing
        return @{ Code = [int]$r.StatusCode; Body = $r.Content }
    } catch {
        if ($_.Exception.Response) {
            $code = [int]$_.Exception.Response.StatusCode.value__
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $body = $reader.ReadToEnd()
            return @{ Code = $code; Body = $body }
        }
        return @{ Code = 0; Body = $_.Exception.Message }
    }
}

Write-Host "==> Go API: https://vision.confmonit2.com.br/vis_health"
$api = Get-Http "https://vision.confmonit2.com.br/vis_health"
if ($api.Code -eq 200) { Write-Host "    OK $($api.Body)" } else { Write-Host "    FAIL $($api.Code)"; $fail = $true }

Write-Host "==> Rust: $RustBase/health"
$rust = Get-Http "$RustBase/health"
if ($rust.Code -eq 200) {
    $h = $rust.Body | ConvertFrom-Json
    Write-Host "    status=$($h.status) cams=$($h.cameras_online)/$($h.cameras_total) capacity=$($h.capacity_state) advisory=$($h.load_advisory)"
} else {
    Write-Host "    FAIL $($rust.Code)"
    $fail = $true
}

Write-Host "==> confvision: https://foxpro-confvision.rkr351.easypanel.host/"
$cv = Get-Http "https://foxpro-confvision.rkr351.easypanel.host/"
Write-Host "    HTTP $($cv.Code)"

Write-Host "==> confvision-worker: https://foxpro-confvision-worker.rkr351.easypanel.host/"
$workerUrl = "https://foxpro-confvision-worker.rkr351.easypanel.host/"
$wBody = & curl.exe -sS --max-time 20 $workerUrl 2>&1 | Out-String
$wCode = & curl.exe -sS -o NUL -w "%{http_code}" --max-time 20 $workerUrl 2>&1
Write-Host "    HTTP $wCode"
if ($wBody -match "Service is not started") {
    Write-Host "    DIAG: servico PARADO no EasyPanel. Acao: Start; se No such image -> Implantar (build) + Start."
    $fail = $true
} elseif ([int]$wCode -ge 502) {
    Write-Host "    DIAG: backend ausente (502/503) - container parado ou crash loop."
    $fail = $true
}

Write-Host ""
Write-Host "NOTA: worker nao expoe HTTP; use Logs EasyPanel + CPU/mem apos Start."
if ($fail) {
    Write-Host "RESULT: FAIL"
    exit 1
}
Write-Host "RESULT: probes OK (worker requer Start manual se parado)."
