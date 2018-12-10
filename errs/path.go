package errs

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

var goPath = build.Default.GOPATH
var srcDir = filepath.Join(goPath, "src")

func trimGoPath(filename string) string {
	return strings.TrimPrefix(filename, fmt.Sprintf("%s%s", srcDir, string(os.PathSeparator)))
}
