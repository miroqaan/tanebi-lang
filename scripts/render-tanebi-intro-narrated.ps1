[CmdletBinding()]
param(
    [string]$Output = (Join-Path (Split-Path -Parent $PSScriptRoot) 'artifacts\videos\tanebi-introduction-ja-narrated-1080p.mp4'),
    [string]$Voice = 'ja-JP-NanamiNeural'
)

$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Drawing

$repositoryRoot = Split-Path -Parent $PSScriptRoot
$workRoot = Join-Path $repositoryRoot 'artifacts\videos\tanebi-intro-narrated'
$frameRoot = Join-Path $workRoot 'frames'
$audioRoot = Join-Path $workRoot 'audio'
$segmentRoot = Join-Path $workRoot 'segments'
New-Item -ItemType Directory -Force -Path $frameRoot, $audioRoot, $segmentRoot, (Split-Path -Parent $Output) | Out-Null

$ffmpeg = (Get-Command ffmpeg -ErrorAction Stop).Source
$ffprobe = (Get-Command ffprobe -ErrorAction Stop).Source
$python = (Get-Command py -ErrorAction Stop).Source
$width = 1280
$height = 720
$renderWidth = 1920
$renderHeight = 1080
$renderScale = $renderWidth / $width

function Get-CodeLines([string]$relativePath, [int]$start, [int]$end) {
    $all = Get-Content -LiteralPath (Join-Path $repositoryRoot $relativePath) -Encoding UTF8
    $result = @()
    for ($line = $start; $line -le $end -and $line -le $all.Count; $line++) {
        $result += [pscustomobject]@{ Number = $line; Text = $all[$line - 1] }
    }
    return $result
}

function New-Font([string]$family, [float]$size, [System.Drawing.FontStyle]$style = [System.Drawing.FontStyle]::Regular) {
    return [System.Drawing.Font]::new($family, $size, $style, [System.Drawing.GraphicsUnit]::Pixel)
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
    $orange = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(238, 82, 45))
    $yellow = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(255, 211, 78))
    $graphics.FillPolygon($orange, $outer)
    $graphics.FillPolygon($yellow, $inner)
    $orange.Dispose()
    $yellow.Dispose()
}

function New-SceneFrame($scene, [int]$index, [int]$count) {
    $bitmap = [System.Drawing.Bitmap]::new($renderWidth, $renderHeight)
    $graphics = [System.Drawing.Graphics]::FromImage($bitmap)
    $graphics.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::AntiAlias
    $graphics.TextRenderingHint = [System.Drawing.Text.TextRenderingHint]::AntiAliasGridFit
    $graphics.Clear([System.Drawing.Color]::FromArgb(8, 14, 28))
    $graphics.ScaleTransform($renderScale, $renderScale)

    $white = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(244, 247, 255))
    $muted = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(154, 171, 203))
    $orange = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(238, 82, 45))
    $teal = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(50, 204, 186))
    $codePanel = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(16, 25, 47))
    $notePanel = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(22, 34, 60))
    $lineBrush = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(40, 55, 88))
    $border = [System.Drawing.Pen]::new([System.Drawing.Color]::FromArgb(55, 76, 119), 2)

    $eyebrowFont = New-Font 'Yu Gothic' 19 ([System.Drawing.FontStyle]::Bold)
    $titleFont = New-Font 'Yu Gothic' 38 ([System.Drawing.FontStyle]::Bold)
    $pathFont = New-Font 'Consolas' 17
    $codeFont = New-Font 'MS Gothic' 21
    $lineNumberFont = New-Font 'Consolas' 18
    $noteTitleFont = New-Font 'Yu Gothic' 21 ([System.Drawing.FontStyle]::Bold)
    $noteFont = New-Font 'Yu Gothic' 20
    $footerFont = New-Font 'Yu Gothic' 15

    $graphics.FillRectangle($orange, 0, 0, 10, $height)
    $graphics.FillRectangle($teal, 10, 0, 4, $height)
    $graphics.DrawString($scene.Eyebrow, $eyebrowFont, $orange, 52, 34)
    $graphics.DrawString($scene.Title, $titleFont, $white, 48, 65)
    Draw-Flame $graphics 1162 32 0.56

    $graphics.FillRectangle($codePanel, 48, 132, 805, 516)
    $graphics.DrawRectangle($border, 48, 132, 805, 516)
    $graphics.FillRectangle($lineBrush, 48, 132, 805, 42)
    $graphics.DrawString($scene.Path, $pathFont, $muted, 68, 143)

    $graphics.SetClip([System.Drawing.Rectangle]::new(49, 175, 803, 472))
    $y = 191
    foreach ($line in $scene.Code) {
        if ($y -gt 606) { break }
        $graphics.DrawString(('{0,3}' -f $line.Number), $lineNumberFont, $muted, 64, $y)
        $brush = $white
        $trimmed = $line.Text.TrimStart()
        if ($trimmed.StartsWith('//') -or $trimmed.StartsWith('#')) { $brush = $muted }
        elseif ($trimmed -match '^(func|type|const|var|if|for|switch|case|return|let|repeat|print|else)\b') { $brush = $teal }
        elseif ($line.Text -match '"') { $brush = $orange }
        $graphics.DrawString($line.Text.Replace("`t", '    '), $codeFont, $brush, 120, $y - 1)
        $y += 27
    }
    $graphics.ResetClip()

    $graphics.FillRectangle($notePanel, 875, 132, 357, 516)
    $graphics.DrawRectangle($border, 875, 132, 357, 516)
    $graphics.DrawString('この場面のポイント', $noteTitleFont, $teal, 900, 156)
    $noteY = 211
    foreach ($note in $scene.Notes) {
        $graphics.FillEllipse($orange, 902, $noteY + 8, 8, 8)
        $format = [System.Drawing.StringFormat]::new()
        $format.Trimming = [System.Drawing.StringTrimming]::Word
        $graphics.DrawString($note, $noteFont, $white, [System.Drawing.RectangleF]::new(925, $noteY, 278, 78), $format)
        $format.Dispose()
        $noteY += 94
    }

    $graphics.DrawString(('AKAGI UNIVERSE  ·  TANEBI  ·  {0}/{1}' -f $index, $count), $footerFont, $muted, 50, 677)

    $path = Join-Path $frameRoot ('scene-{0:d2}.png' -f $index)
    $bitmap.Save($path, [System.Drawing.Imaging.ImageFormat]::Png)
    foreach ($resource in @($eyebrowFont, $titleFont, $pathFont, $codeFont, $lineNumberFont, $noteTitleFont, $noteFont, $footerFont, $white, $muted, $orange, $teal, $codePanel, $notePanel, $lineBrush, $border, $graphics, $bitmap)) {
        $resource.Dispose()
    }
    return $path
}

