// +build js,wasm

package ot

import (
	"fmt"
	"syscall/js"
	"time"
)

func NewTextarea(document, parentElement js.Value, authorityURL, name string) {
	textarea := document.Call("createElement", "textarea")

	var (
		selected = false
	)
	textarea.Call("setAttribute", "autofocus", "")
	textarea.Call("addEventListener", "blur", js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		selected = false
		fmt.Println("selected blur")
		return nil // ignored
	}))
	textarea.Call("addEventListener", "focus", js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		selected = true
		fmt.Println("selected focus")
		return nil // ignored
	}))
	textarea.Call("addEventListener", "select", onSelect(textarea))
	textarea.Call("addEventListener", "input", onInput(textarea))
	textarea.Call("addEventListener", "paste", onPaste(textarea))
	textarea.Call("addEventListener", "keydown", onKeydown(textarea))
	textarea.Call("addEventListener", "cut", onCut(textarea))
	parentElement.Call("appendChild", textarea)
}

func getCaretPosition(textarea js.Value) (int, int) {
	// initial := 0
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		fmt.Printf("getCaretPosition failure: %v", r)
	// 	}
	// }()
	// value := textarea.Get("value")
	// selection := js.Global().Call("getSelection")
	// if selection.Truthy() {
	// 	start := textarea.Get("selectionStart")
	// 	if start.Truthy() {
	// 		return 0, 0
	// 	}
	// 	return int(start.Float()), int(textarea.Get("selectionEnd").Float())
	// }
	// textarea.Call("focus")
	// selectionRange := selection.Call("createRange")
	// rangeLen := selectionRange.Get("text").Length()
	// selectionRange.Call("moveStart", "character", -value.Length())
	// start := selectionRange.Get("text").Length() - rangeLen
	start := textarea.Get("selectionStart").Int()
	end := textarea.Get("selectionEnd").Int()
	return start, end
}

func onInput(textarea js.Value) js.Func {
	return js.FuncOf(func(target js.Value, args []js.Value) interface{} {
		start, end := getCaretPosition(textarea)
		data := target.Get("value").String()
		fmt.Printf("Insert {start: %d, end: %d, data: %s}\n", start, end, data)
		return nil // ignored
	})
}

func onSelect(textarea js.Value) js.Func {
	return js.FuncOf(func(target js.Value, args []js.Value) interface{} {
		start, end := getCaretPosition(textarea)
		val := textarea.Get("value").String()
		if start >= len(val) {
			start = len(val) - 1
		}
		if start < 0 {
			start = 0
		}
		if end >= len(val) {
			end = len(val) - 1
		}
		if end < 0 {
			end = 0
		}
		fmt.Printf("selected: %q\n", val[start:end])
		return nil // ignored
	})
}

func onPaste(textarea js.Value) js.Func {
	return js.FuncOf(func(target js.Value, args []js.Value) interface{} {
		start, end := getCaretPosition(textarea)
		val := textarea.Get("value").String()
		pre := val[:start]
		suf := val[end:]
		go func() {
			// deal with paste that takes a lot of time
			time.Sleep(4 * time.Millisecond)
			pasteVal := textarea.Get("value").String()
			clip := pasteVal[len(pre) : len(pasteVal)-len(suf)]
			fmt.Printf("Insert {start: %d, end: %d, data: %s}\n", start, end, clip)
		}()
		return nil
	})
}

func onKeydown(textarea js.Value) js.Func {
	return js.FuncOf(func(target js.Value, args []js.Value) interface{} {
		event := args[0]
		if key := event.Get("key").String(); key == "Delete" || key == "Backspace" {
			start, end := getCaretPosition(textarea)
			val := textarea.Get("value").String()
			if start == end {
				var data string
				if start > 1 && start < len(val) {
					data = val[start-1 : start]
				}
				fmt.Printf("Delete {start: %d, end: %d, data: %s}\n", start-1, start, data)
			} else {
				fmt.Printf("Delete {start: %d, end: %d, data: %s}\n", start, end, val[start:end])
			}
		}
		return nil
	})
}

func onCut(textarea js.Value) js.Func {
	return js.FuncOf(func(target js.Value, _ []js.Value) interface{} {
		start, end := getCaretPosition(textarea)
		go func() {
			time.Sleep(4 * time.Millisecond)
			value := target.Get("value").String()
			if start > 0 && start < len(value) && end > 0 && end < len(value) {
				value = value[start:end]
			}
			fmt.Printf("Delete {start: %d, end: %d, data: %s}\n", start, end, value)
		}()
		return nil
	})
}

// setCaretPosition := func(ctrl js.Value, start, end int) {
// 	if selectionRange := ctrl.Get("setSelectionRange"); selectionRange.Truthy() {
// 		ctrl.Call("focus")
// 		ctrl.Call("setSelectionRange", start, end)
// 	} else if createTextRange := ctrl.Get("createTextRange"); createTextRange.Truthy() {
// 		selectionRange := ctrl.Call("createTextRange")
// 		selectionRange.Call("collapse", true)
// 		selectionRange.Call("moveEnd", "character", end)
// 		selectionRange.Call("moveStart", "character", start)
// 		selectionRange.Call("select")
// 	}
// }
