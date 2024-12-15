package main
// #include <sys/user.h>
import "C"
import (
	"strconv"
)

type RegStruct_t C.struct_user_regs_struct

func (regs RegStruct_t) RSP() string {
	return strconv.Itoa(int(regs.rsp))
}


