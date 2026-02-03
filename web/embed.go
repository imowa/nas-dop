// Package web holds embedded static files and templates for the app.
package web

import "embed"

//go:embed static
//go:embed templates
var FS embed.FS
