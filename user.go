package main
// #include <sys/user.h>
import "C"
import (
	"strconv"
	"reflect"
	"unsafe"
)

type RegStruct_t C.struct_user_regs_struct

func (regs RegStruct_t) RSP() string {
	return strconv.Itoa(int(regs.rsp))
}

//TODO: error handling / better robustness
func (regs *RegStruct_t) SetByName(rname string, rval int) string {
	field := reflect.ValueOf(regs).Elem().FieldByName(rname)
	vprfield := field.Addr()
	rfield := (*int)(unsafe.Pointer(vprfield.Pointer()))
	*rfield = rval
	return ""
}

