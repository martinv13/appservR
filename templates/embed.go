package templates

import (
	"embed"
	"net/http"
)

//go:embed admin auth shared
var bundled embed.FS

// BundledTemplates serves templates/admin, templates/auth and
// templates/shared embedded into the binary at build time, rooted the same
// way http.Dir("./templates") would be.
var BundledTemplates http.FileSystem = http.FS(bundled)
