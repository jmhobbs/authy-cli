package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func register(authy *Authy) *ffcli.Command {
	return &ffcli.Command{
		Name:       "register",
		ShortUsage: "authy-cli register <country code> <phone>",
		ShortHelp:  "Register as a device on this account",
		Exec: func(_ context.Context, args []string) error {
			// todo: check if already registered
			// todo: check format of args

			if len(args) != 2 {
				return fmt.Errorf("register takes 2 arguments, got %d", len(args))
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

			var registrationStatus RegistrationStatus

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

			// todo: save all this in a better place
			f, err := os.Create("config.json")
			if err != nil {
				return err
			}
			defer f.Close()

			cfg := Config{
				AuthyId: deviceRegistration.AuthyId,
				Device: Device{
					Id:         deviceRegistration.Device.Id,
					SecretSeed: deviceRegistration.Device.SecretSeed,
					ApiKey:     deviceRegistration.Device.ApiKey,
				},
			}

			err = json.NewEncoder(f).Encode(cfg)
			if err != nil {
				return err
			}

			log.Println("Registration complete!")

			return nil
		},
	}
}
