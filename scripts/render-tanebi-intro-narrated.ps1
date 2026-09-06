[CmdletBinding()]
param(
    [string]$Output = (Join-Path (Split-Path -Parent $PSScriptRoot) 'artifacts\videos\tanebi-introduction-ja-compiler-native-1440p.mp4'),
    [string]$Voice = 'ja-JP-NanamiNeural',
    [switch]$FramesOnly
)

$ErrorActionPreference = 'Stop'
if ($Voice -notmatch '^ja-JP-') { throw 'This introduction requires a Japanese narration voice.' }
Add-Type -AssemblyName System.Drawing

$repositoryRoot = Split-Path -Parent $PSScriptRoot
$workRoot = Join-Path $repositoryRoot 'artifacts\videos\tanebi-intro-compiler-native'
$frameRoot = Join-Path $workRoot 'frames'
$audioRoot = Join-Path $workRoot 'audio'
$segmentRoot = Join-Path $workRoot 'segments'
New-Item -ItemType Directory -Force -Path $frameRoot, $audioRoot, $segmentRoot, (Split-Path -Parent $Output) | Out-Null

$ffmpeg = (Get-Command ffmpeg -ErrorAction Stop).Source
$ffprobe = (Get-Command ffprobe -ErrorAction Stop).Source
$python = (Get-Command py -ErrorAction Stop).Source
$width = 1280
$height = 720
$renderWidth = 2560
$renderHeight = 1440
$renderScale = $renderWidth / $width

function Get-CodeLines([string]$relativePath, [int]$start, [int]$end) {
    $all = Get-Content -LiteralPath (Join-Path $repositoryRoot $relativePath) -Encoding UTF8
    $result = @()
    for ($line = $start; $line -le $end -and $line -le $all.Count; $line++) {
        $result += [pscustomobject]@{ Number = $line; Text = $all[$line - 1] }
    }
    return $result
}

function Get-CodeExcerpt([string]$relativePath, [string]$pattern, [int]$count) {
    $all = Get-Content -LiteralPath (Join-Path $repositoryRoot $relativePath) -Encoding UTF8
    for ($index = 0; $index -lt $all.Count; $index++) {
        if ($all[$index] -match $pattern) {
            return Get-CodeLines $relativePath ($index + 1) ($index + $count)
        }
    }
    throw "Code excerpt anchor not found: $relativePath / $pattern"
}

function New-Font([string]$family, [float]$size, [System.Drawing.FontStyle]$style = [System.Drawing.FontStyle]::Regular) {
    return [System.Drawing.Font]::new($family, $size, $style, [System.Drawing.GraphicsUnit]::Pixel)
}

