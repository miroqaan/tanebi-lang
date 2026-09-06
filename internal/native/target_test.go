package native

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestHostOutputExecution(t *testing.T) {
	source := `
fn message()->str { return "種火は消えない"; }
fn main() {
    print("タネビ: ");
    println(message());
    let count:u64=0;
    while count<3 { print(count); count+=1; }
    println("");
    println(18446744073709551615);
    println(true);
    println(false);
    println("{} {name} {:?}");
    println("括弧: } { ; fn true -");
}
`
	code, err := Compile([]Source{{Name: "hello.tanebi", Text: source}}, Options{Target: "host"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(code, "#![no_std]") || strings.Contains(code, "#![no_main]") || strings.Contains(code, "#[panic_handler]") || strings.Contains(code, "fn efi_main(") {
		t.Fatal("host output contains UEFI-only boilerplate")
	}
	if !strings.Contains(code, `::std::print!("{}", `) || !strings.Contains(code, `::std::println!("{}", `) {
		t.Fatal("output must use fixed display format")
	}
	if !strings.Contains(code, "fn main() { unsafe { self::tn_main(); } }") {
		t.Fatal("missing host entry wrapper")
	}
	want := "タネビ: 種火は消えない\n012\n18446744073709551615\ntrue\nfalse\n{} {name} {:?}\n括弧: } { ; fn true -\n"
	if got := runGeneratedHost(t, code); got != want {
		t.Fatalf("output=%q, want=%q", got, want)
	}
}

func TestHostCustomEntryExecution(t *testing.T) {
	code, err := Compile([]Source{{Name: "custom.tanebi", Text: `fn start()->void { println("起動"); return; }`}}, Options{Target: "host", Entry: "start"})
	if err != nil {
		t.Fatal(err)
	}
	if got := runGeneratedHost(t, code); got != "起動\n" {
		t.Fatalf("output=%q", got)
	}
}

func TestTargetAndEntryDiagnostics(t *testing.T) {
	tests := []struct{ name, source, target, entry, want string }{
		{"missing host main", `fn other() {}`, "host", "", "entry function \"main\" not found"},
		{"host main is data", `global main:u64=0;`, "host", "", "entry function \"main\" not found"},
		{"host arguments", `fn main(x:u64) {}`, "host", "", "host entry must have signature"},
		{"host result", `fn main()->u64 { return 0; }`, "host", "", "host entry must have signature"},
		{"missing custom entry", `fn main() {}`, "host", "start", "entry function \"start\" not found"},
		{"missing UEFI entry", `fn main() {}`, "uefi", "", "entry function \"entry\" not found"},
		{"UEFI result", `fn entry(image:u64,st:u64) {}`, "uefi", "", "entry must have signature"},
		{"unknown target", `fn main() {}`, "web", "", "unknown target \"web\""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Compile([]Source{{Name: "entry-error.tanebi", Text: tt.source}}, Options{Target: tt.target, Entry: tt.entry})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error=%v; want %q", err, tt.want)
			}
		})
	}
}

func TestHostOutputDiagnostics(t *testing.T) {
	tests := []struct{ name, source, want string }{
		{"no argument", `fn main() { println(); }`, "println expects 1 arguments, got 0"},
		{"extra argument", `fn main() { print(1,2); }`, "print expects 1 arguments, got 2"},
		{"buffer is not printable", `buffer B[8]; fn main() { println(B); }`, "requires u64, bool, or str, got memory"},
		{"void is not printable", `fn empty() {} fn main() { print(empty()); }`, "requires u64, bool, or str, got void"},
		{"output is not a value", `fn main() { let n:u64=println("x"); }`, "initializer type void, expected u64"},
		{"builtin cannot be replaced", `fn print(value:u64) {} fn main() {}`, "declaration uses builtin name \"print\""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Compile([]Source{{Name: "output-error.tanebi", Text: tt.source}}, Options{Target: "host"})
			if err == nil || !strings.Contains(err.Error(), tt.want) || !strings.Contains(err.Error(), "output-error.tanebi:") {
				t.Fatalf("error=%v; want source diagnostic containing %q", err, tt.want)
			}
		})
	}
	for _, target := range []string{"uefi", "library"} {
		for _, call := range []string{"print", "println"} {
			t.Run(target+"_"+call, func(t *testing.T) {
				source := `fn entry(image:u64,st:u64)->u64 { ` + call + `("test"); return 0; }`
				_, err := Compile([]Source{{Name: "target-error.tanebi", Text: source}}, Options{Target: target})
				if err == nil || !strings.Contains(err.Error(), call+" is available only for target host, not "+target) {
					t.Fatalf("error=%v", err)
				}
			})
		}
	}
}

func TestOldInterpreterGrammarRejected(t *testing.T) {
	tests := map[string]string{
		"top-level let":        `let 名前 = "アカギ"`,
		"untyped let":          `fn main() { let count = 3; }`,
		"repeat":               `fn main() { repeat 3 { println("種火"); } }`,
		"print statement":      `fn main() { print "種火"; }`,
		"floating number":      `fn main() { let value:u64=1.5; }`,
		"string concatenation": `fn main() { println("種"+"火"); }`,
	}
	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := Compile([]Source{{Name: "old-syntax.tanebi", Text: source}}, Options{Target: "host"})
			if err == nil || !strings.Contains(err.Error(), "old-syntax.tanebi:") {
				t.Fatalf("old grammar accepted or missing source diagnostic: %v", err)
			}
		})
	}
}

func TestLibraryTargetAndLegacyCompatibility(t *testing.T) {
	sources := []Source{{Name: "library.tanebi", Text: `fn helper()->u64 { return 42; }`}}
	code, err := Compile(sources, Options{Target: "library"})
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"#![no_std]", "#![no_main]", "fn main()", "fn efi_main(", "#[panic_handler]"} {
		if strings.Contains(code, forbidden) {
			t.Fatalf("library contains %q", forbidden)
		}
	}
	legacy, err := Compile(sources, Options{Library: true})
	if err != nil {
		t.Fatal(err)
	}
	if legacy != code {
		t.Fatal("legacy Library option no longer aliases target library")
	}
	_, err = Compile(sources, Options{Target: "host", Library: true})
	if err == nil || !strings.Contains(err.Error(), "conflicts with target") {
		t.Fatalf("error=%v", err)
	}
	entrySources := []Source{{Name: "efi.tanebi", Text: `fn entry(image:u64,st:u64)->u64 { return 0; }`}}
	defaultCode, err := Compile(entrySources, Options{})
	if err != nil {
		t.Fatal(err)
	}
	explicitCode, err := Compile(entrySources, Options{Target: "uefi"})
	if err != nil {
		t.Fatal(err)
	}
	if defaultCode != explicitCode {
		t.Fatal("empty target no longer preserves UEFI output")
	}
}

func runGeneratedHost(t *testing.T, code string) string {
	t.Helper()
	rustc, err := exec.LookPath("rustc")
	if err != nil {
		t.Skip("rustc unavailable; frontend checks still run")
	}
	dir := t.TempDir()
	source := filepath.Join(dir, "program.rs")
	binary := filepath.Join(dir, "program")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	if err = os.WriteFile(source, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(rustc, "--edition=2024", "-C", "opt-level=2", "-o", binary, source).CombinedOutput(); err != nil {
		t.Fatalf("rustc: %v\n%s", err, output)
	}
	output, err := exec.Command(binary).CombinedOutput()
	if err != nil {
		t.Fatalf("host execution: %v\n%s", err, output)
	}
	return strings.ReplaceAll(string(output), "\r\n", "\n")
}
