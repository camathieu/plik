/**

    Plik upload server

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

package bolt
//
//import (
//	"fmt"
//	"io/ioutil"
//	"os"
//	"testing"
//
//	"github.com/stretchr/testify/require"
//)
//
//func newBackend(t *testing.T) (backend *Backend, cleanup func()) {
//	dir, err := ioutil.TempDir("", "pliktest")
//	require.NoError(t, err, "unable to create temp directory")
//
//	backend, err = NewBackend(&Config{Path: dir + "/plik.db"})
//	require.NoError(t, err, "unable to create bolt metadata backend")
//	cleanup = func() {
//		err := os.RemoveAll(dir)
//		if err != nil {
//			fmt.Println(err)
//		}
//	}
//
//	return backend, cleanup
//}
//
//func TestNewConfig(t *testing.T) {
//	params := make(map[string]interface{})
//	path := "bolt.db"
//	params["Path"] = path
//	config := NewConfig(params)
//	require.Equal(t, path, config.Path)
//}
//
//func TestNewBoltMetadataBackend_InvalidPath(t *testing.T) {
//	_, err := NewBackend(&Config{Path: string([]byte{0})})
//	require.Error(t, err)
//}
//
//func TestNewBoltMetadataBackend(t *testing.T) {
//	backend, cleanup := newBackend(t)
//	defer cleanup()
//	require.NotNil(t, backend)
//}
