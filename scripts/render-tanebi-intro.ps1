[CmdletBinding()]
param(
    [string]$Output = (Join-Path (Split-Path -Parent $PSScriptRoot) 'artifacts\videos\tanebi-introduction-ja-compiler.mp4')
)

$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Drawing

$repositoryRoot = Split-Path -Parent $PSScriptRoot
$frameRoot = Join-Path $repositoryRoot 'artifacts\videos\tanebi-intro-compiler-frames'
New-Item -ItemType Directory -Force -Path $frameRoot, (Split-Path -Parent $Output) | Out-Null

$fontRegular = 'Yu Gothic'
$fontBold = 'Yu Gothic'
$width = 1280
$height = 720

function New-Font([float]$size, [System.Drawing.FontStyle]$style = [System.Drawing.FontStyle]::Regular) {
    return [System.Drawing.Font]::new($fontRegular, $size, $style, [System.Drawing.GraphicsUnit]::Pixel)
}

function Draw-CenteredText($graphics, [string]$text, $font, $brush, [float]$y, [float]$boxHeight) {
    $format = [System.Drawing.StringFormat]::new()
    $format.Alignment = [System.Drawing.StringAlignment]::Center
    $format.LineAlignment = [System.Drawing.StringAlignment]::Center
    $graphics.DrawString($text, $font, $brush, [System.Drawing.RectangleF]::new(80, $y, 1120, $boxHeight), $format)
    $format.Dispose()
}

function Draw-Flame($graphics, [float]$x, [float]$y, [float]$scale) {
    $outer = [System.Drawing.PointF[]]@(
        [System.Drawing.PointF]::new($x + 50 * $scale, $y),
        [System.Drawing.PointF]::new($x + 78 * $scale, $y + 38 * $scale),
        [System.Drawing.PointF]::new($x + 70 * $scale, $y + 78 * $scale),
        [System.Drawing.PointF]::new($x + 40 * $scale, $y + 100 * $scale),
        [System.Drawing.PointF]::new($x + 10 * $scale, $y + 78 * $scale),
        [System.Drawing.PointF]::new($x, $y + 48 * $scale),
        [System.Drawing.PointF]::new($x + 28 * $scale, $y + 65 * $scale),
        [System.Drawing.PointF]::new($x + 24 * $scale, $y + 30 * $scale)
    )
    $inner = [System.Drawing.PointF[]]@(
        [System.Drawing.PointF]::new($x + 45 * $scale, $y + 43 * $scale),
        [System.Drawing.PointF]::new($x + 60 * $scale, $y + 67 * $scale),
        [System.Drawing.PointF]::new($x + 45 * $scale, $y + 91 * $scale),
        [System.Drawing.PointF]::new($x + 27 * $scale, $y + 72 * $scale)
    )
    $orange = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(231, 68, 36))
    $yellow = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(255, 211, 78))
    $graphics.FillPolygon($orange, $outer)
    $graphics.FillPolygon($yellow, $inner)
    $orange.Dispose()
    $yellow.Dispose()
}

function New-Slide([int]$index, [string]$eyebrow, [string]$title, [string[]]$lines, [string[]]$code = @()) {
    $bitmap = [System.Drawing.Bitmap]::new($width, $height)
    $graphics = [System.Drawing.Graphics]::FromImage($bitmap)
    $graphics.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::AntiAlias
    $graphics.TextRenderingHint = [System.Drawing.Text.TextRenderingHint]::AntiAliasGridFit
    $graphics.Clear([System.Drawing.Color]::FromArgb(9, 15, 31))

    $accent = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(231, 68, 36))
    $white = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(244, 247, 255))
    $muted = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(169, 183, 209))
    $teal = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(38, 185, 174))
    $panel = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(20, 30, 54))
    $border = [System.Drawing.Pen]::new([System.Drawing.Color]::FromArgb(49, 67, 105), 2)

    $graphics.FillRectangle($accent, 0, 0, 12, $height)
    $graphics.FillRectangle($teal, 12, 0, 4, $height)
    Draw-Flame $graphics 1120 46 0.72

    $eyebrowFont = New-Font 25 ([System.Drawing.FontStyle]::Bold)
    $titleFont = New-Font 54 ([System.Drawing.FontStyle]::Bold)
    $bodyFont = New-Font 30
    $codeFont = [System.Drawing.Font]::new('MS Gothic', 27, [System.Drawing.FontStyle]::Regular, [System.Drawing.GraphicsUnit]::Pixel)
    $graphics.DrawString($eyebrow, $eyebrowFont, $accent, 76, 62)
    $graphics.DrawString($title, $titleFont, $white, 72, 112)

    if ($code.Count -gt 0) {
        $graphics.FillRectangle($panel, 74, 222, 1130, 380)
        $graphics.DrawRectangle($border, 74, 222, 1130, 380)
        $y = 260
        foreach ($line in $code) {
            $brush = if ($line.TrimStart().StartsWith('#')) { $muted } elseif ($line.TrimStart() -match '^(fn|let|while|if|print|println)\b') { $teal } else { $white }
            $graphics.DrawString($line, $codeFont, $brush, 112, $y)
            $y += 49
        }
    } else {
        $y = 245
        foreach ($line in $lines) {
            $graphics.FillEllipse($accent, 82, $y + 12, 10, 10)
            $graphics.DrawString($line, $bodyFont, $white, 112, $y)
            $y += 72
        }
    }

    $footerFont = New-Font 19
    $graphics.DrawString(('AKAGI UNIVERSE   ·   {0}/6' -f $index), $footerFont, $muted, 76, 660)

    $path = Join-Path $frameRoot ('slide-{0:d2}.png' -f $index)
    $bitmap.Save($path, [System.Drawing.Imaging.ImageFormat]::Png)

    foreach ($resource in @($eyebrowFont, $titleFont, $bodyFont, $codeFont, $accent, $white, $muted, $teal, $panel, $border, $graphics, $bitmap)) {
        $resource.Dispose()
    }
    return $path
}

