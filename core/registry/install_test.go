package registry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/davidahmann/gait/core/jcs"
	schemaregistry "github.com/davidahmann/gait/core/schema/v1/registry"
	"github.com/davidahmann/gait/core/sign"
)

func TestInstallRemoteWithSignatureAndPin(t *testing.T) {
	keyPair, err := sign.GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate key pair: %v", err)
	}
	manifest := schemaregistry.RegistryPack{
		SchemaID:        "gait.registry.pack",
		SchemaVersion:   "1.0.0",
		CreatedAt:       time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		ProducerVersion: "0.0.0-dev",
		PackName:        "baseline-highrisk",
		PackVersion:     "1.1.0",
		RiskClass:       "high",
		UseCase:         "tool-gating",
		Compatibility:   []string{"gait>=1.0.0"},
		Artifacts: []schemaregistry.PackArtifact{{
			Path:   "policy.yaml",
			SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		}},
	}
	signableDigest, err := signableManifestDigest(manifest)
	if err != nil {
		t.Fatalf("signable manifest digest: %v", err)
	}
	signature, err := sign.SignDigestHex(keyPair.Private, signableDigest)
	if err != nil {
		t.Fatalf("sign digest: %v", err)
	}
	manifest.Signatures = []schemaregistry.SignatureRef{{
		Alg:          signature.Alg,
		KeyID:        signature.KeyID,
		Sig:          signature.Sig,
		SignedDigest: signature.SignedDigest,
	}}
	payload, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write(payload)
	}))
	defer server.Close()

	cacheDir := filepath.Join(t.TempDir(), "cache")
	result, err := Install(context.Background(), InstallOptions{
		Source:     server.URL,
		CacheDir:   cacheDir,
		PublicKey:  keyPair.Public,
		AllowHosts: []string{"127.0.0.1"},
		PinDigest:  "sha256:" + signableDigest,
	})
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	if result.PackName != manifest.PackName {
		t.Fatalf("unexpected pack name: %s", result.PackName)
	}
	if result.Digest != signableDigest {
		t.Fatalf("unexpected digest: %s", result.Digest)
	}
	if _, err := os.Stat(result.MetadataPath); err != nil {
		t.Fatalf("metadata path: %v", err)
	}
	if _, err := os.Stat(result.PinPath); err != nil {
		t.Fatalf("pin path: %v", err)
	}
}

func TestInstallRemoteRequiresAllowlist(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{}`))
	}))
	defer server.Close()

	_, err := Install(context.Background(), InstallOptions{
		Source: server.URL,
	})
	if err == nil {
		t.Fatalf("expected allowlist error")
	}
}

func signableManifestDigest(manifest schemaregistry.RegistryPack) (string, error) {
	signable := manifest
	signable.Signatures = nil
	raw, err := json.Marshal(signable)
	if err != nil {
		return "", err
	}
	return jcs.DigestJCS(raw)
}
