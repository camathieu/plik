/**

    Plik upload client

The MIT License (MIT)

Copyright (c) <2015>
	- Mathieu Bodjikian <mathieu@bodjikian.fr>
	- Charles-Antoine Mathieu <skatkatt@root.gg>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
**/

package pgp

import (
	"os"

	"github.com/root-gg/plik/client/Godeps/_workspace/src/github.com/root-gg/utils"
	"github.com/root-gg/plik/client/Godeps/_workspace/src/golang.org/x/crypto/openpgp"
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
func NewPgpBackendConfig(config map[string]interface{}) (pb *BackendConfig) {
	pb = new(BackendConfig)
	pb.Gpg = "/usr/bin/gpg"
	pb.Keyring = os.Getenv("HOME") + "/.gnupg/pubring.gpg"
	utils.Assign(pb, config)
	return
}