$scenes = @(
    [pscustomobject]@{
        Eyebrow = '01 · LANGUAGE OVERVIEW'
        Title = 'TANEBI（種火）とは'
        Path = 'README.md'
        Code = Get-CodeLines 'README.md' 1 15
        Notes = @('Go製のインタープリター', '小さく決定論的', 'Unicode識別子')
        Narration = '種火。一行のコードが、世界に火を灯す。TANEBIは、Goで実装した、小さくて決定論的なインタープリター言語です。'
    },
    [pscustomobject]@{
        Eyebrow = '02 · GO CLI'
        Title = '入口はシンプルなmain関数'
        Path = 'cmd/tanebi/main.go'
        Code = Get-CodeLines 'cmd\tanebi\main.go' 10 27
        Notes = @('Go 1.22以上', '標準ライブラリのみ', 'tanebi.Runを呼び出す')
        Narration = 'CLIのmain関数は、指定されたTANEBIファイルを読み込み、公開関数tanebi.Runへ渡します。外部ライブラリは使わず、Goの標準ライブラリだけで動作します。'
    },
    [pscustomobject]@{
        Eyebrow = '03 · PIPELINE'
        Title = 'Lexer → Parser → Interpreter'
        Path = 'internal/tanebi/interpreter.go'
        Code = Get-CodeLines 'internal\tanebi\interpreter.go' 29 47
        Notes = @('tokenize：字句解析', 'parse：構文解析', 'execute：実行')
        Narration = '実行の中心はRun関数です。tokenizeで字句解析し、parseで抽象構文木を作り、executeで順番に実行します。処理の境界が明確なので、学習にも拡張にも向いています。'
    },
    [pscustomobject]@{
        Eyebrow = '04 · LEXER'
        Title = '文字を意味のあるTokenへ'
        Path = 'internal/tanebi/lexer.go'
        Code = Get-CodeLines 'internal\tanebi\lexer.go' 8 32
        Notes = @('tokenize / number', 'identifier / stringLiteral', 'runeでUnicode対応')
        Narration = 'Lexerでは、tokenize、number、identifier、stringLiteralが文字列をトークンに変換します。ソースをruneとして扱うため、日本語や韓国語を変数名に使用できます。'
    },
    [pscustomobject]@{
        Eyebrow = '05 · PARSER'
        Title = '構文をASTへ組み立てる'
        Path = 'internal/tanebi/parser.go'
        Code = Get-CodeLines 'internal\tanebi\parser.go' 35 62
        Notes = @('statement：文を解析', 'expression：式を解析', '演算子の優先順位を保持')
        Narration = 'Parserのstatement関数は、let、print、repeat、if、代入を構文木へ変換します。expressionからprimaryまで分けた関数が、論理演算、比較、四則演算の優先順位を保ちます。'
    },
    [pscustomobject]@{
        Eyebrow = '06 · INTERPRETER'
        Title = 'ASTを評価して結果を出す'
        Path = 'internal/tanebi/interpreter.go'
        Code = Get-CodeLines 'internal\tanebi\interpreter.go' 48 75
        Notes = @('execute：文の実行', 'evaluate：式の評価', 'evaluateBinary / compareValues')
        Narration = 'Interpreterではexecuteが文を処理し、evaluateとevaluateBinaryが式を評価します。valuesEqualとcompareValuesが比較を担当し、型の暗黙変換を避けて予測可能な結果を守ります。'
    },
    [pscustomobject]@{
        Eyebrow = '07 · REAL PROGRAM'
        Title = '実際のTANEBIコードを実行'
        Path = 'examples/awakening.tanebi'
        Code = Get-CodeLines 'examples\awakening.tanebi' 1 14
        Notes = @('repeatで3回実行', 'ifで条件分岐', '実出力：種火は消えない')
        Narration = 'これは実際のサンプルコードです。Unicodeの変数、repeatによる繰り返し、ifによる条件分岐を使っています。Go CLIで実行すると、アカギの世界が三回目を覚まし、種火は消えないと出力されます。'
    },
    [pscustomobject]@{
        Eyebrow = '08 · WEBASSEMBLY'
        Title = '同じ言語コアをブラウザーへ'
        Path = 'cmd/tanebi-wasm/main.go'
        Code = Get-CodeLines 'cmd\tanebi-wasm\main.go' 14 35
        Notes = @('syscall/jsブリッジ', 'tanebiRunを公開', 'TANEBI 95で再利用')
        Narration = 'WebAssembly版では、run関数をtanebiRunとしてブラウザーへ公開し、出力とエラーをJavaScriptへ返します。この橋によって、同じ言語コアがTANEBI 95のデスクトップ上でも動きます。'
    },
    [pscustomobject]@{
        Eyebrow = '09 · AKAGI UNIVERSE'
        Title = '一行のコードから、世界へ。'
        Path = 'github.com/miroqaan/tanebi-lang'
        Code = Get-CodeLines 'README.md' 17 31
        Notes = @('Lexer', 'Parser / AST', 'Interpreter')
        Narration = 'Lexer、Parser、AST、Interpreter。小さな構造を積み重ね、一行のコードから世界を作る。それがTANEBIです。'
    }
)

