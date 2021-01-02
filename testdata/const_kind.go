package testdata

import (
	_ "go/constant"
)

//go:generate constconv -type=constant.Kind -template=const_kind.tmpl const_kind.go
