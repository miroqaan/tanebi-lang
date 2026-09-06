package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The fake backend and child are one small Go executable. The backend copies
// itself to rustc's requested output; the no-argument child exits normally with
// code 23. No panic, fault, platform error dialog, shell, or network is involved.
const failureBackendSource = `package main
import (
    "fmt"
    "os"
)
func main() {
    if len(os.Args) == 1 {
        if err := os.WriteFile(os.Getenv("TANEBI_TEST_CHILD_MARKER"), []byte("executed"), 0600); err != nil { fmt.Fprintln(os.Stderr,err); os.Exit(91) }
        fmt.Println("safe child executed")
        os.Exit(23)
    }
    output := ""
    for n:=1;n+1<len(os.Args);n++ { if os.Args[n]=="-o" { output=os.Args[n+1] } }
    if output=="" { fmt.Fprintln(os.Stderr,"missing backend output"); os.Exit(92) }
    self,err:=os.Executable(); if err!=nil { fmt.Fprintln(os.Stderr,err); os.Exit(93) }
    data,err:=os.ReadFile(self); if err!=nil { fmt.Fprintln(os.Stderr,err); os.Exit(94) }
    if err=os.WriteFile(output,data,0755);err!=nil { fmt.Fprintln(os.Stderr,err); os.Exit(95) }
    if err=os.WriteFile(os.Getenv("TANEBI_TEST_BACKEND_MARKER"),[]byte(output),0600);err!=nil { fmt.Fprintln(os.Stderr,err); os.Exit(96) }
    if os.Getenv("TANEBI_TEST_BACKEND_MODE")=="fail" { fmt.Fprintln(os.Stderr,"fake backend failure"); os.Exit(42) }
}
`

func TestBackendFailureAndSafeChildExit(t *testing.T) {
	if runtime.GOARCH != "amd64" {
		t.Skip("the current backend supports x86-64 hosts only")
	}
	backendDir := buildFailureBackend(t)
	t.Setenv("PATH", backendDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	source := fixture(t, `fn main() { println("実行"); }`)
	for _, command := range []string{"build", "run"} {
		t.Run("failed_"+command, func(t *testing.T) {
			dir := t.TempDir()
			childMarker := filepath.Join(dir, "child.marker")
			backendMarker := filepath.Join(dir, "backend.marker")
			t.Setenv("TANEBI_TEST_CHILD_MARKER", childMarker)
			t.Setenv("TANEBI_TEST_BACKEND_MARKER", backendMarker)
			t.Setenv("TANEBI_TEST_BACKEND_MODE", "fail")
			artifact := filepath.Join(dir, "existing"+binaryExtension("host"))
			previous := []byte("previous successful artifact")
			if err := os.WriteFile(artifact, previous, 0600); err != nil {
				t.Fatal(err)
			}
			args := []string{command}
			if command == "build" {
				args = append(args, "-o", artifact)
			}
			args = append(args, source)
			out, diagnostics, err := invoke(args...)
			if err == nil || !strings.Contains(err.Error(), "native backend failed") {
				t.Fatalf("error=%v", err)
			}
			if !strings.Contains(diagnostics, "fake backend failure") {
				t.Fatalf("backend stderr lost: %q", diagnostics)
			}
			if out != "" {
				t.Fatalf("failed backend produced program stdout: %q", out)
			}
			if _, err = os.Stat(childMarker); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("program ran after backend failure: %v", err)
			}
			after, err := os.ReadFile(artifact)
			if err != nil || !bytes.Equal(after, previous) {
				t.Fatalf("failed backend altered prior artifact: %q, %v", after, err)
			}
			assertFailureBackendTempRemoved(t, backendMarker)
		})
	}
	t.Run("nonzero_child", func(t *testing.T) {
		dir := t.TempDir()
		childMarker := filepath.Join(dir, "child.marker")
		backendMarker := filepath.Join(dir, "backend.marker")
		t.Setenv("TANEBI_TEST_CHILD_MARKER", childMarker)
		t.Setenv("TANEBI_TEST_BACKEND_MARKER", backendMarker)
		t.Setenv("TANEBI_TEST_BACKEND_MODE", "success")
		out, diagnostics, err := invoke("run", source)
		var childError *exec.ExitError
		if !errors.As(err, &childError) || childError.ExitCode() != 23 {
			t.Fatalf("child exit code not propagated: %v", err)
		}
		if out != "safe child executed\n" || diagnostics != "" {
			t.Fatalf("unexpected streams: stdout=%q, stderr=%q", out, diagnostics)
		}
		marker, err := os.ReadFile(childMarker)
		if err != nil || string(marker) != "executed" {
			t.Fatalf("generated child was not executed: %q, %v", marker, err)
		}
		assertFailureBackendTempRemoved(t, backendMarker)
	})
}

