//go:build js && wasm

package main

import (
	"bytes"
	"syscall/js"

	"github.com/miroqaan/tanebi-lang/internal/tanebi"
)

var runFunction js.Func

func main() {
	runFunction = js.FuncOf(run)
	js.Global().Set("tanebiRun", runFunction)
	select {}
}

func run(_ js.Value, arguments []js.Value) any {
	if len(arguments) != 1 {
		return map[string]any{"output": "", "error": "tanebiRun expects one source string"}
	}

	var output bytes.Buffer
	err := tanebi.Run(arguments[0].String(), &output)
	message := ""
	if err != nil {
		message = err.Error()
	}
	return map[string]any{"output": output.String(), "error": message}
}

