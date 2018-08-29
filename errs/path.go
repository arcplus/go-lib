package errs

import (
	"runtime"
	"strings"
)

// prefixSize is used internally to trim the user specific path from the
// front of the returned filenames from the runtime call stack.
var prefixSize int

// goPath is the deduced path based on the location of this file as compiled.
var goPath string

func init() {
	_, file, _, ok := runtime.Caller(0)
	if file == "?" {
		return
	}
	if ok {
		// We know that the end of the file should be:
		// github.com/arcplus/go-lib/errs/path.go
		size := len(file)
		suffix := len("github.com/arcplus/go-lib/errs/path.go")
		goPath = file[:size-suffix]
		// remove vendor
		if strings.HasSuffix(goPath, "vendor/") {
			goPath = strings.TrimSuffix(goPath, "vendor/")
		}
		prefixSize = len(goPath)
	}
}

func trimGoPath(filename string) string {
	if strings.HasPrefix(filename, goPath) {
		return filename[prefixSize:]
	}
	return filename
}
