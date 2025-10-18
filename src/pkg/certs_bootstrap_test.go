package pkg

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureCertificateFiles_UsesExistingPaths(t *testing.T) {
	dir := t.TempDir()

	p12Path := filepath.Join(dir, "cert.p12")
	rootPath := filepath.Join(dir, "root.cer")

	if err := os.WriteFile(p12Path, []byte("signing"), 0o600); err != nil {
		t.Fatalf("write p12 fixture: %v", err)
	}
	if err := os.WriteFile(rootPath, []byte("root"), 0o600); err != nil {
		t.Fatalf("write root fixture: %v", err)
	}

	cfg := Config{
		DataDir:           t.TempDir(),
		AppleP12Path:      p12Path,
		AppleRootCertPath: rootPath,
	}

	next, err := InitializeCertificates(cfg)
	if err != nil {
		t.Fatalf("EnsureCertificateFiles: %v", err)
	}

	if next.AppleP12Path != p12Path {
		t.Fatalf("expected AppleP12Path to remain %s, got %s", p12Path, next.AppleP12Path)
	}
	if next.AppleRootCertPath != rootPath {
		t.Fatalf("expected AppleRootCert to remain %s, got %s", rootPath, next.AppleRootCertPath)
	}
}

func TestEnsureCertificateFiles_WritesFromBase64(t *testing.T) {
	encodedP12 := base64.StdEncoding.EncodeToString([]byte("signing-data"))
	encodedRoot := base64.StdEncoding.EncodeToString([]byte("root-data"))

	cfg := Config{
		AppleP12Base64:  encodedP12,
		AppleRootBase64: encodedRoot,
	}

	next, err := InitializeCertificates(cfg)
	if err != nil {
		t.Fatalf("EnsureCertificateFiles: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Remove(filepath.Join(certRuntimeDir, "apple-signing.p12"))
		_ = os.Remove(filepath.Join(certRuntimeDir, "apple-root.cer"))
	})

	for _, entry := range []struct {
		name     string
		path     string
		expected []byte
	}{
		{
			name:     "signing",
			path:     next.AppleP12Path,
			expected: []byte("signing-data"),
		},
		{
			name:     "root",
			path:     next.AppleRootCertPath,
			expected: []byte("root-data"),
		},
	} {
		data, err := os.ReadFile(entry.path)
		if err != nil {
			t.Fatalf("read %s certificate: %v", entry.name, err)
		}
		if string(data) != string(entry.expected) {
			t.Fatalf("certificate %s mismatch: want %q got %q", entry.name, entry.expected, data)
		}
		if filepath.Dir(entry.path) != certRuntimeDir {
			t.Fatalf("certificate %s not created under expected directory %s", entry.name, certRuntimeDir)
		}
	}
}

func TestEnsureCertificateFiles_ErrorWhenMissing(t *testing.T) {
	cfg := Config{
		DataDir: t.TempDir(),
	}

	_, err := InitializeCertificates(cfg)
	if err == nil {
		t.Fatal("expected error when certificates are not configured")
	}
}
