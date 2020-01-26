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
				context.Fail(ctx, req, resp, "Unable to parse source IP address", http.StatusInternalServerError)
				return
			}
		}

		// Parse source IP address
		sourceIP := net.ParseIP(sourceIPstr)
		if sourceIP == nil {
			log.Warningf("Unable to parse source IP address %s", sourceIPstr)
			context.Fail(ctx, req, resp, "Unable to parse source IP address", http.StatusInternalServerError)
			return
		}

		// Save source IP address in the context
		context.SetSourceIP(ctx, sourceIP)

		// Update request logger prefix
		prefix := fmt.Sprintf("%s[%s]", log.Prefix, sourceIP.String())
		log.SetPrefix(prefix)

		next.ServeHTTP(resp, req)
	})
}
