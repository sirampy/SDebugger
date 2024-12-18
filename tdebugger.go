package main

import (
	"syscall"
	"unsafe"
)

type tstate_t int
const (
	DETACHED tstate_t = iota
	TERMINATED 
	PSTOPPED 
	RUNNING 
)

func (state tstate_t) String() string {
	switch state {
	case DETACHED:
		return "DETACHED"
	case TERMINATED:
		return "TERMINATED"
	case PSTOPPED:
		return "PSTOPPED"
	case RUNNING:
		return "RUNNING"
	}
	return ""
}

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
		state: DETACHED,
		siginfo_valid: false,
		regs_valid: false,
	}
}

func NewThreadDebuggerExecve(path string, argv *[]string, envv *[]string) (*ThreadDebugger, error) {
	if envv == nil {
		tmp := syscall.Environ()
		envv = &tmp
	}
	pid, _, err := syscall.Syscall(syscall.SYS_FORK,0,0,0)
	if err != 0 {
		return nil, err
	}
	if pid == 0 {
		_, _, syserr := syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_TRACEME, 0, 0, 0, 0, 0)
		if syserr != 0 {
			syscall.Exit(0)
		}

		_ = syscall.Exec(path, *argv, *envv)
		syscall.Exit(0)

	} else {
		var wstatus syscall.WaitStatus
		rpid, err := syscall.Wait4(int(pid), &wstatus, 0, nil)
		if err != nil {
			return nil, err
		}
		tdb := NewThreadDebugger(rpid)
		//TODO: setp past execve
		tdb.UpdateStateFromWait(wstatus)
		//TODO: update with propper option system
		tdb.setOptions(syscall.PTRACE_O_TRACEEXEC)
		return tdb, nil
	}
	return nil, DBGERR
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
	rpid, err := syscall.Wait4(pid, &wstatus, 0, nil)
	if err != nil {
		return nil, err
	}

	tdb := NewThreadDebugger(rpid)
	tdb.UpdateStateFromWait(wstatus)
	
	return tdb, nil
}
// GUARDS
func (tdb *ThreadDebugger) stateGuard(states ...tstate_t) (bool, error) {
	_, err := tdb.UpdateState()
	if err != nil {
		return false, err
	}
	for _, state := range states {
		if tdb.state == state {
			return true, nil
		}
	}
	return false, nil
}


// HELPERS

func (tdb *ThreadDebugger) UpdateState() (bool, error) {
	var wstatus syscall.WaitStatus
	var newwstatus syscall.WaitStatus
	r, err := syscall.Wait4(tdb.pid, &newwstatus, syscall.WNOHANG, nil)
	if r == 0 {
		return true, nil
	}
	for r == tdb.pid {
		wstatus = newwstatus
		r, err = syscall.Wait4(tdb.pid, &newwstatus, syscall.WNOHANG, nil)
	}
	if err != nil {
		return true, err
	}
	return tdb.UpdateStateFromWait(wstatus), nil
}

func (tdb *ThreadDebugger) UpdateStateFromWait(wstatus syscall.WaitStatus) (bool) {
	if wstatus.Stopped(){
		/* FOR STRECH FEAT
		* Ptrace-stops can be further subdivided into signal-
        * delivery-stop (if no SIGTRAP, else), group-stop, syscall-stop, PTRACE_EVENT stops, and
        * so on. 
	    * For thepurposes of ptrace, a tracee which is blocked in a system call
        * (such as read(2), pause(2), etc.)  is nevertheless considered to
        * be running, even if the tracee is blocked for a long time.
	    * I dont think this is an issue so long as we remember to wait for stop
		*/
		tdb.state = PSTOPPED
	}else if wstatus.Exited() {

		tdb.state = TERMINATED
		return false
	} 
	return true
}

// DEBUG COMMANDS
func (tdb *ThreadDebugger) setOptions(options int) (error) {
	_, _, err := syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_SETOPTIONS, uintptr(tdb.pid),0,0,0,0)
	if err != 0 {
		return err
	}
	return nil
}

