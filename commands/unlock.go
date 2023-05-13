package commands

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"github.com/jmhobbs/authy-cli/store"
	"github.com/peterbourgon/ff/v3/ffcli"
	"golang.org/x/term"
)

func Unlock(db *store.Store) *ffcli.Command {
	return &ffcli.Command{
		Name:       "unlock",
		ShortUsage: "authy-cli ",
		ShortHelp:  "unlock",
		Exec: func(_ context.Context, args []string) error {
			config, err := db.Config()
			if err != nil {
				return fmt.Errorf("unable to get device config: %w", err)
			}

			if !config.IsRegistered() {
				return fmt.Errorf("no device registered")
			}

			fmt.Fprint(os.Stderr, "Enter Backup Password: ")
			passphrase, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Fprintln(os.Stderr)
			if err != nil {
				return err
			}

			config.BackupsPassword = string(passphrase)
			err = db.WriteConfig(*config)
			if err != nil {
				return fmt.Errorf("unable to write config: %w", err)
			}

			return nil
		},
	}
}
