# TANEBI 95

TANEBI 95는 1990년대 데스크톱 인터페이스에서 영감을 얻은 TANEBI 실행 환경입니다. 브라우저 안에서 창을 이동하고 최소화하거나 최대화할 수 있으며, TANEBI Studio에서 실제 TANEBI 코드를 실행합니다.

## 구조

```text
브라우저 데스크톱 (React / vinext)
              ↓
Go WebAssembly 호스트 (cmd/tanebi-wasm)
              ↓
TANEBI 사용자 공간 (Lexer → Parser → AST → Interpreter)
```

이 경계는 `gopher-os`의 Rust 커널과 Go 사용자 공간 분리에서 아이디어를 얻었습니다. 코드는 복사하지 않았으며, TANEBI 95는 실제 운영체제나 Microsoft Windows 배포물이 아닌 독자적인 웹 데스크톱입니다.

## 개발

Node.js 22.13 이상과 Go 1.22 이상이 필요합니다.

```powershell
npm install
npm run wasm
npm run dev
```

## 검증

```powershell
npm run lint
npm run test:wasm
npm run build
npm audit --omit=dev --audit-level=high
```

`public/system.tanebi`는 부팅 시 실행되어 데스크톱 이름, 문구, 아이콘 라벨과 부팅 메시지를 구성합니다.
