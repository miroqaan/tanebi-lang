// tanebi checks, compiles, and runs typed programs. There is no interpreter.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/miroqaan/tanebi-lang/internal/native"
)

const usage = `TANEBI compiler
Usage: tanebi <check|build|run|emit> [options] source.tanebi ...
  check   Validate syntax, names, and types without invoking rustc
  build   Compile a native executable or EFI application
  run     Compile a host program to a temporary executable, then run it
  emit    Emit validated Rust backend source
Options (before source paths):
  -target host|uefi|library  Default: host; library is for check/emit only
  -entry NAME               Default: main (host), entry (uefi)
  -o PATH                   Output path for build/emit (emit defaults to stdout)
  -asset NAME=PATH           Embedded asset; repeatable
  -define NAME=U64           Compile-time configuration; repeatable
  -opt 0|1|2|3|s|z           Backend optimization (default: 2)
build/run require rustc. UEFI builds require x86_64-unknown-uefi.
run always compiles; the former untyped script syntax is not supported.
`

type values map[string]string

func (v values) String() string { return fmt.Sprint(map[string]string(v)) }
func (v values) Set(s string) error {
	key, value, ok := strings.Cut(s, "=")
	if !ok || key == "" || value == "" {
		return fmt.Errorf("expected NAME=VALUE")
	}
	if _, exists := v[key]; exists {
		return fmt.Errorf("duplicate option %q", key)
	}
	v[key] = value
	return nil
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		var child *exec.ExitError
		if errors.As(err, &child) && child.ExitCode() > 0 {
			os.Exit(child.ExitCode())
		}
		fmt.Fprintln(os.Stderr, "TANEBI:", err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("missing command; use 'tanebi help'")
	}
	command := args[0]
	if command == "help" || command == "-h" || command == "--help" {
		_, err := io.WriteString(stdout, usage)
		return err
	}
	switch command {
	case "check", "build", "run", "emit":
	default:
		return fmt.Errorf("unknown command %q; use 'tanebi run source.tanebi' to compile and run", command)
	}
	flags := flag.NewFlagSet("tanebi "+command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() { fmt.Fprint(stderr, usage) }
	output := flags.String("o", "", "output path")
	target := flags.String("target", "host", "host, uefi, or library")
	entry := flags.String("entry", "", "entry function")
	opt := flags.String("opt", "2", "backend optimization")
	assets, defines := values{}, values{}
	flags.Var(assets, "asset", "embedded asset NAME=PATH")
	flags.Var(defines, "define", "configuration NAME=U64")
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() == 0 {
		return fmt.Errorf("no source files")
	}
	if *target != "host" && *target != "uefi" && *target != "library" {
		return fmt.Errorf("unknown target %q; expected host, uefi, or library", *target)
	}
	if command == "run" && *target != "host" {
		return fmt.Errorf("run supports only -target host; build UEFI programs and boot them separately")
	}
	if command == "build" && *target == "library" {
		return fmt.Errorf("use emit -target library to generate code for a native harness")
	}
	if *output != "" && command != "build" && command != "emit" {
		return fmt.Errorf("-o is supported only by build and emit")
	}
	switch *opt {
	case "0", "1", "2", "3", "s", "z":
	default:
		return fmt.Errorf("invalid -opt %q; expected 0, 1, 2, 3, s, or z", *opt)
	}
	opts := native.Options{Target: *target, Entry: *entry, Assets: assets, Defines: map[string]uint64{}}
	for key, text := range defines {
		v, err := parseDefine(text)
		if err != nil {
			return fmt.Errorf("invalid -define %s=%s: %w", key, text, err)
		}
		opts.Defines[key] = v
	}
	sources := make([]native.Source, 0, flags.NArg())
	inputs := make([]string, 0, flags.NArg()+len(assets))
	for _, path := range flags.Args() {
		if strings.HasPrefix(path, "-") {
			return fmt.Errorf("options must precede source paths: %s", path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sources = append(sources, native.Source{Name: path, Text: string(data)})
		inputs = append(inputs, path)
	}
	for _, path := range assets {
		inputs = append(inputs, path)
	}
	code, err := native.Compile(sources, opts)
	if err != nil {
		return err
	}
	if command == "check" {
		_, err = fmt.Fprintf(stdout, "OK (%s): %d source file(s)\n", *target, len(sources))
		return err
	}
	if command == "emit" {
		if *output == "" || *output == "-" {
			_, err = io.WriteString(stdout, code)
			return err
		}
		if err = protectInputs(*output, inputs); err != nil {
			return err
		}
		return writeArtifact(*output, strings.NewReader(code), 0644)
	}
	if *target == "host" && runtime.GOARCH != "amd64" {
		return fmt.Errorf("the current native backend supports x86-64 hosts only")
	}
	if command == "build" {
		if *output == "" {
			base := filepath.Base(sources[0].Name)
			base = strings.TrimSuffix(base, filepath.Ext(base))
			*output = filepath.Join("build", base+binaryExtension(*target))
		}
		if *output == "-" {
			return fmt.Errorf("build requires a file output, not stdout")
		}
		if err = protectInputs(*output, inputs); err != nil {
			return err
		}
	}
	rustc, err := exec.LookPath("rustc")
	if err != nil {
		return fmt.Errorf("rustc is required for build/run; install Rust or use check/emit: %w", err)
	}
	work, err := os.MkdirTemp("", "tanebi-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)
	rustSource := filepath.Join(work, "program.rs")
	if err = os.WriteFile(rustSource, []byte(code), 0600); err != nil {
		return err
	}
	binary := filepath.Join(work, "program"+binaryExtension(*target))
	backendArgs := []string{"--edition=2024", "--crate-name", "tanebi_program", "-C", "opt-level=" + *opt, "-C", "panic=abort", "-C", "codegen-units=1"}
	if *target == "uefi" {
		backendArgs = append(backendArgs, "--target", "x86_64-unknown-uefi")
	}
	backendArgs = append(backendArgs, rustSource, "-o", binary)
	backend := exec.Command(rustc, backendArgs...)
	backend.Stdout, backend.Stderr = stderr, stderr
	if err = backend.Run(); err != nil {
		hint := ""
		if *target == "uefi" {
			hint = "; check rustup target add x86_64-unknown-uefi"
		}
		return fmt.Errorf("native backend failed%s: %v", hint, err)
	}
	if command == "run" {
		child := exec.Command(binary)
		child.Stdin, child.Stdout, child.Stderr = stdin, stdout, stderr
		return child.Run()
	}
	file, err := os.Open(binary)
	if err != nil {
		return err
	}
	defer file.Close()
	if err = writeArtifact(*output, file, 0755); err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, *output)
	return err
}

func parseDefine(text string) (uint64, error) {
	raw, base := strings.ReplaceAll(text, "_", ""), 10
	if strings.HasPrefix(raw, "0x") || strings.HasPrefix(raw, "0X") {
		raw, base = raw[2:], 16
	} else if strings.HasPrefix(raw, "0b") || strings.HasPrefix(raw, "0B") {
		raw, base = raw[2:], 2
	}
	return strconv.ParseUint(raw, base, 64)
}

func binaryExtension(target string) string {
	if target == "uefi" {
		return ".efi"
	}
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func protectInputs(output string, inputs []string) error {
	out, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	outInfo, _ := os.Stat(out)
	for _, input := range inputs {
		in, err := filepath.Abs(input)
		if err != nil {
			return err
		}
		inInfo, _ := os.Stat(in)
		if out == in || (runtime.GOOS == "windows" && strings.EqualFold(out, in)) || (outInfo != nil && inInfo != nil && os.SameFile(outInfo, inInfo)) {
			return fmt.Errorf("output would overwrite input %q", input)
		}
	}
	return nil
}

// Stage beside the destination so a failed build/write cannot truncate it.
func writeArtifact(path string, source io.Reader, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".tanebi-output-*")
	if err != nil {
		return err
	}
	temp := file.Name()
	defer os.Remove(temp)
	_, copyErr := io.Copy(file, source)
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if err = os.Chmod(temp, mode); err != nil {
		return err
	}
	return os.Rename(temp, path)
}
