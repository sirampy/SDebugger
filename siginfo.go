package main
// #include <sys/wait.h>
import "C"

type SigInfo_t C.siginfo_t

func (siginfo SigInfo_t) Signum() int {
	return int(siginfo.si_signo)
}
