package pkg

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const certRuntimeDir = "/tmp/certs"

// InitializeCertificates validates certificate configuration and returns a new
// Config instance with concrete file paths. When base64 content is supplied we
// decode it into a runtime file so downstream code can keep using paths.
func InitializeCertificates(cfg Config) (Config, error) {
	if cfg.AppleP12Path != "" && cfg.AppleRootCertPath != "" {
		return cfg, nil
	}

	if cfg.AppleP12Path == "" && cfg.AppleP12Base64 == "" {
		return Config{}, errors.New("either APPLE_P12_PATH or APPLE_P12_BASE64 must be set")
	}
	if cfg.AppleRootCertPath == "" && cfg.AppleRootBase64 == "" {
		return Config{}, errors.New("either APPLE_ROOT_CERT_PATH or APPLE_ROOT_CERT_BASE64 must be set")
	}

	if err := os.MkdirAll(certRuntimeDir, 0o700); err != nil {
		return Config{}, fmt.Errorf("creting directory %s: %w", certRuntimeDir, err)
	}

	nextCfg := cfg

	p12Path, err := decodeCertFile(
		nextCfg.AppleP12Base64,
		"apple-signing.p12",
	)
	if err != nil {
		return Config{}, fmt.Errorf("apple signing certificate: %w", err)
	}

	rootPath, err := decodeCertFile(
		nextCfg.AppleRootBase64,
		"apple-root.cer",
	)
	if err != nil {
		return Config{}, fmt.Errorf("apple root certificate: %w", err)
	}

	nextCfg.AppleP12Path = p12Path
	nextCfg.AppleRootCertPath = rootPath

	return nextCfg, nil
}

// Returns the decoded cert file path into /tmp
func decodeCertFile(encoded string, fileName string) (string, error) {
	if encoded == "" {
		return "", fmt.Errorf("No encoding cert to decode")
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode base64: %w", err)
	}

	certPath := filepath.Join(certRuntimeDir, fileName)
	if err := os.WriteFile(certPath, data, 0o600); err != nil {
		return "", fmt.Errorf("write file %s: %w", certPath, err)
	}

	return certPath, nil
}
