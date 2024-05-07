package common

import (
	"fmt"
	"os"
	"path/filepath"
)

func CreateDBDirectory(dbDir string) error {
	dir := filepath.Dir(dbDir)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0700)
		if err != nil {
			return fmt.Errorf("could not create HLoad configuration directory: %w", err)
		}
	}

	return nil
}
