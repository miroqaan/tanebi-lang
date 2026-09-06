package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func fixture(t *testing.T, source string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "日本語 sample.tanebi")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func invoke(args ...string) (string, string, error) {
	var out, diagnostics bytes.Buffer
	err := run(args, strings.NewReader(""), &out, &diagnostics)
	return out.String(), diagnostics.String(), err
}

func TestCheckAndEmitDoNotNeedBackend(t *testing.T) {
	path := fixture(t, `fn main() { println("灯る"); }`)
	t.Setenv("PATH", "")
	out, _, err := invoke("check", path)
	if err != nil || !strings.Contains(out, "OK (host)") {
		t.Fatalf("%s: %v", out, err)
	}
	out, _, err = invoke("emit", path)
	if err != nil || !strings.Contains(out, "fn main()") || !strings.Contains(out, "灯る") {
		t.Fatalf("%s: %v", out, err)
	}
	_, _, err = invoke("run", path)
	if err == nil || !strings.Contains(err.Error(), "rustc is required") {
		t.Fatalf("missing backend did not fail: %v", err)
	}
}

func TestInvalidCLIAndOldSyntax(t *testing.T) {
	path := fixture(t, `fn main() {}`)
	cases := [][]string{
		{}, {path}, {"check"}, {"check", "-target", "wasm", path},
		{"run", "-target", "uefi", path}, {"build", "-target", "library", path},
		{"run", "-o", "ignored.exe", path}, {"check", "-o", "ignored.rs", path},
		{"build", "-opt", "9", path}, {"check", "-define", "X=bad", path},
		{"check", "-define", "X=1", "-define", "X=2", path},
		{"check", path, "-target", "uefi"},
	}
	for _, args := range cases {
		if _, _, err := invoke(args...); err == nil {
			t.Fatalf("accepted %q", args)
		}
	}
	old := fixture(t, "let 名前 = \"アカギ\"\nprint 名前\n")
	if _, _, err := invoke("run", old); err == nil {
		t.Fatal("old interpreter grammar accepted")
	}
	bad := fixture(t, "fn main() { let n:u64 = true; }")
	if _, _, err := invoke("check", bad); err == nil {
		t.Fatal("type mismatch accepted")
	}
}

func TestInputProtection(t *testing.T) {
	path := fixture(t, `fn main() {}`)
	for _, command := range []string{"emit", "build"} {
		_, _, err := invoke(command, "-o", path, path)
		if err == nil || !strings.Contains(err.Error(), "overwrite input") {
			t.Fatalf("%s: %v", command, err)
		}
	}
	asset := fixture(t, "asset")
	_, _, err := invoke("emit", "-asset", "EXTRA="+asset, "-o", asset, path)
	if err == nil || !strings.Contains(err.Error(), "overwrite input") {
		t.Fatalf("asset overwrite: %v", err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != `fn main() {}` {
		t.Fatal("source overwritten")
	}
}

func TestBuildRunAndEmit(t *testing.T) {
	if runtime.GOARCH != "amd64" {
		t.Skip("x86-64 backend")
	}
	if _, err := exec.LookPath("rustc"); err != nil {
		t.Skip("rustc unavailable")
	}
	path := fixture(t, `fn main() { print("種火:"); println(config("COUNT")); }`)
	want := "種火:42\n"
	out, diagnostics, err := invoke("run", "-define", "COUNT=0x2a", path)
	if err != nil || out != want {
		t.Fatalf("run: %v\n%s\n%q", err, diagnostics, out)
	}
	output := filepath.Join(t.TempDir(), "nested", "compiled"+binaryExtension("host"))
	for i := 0; i < 2; i++ {
		_, diagnostics, err = invoke("build", "-define", "COUNT=42", "-o", output, path)
		if err != nil {
			t.Fatalf("build %d: %v\n%s", i, err, diagnostics)
		}
	}
	data, err := exec.Command(output).CombinedOutput()
	if err != nil || string(data) != want {
		t.Fatalf("binary: %v, %q", err, data)
	}
	emitted := filepath.Join(t.TempDir(), "source", "program.rs")
	if _, _, err = invoke("emit", "-define", "COUNT=42", "-o", emitted, path); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(emitted)
	if err != nil || !bytes.Contains(data, []byte("42u64")) {
		t.Fatalf("emit: %v", err)
	}
	// A failed frontend must leave an existing successful executable untouched.
	before, _ := os.ReadFile(output)
	bad := fixture(t, `fn main() { missing(); }`)
	if _, _, err = invoke("build", "-o", output, bad); err == nil {
		t.Fatal("invalid build succeeded")
	}
	after, _ := os.ReadFile(output)
	if !bytes.Equal(before, after) {
		t.Fatal("failed build modified the executable")
	}
}

func TestTargetsAndDefinitions(t *testing.T) {
	path := fixture(t, `fn start(image:u64, st:u64)->u64 { return config("X"); }`)
	out, _, err := invoke("emit", "-target", "uefi", "-entry", "start", "-define", "X=0b1_010", path)
	if err != nil || !strings.Contains(out, "efi_main") || !strings.Contains(out, "10u64") {
		t.Fatalf("%v: %s", err, out)
	}
	out, _, err = invoke("emit", "-target", "library", "-define", "X=10", path)
	if err != nil || strings.Contains(out, "efi_main") || strings.Contains(out, "#![no_main]") {
		t.Fatalf("%v: %s", err, out)
	}
	for input, want := range map[string]uint64{"42": 42, "0x2A": 42, "0b101010": 42, "1_000": 1000, "010": 10} {
		if got, err := parseDefine(input); err != nil || got != want {
			t.Fatalf("%s: %d %v", input, got, err)
		}
	}
}
