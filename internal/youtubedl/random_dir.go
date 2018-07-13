package youtubedl

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
)

func randomDir() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("Could not read 16 random bytes")
	}

	path := filepath.Join(".", "download", fmt.Sprintf("%X", b))
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		panic("Download directory could not be created")
	}

	return path
}
