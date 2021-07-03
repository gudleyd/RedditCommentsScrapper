package main

import (
	"fmt"
)

var LogLevelColor = [...]string {"\033[0m", "\033[32m", "\033[33m", "\033[31m"}
const (
	ColorReset   = "\033[0m"
)

func Logf(level int, format string, a ...interface{}) {
	fmt.Print(LogLevelColor[level])
    fmt.Printf(format, a...)
	fmt.Print(ColorReset)
}