function Get-WrappedCode($graphics, $font, $lines) {
    $rows = @()
    foreach ($line in $lines) {
        $remaining = $line.Text.Replace("`t", '    ')
        $number = $line.Number
        do {
            $length = $remaining.Length
            while ($length -gt 0 -and $graphics.MeasureString($remaining.Substring(0, $length), $font).Width -gt 710) {
                $length--
            }
            if ($length -lt $remaining.Length) {
                $space = $remaining.LastIndexOf(' ', [math]::Max(0, $length - 1))
                if ($space -gt ($length / 2)) { $length = $space }
            }
            $rows += [pscustomobject]@{ Number = $number; Text = $remaining.Substring(0, $length) }
            $remaining = $remaining.Substring($length)
            if ($remaining.Length -gt 0) { $remaining = '    ' + $remaining.TrimStart() }
            $number = ''
        } while ($remaining.Length -gt 0)
    }
    if ($rows.Count -gt 18) { throw "Code excerpt has $($rows.Count) visual rows; shorten it instead of clipping." }
    return $rows
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
    $codeFont = New-Font 'Consolas' 19
    $lineNumberFont = New-Font 'Consolas' 18
    $noteTitleFont = New-Font 'Yu Gothic' 21 ([System.Drawing.FontStyle]::Bold)
    $noteFont = New-Font 'Yu Gothic' 20
    $footerFont = New-Font 'Yu Gothic' 15
    $outputFont = New-Font 'MS Gothic' 17

    $graphics.FillRectangle($orange, 0, 0, 10, $height)
    $graphics.FillRectangle($teal, 10, 0, 4, $height)
    $graphics.DrawString($scene.Eyebrow, $eyebrowFont, $orange, 52, 34)
    $graphics.DrawString($scene.Title, $titleFont, $white, 48, 65)
    Draw-Flame $graphics 1162 32 0.56

    $graphics.FillRectangle($codePanel, 48, 132, 805, 516)
    $graphics.DrawRectangle($border, 48, 132, 805, 516)
    $graphics.FillRectangle($lineBrush, 48, 132, 805, 42)
    $graphics.DrawString($scene.Path, $pathFont, $muted, 68, 143)

    $y = 191
    foreach ($line in (Get-WrappedCode $graphics $codeFont $scene.Code)) {
        $graphics.DrawString(('{0,3}' -f $line.Number), $lineNumberFont, $muted, 64, $y)
        $brush = $white
        $trimmed = $line.Text.TrimStart()
        if ($trimmed.StartsWith('//') -or $trimmed.StartsWith('#')) { $brush = $muted }
        elseif ($trimmed -match '^(func|fn|type|const|var|if|for|while|switch|case|return|let|print|println|else)\b') { $brush = $teal }
        elseif ($line.Text -match '"') { $brush = $orange }
        $graphics.DrawString($line.Text, $codeFont, $brush, 120, $y - 1)
        $y += 25
    }

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
        if ($scene.Output) { $noteY += 62 } else { $noteY += 94 }
    }

    if ($scene.Output) {
        $graphics.DrawString('コンパイル後の実行結果', $noteTitleFont, $teal, 900, 425)
        $outputY = 470
        foreach ($outputLine in $scene.Output) {
            if ($graphics.MeasureString($outputLine, $outputFont).Width -gt 310) {
                throw "Output line does not fit: $outputLine"
            }
            $graphics.DrawString($outputLine, $outputFont, $white, 900, $outputY)
            $outputY += 27
        }
    }

    $graphics.DrawString(('AKAGI UNIVERSE  ·  TANEBI  ·  {0}/{1}' -f $index, $count), $footerFont, $muted, 50, 677)

    $path = Join-Path $frameRoot ('scene-{0:d2}.png' -f $index)
    $bitmap.Save($path, [System.Drawing.Imaging.ImageFormat]::Png)
    foreach ($resource in @($eyebrowFont, $titleFont, $pathFont, $codeFont, $lineNumberFont, $noteTitleFont, $noteFont, $footerFont, $outputFont, $white, $muted, $orange, $teal, $codePanel, $notePanel, $lineBrush, $border, $graphics, $bitmap)) {
        $resource.Dispose()
    }
    return $path
}

$goCommand = Get-Command go -ErrorAction SilentlyContinue
$go = if ($goCommand) { $goCommand.Source } else { Join-Path $env:USERPROFILE 'sdk\go1.27.1\bin\go.exe' }
Push-Location $repositoryRoot
try {
    $sampleOutput = @(& $go run ./cmd/tanebi run examples/awakening.tanebi)
    if ($LASTEXITCODE -ne 0) { throw 'Sample execution failed.' }
}
finally { Pop-Location }
if (($sampleOutput -join "`n") -cne "アカギの世界が目を覚ます`nアカギの世界が目を覚ます`nアカギの世界が目を覚ます`n種火は消えない") {
    throw 'Sample output no longer matches the narration.'
}

