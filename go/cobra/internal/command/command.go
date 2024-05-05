package command

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func New(ctx context.Context) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Short: "My CLI",
		Long:  "My CLI is a CLI tool for everything you need",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("cmd: %#v\n", cmd)
			fmt.Printf("args: %#v\n", args)
		},
	}

	return cmd, nil
}
