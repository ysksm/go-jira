//go:build !production

package api

import "net/http"

// staticHandler returns nil in development mode.
// Use vite dev server with proxy instead.
func staticHandler() http.Handler {
	return nil
}

func hasEmbeddedWeb() bool {
	return false
}
