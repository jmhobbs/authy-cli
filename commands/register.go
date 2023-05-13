package commands

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmhobbs/authy-cli/api"
	"github.com/jmhobbs/authy-cli/flags"
	"github.com/jmhobbs/authy-cli/store"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func generateRandom() ([]byte, error) {
	buf := make([]byte, 32)
	_, err := rand.Read(buf)
	return buf, err
}

func Register(db *store.Store, authy *api.Authy) *ffcli.Command {
	fs := flag.NewFlagSet("authy-cli register", flag.ExitOnError)
	flags.Register(fs)

	return &ffcli.Command{
		Name:       "register",
		ShortUsage: "authy-cli register <country code> <phone>",
		ShortHelp:  "Register as a device on this account",
		FlagSet:    fs,
		Exec: func(_ context.Context, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("register takes 2 arguments, got %d", len(args))
			}

			config, err := db.Config()
			if err != nil && err != store.ErrNotFound {
				return fmt.Errorf("unable to get device config: %w", err)
			}

			if config.IsRegistered() {
				// todo: prompt to unregister
				return fmt.Errorf("device already registered")
			}

			deviceUUID := uuid.Must(uuid.NewRandom()).String()

			log.Printf("Checking status of account %s-%s", args[0], args[1])
			status, err := authy.UserStatus(args[0], args[1])
			if err != nil {
				return err
			}
			log.Printf("Account ID is %d", status.AuthyId)

			signatureBytes, err := generateRandom()
			if err != nil {
				return err
			}
			signature := hex.EncodeToString(signatureBytes)

			registrationResult, err := authy.RegistrationStart(status.AuthyId, signature)
			if err != nil {
				return err
			}

			log.Printf("Device registration request sent to other devices via %s", registrationResult.Provider)
			log.Println("Please accept this request.")

			var registrationStatus api.RegistrationStatus

			// retry every ~2s for two minutes
			var accepted = false
			for retry := 0; !accepted && retry < 60; retry++ {
				registrationStatus, err = authy.RegistrationStatus(status.AuthyId, signature, registrationResult.RequestId)
				if err != nil {
					return err
				}
				switch registrationStatus.Status {
				case "pending":
					time.Sleep(2 * time.Second)
					continue
				case "accepted":
					log.Printf("Registration approved!")
					accepted = true
					break
				default:
					return fmt.Errorf("unknown registration status: %q", registrationStatus.Status)
				}
			}

			deviceRegistration, err := authy.RegistrationComplete(status.AuthyId, deviceUUID, registrationStatus.Pin)
			if err != nil {
				return err
			}

			cfg := store.Config{
				AuthyId: deviceRegistration.AuthyId,
				Device: store.Device{
					Id:         deviceRegistration.Device.Id,
					SecretSeed: deviceRegistration.Device.SecretSeed,
					ApiKey:     deviceRegistration.Device.ApiKey,
				},
			}

			err = db.WriteConfig(cfg)
			if err != nil {
				return fmt.Errorf("unable to save config: %w", err)
			}

			log.Println("Registration complete!")

			return nil
		},
	}
}
