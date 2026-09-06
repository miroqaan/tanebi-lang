# TANEBI（種火）

> 一行のコードが、世界に火を灯す。

TANEBIは、一行のコードから世界を作る過程を見せる、アカギユニバースのプログラミング言語プロジェクトです。Goで実装したフロントエンドがソースを解析・型検査し、Rustバックエンドを通してネイティブ実行ファイルへコンパイルします。

処理系は一つのコンパイラーに統一しました。インタープリターは使用しません。`tanebi run`も、ソースを直接評価するのではなく、ホスト用バイナリーをコンパイルしてから実行します。

## 設計原則

- **役割を分ける**：字句解析、構文解析、型検査、コード生成を分離します。
- **世界中の言葉で**：英語のキーワードと、日本語などのUnicode識別子を使えます。
- **外部との境界を明示する**：ホストの標準出力と、システム向けCPU時刻・メモリー・I/O操作を明示的な組み込み関数に限定します。
- **明示的に**：数値・文字列・真偽値を暗黙に型変換しません。
- **小さな処理系**：コンパイラーのフロントエンドは、外部Go依存なしでビルドできます。ネイティブ成果物の生成にはRustツールチェーンを使います。

## サンプル

```tanebi
fn main() {
    let 名前:str = "アカギ";
    let 種火:u64 = 3;
    let 回数:u64 = 0;
    while 回数 < 種火 {
        print(名前);
        println("の世界が目を覚ます");
        回数 = 回数 + 1;
    }
    if 種火 >= 3 {
        println("種火は消えない");
    } else {
        println("もう一度、火を灯す");
    }
}
```

## 検査・ビルド・実行

Go 1.22以上とRustツールチェーンが必要です。`check`はRustバイナリーを生成せず、TANEBIの名前と型を検査します。

```powershell
go run ./cmd/tanebi check examples/awakening.tanebi
go run ./cmd/tanebi run examples/awakening.tanebi
go run ./cmd/tanebi build -o awakening.exe examples/awakening.tanebi
.\awakening.exe
```

テストとビルド：

```powershell
go test ./...
go build ./cmd/tanebi
```

## 一つの言語、複数のコンパイル先

`fn`による関数、型付き変数、`const`、`global`、固定長`buffer`、`embed`、`while`、`if`、`return`、`break`、`continue`に対応します。型は`u64`、`bool`、静的UTF-8の`str`、戻り値なしの`void`です。`u64`の加減乗算は64ビットでラップし、ビット演算とシフトも使えます。文字列の`+`連結はありません。出力を分ける場合は`print`と`println`を組み合わせます。

処理の流れは、TANEBIソース → Lexer → AST → 名前・型の検証 → Rustバックエンド → ネイティブ実行ファイルです。RustやCのコードをTANEBIソースへ直接埋め込む機能はありません。

| ターゲット | 用途と入口 |
|---|---|
| `host`（既定） | 開発用PCの実行ファイル。入口は`fn main()`。`print`・`println`が使用可能 |
| `uefi` | x86-64 UEFIの実行ファイル。入口は`fn entry(image:u64,st:u64)->u64` |
| `library` | ホスト検証用Rustソース。`check`・`emit`で使用し、入口のラッパーは生成しない |

```powershell
go run ./cmd/tanebi check -target uefi examples/native/minimal.tanebi
go run ./cmd/tanebi build -target uefi -o BOOTX64.EFI examples/native/minimal.tanebi
go run ./cmd/tanebi emit -target uefi -o native.rs examples/native/minimal.tanebi
```

Rustの`x86_64-unknown-uefi`ターゲットが必要です。これは最小のUEFIプログラムを生成する例で、完成したデスクトップを含む例ではありません。

複数ソース、`-entry entry`、`-asset MEDIA=path.bin`、`-define TEST=1`も利用できます。UEFI用に生成された`efi_main`は、指定したTANEBIの入口を呼び出します。`emit -target library`では入口のラッパーを省略し、ホスト上で生成コードをテストできます。

CPUポートI/O、volatileなロード・ストア、コピー、ゼロクリア、CPU時刻、`efiapi`関数ポインター呼び出しが汎用組み込み関数です。ウィンドウ、描画、入力、メディア再生などのOSロジックは、これらを使ってTANEBI側に記述します。任意アドレスを扱うシステム言語であり、メモリー安全なサンドボックスではありません。

構文・組み込み関数・制限は[言語リファレンス](./docs/native-language.md)を参照してください。

## 旧構文からの移行

コンパイラー統一に伴う破壊的変更です。トップレベルの実行文は`fn main()`へ移し、`let 名前:str = "アカギ";`のように型を明示します。旧`repeat`はカウンター付き`while`へ、旧出力文は`print(value)`または`println(value)`へ変更します。旧インタープリター、別CLIの`tanebi-native`、Go/WASMブリッジは廃止しました。

## TANEBI 95 — 独立OS

現在の独立OSプロジェクトは、[miroqaan/tanebi-95](https://github.com/miroqaan/tanebi-95)です。QEMUのx86-64 UEFI環境から起動し、ファームウェアのブートサービスを終了してデスクトップを描画します。

同プロジェクトの`system/*.tanebi`に、UEFI呼び出し、起動処理、カーネル、デスクトップ、フォント、入力、動画プレイヤー、音声処理を分けて実装しています。本リポジトリのAOTコンパイラーがそれらをRustへ変換し、ネイティブ実行ファイルを生成します。OSロジックはTANEBI側にあり、Rust側にはその変換結果と、CPU・メモリー・ABI用の汎用組み込み関数を出力します。

Microsoft WindowsそのものやMicrosoftの素材を再配布するプロジェクトではありません。ビルド・起動方法と現在の制限は[OS側のREADME](https://github.com/miroqaan/tanebi-95/blob/main/README.md)を参照してください。

## プロジェクト構成

```text
cmd/tanebi/         check・build・run・emitの統一CLI
internal/native/    Lexer、AST、型検査、Rustコード生成
examples/           ホスト用コンパイル言語のサンプル
examples/native/    UEFI向けサンプル
docs/               言語と制作のドキュメント
apps/tanebi95/      旧Webデスクトップのソース
native/tanebi95/    初期ネイティブプレビューのソース
```

## 保存している初期実装

### 旧Webデモ

`apps/tanebi95/`は初期WebデスクトップのUIアーカイブです。Go/WASMインタープリターと配布バイナリーを削除し、起動設定は固定データ、Studioは型付きサンプルの閲覧専用に変更しました。Web UIのビルドは維持していますが、ブラウザー内でTANEBIコードを実行する機能はありません。独立OS本体とは別の旧実装です。

### 初期ネイティブプレビュー

`native/tanebi95/`には、TANEBIで起動マニフェストだけを生成し、手書きのRust `no_std`カーネルが描画していた初期Stage 1を保存しています。現在の独立プロジェクトが使用する、OSロジック自体をTANEBIで書くAOT構成とは異なります。

```powershell
.\native\tanebi95\scripts\build.ps1
.\native\tanebi95\scripts\run.ps1
```

境界と成果物の詳細は[`native/tanebi95/README.md`](./native/tanebi95/README.md)を参照してください。

既存の紹介動画は制作当時の実装を説明しています。生成スクリプトを更新しても、過去の動画や公開済み動画の内容は自動では更新されません。

## 今後の拡張

- 符号付き整数とユーザー定義型
- ListとMapを扱うスクリプト向け機能
- ゲームイベント・会話スクリプトAPI
- seedを明示する決定論的乱数
- エディター向けシンタックスハイライト
