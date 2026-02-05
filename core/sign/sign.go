package sign

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
)

const AlgEd25519 = "ed25519"

type KeyPair struct {
	Public  ed25519.PublicKey
	Private ed25519.PrivateKey
}

type Signature struct {
	Alg          string `json:"alg"`
	KeyID        string `json:"key_id"`
	Sig          string `json:"sig"`
	SignedDigest string `json:"signed_digest,omitempty"`
}

func GenerateKeyPair() (KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return KeyPair{}, err
	}
	return KeyPair{Public: pub, Private: priv}, nil
}

func KeyID(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	return hex.EncodeToString(sum[:])
}

func SignBytes(priv ed25519.PrivateKey, data []byte) Signature {
	sig := ed25519.Sign(priv, data)
	return Signature{
		Alg:   AlgEd25519,
		KeyID: KeyID(priv.Public().(ed25519.PublicKey)),
		Sig:   base64.StdEncoding.EncodeToString(sig),
	}
}

func VerifyBytes(pub ed25519.PublicKey, sig Signature, data []byte) (bool, error) {
	if sig.Alg != AlgEd25519 {
		return false, fmt.Errorf("unsupported alg: %s", sig.Alg)
	}
	if sig.KeyID != "" && sig.KeyID != KeyID(pub) {
		return false, fmt.Errorf("key id mismatch")
	}
	rawSig, err := base64.StdEncoding.DecodeString(sig.Sig)
	if err != nil {
		return false, fmt.Errorf("decode sig: %w", err)
	}
	if len(rawSig) != ed25519.SignatureSize {
		return false, fmt.Errorf("invalid signature length: %d", len(rawSig))
	}
	return ed25519.Verify(pub, data, rawSig), nil
}

func SignDigestHex(priv ed25519.PrivateKey, digestHex string) (Signature, error) {
	digest, err := hex.DecodeString(digestHex)
	if err != nil {
		return Signature{}, fmt.Errorf("decode digest: %w", err)
	}
	if len(digest) != sha256.Size {
		return Signature{}, fmt.Errorf("invalid digest length: %d", len(digest))
	}
	sig := SignBytes(priv, digest)
	sig.SignedDigest = digestHex
	return sig, nil
}

func VerifyDigestHex(pub ed25519.PublicKey, sig Signature) (bool, error) {
	if sig.SignedDigest == "" {
		return false, fmt.Errorf("missing signed_digest")
	}
	digest, err := hex.DecodeString(sig.SignedDigest)
	if err != nil {
		return false, fmt.Errorf("decode digest: %w", err)
	}
	if len(digest) != sha256.Size {
		return false, fmt.Errorf("invalid digest length: %d", len(digest))
	}
	return VerifyBytes(pub, sig, digest)
}

func LoadPrivateKeyBase64(path string) (ed25519.PrivateKey, error) {
	// #nosec G304 -- caller supplies local key path by design
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	return ParsePrivateKeyBase64(string(bytesTrimSpace(b)))
}

func LoadPublicKeyBase64(path string) (ed25519.PublicKey, error) {
	// #nosec G304 -- caller supplies local key path by design
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}
	return ParsePublicKeyBase64(string(bytesTrimSpace(b)))
}

func ParsePrivateKeyBase64(encoded string) (ed25519.PrivateKey, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode private key: %w", err)
	}
	if l := len(raw); l != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key length: %d", l)
	}
	return ed25519.PrivateKey(raw), nil
}

func ParsePublicKeyBase64(encoded string) (ed25519.PublicKey, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode public key: %w", err)
	}
	if l := len(raw); l != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key length: %d", l)
	}
	return ed25519.PublicKey(raw), nil
}

func bytesTrimSpace(b []byte) []byte {
	i := 0
	j := len(b)
	for i < j && (b[i] == ' ' || b[i] == '\n' || b[i] == '\r' || b[i] == '\t') {
		i++
	}
	for j > i && (b[j-1] == ' ' || b[j-1] == '\n' || b[j-1] == '\r' || b[j-1] == '\t') {
		j--
	}
	return b[i:j]
}