$scenes = @(
    [pscustomobject]@{
        Eyebrow = '01 · 言語の紹介'
        Title = '小さな言語、TANEBI'
        Path = 'examples/awakening.tanebi'
        Code = Get-CodeLines 'examples\awakening.tanebi' 2 16
        Notes = @('Go製フロントエンド', 'Rust経由でネイティブへ', '日本語の変数名')
        Narration = 'タネビは、Goで作ったコンパイラーを持つプログラミング言語です。型を確かめたソースを、ネイティブ実行ファイルへ変換します。実装を見ながら、その仕組みをたどってみましょう。'
    },
    [pscustomobject]@{
        Eyebrow = '02 · コマンドライン'
        Title = '一つのCLIで検査から実行まで'
        Path = 'cmd/tanebi/main.go  |  main（抜粋）'
        Code = Get-CodeExcerpt 'cmd\tanebi\main.go' '^func main\(' 9
        Notes = @('checkで静的検証', 'buildで実行ファイルを生成', 'runもコンパイルして実行')
        Narration = '入口は、統一されたコマンドラインです。チェックで検査し、ビルドで実行ファイルを生成します。ランを選んだ場合も、まずコンパイルし、できたバイナリーを実行します。'
    },
    [pscustomobject]@{
        Eyebrow = '03 · コンパイルの流れ'
        Title = 'ソースからネイティブ実行ファイルへ'
        Path = 'internal/native/compiler.go  |  Compile（抜粋）'
        Code = Get-CodeExcerpt 'internal\native\compiler.go' '^func Compile\(' 9
        Notes = @('字句解析・構文解析', '名前と型を検証', 'Rustコードを生成')
        Narration = 'コンパイル関数は、複数のソースを一つのプログラムとして扱います。構文を読み取り、名前や型を検査してから、Rustのコードを生成します。処理ごとに役割を分けた構成です。'
    },
    [pscustomobject]@{
        Eyebrow = '04 · 字句解析'
        Title = '文字をトークンに分ける'
        Path = 'internal/native/syntax.go  |  lex（抜粋）'
        Code = Get-CodeExcerpt 'internal\native\syntax.go' '^\s*if unicode.IsLetter\(c\)' 8
        Notes = @('lex：文字列をトークンへ', 'Unicodeの文字を判定', '名前と位置を記録')
        Narration = '字句解析のレックス関数は、文字列を意味のある単位に分けます。ここでは文字や数字を読み取り、一つの名前として記録しています。日本語の変数名も、同じ仕組みで扱います。'
    },
    [pscustomobject]@{
        Eyebrow = '05 · 構文解析'
        Title = '文の構造を組み立てる'
        Path = 'internal/native/syntax.go  |  statement（抜粋）'
        Code = Get-CodeExcerpt 'internal\native\syntax.go' '^\s*case "let":' 14
        Notes = @('statement：文を解析', '型付きletを読み取る', 'expression：式を解析')
        Narration = 'ステートメント関数は、変数宣言や条件分岐、繰り返しなどの文を解析します。レットでは変数名だけでなく、型も読み取ります。その後の式と組み合わせて、宣言の構造を作ります。'
    },
    [pscustomobject]@{
        Eyebrow = '06 · コード生成'
        Title = '型を確かめ、Rustへ変換する'
        Path = 'internal/native/emit.go  |  emitExpr（抜粋）'
        Code = Get-CodeExcerpt 'internal\native\emit.go' '^\s*case "binary":' 12
        Notes = @('型検査後にコード生成', 'emitExprで式を変換', 'rustcで実行ファイルへ')
        Narration = '式のコード生成では、型の検査を通った式をRustへ変換します。数値の加算などには、64ビットのラップ演算を使います。その後、Rustのコンパイラーがネイティブ実行ファイルを生成します。'
    },
    [pscustomobject]@{
        Eyebrow = '07 · 実際のプログラム'
        Title = 'サンプルを実行する'
        Path = 'examples/awakening.tanebi'
        Code = Get-CodeLines 'examples\awakening.tanebi' 2 16
        Notes = @('whileで3回出力', 'ifで条件を判定', 'mainから実行')
        Output = $sampleOutput
        Narration = 'サンプルをコンパイルして実行してみましょう。メイン関数から始まり、ワイルで同じメッセージを三回出力します。その後はイフで条件を判定し、最後に、種火は消えない、と表示されました。'
    },
    [pscustomobject]@{
        Eyebrow = '08 · OSへの応用'
        Title = '同じコンパイラーで独立OSへ'
        Path = 'examples/native/minimal.tanebi  |  entry'
        Code = Get-CodeExcerpt 'examples\native\minimal.tanebi' '^fn entry\(' 5
        Notes = @('-target uefiでビルド', '最小エントリーポイントの例', 'OSのロジックもTANEBIで記述')
        Narration = '同じコンパイラーで、UEFI向けの実行ファイルも作れます。これは、入口となる関数の最小の例です。独立OSのタネビ95も、デスクトップや入力などのロジックをタネビで記述しています。'
    },
    [pscustomobject]@{
        Eyebrow = '09 · ここから始める'
        Title = 'コードを手に取り、試す'
        Path = 'ターミナル  |  実行コマンド'
        Code = @(
            [pscustomobject]@{ Number = ''; Text = 'git clone https://github.com/miroqaan/tanebi-lang.git' }
            [pscustomobject]@{ Number = ''; Text = 'cd tanebi-lang' }
            [pscustomobject]@{ Number = ''; Text = '' }
            [pscustomobject]@{ Number = ''; Text = 'go run ./cmd/tanebi check examples/awakening.tanebi' }
            [pscustomobject]@{ Number = ''; Text = 'go run ./cmd/tanebi run examples/awakening.tanebi' }
            [pscustomobject]@{ Number = ''; Text = 'go test ./...' }
            [pscustomobject]@{ Number = ''; Text = '' }
            [pscustomobject]@{ Number = ''; Text = 'https://github.com/miroqaan/tanebi-lang' }
        )
        Notes = @('ソースコードを公開', '短いサンプルから試す', '一行のコードから、世界へ')
        Narration = 'ソースコードとサンプルは、ギットハブで公開しています。まずは短いプログラムを動かし、気になる関数から読んでみてください。一行のコードから、世界へ。'
    }
)

