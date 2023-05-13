package commands

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/jmhobbs/authy-cli/api"
	"github.com/jmhobbs/authy-cli/flags"
	"github.com/jmhobbs/authy-cli/model"
	"github.com/jmhobbs/authy-cli/store"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func Sync(db *store.Store, authy *api.Authy) *ffcli.Command {
	fs := flag.NewFlagSet("authy-cli sync", flag.ExitOnError)
	flags.Register(fs)

	return &ffcli.Command{
		Name:       "sync",
		ShortUsage: "authy-cli sync",
		ShortHelp:  "Sync your tokens from Authy",
		FlagSet:    fs,
		Exec: func(_ context.Context, args []string) error {
			config, err := db.Config()
			if err != nil {
				return fmt.Errorf("unable to get device config: %w", err)
			}

			if !config.IsRegistered() {
				return fmt.Errorf("no device registered")
			}

			// todo: send a list of token ID's we already have

			// sync tokens via api
			tokens, err := authy.AuthenticatorTokens(config.AuthyId, config.Device.Id, config.Device.SecretSeed)
			if err != nil {
				return err
			}
			log.Printf("Synced %d tokens", len(tokens))

			if err = db.WriteTokens(tokens); err != nil {
				return fmt.Errorf("unable to write tokens: %w", err)
			}

			// todo: load apps from store and merge/update

			apps, err := authy.AppSync(config.AuthyId, config.Device.Id, config.Device.SecretSeed, []model.App{})
			if err != nil {
				return err
			}

			log.Printf("Synced %d apps", len(apps))

			// todo: merge with store instead of overwrite

			if err = db.WriteApps(apps); err != nil {
				return fmt.Errorf("unable to write apps: %w", err)
			}

			return nil
		},
	}
}
