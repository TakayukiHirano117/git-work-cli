package main

import (
	"context"
	"fmt"
	"os"

	"totonou/internal/app"
)

func main() {
	cli := app.New("", os.Stdin, os.Stdout, os.Stderr)
	if err := cli.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
