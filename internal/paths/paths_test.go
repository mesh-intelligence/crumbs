// Tests for platform-specific directory paths.
// Implements: prd010-configuration-directories acceptance criteria.
package paths

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDefaultConfigDir(t *testing.T) {
	dir, err := DefaultConfigDir()
	if err != nil {
		t.Fatalf("DefaultConfigDir failed: %v", err)
	}

	if dir == "" {
		t.Error("expected non-empty config dir")
	}

	// Verify path contains "crumbs"
	if !strings.Contains(dir, "crumbs") {
		t.Errorf("expected path to contain 'crumbs', got %q", dir)
	}

	// Platform-specific checks
	switch runtime.GOOS {
	case "darwin":
		if !strings.Contains(dir, "Application Support") {
			t.Errorf("macOS config dir should contain 'Application Support', got %q", dir)
		}
	case "windows":
		// Should be under AppData\Roaming
		if !strings.Contains(strings.ToLower(dir), "appdata") {
			t.Errorf("Windows config dir should contain 'AppData', got %q", dir)
		}
	default:
		// Linux: should be under .config or XDG_CONFIG_HOME
		if !strings.Contains(dir, ".config") && os.Getenv("XDG_CONFIG_HOME") == "" {
			t.Errorf("Linux config dir should contain '.config' by default, got %q", dir)
		}
	}
}

func TestDefaultDataDir(t *testing.T) {
	dir, err := DefaultDataDir()
	if err != nil {
		t.Fatalf("DefaultDataDir failed: %v", err)
	}

	if dir == "" {
		t.Error("expected non-empty data dir")
	}

	// Verify path contains "crumbs"
	if !strings.Contains(dir, "crumbs") {
		t.Errorf("expected path to contain 'crumbs', got %q", dir)
	}

	// Platform-specific checks
	switch runtime.GOOS {
	case "darwin":
		if !strings.Contains(dir, "Application Support") {
			t.Errorf("macOS data dir should contain 'Application Support', got %q", dir)
		}
		if !strings.HasSuffix(dir, "data") {
			t.Errorf("macOS data dir should end with 'data', got %q", dir)
		}
	case "windows":
		// Should be under AppData\Local
		if !strings.Contains(strings.ToLower(dir), "local") {
			t.Errorf("Windows data dir should contain 'Local', got %q", dir)
		}
	default:
		// Linux: should be under .local/share or XDG_DATA_HOME
		if !strings.Contains(dir, ".local/share") && os.Getenv("XDG_DATA_HOME") == "" {
			t.Errorf("Linux data dir should contain '.local/share' by default, got %q", dir)
		}
	}
}

func TestResolveConfigDir_FlagPrecedence(t *testing.T) {
	flagValue := "/custom/config/path"
	dir, err := ResolveConfigDir(flagValue, "")
	if err != nil {
		t.Fatalf("ResolveConfigDir failed: %v", err)
	}

	if dir != flagValue {
		t.Errorf("expected flag value %q, got %q", flagValue, dir)
	}
}

func TestResolveConfigDir_EnvPrecedence(t *testing.T) {
	// Set up environment variable
	os.Setenv("TEST_CRUMBS_CONFIG_DIR", "/env/config/path")
	defer os.Unsetenv("TEST_CRUMBS_CONFIG_DIR")

	// Empty flag, should use env
	dir, err := ResolveConfigDir("", "TEST_CRUMBS_CONFIG_DIR")
	if err != nil {
		t.Fatalf("ResolveConfigDir failed: %v", err)
	}

	if dir != "/env/config/path" {
		t.Errorf("expected env value '/env/config/path', got %q", dir)
	}
}

func TestResolveConfigDir_Default(t *testing.T) {
	// Empty flag and no matching env var
	dir, err := ResolveConfigDir("", "NONEXISTENT_VAR")
	if err != nil {
		t.Fatalf("ResolveConfigDir failed: %v", err)
	}

	defaultDir, _ := DefaultConfigDir()
	if dir != defaultDir {
		t.Errorf("expected default dir %q, got %q", defaultDir, dir)
	}
}

func TestResolveDataDir_FlagPrecedence(t *testing.T) {
	flagValue := "/custom/data/path"
	dir, err := ResolveDataDir(flagValue, "")
	if err != nil {
		t.Fatalf("ResolveDataDir failed: %v", err)
	}

	if dir != flagValue {
		t.Errorf("expected flag value %q, got %q", flagValue, dir)
	}
}

func TestResolveDataDir_ConfigPrecedence(t *testing.T) {
	configValue := "/config/data/path"
	dir, err := ResolveDataDir("", configValue)
	if err != nil {
		t.Fatalf("ResolveDataDir failed: %v", err)
	}

	if dir != configValue {
		t.Errorf("expected config value %q, got %q", configValue, dir)
	}
}

func TestResolveDataDir_Default(t *testing.T) {
	dir, err := ResolveDataDir("", "")
	if err != nil {
		t.Fatalf("ResolveDataDir failed: %v", err)
	}

	defaultDir, _ := DefaultDataDir()
	if dir != defaultDir {
		t.Errorf("expected default dir %q, got %q", defaultDir, dir)
	}
}

func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "test", "nested", "dir")

	err := EnsureDir(testDir)
	if err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}

	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("failed to stat created dir: %v", err)
	}

	if !info.IsDir() {
		t.Error("expected directory to be created")
	}
}

func TestEnsureDir_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()

	// First call
	err := EnsureDir(tmpDir)
	if err != nil {
		t.Fatalf("first EnsureDir failed: %v", err)
	}

	// Second call (should not error)
	err = EnsureDir(tmpDir)
	if err != nil {
		t.Fatalf("second EnsureDir failed: %v", err)
	}
}

func TestXDGConfigHomeOverride(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		t.Skip("XDG_CONFIG_HOME only applies to Linux")
	}

	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	dir, err := DefaultConfigDir()
	if err != nil {
		t.Fatalf("DefaultConfigDir failed: %v", err)
	}

	expected := filepath.Join(tmpDir, "crumbs")
	if dir != expected {
		t.Errorf("expected %q with XDG_CONFIG_HOME set, got %q", expected, dir)
	}
}

func TestXDGDataHomeOverride(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		t.Skip("XDG_DATA_HOME only applies to Linux")
	}

	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", oldXDG)

	dir, err := DefaultDataDir()
	if err != nil {
		t.Fatalf("DefaultDataDir failed: %v", err)
	}

	expected := filepath.Join(tmpDir, "crumbs")
	if dir != expected {
		t.Errorf("expected %q with XDG_DATA_HOME set, got %q", expected, dir)
	}
}
