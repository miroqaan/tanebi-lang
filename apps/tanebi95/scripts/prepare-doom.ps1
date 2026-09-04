[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'

$appRoot = Split-Path -Parent $PSScriptRoot
$repositoryRoot = Split-Path -Parent (Split-Path -Parent $appRoot)
$buildRoot = Join-Path $repositoryRoot 'artifacts\doom-wasm-engine'
$publicDoom = Join-Path $appRoot 'public\doom'
$engineRoot = Join-Path $publicDoom 'engine'
$licenseRoot = Join-Path $publicDoom 'licenses'
New-Item -ItemType Directory -Force -Path $buildRoot, $publicDoom, $engineRoot, $licenseRoot | Out-Null

$commit = '64de0924591dec59a7d49a7d10467e125b50ea99'
$engineFiles = @(
    @{ Name = 'chocolate-doom.js'; Sha256 = 'C5ED72BFC6875566AFF22244083907CC5111AA87E33F4007E420368102FAF51B' },
    @{ Name = 'chocolate-doom.wasm'; Sha256 = '6922E97650FF306D0FEF584EB515D52FD1E09BB77BC7E8287B658E09977B54F7' }
)

foreach ($file in $engineFiles) {
    $destination = Join-Path $engineRoot $file.Name
    $url = "https://raw.githubusercontent.com/gabrielbotandev/doom-wasm/$commit/web/public/engine/$($file.Name)"
    if (-not (Test-Path -LiteralPath $destination)) {
        Invoke-WebRequest -Uri $url -OutFile $destination
    }
    $actual = (Get-FileHash -LiteralPath $destination -Algorithm SHA256).Hash
    if ($actual -ne $file.Sha256) {
        throw "Checksum mismatch for $destination."
    }
}

$dataName = 'chocolate-doom.data'
$dataSource = Join-Path $buildRoot $dataName
$dataUrl = "https://raw.githubusercontent.com/gabrielbotandev/doom-wasm/$commit/web/public/engine/$dataName"
if (-not (Test-Path -LiteralPath $dataSource)) {
    Invoke-WebRequest -Uri $dataUrl -OutFile $dataSource
}
if ((Get-FileHash -LiteralPath $dataSource -Algorithm SHA256).Hash -ne 'A8772E088847032510D97BA2312406A6998F21CBAB44D4FF10696FAA9C0ECD4B') {
    throw "Checksum mismatch for $dataSource."
}

# Cloudflare static assets have a per-file size ceiling. Preserve the exact
# upstream bytes while serving the 28.8 MB Emscripten package as four chunks.
$partChecksums = @(
    '5BCB2D195F31A73398337299DEA1CCDD0D9FC26A144B9FC43A0A8010F2270D6B',
    'CC2534E3FED4BF548CB7502A1F1608A176A8FAEA41EB88004449D7D54ABAD7AE',
    '2E3E564522C2D309AE4EB6B73456B5C0E445CC3C820A347E4413CB0F548EF88F',
    '781455B1A6C81823CA1C5D9A2275E371184501A0BA1933064197AC409DC09B36'
)
$sourceStream = [System.IO.File]::OpenRead($dataSource)
try {
    $buffer = New-Object byte[] (8MB)
    $index = 0
    while (($read = $sourceStream.Read($buffer, 0, $buffer.Length)) -gt 0) {
        $partPath = Join-Path $engineRoot "$dataName.part$index"
        $partStream = [System.IO.File]::Create($partPath)
        try { $partStream.Write($buffer, 0, $read) } finally { $partStream.Dispose() }
        if ((Get-FileHash -LiteralPath $partPath -Algorithm SHA256).Hash -ne $partChecksums[$index]) {
            throw "Checksum mismatch for $partPath."
        }
        $index++
    }
} finally {
    $sourceStream.Dispose()
}

$licenseFiles = @(
    'CHOCOLATE-DOOM-GPL-2.0.md',
    'FREEDOOM-COPYING.txt',
    'CREDITS.txt',
    'CREDITS-MUSIC.txt',
    'EMSCRIPTEN-LICENSE.txt',
    'SDL2-LICENSE.txt',
    'THIRD_PARTY_NOTICES.md'
)

foreach ($name in $licenseFiles) {
    $destination = Join-Path $licenseRoot $name
    $url = "https://raw.githubusercontent.com/gabrielbotandev/doom-wasm/$commit/licenses/$name"
    Invoke-WebRequest -Uri $url -OutFile $destination
}

$notice = @'
TANEBI 95 DOOM player

This distribution contains no proprietary DOOM game data.

Engine: Chocolate Doom, GPL-2.0-or-later
Upstream: https://github.com/chocolate-doom/chocolate-doom

Browser build integration: doom-wasm, GPL-2.0-or-later
Pinned source: https://github.com/gabrielbotandev/doom-wasm/tree/64de0924591dec59a7d49a7d10467e125b50ea99

Game data: Freedoom 0.13.0 Phase 2
Upstream: https://github.com/freedoom/freedoom/tree/v0.13.0

The full license and attribution texts are in /doom/licenses/.
DOOM is a trademark of id Software. TANEBI 95 is not affiliated with or
endorsed by id Software or Microsoft.
'@
Set-Content -LiteralPath (Join-Path $publicDoom 'NOTICE.txt') -Value $notice -Encoding UTF8

$totalSize = (Get-ChildItem -LiteralPath $engineRoot -File | Where-Object Name -ne $dataName | Measure-Object -Property Length -Sum).Sum
Write-Host "Prepared legal Chocolate Doom WebAssembly engine ($totalSize bytes)."
