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

type onlineConfig struct {
	GitHub    string `json:"github"`
	Installer string `json:"installer"`
}

type moduleEntry struct {
	Name   string `json:"name"`
	Active *bool  `json:"active,omitempty"`
}

type setupConfig struct {
	Modules []moduleEntry `json:"modules"`
}

type Module struct {
	Name string
}

func PrepareRoot() (string, error) {
	guid := uuid.New().String()
	root := filepath.Join(os.TempDir(), "WebGainInstaller", guid)
	if err := os.MkdirAll(root, 0755); err != nil {
		return "", fmt.Errorf("impossibile creare cartella temporanea: %w", err)
	}
	return root, nil
}

// VerifyModules scarica o usa l'embedded setup.json.
// Restituisce true se il file proviene da download online.
func VerifyModules(configFS fs.FS, webgainRoot string) (bool, error) {
	destPath := filepath.Join(webgainRoot, "setup.json")

	downloadURL := buildInstallerURL(configFS)
	if downloadURL != "" {
		logger.Info("URL installer composto: %s", downloadURL)
		logger.Info("Tentativo download setup.json online...")
		if data, err := downloadWithRetry(downloadURL, 3, 30*time.Second); err == nil {
			if isValidJSON(data) {
				logger.Info("Download riuscito, JSON valido (%d bytes), salvataggio in %s", len(data), destPath)
				return true, os.WriteFile(destPath, data, 0644)
			}
			logger.Warn("Download riuscito ma JSON non valido (%d bytes), passaggio a fallback embedded", len(data))
		} else {
			logger.Warn("Download fallito: %v, passaggio a fallback embedded", err)
		}
	} else {
		logger.Warn("Impossibile comporre URL installer, passaggio diretto a fallback embedded")
	}

	logger.Info("Lettura setup.json embedded...")
	embeddedData, err := fs.ReadFile(configFS, "setup.json")
	if err != nil {
		logger.Error("Lettura setup.json embedded fallita: %v", err)
		return false, fmt.Errorf("setup.json non valido")
	}

	if isValidJSON(embeddedData) {
		logger.Info("Setup.json embedded valido (%d bytes), salvataggio in %s", len(embeddedData), destPath)
		return false, os.WriteFile(destPath, embeddedData, 0644)
	}

	logger.Error("Setup.json embedded non e' un JSON valido (%d bytes)", len(embeddedData))
	return false, fmt.Errorf("setup.json non valido")
}

// InitModules valida il setup.json in WEBGAINROOT.
// Se webgainOnline è true e la validazione fallisce, ritenta con l'embedded.
func InitModules(configFS fs.FS, webgainRoot string, webgainOnline bool) ([]Module, error) {
	destPath := filepath.Join(webgainRoot, "setup.json")
	logger.Info("Inizializzazione moduli da %s (online=%v)", destPath, webgainOnline)

	modules, err := parseAndValidateSetup(destPath)
	if err == nil {
		logger.Info("Validazione setup.json riuscita: %d moduli attivi", len(modules))
		return modules, nil
	}
	logger.Warn("Validazione setup.json fallita: %v", err)

	if !webgainOnline {
		logger.Error("File proveniente da embedded, nessun fallback disponibile")
		return nil, err
	}

	logger.Info("File proveniente da online, tentativo fallback con embedded...")
	embeddedData, readErr := fs.ReadFile(configFS, "setup.json")
	if readErr != nil {
		logger.Error("Lettura setup.json embedded fallita: %v", readErr)
		return nil, fmt.Errorf("setup.json embedded non disponibile")
	}

	if !isValidJSON(embeddedData) {
		logger.Error("Setup.json embedded non e' un JSON valido")
		return nil, fmt.Errorf("setup.json embedded non valido")
	}

	logger.Info("Setup.json embedded JSON valido, sostituzione in WEBGAINROOT...")
	if writeErr := os.WriteFile(destPath, embeddedData, 0644); writeErr != nil {
		logger.Error("Scrittura fallback embedded fallita: %v", writeErr)
		return nil, fmt.Errorf("impossibile scrivere fallback: %w", writeErr)
	}

	modules, err = parseAndValidateSetup(destPath)
	if err != nil {
		logger.Error("Validazione fallback embedded fallita: %v", err)
		return nil, err
	}

	logger.Info("Fallback embedded valido: %d moduli attivi", len(modules))
	return modules, nil
}

func parseAndValidateSetup(path string) ([]Module, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("impossibile leggere %s: %w", path, err)
	}

	logger.Info("Parsing setup.json (%d bytes)...", len(data))

	var cfg setupConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("JSON non valido: %w", err)
	}

	if cfg.Modules == nil {
		logger.Error("Proprieta 'modules' mancante")
		return nil, fmt.Errorf("proprieta 'modules' mancante")
	}

	if len(cfg.Modules) == 0 {
		logger.Error("Array 'modules' vuoto")
		return nil, fmt.Errorf("array 'modules' vuoto")
	}

	logger.Info("Trovati %d elementi nell'array modules", len(cfg.Modules))

	var active []Module
	seen := make(map[string]bool)

	for i, entry := range cfg.Modules {
		name := strings.TrimSpace(entry.Name)
		if name == "" {
			logger.Error("Modulo [%d]: proprieta 'name' mancante o blank", i)
			return nil, fmt.Errorf("modulo [%d]: 'name' mancante o blank", i)
		}

		if entry.Active != nil && !*entry.Active {
			logger.Info("Modulo [%d] '%s': disattivato, scartato", i, name)
			continue
		}

		if seen[name] {
			logger.Error("Modulo [%d] '%s': nome duplicato", i, name)
			return nil, fmt.Errorf("modulo '%s' duplicato", name)
		}
		seen[name] = true

		active = append(active, Module{Name: name})
		logger.Info("Modulo [%d] '%s': attivo", i, name)
	}

	if len(active) == 0 {
		logger.Error("Nessun modulo attivo dopo il filtraggio")
		return nil, fmt.Errorf("nessun modulo attivo")
	}

	logger.Info("Moduli validi: %d su %d totali", len(active), len(cfg.Modules))
	return active, nil
}

func buildInstallerURL(configFS fs.FS) string {
	data, err := fs.ReadFile(configFS, "online.json")
	if err != nil {
		logger.Warn("Impossibile leggere online.json: %v", err)
		return ""
	}

	var cfg onlineConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		logger.Warn("online.json non e' un JSON valido: %v", err)
		return ""
	}

	if cfg.GitHub == "" || cfg.Installer == "" {
		logger.Warn("online.json incompleto: github=%q, installer=%q", cfg.GitHub, cfg.Installer)
		return ""
	}

	logger.Info("online.json letto: github=%s, installer=%s", cfg.GitHub, cfg.Installer)

	baseURL := toRawBaseURL(cfg.GitHub)
	finalURL := baseURL + "repo/config/" + cfg.Installer

	logger.Info("Base URL raw: %s", baseURL)
	logger.Info("URL finale composto: %s", finalURL)

	return finalURL
}

func toRawBaseURL(githubURL string) string {
	url := strings.TrimRight(githubURL, "/")
	url = strings.Replace(url, "github.com/", "raw.githubusercontent.com/", 1)
	url = strings.Replace(url, "/blob/", "/", 1)
	if !strings.HasSuffix(url, "/main") && !strings.HasSuffix(url, "/master") {
		url += "/main"
	}
	return url + "/"
}

// ToRawURL converte un URL GitHub generico nella versione raw.
func ToRawURL(url string) string {
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
