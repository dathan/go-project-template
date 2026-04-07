package server

import (
	"io/fs"
	"net/http"
)

// frontendHandler serves the embedded React SPA. Unknown paths fall back to
// index.html so React Router can handle client-side navigation.
func frontendHandler(assets fs.FS) http.Handler {
	fsHandler := http.FileServer(http.FS(assets))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to open the requested path.
		if _, err := assets.Open(r.URL.Path); err != nil {
			// Not found — serve index.html for SPA fallback routing.
			http.ServeFileFS(w, r, assets, "index.html")
			return
		}
		fsHandler.ServeHTTP(w, r)
	})
}
