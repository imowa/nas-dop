package server

import (
	"io/fs"
	"net/http"
	"strconv"

	"nas-dop/web"
)

// staticHandler returns a handler that serves embedded static files with Cache-Control
// (optimization-recommendations.md: static assets long cache, versioned when changed).
func (s *Server) staticHandler() (http.Handler, error) {
	sub, err := fs.Sub(web.FS, "static")
	if err != nil {
		return nil, err
	}
	maxAge := s.cfg.StaticCacheMaxAge
	if maxAge <= 0 {
		maxAge = 86400
	}
	h := http.StripPrefix("/static/", http.FileServer(http.FS(sub)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private, max-age="+strconv.Itoa(maxAge))
		h.ServeHTTP(w, r)
	}), nil
}