func (tdb *ThreadDebugger) Detach() (syscall.Errno, error) {
	ok, syserr := tdb.stateGuard(PSTOPPED)
	if !ok {
		if syserr != nil {
			return 0, syserr
		}
		myerr := DBGERR_STATE_NOT_SUPPORTED
		return 0, &myerr
	}

	_, _, err := syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_DETACH, uintptr(tdb.pid),0,0,0,0)
	if err != 0 {
		return err, nil
	}
	tdb.state = DETACHED
	return 0, nil
}

func (tdb *ThreadDebugger) Kill() error {
	ok, syserr := tdb.stateGuard(PSTOPPED, RUNNING)
	if !ok {
		if syserr != nil {
			return syserr
		}
		myerr := DBGERR_STATE_NOT_SUPPORTED
		return &myerr
	}
	err := syscall.Kill(tdb.pid, syscall.SIGKILL)
	if	err ==  nil {
		tdb.state = PSTOPPED
	}
	return err
}

func (tdb *ThreadDebugger) Step() (error) {
	ok, err := tdb.stateGuard(PSTOPPED)
	if !ok {
		if err != nil {
			return err
		}
		myerr := DBGERR_STATE_NOT_SUPPORTED
		return &myerr
	}

	_, _, errno := syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_SINGLESTEP, 
	uintptr(tdb.pid),
	uintptr(0),
	uintptr(0), 
	0, 0)
	if errno == 0 {
		return nil
	}
	return &errno
}

func (tdb *ThreadDebugger) Cont() (error) {
	ok, err := tdb.stateGuard(PSTOPPED)
	if !ok {
		if err != nil {
			return err
		}
		myerr := DBGERR_STATE_NOT_SUPPORTED
		return &myerr
	}

	_, _, errno := syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_CONT, 
	uintptr(tdb.pid),
	uintptr(0),
	uintptr(0), // signal
	0, 0)
	if errno == 0 {
		tdb.state = RUNNING
		return nil
	}
	return &errno
}

func (tdb *ThreadDebugger) Int() (error) {
	ok, err := tdb.stateGuard(RUNNING)
	if !ok {
		if err != nil {
			return err
		}
		myerr := DBGERR_STATE_NOT_SUPPORTED
		return &myerr
	}

	// have to use kill as PTRACE_INTERRUPT isnt suported on my hardware - could do feature check as PTRACE_INTERRUPT is better (interrupts through syscalls)
	err = syscall.Kill(tdb.pid, syscall.SIGCHLD)
	if	err ==  nil {
		tdb.state = PSTOPPED
	}
	return err
}

func (tdb *ThreadDebugger) GetRegs() (error) {
	ok, err := tdb.stateGuard(PSTOPPED)
	if !ok {
		if err != nil {
			return err
		}
		myerr := DBGERR_STATE_NOT_SUPPORTED
		return &myerr
	}

	_, _, errno := syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_GETREGS, 
		uintptr(tdb.pid), 
		uintptr(0),
		uintptr(unsafe.Pointer(&tdb.regs)), 
		0, 0)
	if errno == 0 {
		tdb.regs_valid = true
		return nil
	}
	return errno
}

// its probably easier to use /proc/pid/mem
func (tdb *ThreadDebugger) Peek(addr int) (int, error) {
	ok, err := tdb.stateGuard(PSTOPPED)
	if !ok {
		if err != nil {
			return 0, err
		}
		myerr := DBGERR_STATE_NOT_SUPPORTED
		return 0, &myerr
	}

	r1, _, errno := syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_PEEKDATA, 
	uintptr(tdb.pid),
	uintptr(addr),
	uintptr(0), 0, 0)
	if errno  == 0{
		return int(r1), nil
	}
	return 0, err
}

func (tdb *ThreadDebugger) GetSigInfo() (error) {
	ok, err := tdb.stateGuard(PSTOPPED)
	if !ok {
		if err != nil {
			return err
		}
		myerr := DBGERR_STATE_NOT_SUPPORTED
		return &myerr
	}

	_, _, errno := syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_GETSIGINFO, 
		uintptr(tdb.pid), 
		uintptr(unsafe.Pointer(&tdb.siginfo)), 
		uintptr(0), 
		0, 0)
	if errno ==  0 {
		tdb.siginfo_valid = true
		return nil
	}
	return errno
}
