package main

import (
	"strconv"
	"syscall"
	"fmt"
)

type Debugger struct{
	threads []*ThreadDebugger
	ctx_thread int // context::thread - may break out into more complex ctx struct down the line
}

type dbg_error int

const (
	DBGERR dbg_error = iota + 1 		// general error
	DBGERR_CTX_NONEXISTENT
	DBGERR_STATE_NOT_SUPPORTED
)

func (err dbg_error) Error() string {
	switch err {
	case DBGERR:
		return "Failed"
	case DBGERR_CTX_NONEXISTENT:
		return "Context dosn't exist"
	case DBGERR_STATE_NOT_SUPPORTED:
		return "Invalid tracee state for opperation"
	}
	return "Unknwn Error"
}

// HELPERS
func newDebugger() *Debugger {
	return &Debugger{threads: make([]*ThreadDebugger, 0, 1), ctx_thread: -1}
}

func (dbg *Debugger) CtxThread() *ThreadDebugger {
	if dbg.ctx_thread == -1 {
		return nil
	}
	return dbg.threads[dbg.ctx_thread]
}

func (dbg *Debugger) deleteThread(ctx_thread int) {
	dbg.threads = append(dbg.threads[:ctx_thread], dbg.threads[ctx_thread+1:]...)
}

//TODO: unusued
func (dbg *Debugger) UpdateCurrentThreadState() (/* deleted */ bool, error){
	del, err := dbg.CtxThread().UpdateState()
	if err != nil {
		return del, err
	} else if del {
		dbg.deleteThread(dbg.ctx_thread)
		dbg.ctx_thread = -1
	}
	return del, nil
}

func (dbg *Debugger) CtxSwitch(new_ctx int) (error) {
	if new_ctx >= len(dbg.threads)  || new_ctx < 0 {
		return DBGERR_CTX_NONEXISTENT
	}
	dbg.ctx_thread = new_ctx
	return nil
}

// for testing
func (dbg *Debugger) Wait(pid int, hang bool) string {
	var wstatus syscall.WaitStatus
	rusage := syscall.Rusage{}
	opt := syscall.WNOHANG
	if hang {
		opt = 0
	}
	wpid, werr := syscall.Wait4(pid, &wstatus, opt, &rusage)

	if werr != nil {
		return werr.Error()
	}

	if wpid == 0 {
		return fmt.Sprintf("%d, nothing to wait on", wpid)
	}
	signal := strconv.Itoa(int(wstatus))
	if wstatus.Exited(){
		signal = "exited"
	} else if wstatus.Stopped(){
		signal = "stoped"
	} else if wstatus.Continued() {
		signal = "continued"
	}
	return fmt.Sprintf("pid: %d \nwstatus: %X, sig: %s", wpid,  wstatus, signal)
}


func (dbg *Debugger) Attach(pid int) error {
	tdb, err := NewThreadDebuggerAttach(pid)
	if err == nil {
		dbg.threads = append(dbg.threads, tdb)
		dbg.ctx_thread = len(dbg.threads) - 1
	}
	return err 
}

func (dbg *Debugger) Execve(cmd *[]string) error {
	tdb, err := NewThreadDebuggerExecve((*cmd)[0], cmd, nil)
	if err == nil {
		dbg.threads = append(dbg.threads, tdb)
		dbg.ctx_thread = len(dbg.threads) - 1
	}
	return err 
}

func (state *Debugger) GetUID() string {
	uid := syscall.Getuid()
	return strconv.Itoa(uid)
}
