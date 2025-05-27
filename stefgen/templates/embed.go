package templates

import (
	"embed"
	_ "embed"
)

//go:embed */*.tmpl
var Templates embed.FS