$sceneJson = $scenes | ConvertTo-Json -Depth 8
if ($sceneJson -match '[\u1100-\u11FF\u3130-\u318F\uA960-\uA97F\uAC00-\uD7FF\uFFA0-\uFFDC]|WebAssembly|WASM|ブラウザ|JavaScript|&#x20;|韓国語') {
    throw 'Unexpected content language or excluded topic in the video.'
}
[IO.File]::WriteAllText((Join-Path $workRoot 'scenes.json'), $sceneJson, [Text.UTF8Encoding]::new($false))

$segments = @()
$timeline = @()
$startTime = 0.0
for ($index = 0; $index -lt $scenes.Count; $index++) {
    $number = $index + 1
    $scene = $scenes[$index]
    $frame = New-SceneFrame $scene $number $scenes.Count
    if ($FramesOnly) { continue }
    $audio = Join-Path $audioRoot ('scene-{0:d2}.mp3' -f $number)
    $segment = Join-Path $segmentRoot ('scene-{0:d2}.mp4' -f $number)

    & $python -3.14 -m edge_tts --voice $Voice --rate=-3% --text $scene.Narration --write-media $audio
    if ($LASTEXITCODE -ne 0) { throw "Japanese narration failed for scene $number." }

    $durationText = & $ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 $audio
    $duration = [math]::Ceiling(([double]::Parse($durationText, [Globalization.CultureInfo]::InvariantCulture) + 0.65) * 30) / 30
    & $ffmpeg -hide_banner -loglevel error -y -loop 1 -framerate 30 -i $frame -i $audio -t $duration -vf "format=yuv420p" -af 'loudnorm=I=-18:TP=-2:LRA=7,apad=pad_dur=0.65' -c:v libx264 -preset fast -tune stillimage -crf 12 -colorspace bt709 -color_primaries bt709 -color_trc bt709 -c:a aac -b:a 192k -ar 48000 -ac 2 -shortest $segment
    if ($LASTEXITCODE -ne 0) { throw "Video segment render failed for scene $number." }
    $segments += $segment
    $timeline += [pscustomobject]@{ Scene = $number; Title = $scene.Title; Start = $startTime; Duration = $duration; Frame = $frame; Audio = $audio; Segment = $segment; Narration = $scene.Narration }
    $startTime += $duration
    Write-Host ("[ok] Scene {0}/{1}: {2:N2}s" -f $number, $scenes.Count, $duration)
}

if ($FramesOnly) { Write-Host "Rendered all scene frames: $frameRoot"; return }
[IO.File]::WriteAllText((Join-Path $workRoot 'timeline.json'), ($timeline | ConvertTo-Json -Depth 8), [Text.UTF8Encoding]::new($false))
$concatPath = Join-Path $workRoot 'segments.txt'
$concatLines = $segments | ForEach-Object { "file '$($_.Replace("'", "''"))'" }
Set-Content -LiteralPath $concatPath -Value $concatLines -Encoding UTF8

& $ffmpeg -hide_banner -loglevel error -y -fflags +genpts -f concat -safe 0 -i $concatPath -c:v copy -c:a aac -b:a 192k -ar 48000 -ac 2 -af 'aresample=async=1:first_pts=0' -movflags +faststart $Output
if ($LASTEXITCODE -ne 0) { throw 'Final TANEBI narrated video assembly failed.' }

Write-Host "Rendered narrated TANEBI introduction: $Output"
