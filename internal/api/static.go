//go:build production

package api

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:web_dist
var webDist embed.FS

// staticHandler serves the embedded Svelte build output.
// Falls back to index.html for SPA client-side routing.
func staticHandler() http.Handler {
	sub, err := fs.Sub(webDist, "web_dist")
	if err != nil {
		panic("embedded web_dist not found: " + err.Error())
	}

	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Check if file exists.
		f, err := sub.Open(strings.TrimPrefix(path, "/"))
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fallback to index.html for SPA routing.
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}

func hasEmbeddedWeb() bool {
	return true
}
