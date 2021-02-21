package assets

import (
	"net/http"

	"github.com/martinv13/go-shiny/modules/config"

	"github.com/martinv13/go-shiny/vfsdata"
)

var LocalAssets http.FileSystem = http.Dir(config.ExecutableFolder + "/assets")

var Assets = vfsdata.HybridFileSystem{
	LocalFS:   LocalAssets,
	BundledFS: BundledAssets,
}
