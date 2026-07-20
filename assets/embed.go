package assets

import (
	"embed"
	"net/http"
)

//go:embed css js
var bundled embed.FS

// BundledAssets serves assets/css and assets/js embedded into the binary at
// build time, rooted the same way http.Dir("./assets") would be.
var BundledAssets http.FileSystem = http.FS(bundled)
