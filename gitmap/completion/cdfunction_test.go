package completion

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
)

func TestAppendCDFunctionWritesManagedWrappers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile")

	err := appendCDFunction(constants.CDFuncBash, path)
	if err != nil {
		t.Fatalf("appendCDFunction failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read profile failed: %v", err)
	}

	text := string(data)
	if !strings.Contains(text, constants.CDFuncMarker) {
		t.Fatal("expected managed wrapper marker to be written")
	}
	if !strings.Contains(text, "gitmap() {") {
		t.Fatal("expected gitmap shell wrapper to be written")
	}
	if !strings.Contains(text, "gcd() {") {
		t.Fatal("expected gcd shell wrapper to be written")
	}
}

func TestAppendCDFunctionSkipsManagedWrapperWhenPresent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile")
	block := "\n" + constants.CDFuncMarker + "\n" + constants.CDFuncBash + "\n"

	err := os.WriteFile(path, []byte(block), 0o644)
	if err != nil {
		t.Fatalf("seed profile failed: %v", err)
	}

	err = appendCDFunction(constants.CDFuncBash, path)
	if err != nil {
		t.Fatalf("appendCDFunction failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read profile failed: %v", err)
	}

	if string(data) != block {
		t.Fatal("expected managed wrapper block to remain unchanged")
	}
}

func TestAppendCDFunctionAppendsManagedWrapperAfterLegacyMarker(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile")
	legacy := "\n# gitmap cd wrapper\ngcd() {\n  cd \"$(gitmap cd \"$@\")\"\n}\n"

	err := os.WriteFile(path, []byte(legacy), 0o644)
	if err != nil {
		t.Fatalf("seed profile failed: %v", err)
	}

	err = appendCDFunction(constants.CDFuncBash, path)
	if err != nil {
		t.Fatalf("appendCDFunction failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read profile failed: %v", err)
	}

	text := string(data)
	if !strings.Contains(text, legacy) {
		t.Fatal("expected legacy wrapper to remain for migration safety")
	}
	if !strings.Contains(text, constants.CDFuncMarker) {
		t.Fatal("expected managed wrapper marker to be appended")
	}
	if strings.Count(text, constants.CDFuncMarker) != 1 {
		t.Fatal("expected exactly one managed wrapper marker")
	}
}

func TestAppendCDFunctionCreatesProfileDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Documents", "WindowsPowerShell", "profile.ps1")

	err := appendCDFunction(constants.CDFuncPowerShell, path)
	if err != nil {
		t.Fatalf("appendCDFunction failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read profile failed: %v", err)
	}

	if !strings.Contains(string(data), constants.CDFuncMarker) {
		t.Fatal("expected managed wrapper marker in created profile")
	}
}

func TestAppendCDFunctionsWritesToMultipleProfiles(t *testing.T) {
	base := t.TempDir()
	paths := []string{
		filepath.Join(base, "Documents", "PowerShell", "profile.ps1"),
		filepath.Join(base, "Documents", "WindowsPowerShell", "profile.ps1"),
	}

	err := appendCDFunctions(constants.CDFuncPowerShell, paths)
	if err != nil {
		t.Fatalf("appendCDFunctions failed: %v", err)
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read profile failed for %s: %v", path, err)
		}
		if !strings.Contains(string(data), constants.CDFuncMarker) {
			t.Fatalf("expected managed wrapper marker in %s", path)
		}
	}
}
