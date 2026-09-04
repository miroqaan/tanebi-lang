# TANEBI（種火）

> 一行のコードが、世界に火を灯す。

[한국어](./README.ko.md)

TANEBIは、Goの標準ライブラリだけで実装した、小さく決定論的なインタープリターです。電卓ほどの小さな種火から始まり、条件分岐、繰り返し、関数、そしてゲームスクリプト言語へ成長していく過程そのものを見せる、アカギユニバースのプロジェクトです。

## 設計原則

- **小さくても完全に**：Lexer → Parser → AST → Interpreterを分離します。
- **世界中の言葉で**：英語のキーワードと、日本語・韓国語を含むUnicode識別子を使えます。
- **常に同じ結果を**：時刻、ネットワーク、暗黙の乱数を言語コアに持ち込みません。
- **明示的に**：数値・文字列・真偽値を暗黙に型変換しません。
- **軽く**：外部依存なしの単一Goバイナリとしてビルドできます。

## サンプル

```tanebi
let 名前 = "アカギ"
let 種火 = 3

repeat 種火 {
    print 名前 + "の世界が目を覚ます"
}

if 種火 >= 3 {
    print "種火は消えない"
} else {
    print "もう一度、火を灯す"
}
```

## 現在の言語機能

| 分類 | 構文 |
|---|---|
| 値 | 数値、文字列、`true`、`false` |
| 変数 | `let name = expression`、`name = expression` |
| 出力 | `print expression` |
| 繰り返し | `repeat count { ... }` |
| 条件分岐 | `if condition { ... } else { ... }` |
| 演算子 | `+ - * / %`、`== != < <= > >=`、`&& || !` |
| 文の区切り | 改行または`;` |
| コメント | `#`から行末まで |

`+`は数値同士の加算、または文字列同士の結合に使います。条件式は真偽値、`repeat`の回数は0以上の整数でなければなりません。

## 実行

Go 1.22以上が必要です。

```powershell
go run ./cmd/tanebi ./examples/awakening.tanebi
```

テストとビルド：

```powershell
go test ./...
go build ./cmd/tanebi
```

## プロジェクト構成

```text
cmd/tanebi/       CLIエントリーポイント
internal/tanebi/  Token、Lexer、AST、Parser、Interpreter
examples/         実行可能なTANEBIプログラム
```

## Roadmap

- ユーザー定義関数と`return`
- ListとMap
- ゲームイベント・会話スクリプトAPI
- seedを明示する決定論的乱数
- エディター向けシンタックスハイライト
