package main

import (
	"strconv"
	"syscall"
	"fmt"
	"unsafe"
)

type SDebugger struct{
	attached bool
	pid int
}

func newSDebugger() *SDebugger {
	return &SDebugger{attached: false}
}

func (state *SDebugger) PGetRegs() string {
	var regs RegStruct_t
	r1, r2, err := syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_GETREGS, 
	uintptr(state.pid),
	uintptr(0),
	uintptr(unsafe.Pointer(&regs)), 0, 0)
	if err != 0 {
		return fmt.Sprintf("Err: %d %s", err, err.Error())
	}
	return fmt.Sprintf("hellooor1: %d, r2: %d, err: %d, RSP: %s", r1, r2, err, regs.RSP())
}

func (state *SDebugger) PCont(sig int) string {
	mypid, _, _ := syscall.Syscall(syscall.SYS_GETPID, 0,0,0)
	r1, r2, err := syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_CONT, 
	uintptr(state.pid),
	uintptr(0),
	uintptr(sig), 0, 0)
	if err != 0 {
		return fmt.Sprintf("mypid: %d\nErr: %d %s", mypid, err, err.Error())
	}
	return fmt.Sprintf("mypid: %d\nr1: %d, r2: %d, err: %d", mypid, r1, r2, err)
}

func (state *SDebugger) PDetach() string {
	r1, r2, err := syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_DETACH, 0,0,0,0,0)
	if err != 0 {
		return fmt.Sprintf("Err: %d %s", err, err.Error())
	}
	state.attached = false
	return "Detached succesfully" + "\n" + fmt.Sprintf("r1: %d, r2: %d, err: %d", r1, r2, err)
}

func (state *SDebugger) PPeek(addr int) string {
	r1, r2, err := syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_PEEKDATA, 
	uintptr(state.pid),
	uintptr(addr),
	uintptr(0), 0, 0)
	if err != 0 {
		return fmt.Sprintf("Err: %d %s", err, err.Error())
	}
	return fmt.Sprintf("r1: %d, r2: %d, err: %d", r1, r2, err)
}

func (state *SDebugger) PWait(pid int, hang bool) string {
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
		return fmt.Sprintf("%d, nothing to wait on", pid)
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

func (state *SDebugger) PGetSigInfo(pid int) string {
	var siginfo SigInfo_t
	var out string
	r1, r2, err := syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_GETSIGINFO, 
	uintptr(pid), 
	uintptr(unsafe.Pointer(&siginfo)), 
	uintptr(0), 0, 0)
	if err != 0 {
		return fmt.Sprintf("Err: %d %s", err, err.Error())
	}

	out += fmt.Sprintf("signo: %d\n", siginfo.Signum())
	return fmt.Sprintf("r1: %d, r2: %d, err: %d\n", r1, r2, err)
}

func (state *SDebugger) PAttach(pid int) string {
	_, _, err := syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_ATTACH, uintptr(pid), uintptr(0), uintptr(0), 0, 0)
	if err != 0 {
		return fmt.Sprintf("Err: %d %s", err, err.Error())
	}
	state.attached = true
	state.pid = pid
	out := "Attached"
	return out 
}

func (state *SDebugger) GetUID() string {
	uid := syscall.Getuid()
	return strconv.Itoa(uid)
}
