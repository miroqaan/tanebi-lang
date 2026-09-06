# TANEBI 95

このディレクトリは、1990年代のデスクトップインターフェースに着想を得た TANEBI 95 の Web 版です。ブラウザー内でウィンドウを移動・最小化・最大化し、TANEBI Studio で実際の TANEBI コードを実行できます。QEMU で起動する独立 OS 版は [tanebi-95](https://github.com/miroqaan/tanebi-95) を参照してください。

## 主な機能

- **TANEBI Studio**: Unicode 識別子を使った TANEBI コードを Go/WebAssembly で実行します。
- **TANEBI 95 DOOM**: 公開されている Chocolate Doom エンジンを WebAssembly で動かし、Freedoom Phase 2 のゲームデータを使用します。
- **メディアプレイヤー**: アカギユニバースの指定 YouTube 動画を、16分14秒から音声付きで再生します。
- **システムモニター**: 稼働時間、JavaScript ヒープ、ランタイムの境界、機能ツアーを表示します。

## 構成

```text
ブラウザーデスクトップ (React / vinext)
              ↓
Go WebAssembly ホスト (cmd/tanebi-wasm)
              ↓
TANEBI スクリプト層 (Lexer → Parser → AST → Interpreter)
```

この境界は、`gopher-os` の Rust カーネルと Go ユーザー空間の分離を参考にしています。コードはコピーしていません。この Web 版は独自のブラウザーデスクトップであり、実際に起動する OS や Microsoft Windows の配布物ではありません。

## 開発

Node.js 22.13 以降と Go 1.22 以降が必要です。

```powershell
npm install
npm run wasm
npm run prepare:doom
npm run dev
```

## 検証

```powershell
npm run lint
npm run test:wasm
npm run build
npm audit --omit=dev --audit-level=high
```

`public/system.tanebi` は起動時に実行され、デスクトップ名、キャッチコピー、アイコンのラベル、起動メッセージを設定します。

## DOOM のライセンスと帰属

商用 DOOM のゲームデータは含みません。ブラウザー版は GPL-2.0-or-later の [Chocolate Doom](https://github.com/chocolate-doom/chocolate-doom)、BSD 3-Clause の [Freedoom 0.13.0](https://github.com/freedoom/freedoom/tree/v0.13.0)、Emscripten による WebAssembly ビルドを使用します。公開されている [doom-wasm](https://github.com/gabrielbotandev/doom-wasm) ビルドをコミットと SHA-256 で固定し、ライセンス全文とクレジットを `public/doom/licenses/` に含めています。

DOOM は id Software の商標です。TANEBI 95 は id Software または Microsoft と提携しておらず、両社の承認を受けたプロジェクトではありません。
