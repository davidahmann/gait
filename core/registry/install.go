package registry

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/davidahmann/gait/core/jcs"
	schemaregistry "github.com/davidahmann/gait/core/schema/v1/registry"
	"github.com/davidahmann/gait/core/sign"
)

type InstallOptions struct {
	Source     string
	CacheDir   string
	PublicKey  ed25519.PublicKey
	AllowHosts []string
	PinDigest  string
	HTTPClient *http.Client
}

type InstallResult struct {
	Source       string                      `json:"source"`
	PackName     string                      `json:"pack_name"`
	PackVersion  string                      `json:"pack_version"`
	Digest       string                      `json:"digest"`
	MetadataPath string                      `json:"metadata_path"`
	PinPath      string                      `json:"pin_path"`
	Manifest     schemaregistry.RegistryPack `json:"manifest"`
}

func Install(ctx context.Context, options InstallOptions) (InstallResult, error) {
	source := strings.TrimSpace(options.Source)
	if source == "" {
		return InstallResult{}, fmt.Errorf("source is required")
	}
	allowHosts := normalizeAllowHosts(options.AllowHosts)
	if isRemoteSource(source) {
		if err := enforceAllowHost(source, allowHosts); err != nil {
			return InstallResult{}, err
		}
	}

	rawManifest, err := fetchSource(ctx, source, options.HTTPClient)
	if err != nil {
		return InstallResult{}, err
	}
	manifest, err := parseRegistryManifest(rawManifest)
	if err != nil {
		return InstallResult{}, err
	}
	signableDigest, signableBytes, err := digestSignableManifest(manifest)
	if err != nil {
		return InstallResult{}, err
	}
	if err := enforcePin(options.PinDigest, signableDigest); err != nil {
		return InstallResult{}, err
	}
	if err := verifySignatures(manifest.Signatures, options.PublicKey, signableDigest); err != nil {
		return InstallResult{}, err
	}

	cacheDir, err := resolveCacheDir(options.CacheDir)
	if err != nil {
		return InstallResult{}, err
	}
	metadataPath := filepath.Join(
		cacheDir,
		manifest.PackName,
		manifest.PackVersion,
		signableDigest,
		"registry_pack.json",
	)
	if err := os.MkdirAll(filepath.Dir(metadataPath), 0o750); err != nil {
		return InstallResult{}, fmt.Errorf("mkdir metadata dir: %w", err)
	}
	if err := os.WriteFile(metadataPath, signableBytes, 0o600); err != nil {
		return InstallResult{}, fmt.Errorf("write registry metadata: %w", err)
	}

	pinsDir := filepath.Join(cacheDir, "pins")
	if err := os.MkdirAll(pinsDir, 0o750); err != nil {
		return InstallResult{}, fmt.Errorf("mkdir pins dir: %w", err)
	}
	pinPath := filepath.Join(pinsDir, manifest.PackName+".pin")
	if err := os.WriteFile(pinPath, []byte("sha256:"+signableDigest+"\n"), 0o600); err != nil {
		return InstallResult{}, fmt.Errorf("write pin file: %w", err)
	}

	return InstallResult{
		Source:       source,
		PackName:     manifest.PackName,
		PackVersion:  manifest.PackVersion,
		Digest:       signableDigest,
		MetadataPath: metadataPath,
		PinPath:      pinPath,
		Manifest:     manifest,
	}, nil
}

func normalizeAllowHosts(allowHosts []string) []string {
	out := make([]string, 0, len(allowHosts))
	for _, allowHost := range allowHosts {
		for _, segment := range strings.Split(allowHost, ",") {
			trimmed := strings.ToLower(strings.TrimSpace(segment))
			if trimmed == "" {
				continue
			}
			out = append(out, trimmed)
		}
	}
	sort.Strings(out)
	dedup := make([]string, 0, len(out))
	for _, value := range out {
		if len(dedup) == 0 || dedup[len(dedup)-1] != value {
			dedup = append(dedup, value)
		}
	}
	return dedup
}

func isRemoteSource(source string) bool {
	return strings.HasPrefix(strings.ToLower(source), "https://") ||
		strings.HasPrefix(strings.ToLower(source), "http://")
}

