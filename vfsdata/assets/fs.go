package assets

import (
	"net/http"

	"github.com/martinv13/go-shiny/vfsdata"
)

var LocalAssets http.FileSystem = http.Dir("assets")

var Assets = vfsdata.HybridFileSystem{
	LocalFS:   LocalAssets,
	BundledFS: BundledAssets,
}
