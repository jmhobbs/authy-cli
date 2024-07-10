package commands

import (
	"context"
	"encoding/base32"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/jmhobbs/authy-cli/model"
	"github.com/jmhobbs/authy-cli/store"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type exportToken struct {
	model.Token
	DecryptedSeed string `json:"decrypted_seed,omitempty"`
}

func Export(db *store.Store) *ffcli.Command {
	exportFlags := flag.NewFlagSet("export", flag.ExitOnError)
	var decryptOnExport *bool = exportFlags.Bool("decrypt", false, "decrypt tokens before exporting")

	return &ffcli.Command{
		Name:       "export",
		ShortUsage: "authy-cli export [flags] <file>",
		ShortHelp:  "export [flags] <file>",
		FlagSet:    exportFlags,
		Exec: func(_ context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing file argument, use - for stdout")
			}

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

			var passphrase []byte

			if *decryptOnExport {
				fmt.Fprintln(os.Stderr, "!!! Warning: Decrypting tokens before export !!!")
				passphrase, err = getBackupsPasswordWithCache(config)
				if err != nil {
					return fmt.Errorf("unable to get passphrase: %w", err)
				}
			}

			exportTokens := []exportToken{}
			for _, t := range tokens {
				et := exportToken{Token: t}
				if *decryptOnExport {
					decryptedSeed, err := t.Decrypt(passphrase)
					if err != nil {
						return fmt.Errorf("unable to decrypt token %q: %w", t.Name, err)
					}
					et.DecryptedSeed = base32.StdEncoding.EncodeToString(decryptedSeed)
				}
				exportTokens = append(exportTokens, et)
			}

			out := struct {
				Config store.Config  `json:"config"`
				Tokens []exportToken `json:"tokens"`
				Apps   []model.App   `json:"apps"`
			}{
				Config: *config,
				Tokens: exportTokens,
				Apps:   apps,
			}

			var sink io.Writer
			if args[0] == "-" {
				sink = os.Stdout
			} else {
				f, err := os.Create(args[0])
				if err != nil {
					return fmt.Errorf("unable to create file: %w", err)
				}
				defer f.Close()
				sink = f
			}

			return json.NewEncoder(sink).Encode(out)
		},
	}
}