func enforceAllowHost(source string, allowHosts []string) error {
	parsed, err := url.Parse(source)
	if err != nil {
		return fmt.Errorf("parse source url: %w", err)
	}
	if len(allowHosts) == 0 {
		return fmt.Errorf("remote source requires --allow-host")
	}
	host := strings.ToLower(parsed.Hostname())
	for _, allowed := range allowHosts {
		if host == allowed {
			return nil
		}
	}
	return fmt.Errorf("source host %s is not in allowlist", host)
}

func fetchSource(ctx context.Context, source string, client *http.Client) ([]byte, error) {
	if !isRemoteSource(source) {
		// #nosec G304 -- user-supplied local file path is intentional.
		raw, err := os.ReadFile(source)
		if err != nil {
			return nil, fmt.Errorf("read source file: %w", err)
		}
		return raw, nil
	}

	httpClient := client
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return nil, fmt.Errorf("build source request: %w", err)
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("download source: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download source: unexpected status %d", response.StatusCode)
	}
	raw, err := ioReadAllLimit(response.Body, 5*1024*1024)
	if err != nil {
		return nil, fmt.Errorf("read source response: %w", err)
	}
	return raw, nil
}

func parseRegistryManifest(raw []byte) (schemaregistry.RegistryPack, error) {
	var manifest schemaregistry.RegistryPack
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return schemaregistry.RegistryPack{}, fmt.Errorf("parse registry manifest: %w", err)
	}
	if manifest.SchemaID != "gait.registry.pack" {
		return schemaregistry.RegistryPack{}, fmt.Errorf("unsupported schema_id %q", manifest.SchemaID)
	}
	if manifest.SchemaVersion != "1.0.0" {
		return schemaregistry.RegistryPack{}, fmt.Errorf("unsupported schema_version %q", manifest.SchemaVersion)
	}
	if strings.TrimSpace(manifest.PackName) == "" {
		return schemaregistry.RegistryPack{}, fmt.Errorf("pack_name is required")
	}
	if strings.TrimSpace(manifest.PackVersion) == "" {
		return schemaregistry.RegistryPack{}, fmt.Errorf("pack_version is required")
	}
	return manifest, nil
}

func digestSignableManifest(manifest schemaregistry.RegistryPack) (string, []byte, error) {
	signable := manifest
	signable.Signatures = nil
	raw, err := json.Marshal(signable)
	if err != nil {
		return "", nil, fmt.Errorf("marshal signable registry manifest: %w", err)
	}
	canonical, err := jcs.CanonicalizeJSON(raw)
	if err != nil {
		return "", nil, fmt.Errorf("canonicalize signable registry manifest: %w", err)
	}
	digest, err := jcs.DigestJCS(canonical)
	if err != nil {
		return "", nil, fmt.Errorf("digest signable registry manifest: %w", err)
	}
	return digest, canonical, nil
}

func enforcePin(expected string, actual string) error {
	trimmed := strings.TrimSpace(expected)
	if trimmed == "" {
		return nil
	}
	trimmed = strings.TrimPrefix(strings.ToLower(trimmed), "sha256:")
	if trimmed != strings.ToLower(actual) {
		return fmt.Errorf("pin digest mismatch")
	}
	return nil
}

func verifySignatures(signatures []schemaregistry.SignatureRef, publicKey ed25519.PublicKey, digest string) error {
	if len(signatures) == 0 {
		return fmt.Errorf("registry manifest has no signatures")
	}
	if len(publicKey) == 0 {
		return fmt.Errorf("public key is required for signature verification")
	}
	valid := 0
	for _, signatureRef := range signatures {
		signature := sign.Signature{
			Alg:          signatureRef.Alg,
			KeyID:        signatureRef.KeyID,
			Sig:          signatureRef.Sig,
			SignedDigest: signatureRef.SignedDigest,
		}
		if !strings.EqualFold(signature.SignedDigest, digest) {
			continue
		}
		ok, err := sign.VerifyDigestHex(publicKey, signature)
		if err != nil {
			continue
		}
		if ok {
			valid++
		}
	}
	if valid == 0 {
		return fmt.Errorf("no valid signature for manifest digest")
	}
	return nil
}

func resolveCacheDir(cacheDir string) (string, error) {
	trimmed := strings.TrimSpace(cacheDir)
	if trimmed != "" {
		return filepath.Clean(trimmed), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}
	return filepath.Join(home, ".gait", "registry"), nil
}

func ioReadAllLimit(reader io.Reader, maxBytes int64) ([]byte, error) {
	limited := io.LimitReader(reader, maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("payload too large")
	}
	return data, nil
}
