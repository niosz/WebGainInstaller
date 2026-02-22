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

	"WebGainInstaller/internal/logger"

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
		logger.Info("URL online.txt originale letto, URL raw risultante: %s", onlineURL)
		logger.Info("Tentativo download setup.json online...")
		if data, err := downloadWithRetry(onlineURL, 3, 30*time.Second); err == nil {
			if isValidJSON(data) {
				logger.Info("Download riuscito, JSON valido (%d bytes), salvataggio in %s", len(data), destPath)
				return os.WriteFile(destPath, data, 0644)
			}
			logger.Warn("Download riuscito ma JSON non valido (%d bytes), passaggio a fallback embedded", len(data))
		} else {
			logger.Warn("Download fallito: %v, passaggio a fallback embedded", err)
		}
	} else {
		logger.Warn("URL online.txt vuoto o non leggibile, passaggio diretto a fallback embedded")
	}

	logger.Info("Lettura setup.json embedded...")
	embeddedData, err := fs.ReadFile(configFS, "setup.json")
	if err != nil {
		logger.Error("Lettura setup.json embedded fallita: %v", err)
		return fmt.Errorf("setup.json non valido")
	}

	if isValidJSON(embeddedData) {
		logger.Info("Setup.json embedded valido (%d bytes), salvataggio in %s", len(embeddedData), destPath)
		return os.WriteFile(destPath, embeddedData, 0644)
	}

	logger.Error("Setup.json embedded non e' un JSON valido (%d bytes)", len(embeddedData))
	return fmt.Errorf("setup.json non valido")
}

func readOnlineURL(configFS fs.FS) string {
	data, err := fs.ReadFile(configFS, "online.txt")
	if err != nil {
		logger.Warn("Impossibile leggere online.txt: %v", err)
		return ""
	}
	raw := strings.TrimSpace(string(data))
	converted := toRawURL(raw)
	if raw != converted {
		logger.Info("URL convertito: %s -> %s", raw, converted)
	}
	return converted
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
		logger.Info("Download tentativo %d/%d: %s", i+1, maxRetries, url)
		resp, err := client.Get(url)
		if err != nil {
			lastErr = err
			logger.Warn("Tentativo %d/%d fallito (connessione): %v", i+1, maxRetries, err)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			logger.Warn("Tentativo %d/%d fallito (lettura body): %v", i+1, maxRetries, err)
			continue
		}
		if resp.StatusCode == http.StatusOK {
			logger.Info("Tentativo %d/%d riuscito: HTTP %d, %d bytes ricevuti", i+1, maxRetries, resp.StatusCode, len(body))
			return body, nil
		}
		lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
		logger.Warn("Tentativo %d/%d fallito: HTTP %d", i+1, maxRetries, resp.StatusCode)
	}

	return nil, fmt.Errorf("download fallito dopo %d tentativi: %w", maxRetries, lastErr)
}

func isValidJSON(data []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(data, &js) == nil
}
