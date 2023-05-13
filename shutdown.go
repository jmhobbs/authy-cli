package main

import (
	"log"
	"os"

	"github.com/jmhobbs/authy-cli/api"
	"github.com/jmhobbs/authy-cli/flags"
)

func shutdownHandler(writeHar bool, authy *api.Authy, logFile string) {
	if authy.MadeRequests() && (writeHar || flags.WriteHar()) {
		log.Printf("Writing all HTTP requests to %s", logFile)
		log.Println("!!! THIS HAR MAY CONTAIN SENSITIVE INFORMATION !!!")
		f, err := os.Create(logFile)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		authy.WriteHAR(f)
	}
}
