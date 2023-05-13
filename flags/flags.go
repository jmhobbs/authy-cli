package flags

import (
	"flag"
	"os"
)

var writeHar bool = false
var storagePassword string = ""

func Register(fs *flag.FlagSet) {
	fs.BoolVar(&writeHar, "har", false, "Always write all HTTP requests to a HAR file, not just on error.")
	fs.StringVar(&storagePassword, "password", "", "Password to use for the storage files. Can also be set with AUTHY_CLI_STORAGE_PASSWORD environment variable.")
}

func WriteHar() bool {
	return writeHar
}

func StoragePassword() string {
	if storagePassword == "" {
		return os.Getenv("AUTHY_CLI_STORAGE_PASSWORD")
	}
	return storagePassword
}
