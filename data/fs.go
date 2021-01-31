package data

import (
	"net/http"
)

type HybridFileSystem struct {
	LocalFS   http.FileSystem
	BundledFS http.FileSystem
}

func (hfs *HybridFileSystem) Open(name string) (http.File, error) {
	res, err := hfs.LocalFS.Open(name)
	if err != nil {
		return (hfs.BundledFS.Open(name))
	}
	return res, err
}
