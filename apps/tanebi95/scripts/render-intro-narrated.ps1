[CmdletBinding()]
param(
    [string]$Output = (Join-Path (Split-Path -Parent (Split-Path -Parent (Split-Path -Parent $PSScriptRoot))) 'artifacts\videos\tanebi95-introduction-ja-narrated.mp4'),
    [string]$Voice = 'ja-JP-NanamiNeural'
)

$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Drawing

$appRoot = Split-Path -Parent $PSScriptRoot
$repositoryRoot = Split-Path -Parent (Split-Path -Parent $appRoot)
$workRoot = Join-Path $repositoryRoot 'artifacts\videos\tanebi95-intro'
$screenRoot = Join-Path $workRoot 'screens'
$frameRoot = Join-Path $workRoot 'frames'
$audioRoot = Join-Path $workRoot 'audio'
$segmentRoot = Join-Path $workRoot 'segments'
New-Item -ItemType Directory -Force -Path $frameRoot, $audioRoot, $segmentRoot, (Split-Path -Parent $Output) | Out-Null

$ffmpeg = (Get-Command ffmpeg -ErrorAction Stop).Source
$ffprobe = (Get-Command ffprobe -ErrorAction Stop).Source
$python = (Get-Command py -ErrorAction Stop).Source
$width = 1280
$height = 720

function New-Font([string]$family, [float]$size, [System.Drawing.FontStyle]$style = [System.Drawing.FontStyle]::Regular) {
    return [System.Drawing.Font]::new($family, $size, $style, [System.Drawing.GraphicsUnit]::Pixel)
}

function Get-CodeLines([string]$relativePath, [int]$start, [int]$end) {
    $all = Get-Content -LiteralPath (Join-Path $repositoryRoot $relativePath) -Encoding UTF8
    $result = @()
    for ($line = $start; $line -le $end -and $line -le $all.Count; $line++) {
        $result += [pscustomobject]@{ Number = $line; Text = $all[$line - 1] }
    }
    return $result
}

function New-ScreenFrame($scene, [int]$index, [int]$count) {
    $source = [System.Drawing.Image]::FromFile((Join-Path $screenRoot $scene.Screen))
    $bitmap = [System.Drawing.Bitmap]::new($width, $height)
    $graphics = [System.Drawing.Graphics]::FromImage($bitmap)
    $graphics.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
    $graphics.DrawImage($source, 0, 0, $width, $height)

    $banner = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(220, 7, 16, 31))
    $orange = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(255, 165, 44))
    $white = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::White)
    $muted = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(183, 201, 226))
    $eyebrowFont = New-Font 'Yu Gothic' 17 ([System.Drawing.FontStyle]::Bold)
    $titleFont = New-Font 'Yu Gothic' 33 ([System.Drawing.FontStyle]::Bold)
    $footerFont = New-Font 'Consolas' 14

    $graphics.FillRectangle($banner, 0, 0, $width, 104)
    $graphics.FillRectangle($orange, 0, 0, 10, 104)
    $graphics.DrawString($scene.Eyebrow, $eyebrowFont, $orange, 34, 16)
    $graphics.DrawString($scene.Title, $titleFont, $white, 30, 42)
    $graphics.DrawString(('AKAGI UNIVERSE  ·  TANEBI 95  ·  {0}/{1}' -f $index, $count), $footerFont, $muted, 930, 74)

    $path = Join-Path $frameRoot ('scene-{0:d2}.png' -f $index)
    $bitmap.Save($path, [System.Drawing.Imaging.ImageFormat]::Png)
    foreach ($resource in @($source, $banner, $orange, $white, $muted, $eyebrowFont, $titleFont, $footerFont, $graphics, $bitmap)) { $resource.Dispose() }
    return $path
}

