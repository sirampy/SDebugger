package main

import (
	"runtime"
	"fmt"
	"log"
	"strings"
	"strconv"
	"syscall"

	repl "github.com/openengineer/go-repl"
)

// REPL FUNCTIONALITY
const replHelpMsg = `quit | q 		 	exit SDebug
help | h 			display this help message

attach PID 			trace process PID
detach 				detach from tracee
wait 				listen for signal from tracee
getsiginfo 			get info on signal from tracee
pid					PID of tracee`

// REPL HANDLER
type ReplHandler struct {
	r *repl.Repl
	sdb *SDebugger
}

// REPL HANDLER INTERFACE IMPLEMENTATION
func (h *ReplHandler) Prompt() string {
	return "> "
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

	case "attach":
		if len(args) != 1 {
			return "attach expects 1 argument"
		}
		arg1, err := strconv.Atoi(args[0])
		if err != nil {
			return "Failed to convert arg1 into an int"
		}
		return h.sdb.PAttach(arg1)

	case "detach":
		if h.sdb.attached {
			return h.sdb.PDetach()
		}
		return "No tracee attached"

	case "continue", "cont", "c":
		if h.sdb.attached {
			return h.sdb.PCont(0)
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
	
	case "pid":
		if h.sdb.attached {
			return strconv.Itoa(h.sdb.pid)
		}
		return "not currently attached"

	case "regdump":
		return h.sdb.PGetRegs()

	case "getsiginfo":
		if h.sdb.attached {
			return h.sdb.PGetSigInfo(h.sdb.pid)
		}
		return "not currently attached"

	// FOR TESTING
	case "wcont":
		o2 := h.sdb.PWait(h.sdb.pid, true)
		o3 := h.sdb.PCont(0)
		return fmt.Sprintf("%s\n%s\n", o2, o3)

	case "test":
		if len(args) != 1 {
			return "peek expects 1 argument"
		}
		arg1, err := strconv.Atoi(args[0])
		if err != nil {
			return "Failed to convert arg1 into an int"
		}
		o1 := h.sdb.PAttach(arg1)
		o2 := h.sdb.PWait(h.sdb.pid, true)
		o3 := h.sdb.PCont(0)
		return fmt.Sprintf("%s\n%s\n%s\n", o1, o2, o3)

	case "wait":
		if !h.sdb.attached {
			return "not currently attached"
		}
		if len(args) == 1 {
			arg1, err := strconv.Atoi(args[0])
			if err != nil {
				return "Failed to convert arg1 into an int"
			}
			return h.sdb.PWait(arg1, false)
		}
		return h.sdb.PWait(h.sdb.pid, false)

	case "uid":
		return h.sdb.GetUID()

	case "mypid":
		return strconv.Itoa(syscall.Getpid())

	case "arg1":
		if len(args) > 0 {
			return args[0]
		}
		return "not enough args"
	}

	return "unrecognised command"
}

// HANDLER OTHER
func newReplHandler() *ReplHandler {
	handler := &ReplHandler{}
	handler.r = repl.NewRepl(handler)
	handler.sdb = newSDebugger()
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