$segments = @()
for ($index = 0; $index -lt $scenes.Count; $index++) {
    $number = $index + 1
    $scene = $scenes[$index]
    $frame = New-SceneFrame $scene $number $scenes.Count
    $audio = Join-Path $audioRoot ('scene-{0:d2}.mp3' -f $number)
    $segment = Join-Path $segmentRoot ('scene-{0:d2}.mp4' -f $number)

    & $python -m edge_tts --voice $Voice --rate=-4% --text $scene.Narration --write-media $audio
    if ($LASTEXITCODE -ne 0) { throw "Japanese narration failed for scene $number." }

    $durationText = & $ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 $audio
    $duration = [math]::Ceiling(([double]::Parse($durationText, [Globalization.CultureInfo]::InvariantCulture) + 0.45) * 100) / 100
    & $ffmpeg -y -loop 1 -i $frame -i $audio -t $duration -vf "scale=${renderWidth}:${renderHeight},zoompan=z='min(zoom+0.00012,1.012)':d=9999:s=${renderWidth}x${renderHeight}:fps=30,format=yuv420p" -af 'apad=pad_dur=0.45' -c:v libx264 -preset medium -crf 16 -c:a aac -b:a 160k -ar 48000 -ac 2 -shortest $segment
    if ($LASTEXITCODE -ne 0) { throw "Video segment render failed for scene $number." }
    $segments += $segment
}

$concatPath = Join-Path $workRoot 'segments.txt'
$concatLines = $segments | ForEach-Object { "file '$($_.Replace("'", "''"))'" }
Set-Content -LiteralPath $concatPath -Value $concatLines -Encoding UTF8

& $ffmpeg -y -fflags +genpts -f concat -safe 0 -i $concatPath -c:v copy -c:a aac -b:a 160k -ar 48000 -ac 2 -af 'aresample=async=1:first_pts=0' -movflags +faststart $Output
if ($LASTEXITCODE -ne 0) { throw 'Final TANEBI narrated video assembly failed.' }

Write-Host "Rendered narrated TANEBI introduction: $Output"
