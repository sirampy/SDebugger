package main

import (
	"fmt"
	"runtime"
)

func main() {
	runtime.LockOSThread()
	Repl()
	dbg := &SDebugger{}
	fmt.Println(dbg.PAttach(626838))
	fmt.Println(dbg.PPeek(0x479cdb))
}
