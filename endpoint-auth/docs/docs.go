package docs

import (
	_ "embed"

	"github.com/swaggo/swag"
)

//go:embed swagger.json
var doc string

type s struct{}

func (s *s) ReadDoc() string {
	return doc
}

func init() {
	swag.Register(swag.Name, &s{})
}
