package main

import (
	"fmt"
	"log"
	"strings"
	"strconv"
	"syscall"

	repl "github.com/openengineer/go-repl"
)

// REPL FUNCTIONALITY
const replHelpMsg = `quit | q 		 		exit dbg
help | h 				display this help message

ctx CTX					switch to debugging thread of ontext CTX

CONTROL FLOW:
attach PID 				trace process PID
detach 					detach from tracee
interrupt | int			interrupt tracee
continue | cont | c		continue tracee
step | s				continue one instruction
pid						PID of tracee

PEEK TRACEE:
regdump					get register values
getsiginfo 				get info on signal from tracee

DEBUG COMMANDS:
wait 					listen for signal from tracee
dbgprint 				print Debugger struct`

// REPL HANDLER
type ReplHandler struct {
	r *repl.Repl
	dbg *Debugger
}

// REPL HANDLER INTERFACE IMPLEMENTATION
func (h *ReplHandler) Prompt() string {
	if h.dbg.ctx_thread == -1 {
		return "> "
	}
	return fmt.Sprintf("%d> ", h.dbg.ctx_thread)
}

func (h *ReplHandler) Tab(buff string) string {
	return ""
}

func (h *ReplHandler) Eval(buff string) string {

	parts := strings.Fields(buff)
	if len(parts) == 0 {
		return ""
	}
	command, args := parts[0], parts[1:]

	switch command {

	case "quit", "q":
		h.r.Quit()
		return ""

	case "help", "h":
		return replHelpMsg

	case "ctx":
		if len(args) != 1 {
			return "attach expects 1 argument"
		}
		arg1, err := strconv.Atoi(args[0])
		if err != nil {
			return "Failed to convert arg1 into an int"
		}
		err2 := h.dbg.CtxSwitch(arg1)
		if err2 != nil {
			return err2.Error()
		}
		return "Success"
	
	case "pid":
		t := h.dbg.CtxThread()
		if t == nil {
			return "No tracee attached"
		}
		return fmt.Sprintf("PID: %d", t.pid)

	case "attach":
		if len(args) != 1 {
			return "attach expects 1 argument"
		}
		arg1, err := strconv.Atoi(args[0])
		if err != nil {
			return "Failed to convert arg1 into an int"
		}
		err = h.dbg.Attach(arg1)
		if err != nil {
			return err.Error()
		}
		return "Attached"

	case "step", "s":
		t := h.dbg.CtxThread()
		if t == nil {
			return "Not in a thread context"
		}
		err := t.Step()
		if err != 0 {
			return err.Error()
		}
		return "Steping"
	
	case "continue", "cont", "c":
		t := h.dbg.CtxThread()
		if t == nil {
			return "Not in a thread context"
		}
		err := t.Cont()
		if err != 0 {
			return err.Error()
		}
		return "Continuing"

	case "interrupt", "int":
		t := h.dbg.CtxThread()
		if t == nil {
			return "Not in a thread context"
		}
		err := t.Int()
		if err != nil {
			return err.Error()
		}
		return "Interrupted"
		

	case "regdump":
		t := h.dbg.CtxThread()
		if t == nil {
			return "Not in a thread context"
		}
		err := t.GetRegs()
		if err != 0 {
			return err.Error()
		}
		return fmt.Sprintf("%+v",t.regs)


	/*
	case "detach":
		if h.sdb.attached {
			return h.sdb.PDetach()
		}
		return "No tracee attached"
		
	case "peek":
		if len(args) != 1 {
			return "peek expects 1 argument"
		}
		arg1, err := strconv.Atoi(args[0])
		if err != nil {
			return "Failed to convert arg1 into an int"
		}
		return h.sdb.PPeek(arg1)
	

	case "getsiginfo":
		if h.sdb.attached {
			return h.sdb.PGetSigInfo(h.sdb.pid)
		}
		return "not currently attached"
	*/

	// FOR TESTING

	case "wait":
		if len(args) == 1 {
			arg1, err := strconv.Atoi(args[0])
			if err != nil {
				return "Failed to convert arg1 into an int"
			}
			return h.dbg.Wait(arg1, false)
		}
		return h.dbg.Wait(h.dbg.threads[h.dbg.ctx_thread].pid, false)

	case "dbgprint":
		return fmt.Sprintf("%+v", h.dbg.threads[h.dbg.ctx_thread])

	case "mypid":
		return strconv.Itoa(syscall.Getpid())
	}

	return "unrecognised command"
}

// HANDLER OTHER
func newReplHandler() *ReplHandler {
	handler := &ReplHandler{}
	handler.r = repl.NewRepl(handler)
	handler.dbg = newDebugger()
	return handler
}

// REPL RUN
func Repl() {
	fmt.Println("Welcome to SDebug. run help to get started or q to quit")
	handler := newReplHandler()
	
	if err := handler.r.Loop(); err != nil {
		log.Fatal(err)
	}
}
