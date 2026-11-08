package httpapi

import (
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func init() { _ = mime.AddExtensionType(".webmanifest", "application/manifest+json") }

// serveWeb serves the built web app from dir. Unknown paths get index.html so
// client-side routes load; hashed /assets/ files are cached forever.
func serveWeb(dir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := path.Clean("/" + c.Request.URL.Path) // rooted, so ".." cannot leave dir
		if (c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead) || strings.HasPrefix(p, "/api/") {
			abort(c, http.StatusNotFound, CodeNotFound, "Not found.")
			return
		}
		name := filepath.Join(dir, filepath.FromSlash(p))
		if fi, err := os.Stat(name); err != nil || fi.IsDir() {
			name = filepath.Join(dir, "index.html")
		}
		if strings.HasPrefix(p, "/assets/") {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			c.Header("Cache-Control", "no-cache")
		}
		c.File(name)
	}
}
