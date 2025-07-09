package fga

import (
	_ "embed"
)

//go:embed fga.json
var ACLModelBytes []byte

//go:embed fga.md5
var ACLModelMD5 string
