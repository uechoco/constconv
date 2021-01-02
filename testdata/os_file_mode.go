package testdata

import(
	_ "os"
)

//go:generate constconv -type=os.FileMode -template=os_file_mode.tmpl -data=typename=OSFileModeStr os_file_mode.go
