package commands

import (
	"context"
	"fmt"
	"log"
	"syscall"
	"time"

	"github.com/jmhobbs/authy-cli/store"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/xlzd/gotp"
	"golang.org/x/term"
)

var cachedBackupsPassword []byte

func getBackupsPasswordWithCache(config *store.Config) ([]byte, error) {
	if len(cachedBackupsPassword) > 0 {
		return cachedBackupsPassword, nil
	}

	if config.BackupsPassword != "" {
		cachedBackupsPassword = []byte(config.BackupsPassword)
		return []byte(config.BackupsPassword), nil
	} else {
		fmt.Print("Enter Backup Password: ")
		password, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println("")
		if err != nil {
			return nil, err
		}
		cachedBackupsPassword = password
		return password, nil
	}
}

func Token(db *store.Store) *ffcli.Command {
	return &ffcli.Command{
		Name:       "token",
		ShortUsage: "authy-cli token <name or id>",
		ShortHelp:  "Get a OTP",
		Exec: func(_ context.Context, args []string) error {
			config, err := db.Config()
			if err != nil {
				return fmt.Errorf("unable to get device config: %w", err)
			}

			if !config.IsRegistered() {
				return fmt.Errorf("no device registered")
			}

			if len(args) != 1 || args[0] == "" {
				return fmt.Errorf("must provide a token name or id")
			}

			name, totp, err := getTotp(db, config, args[0])
			if err != nil {
				if err == store.ErrNotFound {
					log.Println("unable to find an app or token")
					return nil
				}
				return fmt.Errorf("unable to get token: %w", err)
			}

			otp, expires := totp.NowWithExpiration()
			fmt.Printf("[ %s ]\n\n", name)
			fmt.Printf("Current: %s (%ds)\n", otp, expires-time.Now().Unix())
			fmt.Printf("   Next: %s\n", totp.At(expires))

			return nil
		},
	}
}

func getTotp(db *store.Store, config *store.Config, search string) (string, *gotp.TOTP, error) {
	app, err := db.App(search)
	if err == nil {
		totp, err := app.TOTP()
		return app.Name, totp, err
	}
	if err != store.ErrNotFound {
		return "", nil, fmt.Errorf("unable to get apps: %w", err)
	}

	token, err := db.Token(search)
	if err == nil {
		passphrase, err := getBackupsPasswordWithCache(config)
		if err != nil {
			return "", nil, fmt.Errorf("unable to get backup password: %w", err)
		}

		totp, err := token.TOTP(passphrase)
		return token.Name, totp, err
	}

	return "", nil, err
}
