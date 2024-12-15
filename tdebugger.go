package main

import (
	"syscall"
	"unsafe"
)

type tstate_t int
const (
	DETACHED tstate_t = 0
	INVALID 
	TERMINATED 
	PSTOPPED 
	CONTINUE 
)

type ThreadDebugger struct {
	pid int
	state tstate_t
	siginfo_valid bool
	siginfo SigInfo_t
	regs_valid bool
	regs RegStruct_t
}

func NewThreadDebugger(pid int) (*ThreadDebugger) {
	return &ThreadDebugger {
		pid: pid,
		state: INVALID,
		siginfo_valid: false,
		regs_valid: false,
	}
}

func NewThreadDebuggerAttach(pid int) (*ThreadDebugger, error) {
	_, _, syserr := syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_ATTACH, 
	uintptr(pid), 
	uintptr(0), 
	uintptr(0),
	0, 0)
	if syserr != 0 {
		return nil, syserr
	}

	var wstatus syscall.WaitStatus
	rpid, syserr2 := syscall.Wait4(pid, &wstatus, 0, nil)
	if syserr2 != nil {
		return nil, syserr2
	}

	tdb := NewThreadDebugger(rpid)
	tdb.state = PSTOPPED
	
	return tdb, nil
}

func (tdb *ThreadDebugger) Detach() (error) {
	_, _, err := syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_DETACH, uintptr(tdb.pid),0,0,0,0)
	if err == 0 {
		tdb.state = DETACHED
	}
	return err
}

func (tdb *ThreadDebugger) Cont() (error) {
	_, _, err := syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_CONT, 
	uintptr(tdb.pid),
	uintptr(0),
	uintptr(0), // signal
	0, 0)
	return err
}

func (tdb *ThreadDebugger) GetRegs() (syscall.Errno) {
	_, _, err := syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_GETREGS, 
		uintptr(tdb.pid), 
		uintptr(0),
		uintptr(unsafe.Pointer(&tdb.regs)), 
		0, 0)
	if err == 0 {
		tdb.regs_valid = true
	}
	return err
}

// its probably easier to use /proc/pid/mem
func (tdb *ThreadDebugger) Peek(addr int) (int, error) {
	r1, _, err := syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_PEEKDATA, 
	uintptr(tdb.pid),
	uintptr(addr),
	uintptr(0), 0, 0)
	return int(r1), err
}

func (tdb *ThreadDebugger) GetSigInfo() (error) {
	_, _, err := syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_GETSIGINFO, 
		uintptr(tdb.pid), 
		uintptr(unsafe.Pointer(&tdb.siginfo)), 
		uintptr(0), 
		0, 0)
	if err != 0 {
		return err
	}
	tdb.siginfo_valid = true
	return err
}
