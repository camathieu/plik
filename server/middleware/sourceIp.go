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

package middleware

import (
	"fmt"
	"net"
	"net/http"

	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/context"
)

// SourceIP extract the source IP address from the request and save it to the request context
func SourceIP(ctx *juliet.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		log := context.GetLogger(ctx)
		config := context.GetConfig(ctx)

		var sourceIPstr string
		if config.SourceIPHeader != "" && req.Header.Get(config.SourceIPHeader) != "" {
			// Get source ip from header if behind reverse proxy.
			sourceIPstr = req.Header.Get(config.SourceIPHeader)
		} else {
			var err error
			sourceIPstr, _, err = net.SplitHostPort(req.RemoteAddr)
			if err != nil {
				log.Warningf("Unable to parse source IP address %s", req.RemoteAddr)
				context.Fail(ctx, req, resp, "Unable to parse source IP address", 500)
				return
			}
		}

		// Parse source IP address
		sourceIP := net.ParseIP(sourceIPstr)
		if sourceIP == nil {
			log.Warningf("Unable to parse source IP address %s", sourceIPstr)
			context.Fail(ctx, req, resp, "Unable to parse source IP address", 500)
			return
		}

		// Save source IP address in the context
		ctx.Set("ip", sourceIP)

		// Update request logger prefix
		prefix := fmt.Sprintf("%s[%s]", log.Prefix, sourceIP.String())
		log.SetPrefix(prefix)

		next.ServeHTTP(resp, req)
	})
}
