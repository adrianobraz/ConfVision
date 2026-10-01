# Copia base + documenta overlay (merge manual de chaves simples)
# Uso: .\render-config.ps1 -Environment production
param(
  [ValidateSet("development", "staging", "production", "production-webrtc", "production-srt")]
  [string]$Environment = "production",
  [string]$OutFile = "$PSScriptRoot\..\config\rendered\mediamtx.$Environment.yml"
)
$base = Join-Path $PSScriptRoot "..\config\mediamtx.confvision.yml"
$overlay = Join-Path $PSScriptRoot "..\config\overlays\$Environment.yaml"
$outDir = Split-Path $OutFile -Parent
New-Item -ItemType Directory -Force -Path $outDir | Out-Null
Copy-Item $base $OutFile -Force
if (Test-Path $overlay) {
  Add-Content $OutFile "`n# --- overlay: $Environment ---`n"
  Get-Content $overlay | Add-Content $OutFile
}
Write-Host "Rendered $OutFile (review overlay keys; MediaMTX uses last YAML key wins only if duplicate keys — prefer single merged file for prod)"
