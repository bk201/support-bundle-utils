package utils

import (
	"bufio"
	"fmt"
	"os"
)

func WriteStdout(msg string) {
	stdout := bufio.NewWriter(os.Stdout)
	fmt.Fprintf(stdout, msg)
	stdout.Flush()
}
