// +build ignore

package main

import (
	"fmt"

	"github.com/martinv13/go-shiny/data/assets"
	"github.com/martinv13/go-shiny/data/templates"
	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(assets.LocalAssets, vfsgen.Options{
		PackageName:  "assets",
		Filename:     "data/assets/vfsdata.go",
		VariableName: "BundledAssets",
	})
	if err != nil {
		fmt.Println(err)
	}
	err = vfsgen.Generate(templates.LocalTemplates, vfsgen.Options{
		PackageName:  "templates",
		Filename:     "data/templates/vfsdata.go",
		VariableName: "BundledTemplates",
	})
	if err != nil {
		fmt.Println(err)
	}
}
