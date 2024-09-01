package system

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"

	"github.com/coding-hui/common/util/homedir"

	"github.com/coding-hui/ai-terminal/internal/runner"
)

const (
	DefaultApplicationName = "ai"
	DefaultEditor          = "vim"
)

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
		configFile:      GetConfigFile(),
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
	dist, err := runner.Run("lsb_release", "-sd")
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
		shell, err = runner.Run("echo", os.Getenv("COMSPEC"))
	} else {
		shell, err = runner.Run("echo", os.Getenv("SHELL"))
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
	return homedir.HomeDir()
}

func GetUsername() string {
	name, err := runner.Run("echo", os.Getenv("USER"))
	if err != nil {
		return ""
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name, err = runner.Run("whoami")
		if err != nil {
			return ""
		}
	}

	nameParts := strings.Split(filepath.ToSlash(name), "/")

	return strings.TrimSpace(nameParts[len(nameParts)-1])
}

func GetEditor() string {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
		if editor == "" {
			editor = DefaultEditor
		}
	}
	return editor
}

func GetConfigFile() string {
	defConfFile := viper.GetString("config")
	if defConfFile == "" {
		defConfFile = filepath.Join(GetHomeDirectory(), ".config", "wecoding", strings.ToLower(DefaultApplicationName))
	}
	return defConfFile
}
