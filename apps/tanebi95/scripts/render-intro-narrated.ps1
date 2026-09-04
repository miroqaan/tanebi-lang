[CmdletBinding()]
param(
    [string]$Output = (Join-Path (Split-Path -Parent (Split-Path -Parent (Split-Path -Parent $PSScriptRoot))) 'artifacts\videos\tanebi95-os-introduction-ja-1080p.mp4'),
    [string]$Voice = 'ja-JP-NanamiNeural'
)

$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Drawing

$appRoot = Split-Path -Parent $PSScriptRoot
$repositoryRoot = Split-Path -Parent (Split-Path -Parent $appRoot)
$workRoot = Join-Path $repositoryRoot 'artifacts\videos\tanebi95-os-intro'
$captureRoot = Join-Path $workRoot 'captures'
$decoratedRoot = Join-Path $workRoot 'decorated'
$audioRoot = Join-Path $workRoot 'audio'
$segmentRoot = Join-Path $workRoot 'segments'
$fps = 10.0
$width = 1920
$height = 1080

New-Item -ItemType Directory -Force -Path $decoratedRoot, $audioRoot, $segmentRoot, (Split-Path -Parent $Output) | Out-Null

$ffmpeg = (Get-Command ffmpeg -ErrorAction Stop).Source
$ffprobe = (Get-Command ffprobe -ErrorAction Stop).Source
$python = (Get-Command py -ErrorAction Stop).Source
$invariant = [Globalization.CultureInfo]::InvariantCulture

function New-Font([string]$family, [float]$size, [System.Drawing.FontStyle]$style = [System.Drawing.FontStyle]::Regular) {
    return [System.Drawing.Font]::new($family, $size, $style, [System.Drawing.GraphicsUnit]::Pixel)
}

function Get-Seconds([string]$path) {
    $value = & $ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 $path
    if ($LASTEXITCODE -ne 0) { throw "Could not read media duration: $path" }
    return [double]::Parse($value.Trim(), $invariant)
}

function Draw-RoundedRectangle($graphics, $brush, [float]$x, [float]$y, [float]$w, [float]$h, [float]$radius) {
    $path = [System.Drawing.Drawing2D.GraphicsPath]::new()
    $diameter = $radius * 2
    $path.AddArc($x, $y, $diameter, $diameter, 180, 90)
    $path.AddArc($x + $w - $diameter, $y, $diameter, $diameter, 270, 90)
    $path.AddArc($x + $w - $diameter, $y + $h - $diameter, $diameter, $diameter, 0, 90)
    $path.AddArc($x, $y + $h - $diameter, $diameter, $diameter, 90, 90)
    $path.CloseFigure()
    $graphics.FillPath($brush, $path)
    $path.Dispose()
}

function Add-SceneLabel($graphics, $scene, [int]$number, [double]$frameTime) {
    if ($frameTime -gt 2.4) { return }

    $alpha = if ($frameTime -lt 1.85) { 222 } else { [int](222 * ((2.4 - $frameTime) / 0.55)) }
    $panel = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb([math]::Max(0, $alpha), 8, 18, 38))
    $orange = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb([math]::Max(0, $alpha), 255, 170, 48))
    $white = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb([math]::Max(0, $alpha), 255, 255, 255))
    $muted = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb([math]::Max(0, $alpha), 188, 208, 235))
    $eyebrowFont = New-Font 'Yu Gothic' 19 ([System.Drawing.FontStyle]::Bold)
    $titleFont = New-Font 'Yu Gothic' 34 ([System.Drawing.FontStyle]::Bold)
    $subtitleFont = New-Font 'Yu Gothic' 19

    Draw-RoundedRectangle $graphics $panel 1080 850 800 158 18
    $graphics.FillRectangle($orange, 1080, 850, 9, 158)
    $graphics.DrawString(('TANEBI 95  ·  {0:d2}/07' -f $number), $eyebrowFont, $orange, 1116, 866)
    $graphics.DrawString($scene.Title, $titleFont, $white, 1112, 897)
    $graphics.DrawString($scene.Subtitle, $subtitleFont, $muted, 1115, 951)

    foreach ($resource in @($panel, $orange, $white, $muted, $eyebrowFont, $titleFont, $subtitleFont)) { $resource.Dispose() }
}

