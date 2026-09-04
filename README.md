# Go로 만든 Mini 언어

Go 표준 라이브러리만 사용해 구현한 작은 인터프리터입니다.

## 지원 문법

```mini
# 한글 변수도 사용할 수 있습니다.
let 너비 = 10
let 높이 = 5
print 너비 * 높이
print (너비 + 높이) / 2
```

- `let 이름 = 표현식`: 변수 선언 또는 값 변경
- `print 표현식`: 결과 출력
- 정수와 소수
- `+`, `-`, `*`, `/`, 단항 `+`와 `-`
- 줄바꿈 또는 `;`로 문장 구분
- `#`부터 줄 끝까지 주석

## 실행

Go 1.22 이상이 필요합니다.

```powershell
cd "C:\Users\wavus\IdeaProjects\mini-language-go"
go run . example.mini
```

예상 출력:

```text
3100
5
```

테스트:

```powershell
go test ./...
```

## 구현 구조

```text
Mini 소스 → Lexer → Token → Parser → AST → Interpreter → 출력
```

`main.go` 안에서 각 단계를 순서대로 살펴볼 수 있습니다. 다음 확장 단계로 비교 연산자, `if`, `while`, 함수를 추가할 수 있습니다.
