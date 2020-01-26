package pgp

import (
	"os"

	"github.com/root-gg/utils"
	"golang.org/x/crypto/openpgp"
)

// BackendConfig object
type BackendConfig struct {
	Gpg       string
	Keyring   string
	Recipient string
	Email     string
	Entity    *openpgp.Entity
}

// NewPgpBackendConfig instantiate a new Backend Configuration
// from config map passed as argument
func NewPgpBackendConfig(config map[string]interface{}) (pbc *BackendConfig) {
	pbc = new(BackendConfig)
	pbc.Gpg = "/usr/bin/gpg"
	pbc.Keyring = os.Getenv("HOME") + "/.gnupg/pubring.gpg"
	utils.Assign(pbc, config)
	return
}
