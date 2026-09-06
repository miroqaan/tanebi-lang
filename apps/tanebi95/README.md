# TANEBI 95

このディレクトリは、1990年代のデスクトップに着想を得た旧 Web UI のアーカイブです。TANEBI のコンパイラー一本化に伴い、インタープリター、実行用 Go/WASM ブリッジ、配布バイナリー、起動スクリプトの実行機能を削除しました。独立 OS の現在の実装は [tanebi-95](https://github.com/miroqaan/tanebi-95) を参照してください。これは OS 本体ではありません。

## 主な機能

- **TANEBI Studio**: コンパイラー用の型付きソースと PC での実行コマンドを表示します。閲覧専用で、ブラウザー内では実行しません。
- **TANEBI 95 DOOM**: 公開されている Chocolate Doom エンジンを WebAssembly で動かし、Freedoom Phase 2 のゲームデータを使用します。
- **メディアプレイヤー**: アカギユニバースの指定 YouTube 動画を、16分14秒から音声付きで再生します。
- **システムモニター**: 稼働時間、JavaScript ヒープ、ランタイムの境界、機能ツアーを表示します。

## 構成

```text
ブラウザーデスクトップ (React / vinext)
              ↓
固定システム設定 / 閲覧専用 TANEBI サンプル
```

この境界は、`gopher-os` の Rust カーネルと Go ユーザー空間の分離を参考にしています。コードはコピーしていません。この Web 版は独自のブラウザーデスクトップであり、実際に起動する OS や Microsoft Windows の配布物ではありません。

## 開発

Node.js 22.13 以降が必要です。TANEBI コードの実行には、別途 PC 側の `tanebi` コンパイラーと Rust が必要です。

```powershell
npm install
npm run prepare:doom
npm run dev
```

## 検証

```powershell
npm run lint
npm run build
npm audit --omit=dev --audit-level=high
```

デスクトップの名称とラベルは固定データです。TANEBI の字句解析・構文解析・評価処理を JavaScript で代替する経路はありません。DOOM 用の第三者 WebAssembly エンジンは TANEBI の処理系とは別のため、そのライセンスと資産は保持しています。

## DOOM のライセンスと帰属

商用 DOOM のゲームデータは含みません。ブラウザー版は GPL-2.0-or-later の [Chocolate Doom](https://github.com/chocolate-doom/chocolate-doom)、BSD 3-Clause の [Freedoom 0.13.0](https://github.com/freedoom/freedoom/tree/v0.13.0)、Emscripten による WebAssembly ビルドを使用します。公開されている [doom-wasm](https://github.com/gabrielbotandev/doom-wasm) ビルドをコミットと SHA-256 で固定し、ライセンス全文とクレジットを `public/doom/licenses/` に含めています。

DOOM は id Software の商標です。TANEBI 95 は id Software または Microsoft と提携しておらず、両社の承認を受けたプロジェクトではありません。
