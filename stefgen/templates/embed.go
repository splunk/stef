package templates

import (
	"embed"
	_ "embed"
)

//go:embed *.go.tmpl
var Templates embed.FS
