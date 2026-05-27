package api

import (
	"crypto/subtle"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const tokenFileName = "api.token"

// LoadToken reads the optional API token from .asagiri/runtime/api.token.
func LoadToken(repoRoot string) (string, error) {
	path := filepath.Join(repoRoot, ".asagiri/runtime", tokenFileName)
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(raw)), nil
}

func authMiddleware(token string, next http.Handler) http.Handler {
	if token == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("Authorization")
		if strings.HasPrefix(got, "Bearer ") {
			got = strings.TrimPrefix(got, "Bearer ")
		}
		if got == "" {
			got = r.Header.Get("X-Asagiri-Token")
		}
		if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
