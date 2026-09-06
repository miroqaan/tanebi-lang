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

このディレクトリは初期Stage 1の保存版です。UEFIのブートサービスを終了した後、手書きRustカーネルがフレームバッファーへ描画し、PS/2ポートから入力を読み取ります。ビルド時の`system.tanebi`は統一CLIの`tanebi run`でネイティブ実行ファイルへコンパイルしてから実行し、出力マニフェストを組み込みます。旧インタープリターは使用しません。現在の[独立OSプロジェクト](https://github.com/miroqaan/tanebi-95)では、マニフェストだけでなくOSロジック自体をTANEBIでコンパイルします。Ring 3のプロセス分離やセルフホスティングを実装済みという意味ではありません。

Microsoft WindowsやMicrosoftの素材を含まない、独自の95風デスクトップです。
