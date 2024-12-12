package main

import (
	"fmt"
	"log"
	"strings"

	repl "github.com/openengineer/go-repl"
)

// REPL FUNCTIONALITY
const replHelpMsg = `quit | q 		 	exit SDebug
help | h 			display this help message`

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
	command, args := parts[0], parts[1:]

	switch command {

	case "quit", "q":
		h.r.Quit()
		return ""

	case "help", "h":
		return replHelpMsg


	// FOR TESTING
	case "uid":
		return GetUID()

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
