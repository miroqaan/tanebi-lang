package main

import (
	"fmt"
	"os"

	"github.com/miroqaan/tanebi-lang/internal/tanebi"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: tanebi <file.tanebi>")
		os.Exit(2)
	}

	source, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "TANEBI file error: %v\n", err)
		os.Exit(1)
	}

	if err := tanebi.Run(string(source), os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "TANEBI error: %v\n", err)
		os.Exit(1)
	}
}
