package main

import (
	"fmt"
	"os"

	"github.com/your-moon/mn_compiler_go_version/cli"
)

func main() {
	compiler := cli.New()
	if err := compiler.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
