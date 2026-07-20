// +build tools

// Package tools records tool dependencies. It cannot actually be compiled.
package tools

import (
	_ "github.com/google/wire/cmd/wire"
)
