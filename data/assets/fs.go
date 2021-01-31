package assets

import (
	"net/http"

	"github.com/martinv13/go-shiny/data"
)

var LocalAssets http.FileSystem = http.Dir("assets")

var Assets = data.HybridFileSystem{
	LocalFS:   LocalAssets,
	BundledFS: BundledAssets,
}