function New-CodeFrame($scene, [int]$index, [int]$count) {
    $bitmap = [System.Drawing.Bitmap]::new($width, $height)
    $graphics = [System.Drawing.Graphics]::FromImage($bitmap)
    $graphics.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::AntiAlias
    $graphics.TextRenderingHint = [System.Drawing.Text.TextRenderingHint]::AntiAliasGridFit
    $graphics.Clear([System.Drawing.Color]::FromArgb(8, 14, 28))

    $white = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(244, 247, 255))
    $muted = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(154, 171, 203))
    $orange = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(238, 82, 45))
    $teal = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(50, 204, 186))
    $codePanel = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(16, 25, 47))
    $notePanel = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(22, 34, 60))
    $lineBrush = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(40, 55, 88))
    $border = [System.Drawing.Pen]::new([System.Drawing.Color]::FromArgb(55, 76, 119), 2)
    $eyebrowFont = New-Font 'Yu Gothic' 18 ([System.Drawing.FontStyle]::Bold)
    $titleFont = New-Font 'Yu Gothic' 34 ([System.Drawing.FontStyle]::Bold)
    $pathFont = New-Font 'Consolas' 15
    $codeFont = New-Font 'MS Gothic' 17
    $lineNumberFont = New-Font 'Consolas' 15
    $noteTitleFont = New-Font 'Yu Gothic' 19 ([System.Drawing.FontStyle]::Bold)
    $noteFont = New-Font 'Yu Gothic' 18
    $footerFont = New-Font 'Consolas' 14

    $graphics.FillRectangle($orange, 0, 0, 9, $height)
    $graphics.FillRectangle($teal, 9, 0, 4, $height)
    $graphics.DrawString($scene.Eyebrow, $eyebrowFont, $orange, 46, 25)
    $graphics.DrawString($scene.Title, $titleFont, $white, 43, 54)
    $graphics.FillRectangle($codePanel, 44, 124, 870, 520)
    $graphics.DrawRectangle($border, 44, 124, 870, 520)
    $graphics.FillRectangle($lineBrush, 44, 124, 870, 39)
    $graphics.DrawString($scene.Path, $pathFont, $muted, 62, 135)
    $graphics.SetClip([System.Drawing.Rectangle]::new(45, 164, 868, 479))
    $y = 178
    foreach ($line in $scene.Code) {
        if ($y -gt 615) { break }
        $graphics.DrawString(('{0,3}' -f $line.Number), $lineNumberFont, $muted, 59, $y)
        $brush = $white
        $trimmed = $line.Text.TrimStart()
        if ($trimmed.StartsWith('//') -or $trimmed.StartsWith('#')) { $brush = $muted }
        elseif ($trimmed -match '^(export|async|function|const|let|if|for|return|try|catch|switch|case)\b') { $brush = $teal }
        elseif ($line.Text.Contains('"') -or $line.Text.Contains("'")) { $brush = $orange }
        $graphics.DrawString($line.Text.Replace("`t", '    '), $codeFont, $brush, 108, $y - 1)
        $y += 23
    }
    $graphics.ResetClip()
    $graphics.FillRectangle($notePanel, 936, 124, 300, 520)
    $graphics.DrawRectangle($border, 936, 124, 300, 520)
    $graphics.DrawString('主な関数と技術', $noteTitleFont, $teal, 957, 150)
    $noteY = 210
    foreach ($note in $scene.Notes) {
        $graphics.FillEllipse($orange, 959, $noteY + 7, 8, 8)
        $graphics.DrawString($note, $noteFont, $white, [System.Drawing.RectangleF]::new(981, $noteY, 228, 82))
        $noteY += 103
    }
    $graphics.DrawString(('AKAGI UNIVERSE  ·  TANEBI 95  ·  {0}/{1}' -f $index, $count), $footerFont, $muted, 46, 680)

    $path = Join-Path $frameRoot ('scene-{0:d2}.png' -f $index)
    $bitmap.Save($path, [System.Drawing.Imaging.ImageFormat]::Png)
    foreach ($resource in @($white, $muted, $orange, $teal, $codePanel, $notePanel, $lineBrush, $border, $eyebrowFont, $titleFont, $pathFont, $codeFont, $lineNumberFont, $noteTitleFont, $noteFont, $footerFont, $graphics, $bitmap)) { $resource.Dispose() }
    return $path
}

