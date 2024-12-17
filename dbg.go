package main

import (
	"runtime"
)

func main() {
	runtime.LockOSThread()
	Repl()
}
