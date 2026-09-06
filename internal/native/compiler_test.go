package native

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestStringsAreNotSyntaxTokens(t *testing.T) {
	for _, text := range []string{"}", "{", "fn", ";", "-", "+", "!", "~", "(", ")", "true", "false", "if", "else", "return", "||", "=", "0x123"} {
		t.Run(text, func(t *testing.T) {
			source := fmt.Sprintf("fn f()->str { if true { return %q; } return %q; }", text, text)
			code := compileTest(t, source)
			if !strings.Contains(code, "return "+rustString(text)+";") {
				t.Fatalf("string was altered: %s", text)
			}
		})
	}
}

func TestIdentifierEncodingIsCollisionFree(t *testing.T) {
	names := []string{"日", "_u65e5_", "a日", "a_u65e5_", "__u_e697a5", "__a_5f5f", "type", "loop", "名前", "mouse_decode"}
	seen := map[string]string{}
	for _, name := range names {
		encoded := rustName(name)
		if previous, ok := seen[encoded]; ok {
			t.Fatalf("%q and %q collide as %q", previous, name, encoded)
		}
		seen[encoded] = name
		if !strings.HasPrefix(encoded, "tn_") {
			t.Fatalf("wrong prefix: %s", encoded)
		}
	}
	if rustName("mouse_decode") != "tn_mouse_decode" {
		t.Fatal("ordinary ASCII ABI name changed")
	}
}

func TestIOPrimitivesRetainCompilerMemoryBarrier(t *testing.T) {
	code := compileTest(t, `fn f() {}`)
	for _, line := range strings.Split(code, "\n") {
		if strings.Contains(line, "unsafe fn __tn_in8(") || strings.Contains(line, "unsafe fn __tn_out8(") || strings.Contains(line, "unsafe fn __tn_out32(") || strings.Contains(line, "unsafe fn __tn_irq_disable(") {
			if strings.Contains(line, "nomem") || strings.Contains(line, "readonly") {
				t.Fatalf("I/O must retain memory clobber: %s", line)
			}
		}
	}
}

func compileTest(t *testing.T, source string) string {
	t.Helper()
	code, err := Compile([]Source{{Name: "test.tanebi", Text: source}}, Options{Target: "library"})
	if err != nil {
		t.Fatal(err)
	}
	return code
}

func TestSourceDiagnostics(t *testing.T) {
	tests := []struct{ name, source, want string }{
		{"unknown symbol", "fn f()->u64 { return missing; }", "unknown symbol"},
		{"duplicate declaration", "global X:u64=0; const X:u64=1;", "duplicate declaration"},
		{"duplicate parameter", "fn f(x:u64,x:bool) {}", "duplicate parameter"},
		{"duplicate local", "fn f() { let x:u64=0; let x:u64=1; }", "duplicate local"},
		{"unknown function", "fn f() { raw_rust(\"unsafe {}\"); }", "unknown function"},
		{"no foreign block", "rust { fn exploit() {} }", "expected fn"},
		{"arity", "fn f() { out8(1); }", "expects 2 arguments"},
		{"argument type", "fn f() { out8(true,1); }", "expected u64"},
		{"return type", "fn f()->u64 { return true; }", "return has type bool"},
		{"missing return", "fn f()->u64 { let x:u64=1; }", "without returning"},
		{"condition type", "fn f() { if 1 {} }", "condition must be bool"},
		{"unknown type", "fn f(x:i64) {}", "unknown type"},
		{"break context", "fn f() { break; }", "outside a loop"},
		{"const assignment", "const X:u64=1; fn f() { X=2; }", "cannot assign to const"},
		{"memory assignment", "buffer B[16]; fn f() { B=2; }", "cannot assign to buffer"},
		{"const cycle", "const A:u64=B; const B:u64=A;", "cyclic constant"},
		{"const division", "const A:u64=1/0;", "division by zero"},
		{"buffer limit", "buffer B[0];", "buffer size must"},
		{"number overflow", "const A:u64=18446744073709551616;", "invalid u64 literal"},
		{"runtime config", "fn f(s:str)->u64 { return config(s); }", "string literal"},
		{"address type", "fn f()->u64 { return addr(42); }", "requires a buffer"},
		{"unterminated string", "fn f(){ let s:str=\"x; }", "unterminated string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Compile([]Source{{Name: "bad.tanebi", Text: tt.source}}, Options{Target: "library"})
			if err == nil || !strings.Contains(err.Error(), tt.want) || !strings.Contains(err.Error(), "bad.tanebi:") {
				t.Fatalf("error=%v; want source diagnostic containing %q", err, tt.want)
			}
		})
	}
}

func TestMultiFileAndEFIEntry(t *testing.T) {
	code, err := Compile([]Source{{Name: "entry.tanebi", Text: "fn entry(image:u64,st:u64)->u64 { return later(image)+st; }"}, {Name: "later.tanebi", Text: "fn later(x:u64)->u64 { return x+1; }"}}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"#![no_std]", "#![no_main]", "extern \"efiapi\" fn efi_main", "tn_entry(image as u64, st as u64)", "pub unsafe fn tn_later"} {
		if !strings.Contains(code, want) {
			t.Fatalf("missing %q", want)
		}
	}
	_, err = Compile([]Source{{Name: "wrong.tanebi", Text: "fn entry()->u64 { return 0; }"}}, Options{})
	if err == nil || !strings.Contains(err.Error(), "entry must have signature") {
		t.Fatalf("error=%v", err)
	}
}

