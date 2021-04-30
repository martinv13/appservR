package templates

import (
	"net/http"
)

var LocalTemplates http.FileSystem = http.Dir("./templates")
