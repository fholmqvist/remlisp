package cli

import (
	"os"
)

func createFile(filename, out string) error {
	return os.WriteFile(filename, []byte(out), os.ModePerm)
}