$slides = @()
$slides += New-Slide 1 'プログラミング言語' 'TANEBI' @('一行のコードが、世界に火を灯す。', 'Goで解析し、ネイティブ実行ファイルへ。')
$slides += New-Slide 2 'コンパイラーの構成' '言語の中身を見せる。' @('字句解析 → 構文解析 → 型検査', 'GoフロントエンドからRustを生成', 'rustcでネイティブ実行ファイルへ')
$slides += New-Slide 3 'Unicode識別子' '世界中の言葉で書ける。' @() @(
    'fn main() {',
    '    let 名前:str = "アカギ";',
    '    let 種火:u64 = 3;',
    '    print(名前);',
    '    println("の世界が目を覚ます");',
    '}'
)
$slides += New-Slide 4 '明示的な型' '型を確かめてから動かす。' @('u64・bool・strを明示的に扱う', '関数・条件分岐・whileによる繰り返し', 'メモリーとI/Oは組み込み関数で操作')
$slides += New-Slide 5 'ホストから独立OSへ' '一つの言語、一つのCLI。' @('check・build・run・emitに統一', 'runはコンパイル後のバイナリーを実行', 'UEFI向けにTANEBI 95のOSロジックも生成')
$slides += New-Slide 6 'AKAGI UNIVERSE' '一行のコードから、世界へ。' @('go run ./cmd/tanebi run examples/awakening.tanebi', 'github.com/miroqaan/tanebi-lang', '短いサンプルから、コンパイラーの中へ。')

$ffmpeg = Get-Command ffmpeg -ErrorAction Stop
$inputs = @()
foreach ($slide in $slides) {
    $inputs += @('-loop', '1', '-t', '6', '-i', $slide)
}

$videoFilter = @(
    '[0:v]scale=1280:720,zoompan=z=min(zoom+0.00035\,1.03):d=180:s=1280x720:fps=30[v0]',
    '[1:v]scale=1280:720,zoompan=z=min(zoom+0.00035\,1.03):d=180:s=1280x720:fps=30[v1]',
    '[2:v]scale=1280:720,zoompan=z=min(zoom+0.00035\,1.03):d=180:s=1280x720:fps=30[v2]',
    '[3:v]scale=1280:720,zoompan=z=min(zoom+0.00035\,1.03):d=180:s=1280x720:fps=30[v3]',
    '[4:v]scale=1280:720,zoompan=z=min(zoom+0.00035\,1.03):d=180:s=1280x720:fps=30[v4]',
    '[5:v]scale=1280:720,zoompan=z=min(zoom+0.00035\,1.03):d=180:s=1280x720:fps=30[v5]',
    '[v0][v1]xfade=transition=fade:duration=0.7:offset=5.3[x1]',
    '[x1][v2]xfade=transition=fade:duration=0.7:offset=10.6[x2]',
    '[x2][v3]xfade=transition=fade:duration=0.7:offset=15.9[x3]',
    '[x3][v4]xfade=transition=fade:duration=0.7:offset=21.2[x4]',
    '[x4][v5]xfade=transition=fade:duration=0.7:offset=26.5,format=yuv420p[v]'
) -join ';'

$audioFilter = 'sine=frequency=110:sample_rate=48000:duration=32.5,volume=0.035[a0];sine=frequency=164.81:sample_rate=48000:duration=32.5,volume=0.018[a1];sine=frequency=220:sample_rate=48000:duration=32.5,volume=0.012[a2];[a0][a1][a2]amix=inputs=3,afade=t=in:st=0:d=1.5,afade=t=out:st=30:d=2.5[a]'
$filter = $videoFilter + ';' + $audioFilter

& $ffmpeg.Source -y @inputs -filter_complex $filter -map '[v]' -map '[a]' -c:v libx264 -preset medium -crf 18 -c:a aac -b:a 160k -movflags +faststart -shortest $Output
if ($LASTEXITCODE -ne 0) { throw 'ffmpeg failed to render the TANEBI introduction video.' }

Write-Host "Rendered: $Output"