func TestConstantShortCircuitAndLengths(t *testing.T) {
	code := compileTest(t, `const A:bool=false && (1/0==1); const B:bool=true || (1/0==1); const N:u64=len("種火"); buffer DATA[N*2]; const SIZE:u64=len(DATA);`)
	for _, want := range []string{"tn_A: bool = false", "tn_B: bool = true", "tn_N: u64 = 6u64", "tn_SIZE: u64 = 12u64"} {
		if !strings.Contains(code, want) {
			t.Fatalf("missing %q", want)
		}
	}
	_, err := Compile([]Source{{Name: "bad.tanebi", Text: `const X:bool=false && missing;`}}, Options{Target: "library"})
	if err == nil || !strings.Contains(err.Error(), "unknown symbol") {
		t.Fatalf("skipped invalid operand: %v", err)
	}
}

func TestConfigAssetsAndStringEscaping(t *testing.T) {
	asset := filepath.Join(t.TempDir(), "media.bin")
	if err := os.WriteFile(asset, []byte{1, 2, 3, 4}, 0600); err != nil {
		t.Fatal(err)
	}
	src := `embed MEDIA; buffer B[(1 << 4)+config("EXTRA")]; const TEXT:str="種火\n\0\""; fn f()->u64 { return len(MEDIA)+len(B)+byte(TEXT,0)+config("MISSING"); }`
	code, err := Compile([]Source{{Name: "assets.tanebi", Text: src}}, Options{Target: "library", Assets: map[string]string{"media": asset}, Defines: map[string]uint64{"EXTRA": 2}})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"align(16)", "[u8; 18]", "tn_MEDIA: &[u8; 4]", "include_bytes!(", `種火\n\u{0}\"`, "4u64", "18u64"} {
		if !strings.Contains(code, want) {
			t.Fatalf("missing %q", want)
		}
	}
}

func TestGeneratedRustHostExecution(t *testing.T) {
	rustc, err := exec.LookPath("rustc")
	if err != nil {
		t.Skip("rustc unavailable; frontend tests still run")
	}
	code := compileTest(t, `
const MASK:u64=0xffff_ffff_ffff_ffff;
global COUNTER:u64=0;
global STATE:u64=99;
buffer MEM[128];
fn 名前(x:u64)->u64 { return x+1; }
fn 日()->u64 { return 2; }
fn _u65e5_()->u64 { return 3; }
fn __u_e697a5()->u64 { return 4; }
fn type()->u64 { return 5; }
fn shadow(名前:u64)->u64 { return 名前(名前); }
fn shadow_global(STATE:u64)->u64 {
    let MASK:u64=2;
    if true { let STATE:u64=5; STATE+=1; MASK+=STATE; }
    return STATE+MASK;
}
fn pick(x:bool)->str { if x { return "種火"; } else { return "Go"; } }
fn probe()->u64 {
    let p:u64=addr(MEM);
    if p % 16 != 0 { return 1; }
    zero(p,len(MEM));
    store64(p+1,0x8877665544332211);
    if load64(p+1)!=0x8877665544332211 { return 2; }
    if load16(p+2)!=0x3322 { return 3; }
    store32(p+16,0xaabbccdd);
    if load32(p+16)!=0xaabbccdd { return 4; }
    copy(p+40,p+1,8);
    copy(p+42,p+40,8);
    if load64(p+42)!=0x8877665544332211 { return 5; }
    let n:u64=0;
    while n<10 { n+=1; if n==2 { continue; } if n==8 { break; } COUNTER+=n; }
    if COUNTER!=26 { return 6; }
    let bits:u64=0b1010; bits|=1; bits&=0xf; bits-=1;
    if bits!=10 || ((MASK+1)!=0) || ((0-1)!=MASK) || ((1<<65)!=2) { return 7; }
    if len(pick(true))!=6 || byte(pick(false),0)!=71 || byte("A",9)!=0 { return 8; }
    if false && load8(0)==1 { return 9; }
    if 日()+_u65e5_()+__u_e697a5()+type()+shadow(5) != 20 { return 10; }
    if shadow_global(10)!=18 || STATE!=99 { return 12; }
    if true || load8(0)==1 { return 名前(0)-1; } else if false { return 10; }
    return 11;
}
`)
	code += "\nfn main() { unsafe { assert_eq!(tn_probe(), 0); assert_eq!(__tn_call2(add as *const () as u64, 12, 30), 42); } println!(\"native compiler host tests passed\"); }\nextern \"efiapi\" fn add(a:usize,b:usize)->usize { a+b }\n"
	dir := t.TempDir()
	src := filepath.Join(dir, "native.rs")
	bin := filepath.Join(dir, "native-test")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if err = os.WriteFile(src, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(rustc, "--edition=2024", "-C", "opt-level=2", "-o", bin, src).CombinedOutput(); err != nil {
		t.Fatalf("rustc: %v\n%s", err, out)
	}
	if out, err := exec.Command(bin).CombinedOutput(); err != nil {
		t.Fatalf("host: %v\n%s", err, out)
	} else {
		t.Log(string(out))
	}
}
