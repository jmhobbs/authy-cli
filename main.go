package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/jmhobbs/authy-cli/api"
	"github.com/jmhobbs/authy-cli/commands"
	"github.com/jmhobbs/authy-cli/flags"
	"github.com/jmhobbs/authy-cli/store"
	"github.com/peterbourgon/ff/v3/ffcli"
	"golang.org/x/term"
)

func main() {
	logFile := time.Now().Format("2006-01-02T15:04:05") + ".har"
	authy := api.New("", "")

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	db, err := store.New(path.Join(home, ".authy-cli"), func() (string, error) {
		var passphrase string = flags.StoragePassword()
		if passphrase != "" {
			return passphrase, nil
		}
		fmt.Fprint(os.Stderr, "Enter Storage Password: ")
		passphraseBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Fprintln(os.Stderr)
		return string(passphraseBytes), err
	})
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	defer shutdownHandler(false, authy, logFile)

	// defer won't run in the case of ctrl-c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		shutdownHandler(true, authy, logFile)
		os.Exit(1)
	}()

	var rootFlagSet = flag.NewFlagSet("authy-cli", flag.ExitOnError)

	flags.Register(rootFlagSet)

	root := &ffcli.Command{
		ShortUsage: "authy-cli [flags] <subcommand>",
		FlagSet:    rootFlagSet,
		Subcommands: []*ffcli.Command{
			commands.Register(db, authy),
			commands.Sync(db, authy),
			commands.Export(db),
			commands.Token(db),
			commands.List(db),
			commands.Unlock(db),
		},
		Exec: func(context.Context, []string) error {
			return flag.ErrHelp
		},
	}

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			return
		}
		log.Printf("error: %v", err)
		shutdownHandler(true, authy, logFile)
	}
}
