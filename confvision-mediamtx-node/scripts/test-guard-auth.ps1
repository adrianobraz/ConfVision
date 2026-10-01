# Testes de rejeicao/aceite no POST /auth (sem RTMP real)
param(
    [string]$GuardBase = "http://127.0.0.1:8100"
)

function Invoke-Auth($payload) {
    $json = $payload | ConvertTo-Json -Compress
    try {
        $r = Invoke-WebRequest -Method Post -Uri "$GuardBase/auth" -Body $json -ContentType "application/json" -UseBasicParsing -TimeoutSec 15
        return @{ Code = $r.StatusCode; Body = ($r.Content | ConvertFrom-Json) }
    } catch {
        $resp = $_.Exception.Response
        if ($resp) {
            $reader = New-Object System.IO.StreamReader($resp.GetResponseStream())
            $raw = $reader.ReadToEnd()
            $body = $null
            try { $body = $raw | ConvertFrom-Json } catch {}
            return @{ Code = [int]$resp.StatusCode; Body = $body }
        }
        throw
    }
}

Write-Host "=== test-guard-auth @ $GuardBase ==="

$cases = @(
    @{ Name = "path_invalido"; Payload = @{ action = "publish"; ip = "203.0.113.1"; path = "live/foo" }; Expect = 403 },
    @{ Name = "path_curto"; Payload = @{ action = "publish"; ip = "203.0.113.2"; path = "cam/abc" }; Expect = 403 },
    @{ Name = "action_desconhecida"; Payload = @{ action = "foo"; ip = "203.0.113.3"; path = "cam/0123456789ab" }; Expect = 401 }
)

$fail = 0
foreach ($c in $cases) {
    $res = Invoke-Auth $c.Payload
    $motivo = $res.Body.motivo
    if ($res.Code -eq $c.Expect) {
        Write-Host "[OK] $($c.Name) HTTP $($res.Code) motivo=$motivo"
    } else {
        Write-Host "[FALHA] $($c.Name) esperado $($c.Expect) obteve $($res.Code) motivo=$motivo"
        $fail++
    }
}

Write-Host ""
Write-Host "Cenarios camera_inativa / ip_banido / aceite valido exigem RTMP_PUBLISH_SECRET + API Go (teste manual ou integracao)."
if ($fail -gt 0) { exit 1 }
exit 0
