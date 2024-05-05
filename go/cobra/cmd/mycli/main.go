package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/nownabe/playground/go/cobra/internal/command"
)

type exitCode int

const (
	exitOK    exitCode = 0
	exitError exitCode = 1
)

func main() {
	code := run(os.Stdout, os.Stderr)
	os.Exit(int(code))
}

func run(stdout, stderr io.Writer) exitCode {
	ctx := context.Background()

	cmd, err := command.New(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "command.New failed: %v\n", err)
		return exitError
	}

	if cmd, err := cmd.ExecuteContextC(ctx); err != nil {
		fmt.Fprintf(stderr, "cmd.Execute failed: %v\n", err)
		fmt.Fprintf(stderr, cmd.UsageString())
		return exitError
	}

	return exitOK
}
