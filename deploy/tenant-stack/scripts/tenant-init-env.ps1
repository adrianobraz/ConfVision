# Gera .env.tenant / .env.mtx / .env.publisher / .env.rust (Windows)
param(
    [string]$TenantId = "ct_cli_example",
    [int]$MaxCameras = 200
)

$ErrorActionPreference = "Stop"
Set-Location (Join-Path $PSScriptRoot "..")

function Set-LineInFile($path, $pattern, $replacement) {
    $c = Get-Content $path -Raw
    $c = $c -replace "(?m)^$pattern.*$", $replacement
    Set-Content -Path $path -Value $c -NoNewline
}

Copy-Item env/tenant.env.example .env.tenant -Force
Set-LineInFile .env.tenant "TENANT_ID" "TENANT_ID=$TenantId"
Set-LineInFile .env.tenant "MAX_CAMERAS" "MAX_CAMERAS=$MaxCameras"

Copy-Item env/mtx.env.example .env.mtx -Force
Copy-Item env/publisher.env.example .env.publisher -Force
Copy-Item env/rust-processor.env.example .env.rust -Force

$p = Get-Content .env.publisher -Raw -replace '\$\{TENANT_ID\}', $TenantId
Set-Content .env.publisher $p -NoNewline
$r = Get-Content .env.rust -Raw -replace '\$\{TENANT_ID\}', $TenantId
$r = $r -replace '(?m)^MAX_CAMERAS=.*', "MAX_CAMERAS=$MaxCameras"
$r = $r -replace '(?m)^PROCESSOR_ID=.*', "PROCESSOR_ID=$TenantId"
$r = $r -replace '(?m)^WORKER_ID=.*', "WORKER_ID=$TenantId"
Set-Content .env.rust $r -NoNewline

Write-Host "Criados .env.* para tenant $TenantId" -ForegroundColor Green
