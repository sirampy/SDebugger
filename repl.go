package main

import (
	"fmt"
	"log"
	"strings"
	"strconv"
	"syscall"
	"reflect"

	repl "github.com/openengineer/go-repl"
)

// REPL FUNCTIONALITY
const replHelpMsg = `quit | q 		 		exit dbg
help | h 				display this help message

ctx CTX					switch to debugging thread of ontext CTX

CONTROL FLOW:
attach PID 				trace process PID
trace PROGRAM ARGS		execute and trace PROGRAM
detach 					detach from tracee
kill					kill child process

interrupt | int			interrupt tracee
continue | cont | c		continue tracee
step | s				continue one instruction
stepsc | ssc			continue until next syscall
pid						PID of tracee

PEEK TRACEE:
regdump | reg			get register values
peek ADDR				get data stored in ADDR 

DEBUG COMMANDS:
wait 					listen for signal from tracee
dbgprint 				print Debugger struct`
const delMsg = `FAILED: Thread terminated`
const ctxMsg = `FAILED: Not in a thread context`
const stateMsg = "FAILED: required tracee state "
const invStateMsg = "FAILED: not supported for state "

// REPL HANDLER
type ReplHandler struct {
	r *repl.Repl
	dbg *Debugger
}

// REPL HANDLER INTERFACE IMPLEMENTATION
//NOTE: state hints are hints and only hints, as states are checked at command execute time
func (h *ReplHandler) Prompt() string {
	t := h.dbg.CtxThread()
	if t == nil {
		return "> "
	}
	statesymbol := "> "
	if t.state == PSTOPPED {
		statesymbol = ": "
	}
	return strconv.Itoa(h.dbg.ctx_thread) + statesymbol
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

	case "trace":
		if len(args) == 0 {
			return "trace expects a program to execute"
		}
		err := h.dbg.Execve(&args)
		if err != nil {
			return err.Error()
		}
		return "Attached"

	case "detach":
		t := h.dbg.CtxThread()
		if t == nil {
			return stateMsg
		}
		
		interr, syserr := t.Detach()
		if syserr != nil {
			return syserr.Error()
		} else if interr != 0 {
			return interr.Error()
		}
		return "Detached"

	case "kill":
		t := h.dbg.CtxThread()
		if t == nil {
			return stateMsg
		}
		
		 err := t.Kill()
		 if err != nil {
			 return err.Error()
		 }
		 return "Process Killed"


	case "step", "s":
		t := h.dbg.CtxThread()
		if t == nil {
			return "Not in a thread context"
		}

		err := t.StepSyscall()
		if err != nil {
			return err.Error()
		}
		return "Steping"

	case "stepsc", "ssc":
		t := h.dbg.CtxThread()
		if t == nil {
			return "Not in a thread context"
		}

		err := t.Step()
		if err != nil {
			return err.Error()
		}
		return "Steping"
	
	case "continue", "cont", "c":
		t := h.dbg.CtxThread()
		if t == nil {
			return "Not in a thread context"
		}
		err := t.Cont()
		if err != nil {
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
		

	case "regdump", "reg":
		t := h.dbg.CtxThread()
		if t == nil {
			return "Not in a thread context"
		}
		err := t.GetRegs()
		if err != nil {
			return err.Error()
		}
		return fmt.Sprintf("%+v",t.regs)
		
	case "peek":
		if len(args) != 1 {
			return "peek expects 1 argument"
		}
		arg1, err := strconv.Atoi(args[0])
		if err != nil {
			return "Failed to convert arg1 into an int"
		}
		t := h.dbg.CtxThread()
		if t == nil {
			return "Not in a thread context"
		}
		data, err2 := t.Peek(arg1)
		if err2 != nil {
			return err2.Error()
		}
		return fmt.Sprintf("Data: 0x%X", data)
	
	case "poke":
		if len(args) != 2 {
			return "poke expects 2 argument"
		}
		arg1, err := strconv.Atoi(args[0])
		if err != nil {
			return "Failed to convert arg1 into an int"
		}
		arg2, err := strconv.Atoi(args[1])
		if err != nil {
			return "Failed to convert arg2 into an int"
		}
		t := h.dbg.CtxThread()
		if t == nil {
			return "Not in a thread context"
		}
		err2 := t.Poke(arg1, arg2)
		if err2 != nil {
			return err2.Error()
		}
		return "Write successful"

	case "pokereg", "preg" :
		if len(args) != 2 {
			return "poke expects 2 argument"
		}
		t := h.dbg.CtxThread()
		if t == nil {
			return "Not in a thread context"
		}
		err := t.GetRegs()
		if err != nil {
			return err.Error()
		}

		_, valid := reflect.TypeOf(t.regs).FieldByName(args[0])
		if !valid {
			return "Invalid register name"
		}
		arg2, err2 := strconv.Atoi(args[1])
		if err2 != nil {
			return "Failed to convert arg2 into an int"
		}
		
		t.regs.SetByName(args[0], arg2)
		fmt.Printf("%+v",t.regs)
		err3 := t.SetRegs(&t.regs)
		if err3 != nil {
			return err3.Error()
		}
		return "Write successful"

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
