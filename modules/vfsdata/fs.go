package vfsdata

import (
	"net/http"

	"github.com/martinv13/go-shiny/modules/config"
	"github.com/martinv13/go-shiny/modules/vfsdata/assets"
	"github.com/martinv13/go-shiny/modules/vfsdata/templates"
)

type HybridFileSystem struct {
	LocalFS   http.FileSystem
	BundledFS http.FileSystem
}

func (hfs HybridFileSystem) Open(name string) (http.File, error) {
	res, err := hfs.LocalFS.Open(name)
	if err != nil {
		return (hfs.BundledFS.Open(name))
	}
	return res, err
}

type StaticPaths struct {
	Assets    HybridFileSystem
	Templates HybridFileSystem
}

func NewStaticPaths(c *config.Config) *StaticPaths {
	return &StaticPaths{
		Assets: HybridFileSystem{
			LocalFS:   http.Dir(c.ExecutableFolder + "/assets"),
			BundledFS: assets.BundledAssets,
		},
		Templates: HybridFileSystem{
			LocalFS:   http.Dir(c.ExecutableFolder + "/templates"),
			BundledFS: templates.BundledTemplates,
		},
	}
}
