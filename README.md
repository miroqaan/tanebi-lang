# TANEBI 95

TANEBI 95は、1990年代のデスクトップインターフェースから着想を得たTANEBIの実行環境です。ブラウザー内でウィンドウを移動、最小化、最大化でき、TANEBI Studioで実際のTANEBIコードを実行します。

## 構成

```text
ブラウザーのデスクトップ (React / vinext)
              ↓
Go WebAssemblyホスト (cmd/tanebi-wasm)
              ↓
TANEBIユーザー空間 (Lexer → Parser → AST → Interpreter)
```

この境界設計は、`gopher-os`におけるRustカーネルとGoユーザー空間の分離から着想を得ています。コードはコピーしておらず、TANEBI 95は実際のOSやMicrosoft Windowsの配布物ではなく、独自のウェブデスクトップです。

## 開発

Node.js 22.13以降とGo 1.22以降が必要です。

```powershell
npm install
npm run wasm
npm run dev
```

## 検証

```powershell
npm run lint
npm run test:wasm
npm run build
npm audit --omit=dev --audit-level=high
```

`public/system.tanebi`は起動時に実行され、デスクトップ名、文言、アイコンのラベル、起動メッセージを設定します。
