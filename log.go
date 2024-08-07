package main

import (
	"fmt"
)

// log writes a log.
func log(a ...interface{}) {
	fmt.Println(a...)
}

// logf writes a formatted log.
func logf(format string, a ...interface{}) {
	format += "\n"

	fmt.Printf(format, a...)
}