function Add-ClickMarker($graphics, $click, [double]$frameTime) {
    $elapsed = $frameTime - [double]$click.Time
    if ($elapsed -lt -0.22 -or $elapsed -gt 0.72) { return }

    $phase = [math]::Max(0.0, [math]::Min(1.0, ($elapsed + 0.22) / 0.94))
    $radius = [float](20 + (44 * $phase))
    $alpha = [int](235 * (1.0 - (0.72 * $phase)))
    $ring = [System.Drawing.Pen]::new([System.Drawing.Color]::FromArgb([math]::Max(45, $alpha), 255, 169, 40), 6)
    $graphics.DrawEllipse($ring, [float]$click.X - $radius, [float]$click.Y - $radius, $radius * 2, $radius * 2)

    $cursorPath = [System.Drawing.Drawing2D.GraphicsPath]::new()
    $cursorPath.AddPolygon([System.Drawing.PointF[]]@(
        [System.Drawing.PointF]::new([float]$click.X, [float]$click.Y),
        [System.Drawing.PointF]::new([float]$click.X + 15, [float]$click.Y + 43),
        [System.Drawing.PointF]::new([float]$click.X + 24, [float]$click.Y + 28),
        [System.Drawing.PointF]::new([float]$click.X + 42, [float]$click.Y + 30)
    ))
    $cursorFill = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::White)
    $cursorOutline = [System.Drawing.Pen]::new([System.Drawing.Color]::FromArgb(220, 0, 0, 0), 3)
    $graphics.FillPath($cursorFill, $cursorPath)
    $graphics.DrawPath($cursorOutline, $cursorPath)

    if ($elapsed -ge -0.05 -and $elapsed -le 0.5) {
        $labelFont = New-Font 'Yu Gothic' 18 ([System.Drawing.FontStyle]::Bold)
        $labelSize = $graphics.MeasureString($click.Label, $labelFont)
        $labelX = [math]::Min($width - $labelSize.Width - 42, [float]$click.X + 52)
        $labelY = [math]::Min($height - 56, [float]$click.Y + 34)
        $labelPanel = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(218, 7, 17, 34))
        $labelBrush = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::White)
        Draw-RoundedRectangle $graphics $labelPanel $labelX $labelY ($labelSize.Width + 28) 42 10
        $graphics.DrawString($click.Label, $labelFont, $labelBrush, $labelX + 14, $labelY + 7)
        foreach ($resource in @($labelFont, $labelPanel, $labelBrush)) { $resource.Dispose() }
    }

    foreach ($resource in @($ring, $cursorPath, $cursorFill, $cursorOutline)) { $resource.Dispose() }
}

function New-DecoratedCapture($scene, [int]$number) {
    $sourceDir = Join-Path $captureRoot $scene.Capture
    $destinationDir = Join-Path $decoratedRoot $scene.Capture
    if (-not (Test-Path -LiteralPath $sourceDir)) { throw "Missing recorded scene: $sourceDir" }
    New-Item -ItemType Directory -Force -Path $destinationDir | Out-Null

    $frames = @(Get-ChildItem -LiteralPath $sourceDir -Filter 'frame-*.png' | Sort-Object Name)
    if ($frames.Count -eq 0) { throw "Recorded scene has no frames: $sourceDir" }

    for ($index = 0; $index -lt $frames.Count; $index++) {
        $frameTime = $index / $fps
        $hasLabel = $frameTime -le 2.4
        $nearClick = $false
        foreach ($click in $scene.Clicks) {
            if ($frameTime -ge ([double]$click.Time - 0.22) -and $frameTime -le ([double]$click.Time + 0.72)) {
                $nearClick = $true
                break
            }
        }

        $destination = Join-Path $destinationDir $frames[$index].Name
        # Browser captures may contain JPEG bytes under a .png filename. Re-save
        # every frame so ffmpeg receives a homogeneous, standards-compliant PNG sequence.
        $source = [System.Drawing.Image]::FromFile($frames[$index].FullName)
        $bitmap = [System.Drawing.Bitmap]::new($width, $height)
        $graphics = [System.Drawing.Graphics]::FromImage($bitmap)
        $graphics.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::AntiAlias
        $graphics.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
        $graphics.TextRenderingHint = [System.Drawing.Text.TextRenderingHint]::AntiAliasGridFit
        $graphics.DrawImage($source, 0, 0, $width, $height)

        Add-SceneLabel $graphics $scene $number $frameTime
        foreach ($click in $scene.Clicks) { Add-ClickMarker $graphics $click $frameTime }

        $bitmap.Save($destination, [System.Drawing.Imaging.ImageFormat]::Png)
        foreach ($resource in @($source, $graphics, $bitmap)) { $resource.Dispose() }
    }

    return [pscustomobject]@{ Directory = $destinationDir; FrameCount = $frames.Count }
}

