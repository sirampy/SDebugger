# basic singlethread debugging / internals
- [x] attach by pid
- [x] attach by execve 
- [x] detach / kill tracee
- [x] setp exec
- [x] setp syscall exec ( strech - verry easy )
- [x] int / cont exec 
- [x] get registers
- [x] peek data
- [x] poke data ( not fully tested )
- [x] poke regs
- [ ] peek / poke other registers ( strech - i suspect regset NT_foo can do some realy cool stuff like get taskstruct or perf regs )

# multithread / state handling
- [x] pre-command state checks
- [ ] wait all / thread status commands
- [ ] better ThreadDebugger container ( strech )
- [ ] async wait ( strech - requires rewriting repl )
- [ ] connect all ( strech ) 
- [ ] fork / execve ...etc following ( strech )
- [ ] feat. exploration ( strech )
- [ ] cleanup on exit ( kill execve'd threads )

# breakpoints
- [ ] software breakpoints
- [ ] hardwate breakpoints ( strech - quite cool, gets arch specific fast: https://pdos.csail.mit.edu/6.828/2004/readings/i386/s12_02.htm )

# config
- [ ] read config file ( strech )
- [ ] set ptrace options ( strech - i need option setting regardless, but by this i really mean a full blown option system )

# debuginfo usage ( strech )
- [ ] source code instruction referencing
- [ ] source code variable referencing
- [ ] advanced peeking / poking
- [ ] source code injection ( quite advanced, but verry cool - uses temporary executable memory )
- [ ] feat exploration

# interception / jailing ( strech )
- [ ] signal interception (siginfo)
- [ ] sigmask
- [ ] syscall interception (syscall stop states)

# quality of life / maintainability ( strech )
- [ ] split packages / create and document module ( strech - this is just boring code quality refactoring, but I do feel like I have implemented a usable debuging library )
- [ ] verify correctness / testing EG: verify that using the i386 user.h works correctly 
- [ ] linux capabilities ( aka give CAP_SYS_PTRACE )

# feaure exploration ( strech )
- [ ] berkly packet filters
- [ ] berkly packet filters
- [ ] thread local storage
- [ ] /procfs

# Apendix
I inted to implement all non-streach features. Strech features are all things I want to do, but likely wont to keep the scope of this project in check. The featurs are in rough order of priority ( for me ). I chose which features to keep as strech based on a combination of how interesting I think they are and how much they expand the scope of the project. Usefulness is a relatively low priority (although i do stil think most of the core features are very useful)
