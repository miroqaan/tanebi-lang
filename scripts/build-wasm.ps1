[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$siteRoot = Split-Path -Parent $PSScriptRoot
$repositoryRoot = (Resolve-Path (Join-Path $siteRoot '..\..')).Path
$publicDirectory = Join-Path $siteRoot 'public'

$goCommand = Get-Command go -ErrorAction SilentlyContinue
$go = if ($goCommand) { $goCommand.Source } else {
    Get-Item -Path (Join-Path $env:USERPROFILE 'sdk\go*\bin\go.exe') -ErrorAction SilentlyContinue |
        Sort-Object FullName -Descending |
        Select-Object -First 1 -ExpandProperty FullName
}
if (-not $go) { throw 'Go SDK was not found.' }

$goRoot = (& $go env GOROOT).Trim()
$wasmExecCandidates = @(
    (Join-Path $goRoot 'lib\wasm\wasm_exec.js'),
    (Join-Path $goRoot 'misc\wasm\wasm_exec.js')
)
$wasmExec = $wasmExecCandidates | Where-Object { Test-Path -LiteralPath $_ } | Select-Object -First 1
if (-not $wasmExec) { throw "wasm_exec.js was not found below $goRoot." }

New-Item -ItemType Directory -Path $publicDirectory -Force | Out-Null
$previousGoOS = $env:GOOS
$previousGoArch = $env:GOARCH
$previousCgo = $env:CGO_ENABLED
try {
    $env:GOOS = 'js'
    $env:GOARCH = 'wasm'
    $env:CGO_ENABLED = '0'
    Push-Location $repositoryRoot
    try {
        & $go build -trimpath -ldflags '-s -w' -o (Join-Path $publicDirectory 'tanebi.wasm') ./cmd/tanebi-wasm
        if ($LASTEXITCODE -ne 0) { throw 'TANEBI WebAssembly build failed.' }
    }
    finally {
        Pop-Location
    }
}
finally {
    $env:GOOS = $previousGoOS
    $env:GOARCH = $previousGoArch
    $env:CGO_ENABLED = $previousCgo
}

Copy-Item -LiteralPath $wasmExec -Destination (Join-Path $publicDirectory 'wasm_exec.js') -Force
Write-Host "Built TANEBI WebAssembly in $publicDirectory"

