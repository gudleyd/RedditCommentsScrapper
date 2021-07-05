package mylog

import (
	"fmt"
)

var LogLevelColor = [...]string {"\033[0m", "\033[32m", "\033[33m", "\033[31m"}
const (
	ColorReset   = "\033[0m"
)

func Logf(level int, format string, a ...interface{}) {
    fmt.Printf(LogLevelColor[level] + format + ColorReset, a...)
}
