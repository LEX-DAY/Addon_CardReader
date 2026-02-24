package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const nativeHostName = "com.cardreader.bridge"

type installConfig struct {
	DoInstall   bool
	ExtensionID string
	HostPath    string
	Browser     string
}

func parseFlags() installConfig {
	cfg := installConfig{}
	flag.BoolVar(&cfg.DoInstall, "install", false, "install native messaging manifest for current user")
	flag.StringVar(&cfg.ExtensionID, "extension-id", "", "Chrome extension ID for allowed_origins")
	flag.StringVar(&cfg.HostPath, "host-path", "", "absolute path to host executable (default: current executable)")
	flag.StringVar(&cfg.Browser, "browser", "chrome", "target browser: chrome|chromium|edge")
	flag.Parse()
	return cfg
}

func runInstall(cfg installConfig) error {
	if cfg.ExtensionID == "" {
		return errors.New("--extension-id is required")
	}

	hostPath := cfg.HostPath
	if hostPath == "" {
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("resolve executable path: %w", err)
		}
		hostPath = exe
	}
	absHost, err := filepath.Abs(hostPath)
	if err != nil {
		return fmt.Errorf("resolve host absolute path: %w", err)
	}

	browser := strings.ToLower(strings.TrimSpace(cfg.Browser))

	manifestPath, err := nativeManifestPath(browser)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		return fmt.Errorf("create manifest dir: %w", err)
	}

	payload := map[string]any{
		"name":        nativeHostName,
		"description": "Card reader bridge",
		"path":        absHost,
		"type":        "stdio",
		"allowed_origins": []string{
			fmt.Sprintf("chrome-extension://%s/", cfg.ExtensionID),
		},
	}
	buf, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	if err := os.WriteFile(manifestPath, append(buf, '\n'), 0o644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	if runtime.GOOS == "windows" {
		if err := registerWindowsNativeHost(browser, manifestPath); err != nil {
			return err
		}
	}

	fmt.Printf("Native host manifest installed: %s\n", manifestPath)
	fmt.Printf("Host executable path: %s\n", absHost)
	return nil
}

func nativeManifestPath(browser string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}

	switch runtime.GOOS {
	case "linux":
		switch browser {
		case "chrome":
			return filepath.Join(home, ".config", "google-chrome", "NativeMessagingHosts", nativeHostName+".json"), nil
		case "chromium":
			return filepath.Join(home, ".config", "chromium", "NativeMessagingHosts", nativeHostName+".json"), nil
		case "edge":
			return filepath.Join(home, ".config", "microsoft-edge", "NativeMessagingHosts", nativeHostName+".json"), nil
		default:
			return "", fmt.Errorf("unsupported browser: %s", browser)
		}
	case "windows":
		appData := os.Getenv("LOCALAPPDATA")
		if appData == "" {
			return "", errors.New("LOCALAPPDATA is not set")
		}
		switch browser {
		case "chrome":
			return filepath.Join(appData, "Google", "Chrome", "User Data", "NativeMessagingHosts", nativeHostName+".json"), nil
		case "chromium":
			return filepath.Join(appData, "Chromium", "User Data", "NativeMessagingHosts", nativeHostName+".json"), nil
		case "edge":
			return filepath.Join(appData, "Microsoft", "Edge", "User Data", "NativeMessagingHosts", nativeHostName+".json"), nil
		default:
			return "", fmt.Errorf("unsupported browser: %s", browser)
		}
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func registerWindowsNativeHost(browser, manifestPath string) error {
	var key string
	switch browser {
	case "chrome":
		key = `HKCU\Software\Google\Chrome\NativeMessagingHosts\` + nativeHostName
	case "chromium":
		key = `HKCU\Software\Chromium\NativeMessagingHosts\` + nativeHostName
	case "edge":
		key = `HKCU\Software\Microsoft\Edge\NativeMessagingHosts\` + nativeHostName
	default:
		return fmt.Errorf("unsupported browser: %s", browser)
	}

	cmd := exec.Command("reg", "add", key, "/ve", "/t", "REG_SZ", "/d", manifestPath, "/f")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("register host in windows registry: %w (%s)", err, string(out))
	}
	return nil
}
