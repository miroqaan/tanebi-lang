# TANEBI 95

TANEBI 95는 1990년대 데스크톱 인터페이스에서 영감을 얻은 TANEBI 실행 환경입니다. 브라우저 안에서 창을 이동하고 최소화하거나 최대화할 수 있으며, TANEBI Studio에서 실제 TANEBI 코드를 실행합니다.

## 주요 기능

- **TANEBI Studio**: Unicode 식별자를 사용하는 TANEBI 코드를 Go/WebAssembly로 직접 실행합니다.
- **TANEBI 95 DOOM**: 공개된 Chocolate Doom 엔진을 WebAssembly로 구동하고 Freedoom Phase 2 데이터를 사용합니다.
- **미디어 플레이어**: 아카기유니버스의 지정 YouTube 영상을 16분 14초부터 소리와 함께 재생합니다.
- **시스템 모니터**: 실행 시간, JavaScript 힙, 런타임 경계와 기능 투어를 표시합니다.

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
npm run prepare:doom
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

## DOOM 라이선스 경계

상용 DOOM 게임 데이터는 포함하지 않습니다. 브라우저 게임은 GPL-2.0-or-later의 [Chocolate Doom](https://github.com/chocolate-doom/chocolate-doom), BSD 3-Clause의 [Freedoom 0.13.0](https://github.com/freedoom/freedoom/tree/v0.13.0), Emscripten WebAssembly 빌드를 사용합니다. 공개된 [doom-wasm](https://github.com/gabrielbotandev/doom-wasm) 빌드를 커밋과 SHA-256으로 고정하며, 전체 라이선스와 크레딧은 `public/doom/licenses/`에 포함합니다.

DOOM은 id Software의 상표이며, TANEBI 95는 id Software 또는 Microsoft와 제휴하거나 보증받은 프로젝트가 아닙니다.
