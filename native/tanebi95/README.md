# TANEBI 95 Native

x86-64 UEFI環境から起動する、TANEBI 95のネイティブプレビューです。`gopher-os` Stage 1におけるRust `no_std`とUEFIの役割分担を参考に、ビルド時にTANEBIプログラムを実行してシステムマニフェストを生成します。カーネルはGOPからフレームバッファーのアドレスを取得し、`ExitBootServices`でファームウェアのブートサービスを終了してから、デスクトップを直接描画します。

## ビルド

```powershell
.\native\tanebi95\scripts\build.ps1
```

生成されるファイル：

- `build/native/esp/EFI/BOOT/BOOTX64.EFI`
- `build/native/tanebi95.img` — 32 MiBのFAT16形式UEFI起動イメージ

## 実行

QEMUとx86-64用のEDK2ファームウェアが必要です。

```powershell
.\native\tanebi95\scripts\run.ps1
```

キーボード操作：

- `S`：スタートメニュー
- `T`：TANEBI Studioウィンドウの開閉
- `Esc`：電源画面

## 現在の実装範囲

このバージョンは、UEFIから起動した後にファームウェアのブートサービスを終了する、ベアメタルのStage 1です。以降はフレームバッファーメモリーに直接書き込んで画面を描画し、PS/2のI/Oポートからキーボードのスキャンコードを読み取ります。TANEBIの言語コアは、まだRing 3のプロセスとしては動作しません。現時点では、同じリポジトリ内のGoインタープリターでビルド時に`system.tanebi`を実行し、その決定論的な出力をカーネルに組み込んでいます。Ring 3で動作するTANEBIランタイムは、ページテーブル、システムコール、プロセスローダーの実装後に取り組む段階です。

Microsoft WindowsやMicrosoftの素材を含まない、独自の95風デスクトップです。
