package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func generateRandom() ([]byte, error) {
	buf := make([]byte, 32)
	_, err := rand.Read(buf)
	return buf, err
}

func loadConfig() (Config, error) {
	var cfg Config

	f, err := os.Open("config.json")
	if err != nil {
		return cfg, err
	}
	defer f.Close()
	return cfg, json.NewDecoder(f).Decode(&cfg)
}

func main() {
	logFile := time.Now().Format("2006-01-02T15:04:05") + ".log"
	log.Printf("Logging all HTTP requests to %s", logFile)
	f, err := os.Create(logFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	authy := NewAuthy(f, "", "")

	var (
		rootFlagSet = flag.NewFlagSet("authy-cli", flag.ExitOnError)
		//		verbose     = rootFlagSet.Bool("v", false, "increase log verbosity")
	)

	sync := &ffcli.Command{
		Name:       "sync",
		ShortUsage: "authy-cli sync",
		ShortHelp:  "Sync your tokens from Authy",
		Exec: func(_ context.Context, args []string) error {
			// open config
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			// todo: check we're registered

			// sync tokens via api
			tokens, err := authy.AuthenticatorTokens(cfg.AuthyId, cfg.Device.Id, cfg.Device.SecretSeed)
			if err != nil {
				return err
			}
			log.Printf("Synced %d tokens", len(tokens))

			// write to disk
			f, err := os.Create("tokens.json")
			if err != nil {
				return err
			}
			defer f.Close()
			err = json.NewEncoder(f).Encode(tokens)
			if err != nil {
				return err
			}

			return nil

			apps, err := authy.AppSync(cfg.AuthyId, cfg.Device.Id, cfg.Device.SecretSeed)
			if err != nil {
				return err
			}

			log.Printf("Synced %d apps", len(apps))

			fa, err := os.Create("apps.json")
			if err != nil {
				return err
			}
			defer fa.Close()

			return json.NewEncoder(fa).Encode(apps)
		},
	}

	export := &ffcli.Command{
		Name:       "",
		ShortUsage: "authy-cli ",
		ShortHelp:  "",
		Exec: func(_ context.Context, args []string) error {
			fmt.Println("yolo")
			return nil
		},
	}

	token := &ffcli.Command{
		Name:       "token",
		ShortUsage: "authy-cli token <name or id>",
		ShortHelp:  "Get a OTP",
		Exec: func(_ context.Context, args []string) error {
			fmt.Println("yolo")
			return nil
		},
	}

	list := &ffcli.Command{
		Name:       "list",
		ShortUsage: "authy-cli list",
		ShortHelp:  "List out all known tokens",
		Exec: func(_ context.Context, args []string) error {
			f, err := os.Open("tokens.json")
			if err != nil {
				return err
			}
			defer f.Close()

			var tokens []AuthenticatorToken

			err = json.NewDecoder(f).Decode(&tokens)
			if err != nil {
				return err
			}

			sort.Slice(tokens, func(i, j int) bool {
				return tokens[i].Name < tokens[j].Name
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

			return nil
		},
	}

	root := &ffcli.Command{
		ShortUsage: "authy-cli [flags] <subcommand>",
		FlagSet:    rootFlagSet,
		Subcommands: []*ffcli.Command{
			register(authy), sync, export, token, list},
	}

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
