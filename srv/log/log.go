package logsrv

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var out io.Writer = os.Stdout

// Log writes a log.
func Log(a ...interface{}) {
	var builder strings.Builder

	fmt.Fprintln(&builder, a...)

	io.WriteString(out, builder.String())
}

// Logf writes a formatted log.
func Logf(format string, a ...interface{}) {
	var builder strings.Builder

	format += "\n"

	fmt.Fprintf(&builder, format, a...)

	io.WriteString(out, builder.String())
}