$scenes = @(
    [pscustomobject]@{ Kind='screen'; Eyebrow='01 · DESKTOP'; Title='TANEBI 95 — 種火のユーザー空間'; Screen='01-desktop.png'; Narration='TANEBI 95。これはWindows 95の視覚言語を借りて、TANEBIの実行環境をデスクトップとして再構成した、ブラウザー上のユーザー空間です。タイトル、アイコン、タスクバーまでTANEBI 95として統一しています。' },
    [pscustomobject]@{ Kind='code'; Eyebrow='02 · WINDOW SYSTEM'; Title='Reactで作るデスクトップとウィンドウ'; Path='apps/tanebi95/components/tanebi95/desktop-shell.tsx'; Code=(Get-CodeLines 'apps\tanebi95\components\tanebi95\desktop-shell.tsx' 140 166); Notes=@('openWindow：起動と前面化','focusWindow：Z順制御','WindowFrame：移動・最大化'); Narration='デスクトップの中心はReactのDesktopShellです。openWindowがアプリを起動し、focusWindowが前面へ移動させます。各アプリはWindowFrameに入り、移動、最小化、最大化、タスクバー連携を同じ状態モデルで処理します。' },
    [pscustomobject]@{ Kind='screen'; Eyebrow='03 · TANEBI STUDIO'; Title='実際のTANEBIコードを、その場で実行'; Screen='02-studio.png'; Narration='TANEBI Studioでは、実際のTANEBIコードを編集して実行できます。日本語の識別子、repeat、ifを含むプログラムが、三回の出力と、種火は消えないという結果を返しました。これは画像だけの演出ではなく、本物の言語処理系です。' },
    [pscustomobject]@{ Kind='code'; Eyebrow='04 · GO / WEBASSEMBLY'; Title='Go製インタープリターをブラウザーへ'; Path='apps/tanebi95/lib/tanebi-runtime.ts'; Code=(Get-CodeLines 'apps\tanebi95\lib\tanebi-runtime.ts' 36 64); Notes=@('instantiateGoWasm','loadTanebi','tanebiRunブリッジ'); Narration='instantiateGoWasmはtanebi.wasmをストリーミングで読み込み、失敗時は通常のinstantiateへフォールバックします。loadTanebiはGoのブリッジを起動し、公開関数tanebiRunが準備できるまで待機します。同じGo製インタープリターを、CLIとブラウザーで共有する構成です。' },
    [pscustomobject]@{ Kind='screen'; Eyebrow='05 · SYSTEM MONITOR'; Title='実行中の技術スタックを見える化'; Screen='03-monitor.png'; Narration='システムモニターは、React、vinext、Go WebAssembly、Cloudflare Runtimeまでの流れを表示します。稼働時間とJavaScriptヒープも更新し、Studio、DOOM、メディアプレイヤーへ直接移動できる機能ツアーを備えています。' },
    [pscustomobject]@{ Kind='code'; Eyebrow='06 · MEDIA PLAYER'; Title='指定映像を16分14秒から再生'; Path='apps/tanebi95/components/tanebi95/media-player.tsx'; Code=(Get-CodeLines 'apps\tanebi95\components\tanebi95\media-player.tsx' 8 29); Notes=@('YouTube nocookie埋め込み','start=974秒','音声・全画面対応'); Narration='メディアプレイヤーは、指定されたアカギユニバースの映像を16分14秒、つまり974秒から開きます。youtube nocookieの埋め込みを使い、再生、音声調整、ピクチャーインピクチャー、全画面表示に対応します。' },
    [pscustomobject]@{ Kind='screen'; Eyebrow='07 · MEDIA PLAYBACK'; Title='アカギユニバース映像を音声付きで'; Screen='05-media-playing.png'; Narration='再生ボタンを押すと、プレイヤーがウィンドウの中に展開されます。外部のYouTubeページを開く導線も残し、TANEBI 95の中でも、通常のプレイヤーとしても利用できます。' },
    [pscustomobject]@{ Kind='code'; Eyebrow='08 · DOOM WINDOW'; Title='DOOMを独立プロセスのように隔離'; Path='apps/tanebi95/components/tanebi95/doom-player.tsx'; Code=(Get-CodeLines 'apps\tanebi95\components\tanebi95\doom-player.tsx' 40 70); Notes=@('start / stop','togglePause','sendGameAction'); Narration='DOOMウィンドウでは、startがゲーム用iframeを生成し、stopがフレームごと破棄します。togglePauseはメッセージで一時停止と再開を伝え、sendGameActionは前進と攻撃を送ります。ゲームを独立したプロセスのように隔離する設計です。' },
    [pscustomobject]@{ Kind='code'; Eyebrow='09 · DOOM ENGINE'; Title='Chocolate DoomをWebAssemblyで起動'; Path='apps/tanebi95/public/doom/player.js'; Code=(Get-CodeLines 'apps\tanebi95\public\doom\player.js' 21 48); Notes=@('4分割データを結合','createDoomModule','Freedoom Phase 2'); Narration='ゲーム本体は公開GPLエンジンのChocolate Doomです。28メガバイトのFreedoomデータを四つに分けて並列取得し、元のバイト列へ結合してcreateDoomModuleへ渡します。相互運用可能なWADとして、自由に配布できるFreedoom Phase 2を使います。' },
    [pscustomobject]@{ Kind='screen'; Eyebrow='10 · DOOM PLAY'; Title='MAP01をWebAssemblyで直接起動'; Screen='06-doom-start.png'; Narration='起動すると、Chocolate DoomはWebAssembly上でFreedoomのMAP01へ直接入ります。矢印キー、Control、Space、Shiftに加えて、ウィンドウ上の前進と攻撃ボタンでも操作できます。' },
    [pscustomobject]@{ Kind='screen'; Eyebrow='11 · GAME INPUT'; Title='前進 — 画面もゲーム状態も動く'; Screen='07-doom-move.png'; Narration='前進ボタンを連続して送ると、プレイヤーは通路の奥へ移動しました。画面転送のデモではなく、敵、マップ、入力、ゲームループを含むDOOMエンジンそのものが動いています。' },
    [pscustomobject]@{ Kind='screen'; Eyebrow='12 · GAMEPLAY'; Title='攻撃と一時停止もウィンドウから'; Screen='08-doom-fire.png'; Narration='攻撃入力と一時停止も、TANEBI 95のツールバーから制御できます。テスト時の音量はゼロに固定しましたが、通常起動では事前に音量を設定し、ゲーム音声も利用できます。' },
    [pscustomobject]@{ Kind='screen'; Eyebrow='13 · AKAGI UNIVERSE'; Title='一行のコードから、遊べる世界へ。'; Screen='01-desktop.png'; Narration='TANEBIのコード実行、Windows 95風のデスクトップ、システムモニター、映像、そしてDOOM。すべてのソースとライセンス表示を含めて公開リポジトリに保存しました。一行のコードから、遊べる世界へ。TANEBI 95です。' }
)

