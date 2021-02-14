package templates

import (
	"net/http"

	"github.com/martinv13/go-shiny/vfsdata"
)

var LocalTemplates http.FileSystem = http.Dir("templates")

var Templates = vfsdata.HybridFileSystem{
	LocalFS:   LocalTemplates,
	BundledFS: BundledTemplates,
}
