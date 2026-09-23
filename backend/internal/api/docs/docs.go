package docs

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed static
var staticFS embed.FS

//go:embed openapi.json
var openAPISpecJSON []byte

// Handler serves the OpenAPI 3.1 specification and the Swagger UI for the
// external developer API (/api/v1/*). The doc routes are intentionally
// public: documentation must not require an API key, and these routes are not
// wrapped by RequireAPIKey or WithUsage.
//
// Served under GET /docs/{path...}:
//   - /docs/openapi.json        -> the OpenAPI 3.1 specification
//   - /docs/ and /docs/index.html -> Swagger UI entry point
//   - /docs/<asset>             -> Swagger UI static assets
func Handler() http.Handler {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServerFS(sub)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/docs")
		if path == "" || path == "/index.html" {
			path = "/"
		}

		if path == "/openapi.json" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("Cache-Control", "public, max-age=3600")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(openAPISpecJSON)
			return
		}

		if strings.HasSuffix(path, ".json") || strings.HasSuffix(path, ".js") ||
			strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".html") {
			w.Header().Set("Cache-Control", "public, max-age=3600")
		}

		r2 := r.Clone(r.Context())
		r2.URL.Path = path
		fileServer.ServeHTTP(w, r2)
	})
}
