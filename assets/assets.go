// Package assets carries the browser libraries every panel in the suite loads.
//
// htmx and Alpine were byte-identical copies in each consumer — 95 KB of
// vendored JavaScript committed twice, and a third copy about to appear. They
// are not a design decision and nobody edits them, which makes them the clearest
// possible case for living in one place.
//
// The stylesheet is deliberately not here. Tailwind output depends on which
// classes the consumer's own templates use, so it has to be built per product;
// what the kit can offer is the @source path that lets that build see the kit's
// templates — see cmd/kitsources.
package assets

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed js
var files embed.FS

// FS is the kit's browser assets, rooted so paths read as "js/htmx.min.js".
func FS() fs.FS { return files }

// Handler serves them under a prefix a consumer chooses.
//
// Long cache with a version query, the same bargain every panel already makes:
// the files never change within a build, and the ?v= on the URL is what makes a
// new build's assets actually load.
func Handler(prefix string) http.Handler {
	h := http.FileServer(http.FS(files))
	return http.StripPrefix(prefix, cacheForever(h))
}

func cacheForever(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}
