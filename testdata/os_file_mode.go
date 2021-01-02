package testdata

import(
	_ "os"
)

// go:generate constconv -type=os.FileMode -template=os_file_mode.tmpl -data="typename=OSFileModeStr" testdata/os_file_mode.go