package commands

import (
	"context"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"
)

var Version *ffcli.Command = &ffcli.Command{
	Name:       "version",
	ShortUsage: "authy-cli version",
	ShortHelp:  "Print authy-cli version",
	Exec: func(_ context.Context, args []string) error {
		fmt.Println("authy-cli version 0.1.0")
		return nil
	},
}
