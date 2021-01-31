package templates

import (
	"net/http"

	"github.com/martinv13/go-shiny/data"
)

var LocalTemplates http.FileSystem = http.Dir("templates")

var Templates = data.HybridFileSystem{
	LocalFS:   LocalTemplates,
	BundledFS: BundledTemplates,
}
