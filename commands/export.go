package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jmhobbs/authy-cli/model"
	"github.com/jmhobbs/authy-cli/store"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func Export(db *store.Store) *ffcli.Command {
	return &ffcli.Command{
		Name:       "export",
		ShortUsage: "authy-cli ",
		ShortHelp:  "export <file>",
		Exec: func(_ context.Context, args []string) error {
			// todo: optionally decrypt on export
			// todo: export formats for import elsewhere
			config, err := db.Config()
			if err != nil {
				return fmt.Errorf("unable to get device config: %w", err)
			}

			if !config.IsRegistered() {
				return fmt.Errorf("no device registered")
			}

			tokens, err := db.Tokens()
			if err != nil {
				return fmt.Errorf("unable to get tokens: %w", err)
			}

			apps, err := db.Apps()
			if err != nil {
				return fmt.Errorf("unable to get apps: %w", err)
			}

			out := struct {
				Config store.Config
				Tokens []model.Token
				Apps   []model.App
			}{
				Config: *config,
				Tokens: tokens,
				Apps:   apps,
			}

			return json.NewEncoder(os.Stdout).Encode(out)
		},
	}
}
