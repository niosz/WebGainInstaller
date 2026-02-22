package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"WebGainInstaller/internal/module"

	"golang.org/x/sys/windows/registry"
)

func executeStep(step module.Step, workDir string) error {
	switch step.Type {
	case "exe":
		return runExe(step, workDir)
	case "msi":
		return runMsi(step, workDir)
	case "powershell":
		return runPowerShellCommand(step)
	case "powershell_script":
		return runPowerShellScript(step, workDir)
	case "powershell_module":
		return runPowerShellModule(step)
	case "batch":
		return runBatch(step, workDir)
	case "env_path":
		return setEnvPath(step)
	case "env_set":
		return setEnvVariable(step)
	case "shell_config":
		return configureShell(step)
	case "registry":
		return setRegistry(step)
	case "copy":
		return copyFiles(step, workDir)
	case "service":
		return manageService(step)
	case "verify":
		return verifyInstall(step)
	default:
		return fmt.Errorf("tipo di step sconosciuto: %s", step.Type)
	}
}

func runExe(step module.Step, workDir string) error {
	exePath := filepath.Join(workDir, step.File)
	args := parseArgs(step.Args)
	cmd := exec.Command(exePath, args...)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("esecuzione %s fallita: %w\nOutput: %s", step.File, err, string(output))
	}
	return nil
}

func runMsi(step module.Step, workDir string) error {
	msiPath := filepath.Join(workDir, step.File)
	baseArgs := []string{"/i", msiPath}
	baseArgs = append(baseArgs, parseArgs(step.Args)...)
	cmd := exec.Command("msiexec.exe", baseArgs...)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installazione MSI %s fallita: %w\nOutput: %s", step.File, err, string(output))
	}
	return nil
}

func runPowerShellCommand(step module.Step) error {
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", step.Command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("comando PowerShell fallito: %w\nOutput: %s", err, string(output))
	}
	return nil
}

func runPowerShellScript(step module.Step, workDir string) error {
	scriptPath := filepath.Join(workDir, step.File)
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("script PowerShell %s fallito: %w\nOutput: %s", step.File, err, string(output))
	}
	return nil
}

func runPowerShellModule(step module.Step) error {
	installCmd := fmt.Sprintf("Install-Module -Name %s -Force -AllowClobber -Scope AllUsers", step.Value)
	if step.Command != "" {
		installCmd = step.Command
	}
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", installCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installazione modulo PowerShell fallita: %w\nOutput: %s", err, string(output))
	}
	return nil
}

func runBatch(step module.Step, workDir string) error {
	batPath := filepath.Join(workDir, step.File)
	cmd := exec.Command("cmd.exe", "/C", batPath)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("script batch %s fallito: %w\nOutput: %s", step.File, err, string(output))
	}
	return nil
}

func setEnvPath(step module.Step) error {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("impossibile aprire chiave registro Environment: %w", err)
	}
	defer key.Close()

	currentPath, _, err := key.GetStringValue("Path")
	if err != nil {
		return fmt.Errorf("impossibile leggere PATH: %w", err)
	}

	expandedValue := os.ExpandEnv(step.Value)

	if strings.Contains(strings.ToLower(currentPath), strings.ToLower(expandedValue)) {
		return nil
	}

	var newPath string
	switch step.Action {
	case "append":
		newPath = currentPath + ";" + expandedValue
	case "prepend":
		newPath = expandedValue + ";" + currentPath
	default:
		newPath = currentPath + ";" + expandedValue
	}

	if err := key.SetExpandStringValue("Path", newPath); err != nil {
		return fmt.Errorf("impossibile aggiornare PATH: %w", err)
	}

	broadcastEnvironmentChange()
	return nil
}

func setEnvVariable(step module.Step) error {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("impossibile aprire chiave registro Environment: %w", err)
	}
	defer key.Close()

	expandedValue := os.ExpandEnv(step.Value)
	if err := key.SetExpandStringValue(step.Variable, expandedValue); err != nil {
		return fmt.Errorf("impossibile impostare variabile %s: %w", step.Variable, err)
	}

	broadcastEnvironmentChange()
	return nil
}

