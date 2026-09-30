# Opção A — gerar CSV MediaMTX (RTSP cam/{hash}) + aplicar via API Go
param(
    [string]$VisWorkerKey = $env:VIS_WORKER_API_KEY,
    [int[]]$CameraIds = @(),
    [switch]$AllCameras,
    [switch]$DespausarAnalitico,
    [switch]$SkipBuild
)

$ErrorActionPreference = "Stop"
if (-not $VisWorkerKey) {
    $envFile = "C:\sistemaconfmonit\core4\home\confmonit\v4.0\confvision\.env"
    if (Test-Path $envFile) {
        $VisWorkerKey = (Select-String -Path $envFile -Pattern '^VIS_WORKER_API_KEY=(.+)$').Matches.Groups[1].Value.Trim()
    }
}
if (-not $VisWorkerKey) { throw "Defina VIS_WORKER_API_KEY" }

$csv = Join-Path $PSScriptRoot "..\sql\opcao_a_mediamtx.local.csv"
if (-not $SkipBuild) {
    $buildArgs = @{
        OutCsv              = $csv
        VisWorkerKey        = $VisWorkerKey
        DespausarAnalitico  = $DespausarAnalitico
    }
    if ($AllCameras) {
        $buildArgs.AllCameras = $true
    } elseif ($CameraIds.Count -gt 0) {
        $buildArgs.CameraIds = $CameraIds
    }
    & (Join-Path $PSScriptRoot "opcao-a-build-mediamtx-csv.ps1") @buildArgs
}

& (Join-Path $PSScriptRoot "opcao-a-run.ps1") -VisWorkerKey $VisWorkerKey -CsvPath $csv -Apply
