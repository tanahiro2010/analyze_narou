package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnvSetsMissingVariables(t *testing.T) {
	t.Setenv("EXISTING_KEY", "already-set")

	path := filepath.Join(t.TempDir(), ".env")
	content := []byte(`
# comment
NEW_KEY=new-value
QUOTED_KEY="quoted value"
SINGLE_QUOTED_KEY='single quoted value'
SPACED_KEY = spaced value
EXISTING_KEY=from-file
EMPTY_LINE_WITHOUT_EQUALS
=missing-key
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}

	LoadDotEnv(path)

	tests := map[string]string{
		"NEW_KEY":           "new-value",
		"QUOTED_KEY":        "quoted value",
		"SINGLE_QUOTED_KEY": "single quoted value",
		"SPACED_KEY":        "spaced value",
		"EXISTING_KEY":      "already-set",
	}

	for key, want := range tests {
		if got := os.Getenv(key); got != want {
			t.Fatalf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestLoadDotEnvIgnoresMissingFile(t *testing.T) {
	LoadDotEnv(filepath.Join(t.TempDir(), "missing.env"))
}