func configureShell(step module.Step) error {
	var profilePath string
	switch step.Target {
	case "powershell_profile":
		profileDir := filepath.Join(os.Getenv("ProgramFiles"), "PowerShell", "7")
		if _, err := os.Stat(profileDir); os.IsNotExist(err) {
			profileDir = filepath.Join(os.Getenv("WINDIR"), "System32", "WindowsPowerShell", "v1.0")
		}
		profilePath = filepath.Join(profileDir, "profile.ps1")
	default:
		return fmt.Errorf("target shell sconosciuto: %s", step.Target)
	}

	if err := os.MkdirAll(filepath.Dir(profilePath), 0755); err != nil {
		return fmt.Errorf("impossibile creare directory profilo: %w", err)
	}

	existing, _ := os.ReadFile(profilePath)
	if strings.Contains(string(existing), step.Content) {
		return nil
	}

	content := step.Content
	if len(existing) > 0 {
		content = "\n" + content
	}

	f, err := os.OpenFile(profilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("impossibile aprire profilo shell: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("impossibile scrivere profilo shell: %w", err)
	}
	return nil
}

func setRegistry(step module.Step) error {
	parts := strings.SplitN(step.Key, `\`, 2)
	if len(parts) != 2 {
		return fmt.Errorf("chiave di registro non valida: %s", step.Key)
	}

	var rootKey registry.Key
	switch strings.ToUpper(parts[0]) {
	case "HKLM", "HKEY_LOCAL_MACHINE":
		rootKey = registry.LOCAL_MACHINE
	case "HKCU", "HKEY_CURRENT_USER":
		rootKey = registry.CURRENT_USER
	case "HKCR", "HKEY_CLASSES_ROOT":
		rootKey = registry.CLASSES_ROOT
	default:
		return fmt.Errorf("root key sconosciuta: %s", parts[0])
	}

	key, _, err := registry.CreateKey(rootKey, parts[1], registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("impossibile creare/aprire chiave %s: %w", step.Key, err)
	}
	defer key.Close()

	if err := key.SetStringValue(step.Variable, step.Value); err != nil {
		return fmt.Errorf("impossibile impostare valore %s: %w", step.Variable, err)
	}
	return nil
}

func copyFiles(step module.Step, workDir string) error {
	src := filepath.Join(workDir, step.File)
	dest := os.ExpandEnv(step.Dest)

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("impossibile creare directory destinazione: %w", err)
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("impossibile leggere file sorgente %s: %w", src, err)
	}
	if err := os.WriteFile(dest, data, 0644); err != nil {
		return fmt.Errorf("impossibile copiare in %s: %w", dest, err)
	}
	return nil
}

func manageService(step module.Step) error {
	var args []string
	switch step.Action {
	case "start":
		args = []string{"start", step.Value}
	case "stop":
		args = []string{"stop", step.Value}
	case "restart":
		exec.Command("sc.exe", "stop", step.Value).Run()
		args = []string{"start", step.Value}
	default:
		return fmt.Errorf("azione servizio sconosciuta: %s", step.Action)
	}

	cmd := exec.Command("sc.exe", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gestione servizio %s fallita: %w\nOutput: %s", step.Value, err, string(output))
	}
	return nil
}

func verifyInstall(step module.Step) error {
	cmd := exec.Command("cmd.exe", "/C", step.Command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("verifica fallita (%s): %w\nOutput: %s", step.Command, err, string(output))
	}
	return nil
}

func parseArgs(args string) []string {
	if args == "" {
		return nil
	}
	return strings.Fields(args)
}

func broadcastEnvironmentChange() {
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command",
		`[System.Environment]::SetEnvironmentVariable("_WGI_REFRESH","1","Process"); `+
			`Add-Type -Namespace Win32 -Name NativeMethods -MemberDefinition '[DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)] public static extern IntPtr SendMessageTimeout(IntPtr hWnd, uint Msg, UIntPtr wParam, string lParam, uint fuFlags, uint uTimeout, out UIntPtr lpdwResult);'; `+
			`$result = [UIntPtr]::Zero; `+
			`[Win32.NativeMethods]::SendMessageTimeout([IntPtr]0xFFFF, 0x1A, [UIntPtr]::Zero, "Environment", 2, 5000, [ref]$result)`)
	cmd.Run()
}
