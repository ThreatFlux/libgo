package exec

import (
	"os/exec"
)

// LookPath wraps exec.LookPath to search for an executable
func LookPath(file string) (string, error) {
	return exec.LookPath(file)
}
