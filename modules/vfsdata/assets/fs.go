package assets

import (
	"net/http"
)

var LocalAssets http.FileSystem = http.Dir("./assets")