$scenes = @(
    [pscustomobject]@{
        Capture='01-boot'; Title='起動シーケンス'; Subtitle='Go/WASMユーザー空間を準備'; Speed=1.20
        Narration='TANEBI 95。種火の実行環境を、実際に触って使えるデスクトップとして再構成しました。起動シーケンスが完了すると、すぐにユーザー空間へ入ります。'
        Clicks=@()
    },
    [pscustomobject]@{
        Capture='02-shell-b'; Title='デスクトップとウィンドウ'; Subtitle='スタート・タスクバー・最小化・最大化'; Speed=1.45
        Narration='左下のスタートボタンからアプリを開きます。開いたウィンドウは最小化でき、タスクバーのボタンから元に戻せます。最大化と通常表示の切り替えも、デスクトップ上でそのまま操作できます。'
        Clicks=@(
            [pscustomobject]@{Time=0.5; X=42; Y=1050; Label='スタート'},
            [pscustomobject]@{Time=1.7; X=160; Y=932; Label='TANEBI 95について'},
            [pscustomobject]@{Time=3.0; X=795; Y=141; Label='最小化'},
            [pscustomobject]@{Time=4.3; X=325; Y=1050; Label='タスクバーから復元'},
            [pscustomobject]@{Time=5.6; X=820; Y=141; Label='最大化'},
            [pscustomobject]@{Time=6.9; X=1866; Y=21; Label='通常表示'}
        )
    },
    [pscustomobject]@{
        Capture='03-studio'; Title='TANEBI Studio'; Subtitle='言語を実行し、結果をウィンドウに表示'; Speed=1.40
        Narration='TANEBI Studioをダブルクリックすると、言語の実行画面が開きます。実行ボタンを押すと、TANEBIプログラムの結果がすぐに表示されます。作業に合わせてウィンドウを最大化することもできます。'
        Clicks=@(
            [pscustomobject]@{Time=0.5; X=64; Y=124; Label='TANEBI Studio'},
            [pscustomobject]@{Time=2.3; X=310; Y=158; Label='実行'},
            [pscustomobject]@{Time=5.2; X=1000; Y=114; Label='最大化'}
        )
    },
    [pscustomobject]@{
        Capture='04-monitor'; Title='システムモニター'; Subtitle='稼働時間・メモリー・構成をリアルタイム表示'; Speed=1.35
        Narration='システムモニターでは、稼働時間、メモリー使用量、WebAssemblyの準備状態を確認できます。デスクトップ、Studio、メディア、DOOMまで、搭載機能の状態を一つの画面で見渡せます。'
        Clicks=@(
            [pscustomobject]@{Time=0.5; X=65; Y=377; Label='システムモニター'},
            [pscustomobject]@{Time=2.6; X=992; Y=109; Label='最大化'},
            [pscustomobject]@{Time=5.1; X=1865; Y=21; Label='通常表示'}
        )
    },
    [pscustomobject]@{
        Capture='05-media'; Title='メディアプレイヤー'; Subtitle='指定映像を16分14秒から再生'; Speed=1.00
        Narration='メディアプレイヤーを開いて最大化します。再生ボタンを押すと、指定されたアカギユニバースの映像が16分14秒から実際に始まります。通常利用では音量調整と全画面表示にも対応します。'
        Clicks=@(
            [pscustomobject]@{Time=0.5; X=65; Y=293; Label='メディア'},
            [pscustomobject]@{Time=1.8; X=1030; Y=72; Label='最大化'},
            [pscustomobject]@{Time=3.2; X=78; Y=82; Label='16:14から再生'}
        )
    },
    [pscustomobject]@{
        Capture='06-doom'; Title='DOOMを実際にプレイ'; Subtitle='起動・移動・攻撃・一時停止を実操作'; Speed=1.00
        Narration='DOOMアイコンからゲームを起動します。Freedoomのマップが読み込まれ、前進すると視点が本当に移動します。攻撃、ゲームの一時停止、再開もウィンドウ上のボタンから操作できます。これは画像の切り替えではなく、実際に動くWebAssembly版DOOMです。収録時のゲーム音量はゼロにしています。'
        Clicks=@(
            [pscustomobject]@{Time=0.5; X=65; Y=208; Label='DOOM'},
            [pscustomobject]@{Time=1.8; X=1032; Y=71; Label='最大化'},
            [pscustomobject]@{Time=3.2; X=52; Y=57; Label='DOOMを起動'},
            [pscustomobject]@{Time=16.0; X=277; Y=57; Label='前進'},
            [pscustomobject]@{Time=19.0; X=350; Y=57; Label='攻撃'},
            [pscustomobject]@{Time=21.5; X=132; Y=57; Label='一時停止'},
            [pscustomobject]@{Time=23.5; X=132; Y=57; Label='再開'},
            [pscustomobject]@{Time=27.0; X=350; Y=57; Label='攻撃'}
        )
    },
    [pscustomobject]@{
        Capture='07-shutdown'; Title='終了と再起動'; Subtitle='スタートメニューから安全に切り替え'; Speed=1.25
        Narration='最後はスタートメニューから終了を選びます。終了画面の再起動ボタンを押すと、デスクトップへ戻ります。起動からアプリ、ゲーム、メディア、終了まで、これが実際に操作できるTANEBI 95です。'
        Clicks=@(
            [pscustomobject]@{Time=0.5; X=42; Y=1050; Label='スタート'},
            [pscustomobject]@{Time=1.7; X=150; Y=1006; Label='終了'},
            [pscustomobject]@{Time=4.5; X=960; Y=626; Label='再起動'}
        )
    }
)