func buildFailureBackend(t *testing.T) string {
	t.Helper()
	goCommand, err := exec.LookPath("go")
	if err != nil {
		t.Skip("Go executable unavailable for the local backend fixture")
	}
	dir := t.TempDir()
	source := filepath.Join(dir, "fake_backend.go")
	if err = os.WriteFile(source, []byte(failureBackendSource), 0600); err != nil {
		t.Fatal(err)
	}
	name := "rustc"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	command := exec.Command(goCommand, "build", "-o", filepath.Join(dir, name), source)
	command.Dir = dir
	command.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off", "CGO_ENABLED=0", "GOOS="+runtime.GOOS, "GOARCH="+runtime.GOARCH)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build local fake backend: %v\n%s", err, output)
	}
	return dir
}

func assertFailureBackendTempRemoved(t *testing.T, marker string) {
	t.Helper()
	output, err := os.ReadFile(marker)
	if err != nil || len(output) == 0 {
		t.Fatalf("backend was not invoked: %q, %v", output, err)
	}
	if _, err = os.Stat(filepath.Dir(string(output))); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("compiler temporary directory was not cleaned: %v", err)
	}
}

func TestLinkedInputProtection(t *testing.T) {
	for _, kind := range []string{"hardlink", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			source := fixture(t, `fn main() {}`)
			dir := t.TempDir()
			alias := filepath.Join(dir, "source-alias.tanebi")
			link := os.Link
			if kind == "symlink" {
				link = os.Symlink
			}
			if err := link(source, alias); err != nil {
				t.Skipf("%s is unavailable: %v", kind, err)
			}
			asset := filepath.Join(dir, "media.dat")
			assetAlias := filepath.Join(dir, "media-alias.dat")
			assetBytes := []byte{10, 20, 30, 40}
			if err := os.WriteFile(asset, assetBytes, 0600); err != nil {
				t.Fatal(err)
			}
			if err := link(asset, assetAlias); err != nil {
				t.Skipf("asset %s is unavailable: %v", kind, err)
			}
			// A missing backend proves rejection happens before compilation.
			t.Setenv("PATH", "")
			for _, command := range []string{"emit", "build"} {
				for _, pair := range [][2]string{{alias, source}, {source, alias}} {
					_, _, err := invoke(command, "-o", pair[0], pair[1])
					if err == nil || !strings.Contains(err.Error(), "overwrite input") {
						t.Fatalf("%s source alias was not protected: %v", command, err)
					}
				}
				_, _, err := invoke(command, "-asset", "MEDIA="+asset, "-o", assetAlias, source)
				if err == nil || !strings.Contains(err.Error(), "overwrite input") {
					t.Fatalf("%s asset alias was not protected: %v", command, err)
				}
			}
			for _, path := range []string{source, alias} {
				data, err := os.ReadFile(path)
				if err != nil || string(data) != `fn main() {}` {
					t.Fatalf("source changed: %q, %v", data, err)
				}
			}
			for _, path := range []string{asset, assetAlias} {
				data, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(data, assetBytes) {
					t.Fatalf("asset changed: %v, %v", data, err)
				}
			}
		})
	}
}
