package sign

import (
	"crypto/ed25519"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestSignVerifyDigestHex(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	sig, err := SignDigestHex(kp.Private, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatalf("sign digest: %v", err)
	}
	ok, err := VerifyDigestHex(kp.Public, sig)
	if err != nil {
		t.Fatalf("verify digest: %v", err)
	}
	if !ok {
		t.Fatalf("expected signature to verify")
	}
}

func TestVerifyBytesWrongKey(t *testing.T) {
	kp1, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	kp2, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	sig := SignBytes(kp1.Private, []byte("hello"))
	ok, err := VerifyBytes(kp2.Public, sig, []byte("hello"))
	if err == nil {
		t.Fatalf("expected error for key id mismatch")
	}
	if ok {
		t.Fatalf("expected verification to fail with wrong key")
	}
}

func TestParseKeyBase64Invalid(t *testing.T) {
	if _, err := ParsePrivateKeyBase64("not-base64"); err == nil {
		t.Fatalf("expected error for invalid private key")
	}
	if _, err := ParsePublicKeyBase64("not-base64"); err == nil {
		t.Fatalf("expected error for invalid public key")
	}
	short := base64.StdEncoding.EncodeToString([]byte("short"))
	if _, err := ParsePrivateKeyBase64(short); err == nil {
		t.Fatalf("expected error for short private key")
	}
	if _, err := ParsePublicKeyBase64(short); err == nil {
		t.Fatalf("expected error for short public key")
	}
}

func TestParseKeyBase64RoundTrip(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	privEnc := base64.StdEncoding.EncodeToString(kp.Private)
	pubEnc := base64.StdEncoding.EncodeToString(kp.Public)

	priv, err := ParsePrivateKeyBase64(privEnc)
	if err != nil {
		t.Fatalf("parse private: %v", err)
	}
	pub, err := ParsePublicKeyBase64(pubEnc)
	if err != nil {
		t.Fatalf("parse public: %v", err)
	}
	if !ed25519.PublicKey(pub).Equal(kp.Public) {
		t.Fatalf("public key mismatch")
	}
	if !ed25519.PrivateKey(priv).Equal(kp.Private) {
		t.Fatalf("private key mismatch")
	}
}

func TestKeyIDLength(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	id := KeyID(kp.Public)
	if len(id) != 64 {
		t.Fatalf("expected 64 hex chars, got %d", len(id))
	}
}

func TestVerifyBytesInvalidAlg(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	sig := SignBytes(kp.Private, []byte("hello"))
	sig.Alg = "none"
	if _, err := VerifyBytes(kp.Public, sig, []byte("hello")); err == nil {
		t.Fatalf("expected error for unsupported alg")
	}
}

func TestVerifyBytesInvalidSig(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	sig := SignBytes(kp.Private, []byte("hello"))
	sig.Sig = "%%%notbase64"
	if _, err := VerifyBytes(kp.Public, sig, []byte("hello")); err == nil {
		t.Fatalf("expected error for invalid signature")
	}
}

func TestVerifyBytesKeyIDMismatch(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	sig := SignBytes(kp.Private, []byte("hello"))
	sig.KeyID = "deadbeef"
	if _, err := VerifyBytes(kp.Public, sig, []byte("hello")); err == nil {
		t.Fatalf("expected error for key id mismatch")
	}
}

func TestVerifyBytesSignatureLength(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	sig := SignBytes(kp.Private, []byte("hello"))
	sig.Sig = base64.StdEncoding.EncodeToString([]byte("short"))
	if _, err := VerifyBytes(kp.Public, sig, []byte("hello")); err == nil {
		t.Fatalf("expected error for invalid signature length")
	}
}

func TestSignDigestHexInvalid(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	if _, err := SignDigestHex(kp.Private, "not-hex"); err == nil {
		t.Fatalf("expected error for invalid digest")
	}
	if _, err := SignDigestHex(kp.Private, "aa"); err == nil {
		t.Fatalf("expected error for invalid digest length")
	}
}

func TestVerifyDigestHexMissing(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	sig := SignBytes(kp.Private, []byte("hello"))
	if _, err := VerifyDigestHex(kp.Public, sig); err == nil {
		t.Fatalf("expected error for missing signed_digest")
	}
}

func TestVerifyDigestHexLength(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	sig := SignBytes(kp.Private, []byte("hello"))
	sig.SignedDigest = "aa"
	if _, err := VerifyDigestHex(kp.Public, sig); err == nil {
		t.Fatalf("expected error for invalid digest length")
	}
}

func TestLoadKeyBase64Files(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	dir := t.TempDir()
	privPath := filepath.Join(dir, "priv.key")
	pubPath := filepath.Join(dir, "pub.key")

	if err := os.WriteFile(privPath, []byte("  "+base64.StdEncoding.EncodeToString(kp.Private)+"\n"), 0o600); err != nil {
		t.Fatalf("write priv: %v", err)
	}
	if err := os.WriteFile(pubPath, []byte("\n"+base64.StdEncoding.EncodeToString(kp.Public)+"  "), 0o600); err != nil {
		t.Fatalf("write pub: %v", err)
	}
	priv, err := LoadPrivateKeyBase64(privPath)
	if err != nil {
		t.Fatalf("load priv: %v", err)
	}
	pub, err := LoadPublicKeyBase64(pubPath)
	if err != nil {
		t.Fatalf("load pub: %v", err)
	}
	if !ed25519.PrivateKey(priv).Equal(kp.Private) || !ed25519.PublicKey(pub).Equal(kp.Public) {
		t.Fatalf("loaded keys do not match original")
	}
}
