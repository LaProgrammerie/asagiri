package cloud

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/redact"
)

// SaveToken writes the API token to path with mode 0600.
func SaveToken(path, token string) error {
	path = strings.TrimSpace(path)
	token = strings.TrimSpace(token)
	if path == "" {
		return fmt.Errorf("cloud: token_path requis")
	}
	if token == "" {
		return fmt.Errorf("cloud: token requis")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("cloud: mkdir token: %w", err)
	}
	if err := os.WriteFile(path, []byte(token+"\n"), 0o600); err != nil {
		return fmt.Errorf("cloud: écriture token: %w", err)
	}
	return nil
}

// LoadToken reads the API token from path. Missing file returns empty string, no error.
func LoadToken(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("cloud: token_path requis")
	}
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("cloud: lecture token: %w", err)
	}
	return strings.TrimSpace(string(raw)), nil
}

// RemoveToken deletes the token file if present.
func RemoveToken(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("cloud: token_path requis")
	}
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// RedactError masks secrets in HTTP or auth errors for CLI output.
func RedactError(err error) string {
	if err == nil {
		return ""
	}
	return redact.String(err.Error())
}
