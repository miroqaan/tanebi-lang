# TANEBI（種火）

> 한 줄의 코드가 세계에 불을 밝힌다.

[日本語](./README.md)

TANEBI는 Go 표준 라이브러리만으로 만든 작고 결정론적인 인터프리터입니다. 계산기 수준의 불씨에서 시작해 조건문, 반복문, 함수, 게임 스크립팅 언어로 성장하는 과정을 보여 주는 아카기유니버스 프로젝트입니다.

## 언어의 원칙

- **작지만 완전하게**: Lexer → Parser → AST → Interpreter 구조를 분리합니다.
- **세계 어디서나 읽게**: 영문 키워드와 한글·일본어를 포함한 Unicode 식별자를 지원합니다.
- **항상 같은 결과로**: 시간, 네트워크, 암시적 난수를 언어 코어에 두지 않습니다.
- **명시적으로**: 숫자·문자열·불리언 사이의 암시적 타입 변환을 하지 않습니다.
- **가볍게**: 외부 의존성 없이 하나의 Go 실행 파일로 빌드합니다.

## 예제

```tanebi
let 이름 = "아카기"
let 불꽃 = 3

repeat 불꽃 {
    print 이름 + "의 세계가 깨어난다"
}

if 불꽃 >= 3 {
    print "種火は消えない"
} else {
    print "다시 불을 붙인다"
}
```

## 지원 문법

| 영역 | 문법 |
|---|---|
| 값 | 숫자, 문자열, `true`, `false` |
| 변수 | `let 이름 = 표현식`, `이름 = 표현식` |
| 출력 | `print 표현식` |
| 반복 | `repeat 횟수 { ... }` |
| 조건 | `if 조건 { ... } else { ... }` |
| 연산 | `+ - * / %`, `== != < <= > >=`, `&& || !` |
| 구분 | 줄바꿈 또는 `;` |
| 주석 | `#`부터 줄 끝까지 |

`+`는 숫자끼리 더하거나 문자열끼리 연결합니다. 조건식은 반드시 불리언이어야 하며 `repeat` 횟수는 0 이상의 정수여야 합니다.

## 실행

Go 1.22 이상이 필요합니다.

```powershell
go run ./cmd/tanebi ./examples/awakening.tanebi
```

테스트와 빌드:

```powershell
go test ./...
go build ./cmd/tanebi
```

## 프로젝트 구조

```text
cmd/tanebi/       CLI 진입점
internal/tanebi/  Token, Lexer, AST, Parser, Interpreter
examples/         실행 가능한 TANEBI 프로그램
```

## Roadmap

- 사용자 정의 함수와 `return`
- List와 Map
- 게임 이벤트·대사 스크립팅 API
- seed를 명시하는 결정론적 난수
- 에디터 문법 강조
