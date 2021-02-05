package main

import (
	"fmt"
	"os"

	"github.com/symcn/sym-ops/cmd/ops"
)

func main() {
	rootCmd := ops.GetRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		os.Exit(-1)
	}
}
