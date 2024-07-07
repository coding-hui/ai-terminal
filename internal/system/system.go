package system

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mitchellh/go-homedir"

	"github.com/coding-hui/ai-terminal/internal/run"
)

const DefaultApplicationName = "ai"

type Analysis struct {
	operatingSystem OperatingSystem
	distribution    string
	shell           string
	homeDirectory   string
	username        string
	editor          string
	configFile      string
}

func (a *Analysis) GetApplicationName() string {
	return DefaultApplicationName
}

func (a *Analysis) GetOperatingSystem() OperatingSystem {
	return a.operatingSystem
}

func (a *Analysis) GetDistribution() string {
	return a.distribution
}

func (a *Analysis) GetShell() string {
	return a.shell
}

func (a *Analysis) GetHomeDirectory() string {
	return a.homeDirectory
}

func (a *Analysis) GetUsername() string {
	return a.username
}

func (a *Analysis) GetEditor() string {
	return a.editor
}

func (a *Analysis) GetConfigFile() string {
	return a.configFile
}

func Analyse() *Analysis {
	return &Analysis{
		operatingSystem: GetOperatingSystem(),
		distribution:    GetDistribution(),
		shell:           GetShell(),
		homeDirectory:   GetHomeDirectory(),
		username:        GetUsername(),
		editor:          GetEditor(),
		configFile:      GetDefaultConfigFile(),
	}
}

func GetOperatingSystem() OperatingSystem {
	switch runtime.GOOS {
	case "linux":
		return LinuxOperatingSystem
	case "darwin":
		return MacOperatingSystem
	case "windows":
		return WindowsOperatingSystem
	default:
		return UnknownOperatingSystem
	}
}

func GetDistribution() string {
	dist, err := run.Run("lsb_release", "-sd")
	if err != nil {
		return ""
	}

	return strings.Trim(strings.TrimSpace(dist), "\"")
}

func GetShell() string {
	var (
		shell           string
		err             error
		operatingSystem = GetOperatingSystem()
	)

	if operatingSystem == WindowsOperatingSystem {
		shell, err = run.Run("echo", os.Getenv("COMSPEC"))
	} else {
		shell, err = run.Run("echo", os.Getenv("SHELL"))
	}
	if err != nil {
		return ""
	}

	shell = strings.TrimSpace(shell)     // Trims all leading and trailing white spaces
	shellPath := filepath.ToSlash(shell) // Normalize path separators to forward slash
	shellParts := strings.Split(shellPath, "/")

	return shellParts[len(shellParts)-1]
}

func GetHomeDirectory() string {
	homeDir, err := homedir.Dir()
	if err != nil {
		return ""
	}

	return homeDir
}

func GetUsername() string {
	name, err := run.Run("echo", os.Getenv("USER"))
	if err != nil {
		return ""
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name, err = run.Run("whoami")
		if err != nil {
			return ""
		}
	}

	nameParts := strings.Split(filepath.ToSlash(name), "/")

	return strings.TrimSpace(nameParts[len(nameParts)-1])
}

func GetEditor() string {
	name, err := run.Run("echo", os.Getenv("EDITOR"))
	if err != nil {
		return "nano"
	}

	return strings.TrimSpace(name)
}

func GetDefaultConfigFile() string {
	return filepath.Join(GetHomeDirectory(), ".config", strings.ToLower(DefaultApplicationName))
}
