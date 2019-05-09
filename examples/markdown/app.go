// +build js,wasm

package main

import (
	"syscall/js"

	"github.com/crhntr/ot"
)

func main() {
	window := js.Global()
	document := window.Get("document")
	body := document.Get("body")
	ot.NewTextarea(document, body, "ws://localhost:8080", "crhntr")
	select {}
}
