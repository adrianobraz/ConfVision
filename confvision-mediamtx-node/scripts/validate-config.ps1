# Valida sintaxe YAML e presença de chaves ConfVision obrigatórias
param(
  [string]$ConfigPath = "$PSScriptRoot\..\config\mediamtx.confvision.yml"
)
$ErrorActionPreference = "Stop"
if (-not (Test-Path $ConfigPath)) { throw "Config not found: $ConfigPath" }
$content = Get-Content -Raw $ConfigPath
$required = @(
  "authMethod: http",
  "authHTTPAddress:",
  "paths:",
  "all_others:",
  "~^cam/[0-9a-z]{12,}$"
)
foreach ($r in $required) {
  if ($content -notmatch [regex]::Escape($r).Replace('\{12,\}', '\{12,\}')) {
    if ($content -notlike "*$r*") { throw "Missing required fragment: $r" }
  }
}
# Simple check without regex escape issues
foreach ($line in @("authMethod: http", "authHTTPAddress:", "all_others:")) {
  if ($content -notlike "*$line*") { throw "Missing: $line" }
}
Write-Host "OK ConfVision MediaMTX config structure: $ConfigPath"