$segments = @()
for ($index = 0; $index -lt $scenes.Count; $index++) {
    $number = $index + 1
    $scene = $scenes[$index]
    Write-Host ('[{0}/7] Decorating real interaction capture: {1}' -f $number, $scene.Title)
    $capture = New-DecoratedCapture $scene $number

    $audio = Join-Path $audioRoot ('scene-{0:d2}.mp3' -f $number)
    & $python -m edge_tts --voice $Voice --rate=-2% --text $scene.Narration --write-media $audio
    if ($LASTEXITCODE -ne 0) { throw "Japanese narration failed for scene $number." }

    $audioDuration = Get-Seconds $audio
    $videoDuration = ($capture.FrameCount / $fps) * [double]$scene.Speed
    $finalDuration = [math]::Max($videoDuration, $audioDuration + 0.55)
    $videoPad = [math]::Max(0.0, $finalDuration - $videoDuration)
    $audioPad = [math]::Max(0.0, $finalDuration - $audioDuration)
    $speedText = ([double]$scene.Speed).ToString('0.###', $invariant)
    $videoPadText = $videoPad.ToString('0.###', $invariant)
    $audioPadText = $audioPad.ToString('0.###', $invariant)
    $durationText = $finalDuration.ToString('0.###', $invariant)
    $inputPattern = Join-Path $capture.Directory 'frame-%04d.png'
    $segment = Join-Path $segmentRoot ('scene-{0:d2}.mp4' -f $number)
    $videoFilter = "setpts=$speedText*PTS,fps=30,tpad=stop_mode=clone:stop_duration=$videoPadText,format=yuv420p"
    $audioFilter = "apad=pad_dur=$audioPadText"

    & $ffmpeg -y -hide_banner -loglevel warning -framerate 10 -start_number 1 -i $inputPattern -i $audio -t $durationText -vf $videoFilter -af $audioFilter -c:v libx264 -preset slow -crf 14 -pix_fmt yuv420p -r 30 -c:a aac -b:a 192k -ar 48000 -ac 2 $segment
    if ($LASTEXITCODE -ne 0) { throw "Video segment render failed for scene $number." }
    $segments += $segment
}

$concatPath = Join-Path $workRoot 'segments.txt'
$concatLines = $segments | ForEach-Object { "file '$($_.Replace('\', '/').Replace("'", "''"))'" }
[IO.File]::WriteAllLines($concatPath, $concatLines, [Text.UTF8Encoding]::new($false))
& $ffmpeg -y -hide_banner -loglevel warning -fflags +genpts -f concat -safe 0 -i $concatPath -c:v copy -c:a aac -b:a 192k -ar 48000 -ac 2 -af 'aresample=async=1:first_pts=0' -movflags +faststart $Output
if ($LASTEXITCODE -ne 0) { throw 'Final TANEBI 95 operating-system introduction assembly failed.' }

$finalDuration = Get-Seconds $Output
Write-Host ('Rendered TANEBI 95 OS introduction: {0} ({1:N1}s, 1920x1080)' -f $Output, $finalDuration)
