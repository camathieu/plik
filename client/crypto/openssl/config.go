package openssl

import (
	"github.com/root-gg/utils"
)

// BackendConfig object
type BackendConfig struct {
	Openssl    string
	Cipher     string
	Passphrase string
	Options    string
}

// NewOpenSSLBackendConfig instantiate a new Backend Configuration
// from config map passed as argument
func NewOpenSSLBackendConfig(config map[string]interface{}) (ob *BackendConfig) {
	ob = new(BackendConfig)
	ob.Openssl = "/usr/bin/openssl"
	ob.Cipher = "aes256"
	utils.Assign(ob, config)
	return
}
