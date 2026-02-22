package module

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const tempBase = "WebGainInstaller"

func ExtractModule(moduleFS fs.FS, folderName string) (string, error) {
	tempDir := filepath.Join(os.TempDir(), tempBase, folderName)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("impossibile creare cartella temp %s: %w", tempDir, err)
	}

	err := fs.WalkDir(moduleFS, folderName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(folderName, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(tempDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		data, err := fs.ReadFile(moduleFS, path)
		if err != nil {
			return fmt.Errorf("impossibile leggere %s: %w", path, err)
		}
		return os.WriteFile(destPath, data, 0644)
	})

	if err != nil {
		return "", fmt.Errorf("impossibile estrarre modulo %s: %w", folderName, err)
	}

	return tempDir, nil
}

func CleanupModule(folderName string) error {
	tempDir := filepath.Join(os.TempDir(), tempBase, folderName)
	return os.RemoveAll(tempDir)
}

func CleanupAll() error {
	tempDir := filepath.Join(os.TempDir(), tempBase)
	return os.RemoveAll(tempDir)
}
