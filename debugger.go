package main

import (
	"strconv"

	unix "golang.org/x/sys/unix"
)

type SDebugger struct{}

func (state *SDebugger) GetUID() string {
	uid := unix.Getuid()
	return strconv.Itoa(uid)
}

func GetUID() string {
	uid := unix.Getuid()
	return strconv.Itoa(uid)
}
