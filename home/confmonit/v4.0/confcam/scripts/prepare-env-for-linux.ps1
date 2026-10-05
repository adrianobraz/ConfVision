# Normaliza .env para upload no core-4 (LF, UTF-8 sem BOM, sem CR nas chaves).
# Uso: .\scripts\prepare-env-for-linux.ps1

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$envPath = Join-Path $root '.env'

if (-not (Test-Path $envPath)) {
    Write-Error ".env nao encontrado em $root"
}

$raw = [System.IO.File]::ReadAllText($envPath)
$raw = $raw -replace "`r`n", "`n" -replace "`r", "`n"
$lines = $raw -split "`n"

$out = New-Object System.Collections.Generic.List[string]
foreach ($line in $lines) {
    $t = $line.TrimEnd()
    if ($t -match '^(POSTGRES_URL|VIS_WORKER_API_KEY|CONFVISION_GRADE_WORKER_KEY)=') {
        $parts = $t -split '=', 2
        if ($parts.Count -eq 2) {
            $t = $parts[0] + '=' + ($parts[1].Trim())
        }
    }
    $out.Add($t)
}

$text = ($out -join "`n").TrimEnd() + "`n"
[System.IO.File]::WriteAllText($envPath, $text, [System.Text.UTF8Encoding]::new($false))

Write-Host "OK: $envPath"
Write-Host "  encoding: UTF-8 sem BOM"
Write-Host "  line endings: LF"
Write-Host "Suba via FileZilla (modo texto) e no servidor: systemctl restart confmonit4confvision"