$segments = @()
for ($index = 0; $index -lt $scenes.Count; $index++) {
    $number = $index + 1
    $scene = $scenes[$index]
    $frame = if ($scene.Kind -eq 'screen') { New-ScreenFrame $scene $number $scenes.Count } else { New-CodeFrame $scene $number $scenes.Count }
    $audio = Join-Path $audioRoot ('scene-{0:d2}.mp3' -f $number)
    $segment = Join-Path $segmentRoot ('scene-{0:d2}.mp4' -f $number)
    & $python -m edge_tts --voice $Voice --rate=-3% --text $scene.Narration --write-media $audio
    if ($LASTEXITCODE -ne 0) { throw "Japanese narration failed for scene $number." }
    $durationText = & $ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 $audio
    $duration = [math]::Ceiling(([double]::Parse($durationText, [Globalization.CultureInfo]::InvariantCulture) + 0.45) * 100) / 100
    & $ffmpeg -y -loop 1 -i $frame -i $audio -t $duration -vf 'scale=1280:720,format=yuv420p' -af 'apad=pad_dur=0.45' -c:v libx264 -preset medium -crf 18 -c:a aac -b:a 160k -ar 48000 -ac 2 -shortest $segment
    if ($LASTEXITCODE -ne 0) { throw "Video segment render failed for scene $number." }
    $segments += $segment
}

$concatPath = Join-Path $workRoot 'segments.txt'
$concatLines = $segments | ForEach-Object { "file '$($_.Replace("'", "''"))'" }
Set-Content -LiteralPath $concatPath -Value $concatLines -Encoding UTF8
& $ffmpeg -y -fflags +genpts -f concat -safe 0 -i $concatPath -c:v copy -c:a aac -b:a 160k -ar 48000 -ac 2 -af 'aresample=async=1:first_pts=0' -movflags +faststart $Output
if ($LASTEXITCODE -ne 0) { throw 'Final TANEBI 95 narrated video assembly failed.' }
Write-Host "Rendered narrated TANEBI 95 introduction: $Output"
