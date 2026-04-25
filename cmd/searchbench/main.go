package main

import (
	"context"
	"fmt"
	"os"

	"github.com/becker63/searchbench-go/internal/surface/cli"
)

func main() {
	if err := cli.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
