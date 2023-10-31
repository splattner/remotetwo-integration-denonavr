package cmd

import (
	"context"
	"fmt"
	"os"
)

// CheckError prints err to stderr and exits with code 1 if err is not nil. Otherwise, it is a
// no-op.
func CheckError(err error) {
	if err != nil {
		if err != context.Canceled {
			fmt.Fprintf(os.Stderr, "An error occurred: %v\n", err)
		}
		os.Exit(1)
	}
}

// Exit prints msg (with optional args), plus a newline, to stderr and exits with code 1.
func Exit(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
