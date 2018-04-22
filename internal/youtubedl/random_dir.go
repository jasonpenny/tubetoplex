package youtubedl

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
)

func randomDir() string {
	b := make([]byte, 16)
	rand.Read(b)

	path := filepath.Join(".", "download", fmt.Sprintf("%X", b))
	os.MkdirAll(path, os.ModePerm)
	return path
}
