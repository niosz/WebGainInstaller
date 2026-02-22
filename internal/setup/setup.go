package setup

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

func PrepareRoot() (string, error) {
	guid := uuid.New().String()
	root := filepath.Join(os.TempDir(), "WebGainInstaller", guid)
	if err := os.MkdirAll(root, 0755); err != nil {
		return "", fmt.Errorf("impossibile creare cartella temporanea: %w", err)
	}
	return root, nil
}

func VerifyModules(configFS fs.FS, webgainRoot string) error {
	destPath := filepath.Join(webgainRoot, "setup.json")

	onlineURL := readOnlineURL(configFS)
	if onlineURL != "" {
		if data, err := downloadWithRetry(onlineURL, 3, 30*time.Second); err == nil {
			if isValidJSON(data) {
				return os.WriteFile(destPath, data, 0644)
			}
		}
	}

	embeddedData, err := fs.ReadFile(configFS, "setup.json")
	if err == nil && isValidJSON(embeddedData) {
		return os.WriteFile(destPath, embeddedData, 0644)
	}

	return fmt.Errorf("setup.json non valido")
}

func readOnlineURL(configFS fs.FS) string {
	data, err := fs.ReadFile(configFS, "online.txt")
	if err != nil {
		return ""
	}
	return toRawURL(strings.TrimSpace(string(data)))
}

func toRawURL(url string) string {
	if !strings.Contains(url, "github.com/") {
		return url
	}
	url = strings.Replace(url, "github.com/", "raw.githubusercontent.com/", 1)
	url = strings.Replace(url, "/blob/", "/", 1)

	const repoPrefix = "raw.githubusercontent.com/niosz/WebGainInstaller/"
	if idx := strings.Index(url, repoPrefix); idx >= 0 {
		afterRepo := url[idx+len(repoPrefix):]
		if !strings.HasPrefix(afterRepo, "main/") && !strings.HasPrefix(afterRepo, "master/") {
			url = url[:idx+len(repoPrefix)] + "main/" + afterRepo
		}
	}
	return url
}

func downloadWithRetry(url string, maxRetries int, timeout time.Duration) ([]byte, error) {
	client := &http.Client{Timeout: timeout}
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		resp, err := client.Get(url)
		if err != nil {
			lastErr = err
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode == http.StatusOK {
			return body, nil
		}
		lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return nil, fmt.Errorf("download fallito dopo %d tentativi: %w", maxRetries, lastErr)
}

func isValidJSON(data []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(data, &js) == nil
}
