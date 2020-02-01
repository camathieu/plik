package middleware

import (
	"fmt"
	"net"
	"net/http"

	"github.com/root-gg/plik/server/context"
)

// SourceIP extract the source IP address from the request and save it to the request context
func SourceIP(ctx *context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		log := ctx.GetLogger()
		config := ctx.GetConfig()

		var sourceIPstr string
		if config.SourceIPHeader != "" && req.Header.Get(config.SourceIPHeader) != "" {
			// Get source ip from header if behind reverse proxy.
			sourceIPstr = req.Header.Get(config.SourceIPHeader)
		} else {
			var err error
			sourceIPstr, _, err = net.SplitHostPort(req.RemoteAddr)
			if err != nil {
				ctx.InternalServerError(fmt.Errorf("unable to parse source IP address : %s", err))
				return
			}
		}

		// Parse source IP address
		sourceIP := net.ParseIP(sourceIPstr)
		if sourceIP == nil {
			ctx.InvalidParameter("IP address")
			return
		}

		// Save source IP address in the context
		ctx.SetSourceIP(sourceIP)

		// Check if IP is whitelisted
		setWhitelisted(ctx)

		// Update request logger prefix
		prefix := fmt.Sprintf("%s[%s]", log.Prefix, sourceIP.String())
		log.SetPrefix(prefix)

		next.ServeHTTP(resp, req)
	})
}

// IsWhitelisted return true if the IP address in the request context is whitelisted
// TODO : This could be evaluated lazily
func setWhitelisted(ctx *context.Context) {
	uploadWhitelist := ctx.GetConfig().GetUploadWhitelist()

	// Check if the source IP address is in whitelist
	whitelisted := false
	if len(uploadWhitelist) > 0 {
		sourceIP := ctx.GetSourceIP()
		if sourceIP != nil {
			for _, subnet := range uploadWhitelist {
				if subnet.Contains(sourceIP) {
					whitelisted = true
					break
				}
			}
		}
	} else {
		whitelisted = true
	}

	ctx.SetWhitelisted(whitelisted)
}
