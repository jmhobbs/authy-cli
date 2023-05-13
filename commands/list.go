package commands

import (
	"context"
	"fmt"
	"sort"

	"github.com/jmhobbs/authy-cli/store"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func List(db *store.Store) *ffcli.Command {
	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "authy-cli list",
		ShortHelp:  "List out all known tokens",
		Exec: func(_ context.Context, args []string) error {
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

			sort.Slice(tokens, func(i, j int) bool {
				return tokens[i].Name < tokens[j].Name
			})
			sort.Slice(apps, func(i, j int) bool {
				return apps[i].Name < apps[j].Name
			})

			maxType := 4
			maxName := 4
			for _, token := range tokens {
				if len(token.AccountType) > maxType {
					maxType = len(token.AccountType)
				}
				if len(token.Name) > maxName {
					maxName = len(token.Name)
				}
			}

			fmt.Println("[ Authenticator Tokens ]")
			fmt.Println()

			fmtString := fmt.Sprintf("%% %ds | %% %ds | %% 10s\n", maxType, maxName)
			fmt.Printf(fmtString, "Type", "Name", "ID")
			for i := 0; i < maxType; i++ {
				fmt.Print("-")
			}
			fmt.Print("-|-")
			for i := 0; i < maxName; i++ {
				fmt.Print("-")
			}
			fmt.Println("-|-----------")
			for _, token := range tokens {
				fmt.Printf(fmtString, token.AccountType, token.Name, token.UniqueId)
			}

			maxName = 4
			for _, app := range apps {
				if len(app.Name) > maxName {
					maxName = len(app.Name)
				}
			}

			fmt.Println()
			fmt.Println("[ Authy Apps ]")
			fmt.Println()

			fmtString = fmt.Sprintf("%% %ds | %% 24s\n", maxName)
			fmt.Printf(fmtString, "Name", "ID")
			for i := 0; i < maxName; i++ {
				fmt.Print("-")
			}
			fmt.Println("-|-------------------------")
			for _, app := range apps {
				fmt.Printf(fmtString, app.Name, app.Id)
			}

			return nil
		},
	}
}
