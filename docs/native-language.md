# TANEBI コンパイル言語リファレンス

TANEBIは明示的な型を持つAOTコンパイル言語です。Goフロントエンドが字句解析・構文解析・名前と型の検証を行い、Rustバックエンドを通してネイティブ実行ファイルを生成します。ホスト用プログラムとUEFI用プログラムを同じ言語・CLIで扱います。TANEBIソース内にRustやCを直接埋め込む機能も、ソースを逐次評価するインタープリターもありません。

## コンパイル

```powershell
go run ./cmd/tanebi check examples/awakening.tanebi
go run ./cmd/tanebi run examples/awakening.tanebi
go run ./cmd/tanebi build -o awakening.exe examples/awakening.tanebi
go run ./cmd/tanebi check -target uefi examples/native/minimal.tanebi
go run ./cmd/tanebi build -target uefi -o BOOTX64.EFI examples/native/minimal.tanebi
go run ./cmd/tanebi emit -target uefi -o native.rs examples/native/minimal.tanebi
```

複数ファイルを同時に渡せます。関数と定数はファイルをまたいで前方参照できます。`-asset MEDIA=path.bin` で明示したファイルを `embed MEDIA;` に対応付け、`-define TEST=1` を `config("TEST")` から参照します。指定のない設定値は 0 です。オプションはソースファイルより前に指定します。

`check`はTANEBIの静的検証、`build`は実行ファイルの生成、`run`はホスト用バイナリーをコンパイルして実行、`emit`はRustソースの出力です。既定ターゲットは`host`です。`build`は`host`と`uefi`、`check`と`emit`は追加で`library`を受け付けます。`run`はホスト専用です。Goに加えて、`build`と`run`にはRustツールチェーンが必要です。UEFI用ビルドには`x86_64-unknown-uefi`ターゲットも必要です。

ホストの入口は引数なし・戻り値なしの`fn main()`です。UEFIの入口は`fn entry(image:u64,st:u64)->u64`で、`-entry`で名前を指定できます。UEFI出力にはRustの`no_std` / `no_main`を使用します。

`emit -target library`はエントリーポイントとpanicハンドラーを生成せず、ホスト上のRustテストで使用できるソースを出力します。生成関数は`pub unsafe fn tn_name`、グローバル変数は`pub static mut tn_name`です。通常のASCII名には接頭辞`tn_`を付けます。日本語などの名前と、`__`で始まる予約領域のASCII名は、相互に衝突しないUTF-8の十六進表現へ変換します。

旧`tanebi-native` CLIとインタープリター、Go/WASMブリッジは廃止しました。旧トップレベル実行文は`fn main()`へ移し、型付き`let`、`while`、関数形式の`print`・`println`へ書き換えてください。旧`repeat`や文字列の`+`連結はサポートしません。

## 宣言と文

```text
const WIDTH:u64 = 640;
global frame:u64 = 0;
buffer PIXELS[WIDTH * 480 * 4];
embed MEDIA;

fn advance(amount:u64)->u64 {
    let next:u64 = frame + amount;
    if next > 100 { next = 0; } else if next == 0 { next = 1; }
    frame = next;
    return frame;
}
```

- 型：`u64`、`bool`、`str`。戻り値なしは `-> void`、または戻り値の型を省略します。
- 文：`let`、代入、関数呼び出し、`if` / `else if` / `else`、`while`、`return`、`break`、`continue`。
- 代入：`=`、`+=`、`-=`、`|=`、`&=`。後ろの四つは `u64` に使用します。
- 条件は `bool` です。整数を条件にする場合は `value != 0` のように比較します。
- 数値リテラルは常に `u64`。十進、`0x` の十六進、`0b` の二進と `_` による区切りに対応します。
- 数値演算：`+ - * / % << >> & | ^`。単項演算：`+ - ~`。加減乗算は 64 ビットでラップし、シフト量は下位 6 ビットを使用します。実行時のゼロ除算は panic になります。
- 比較：`== != < <= > >=`。`bool` と `str` は等値・非等値比較に対応します。
- 論理演算：`! && ||`。`&&` と `||` は短絡評価です。
- 文字列は静的 UTF-8。日本語と `\n \r \t \0 \\ \"` に対応します。`len` と `byte` は文字数ではなくバイト単位です。
- コメントは `#` から行末まで。文の区切りには `;` を推奨します。

グローバル変数と定数の初期値、バッファー長はコンパイル時に評価できる式に限定します。固定長バッファーは 16 バイト境界に配置され、ゼロで初期化されます。メモリー領域自体は代入できません。`addr` とロード・ストアを使って内容を扱います。

## ホスト用出力

`print(value)`は改行なし、`println(value)`は改行付きで標準出力へ書き込みます。引数は`u64`、`bool`、`str`のいずれか一つです。この二つは`host`ターゲット専用で、UEFI・libraryには使用できません。文字列を連結する代わりに`print("SYSTEM="); println("TANEBI 95");`のように出力を分けられます。

## 汎用組み込み関数

| 関数 | 内容 |
|---|---|
| `in8(port)` | 8 ビットの I/O ポート読み込み |
| `out8(port,value)` / `out32(port,value)` | I/O ポート書き込み |
| `load8/16/32/64(addr)` | 指定幅の volatile 読み込み。戻り値は `u64` |
| `store8/16/32/64(addr,value)` | 指定幅の volatile 書き込み。非整列アドレスにも対応 |
| `copy(dst,src,len)` | 重なる領域にも対応するバイトコピー |
| `zero(dst,len)` | 領域のゼロクリア |
| `addr(buffer/embed/str)` | 先頭アドレス |
| `len(buffer/embed/str)` | バイト数 |
| `byte(str,index)` | UTF-8 の 1 バイト。範囲外は 0 |
| `ticks()` | CPU のタイムスタンプカウンター。秒ではありません |
| `pause()` | CPU のスピンループヒント |
| `irq_disable()` | CPU のマスク可能割り込みを無効化 |
| `call0` ～ `call6` | 先頭引数の関数ポインターを `efiapi` ABI で呼び出す |
| `config("NAME")` | コンパイル時設定。キーは文字列リテラル |

関数ポインターの引数と戻り値、アドレスは `u64` です。`call2(pointer,a,b)` のように使用します。これらは CPU・メモリー・ABI の基本操作のみで、ウィンドウ、描画、入力、動画、音声などの OS 機能は TANEBI ソース側で実装します。

ポートI/Oと割り込み無効化の生成アセンブリーにはメモリーの変更可能性を明示し、DMA用バッファーの書き込みなどをコンパイラーが命令の前後へ移動しないようにします。ロード・ストアはvolatileです。ただし、これらは任意の機器・キャッシュ属性に対する完全な同期APIではありません。対象ハードウェアのメモリー順序とDMAの条件を確認してください。

## 安全性と制限

未知の名前、宣言の重複、引数の数や型、戻り値の型、ループ外の `break`、不正な定数式はファイル名・行・列付きで報告します。構造体、動的メモリー確保、符号付き整数、ユーザー定義型はまだありません。関数定義はトップレベルです。

この言語はメモリー安全なサンドボックスではありません。任意アドレスと特権命令を扱うため、誤ったアドレス・長さ・関数ポインターは破損や停止を引き起こします。I/O や割り込み命令は適切な特権レベルでのみ実行してください。ホストテストでは通常の配列領域と安全なテスト関数だけを使用します。

```powershell
go test ./internal/native ./cmd/tanebi
```

Rust が利用可能な場合、テストは生成コードを実際にコンパイル・実行し、算術、分岐、ループ、UTF-8、メモリー操作と ABI 呼び出しを確認します。
