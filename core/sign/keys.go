package sign

import (
	"crypto/ed25519"
	"fmt"
	"os"
	"strings"
)

type KeyMode string

const (
	ModeDev  KeyMode = "dev"
	ModeProd KeyMode = "prod"
)

const DevKeyWarning = "dev mode: ephemeral keypair generated; signatures will not verify across machines"

type KeyConfig struct {
	Mode           KeyMode
	PrivateKeyPath string
	PublicKeyPath  string
	PrivateKeyEnv  string
	PublicKeyEnv   string
}

func LoadSigningKey(cfg KeyConfig) (KeyPair, []string, error) {
	mode := cfg.Mode
	if mode == "" {
		mode = ModeProd
	}
	switch mode {
	case ModeDev:
		if cfg.hasAnyKeySource() {
			return KeyPair{}, nil, fmt.Errorf("dev mode does not accept explicit key sources")
		}
		kp, err := GenerateKeyPair()
		if err != nil {
			return KeyPair{}, nil, err
		}
		return kp, []string{DevKeyWarning}, nil
	case ModeProd:
		if !cfg.hasPrivateSource() {
			return KeyPair{}, nil, fmt.Errorf("prod mode requires a private key source")
		}
		priv, err := loadPrivateKey(cfg)
		if err != nil {
			return KeyPair{}, nil, err
		}
		pub := priv.Public().(ed25519.PublicKey)
		if cfg.hasPublicSource() {
			loaded, err := loadPublicKey(cfg)
			if err != nil {
				return KeyPair{}, nil, err
			}
			if !loaded.Equal(pub) {
				return KeyPair{}, nil, fmt.Errorf("public key does not match private key")
			}
			pub = loaded
		}
		return KeyPair{Public: pub, Private: priv}, nil, nil
	default:
		return KeyPair{}, nil, fmt.Errorf("unsupported key mode: %q", cfg.Mode)
	}
}

func LoadVerifyKey(cfg KeyConfig) (ed25519.PublicKey, error) {
	if cfg.PublicKeyPath != "" && cfg.PublicKeyEnv != "" {
		return nil, fmt.Errorf("public key source: set either path or env")
	}
	if cfg.PrivateKeyPath != "" && cfg.PrivateKeyEnv != "" {
		return nil, fmt.Errorf("private key source: set either path or env")
	}
	if cfg.hasPublicSource() {
		return loadPublicKey(cfg)
	}
	if cfg.hasPrivateSource() {
		priv, err := loadPrivateKey(cfg)
		if err != nil {
			return nil, err
		}
		return priv.Public().(ed25519.PublicKey), nil
	}
	return nil, fmt.Errorf("public key not configured")
}

func (cfg KeyConfig) hasPrivateSource() bool {
	return cfg.PrivateKeyPath != "" || cfg.PrivateKeyEnv != ""
}

func (cfg KeyConfig) hasPublicSource() bool {
	return cfg.PublicKeyPath != "" || cfg.PublicKeyEnv != ""
}

func (cfg KeyConfig) hasAnyKeySource() bool {
	return cfg.hasPrivateSource() || cfg.hasPublicSource()
}

func loadPrivateKey(cfg KeyConfig) (ed25519.PrivateKey, error) {
	if cfg.PrivateKeyPath != "" && cfg.PrivateKeyEnv != "" {
		return nil, fmt.Errorf("private key source: set either path or env")
	}
	if cfg.PrivateKeyPath != "" {
		return LoadPrivateKeyBase64(cfg.PrivateKeyPath)
	}
	if cfg.PrivateKeyEnv != "" {
		encoded, ok := readEnvValue(cfg.PrivateKeyEnv)
		if !ok {
			return nil, fmt.Errorf("private key env not set: %s", cfg.PrivateKeyEnv)
		}
		return ParsePrivateKeyBase64(encoded)
	}
	return nil, fmt.Errorf("private key not configured")
}

func loadPublicKey(cfg KeyConfig) (ed25519.PublicKey, error) {
	if cfg.PublicKeyPath != "" && cfg.PublicKeyEnv != "" {
		return nil, fmt.Errorf("public key source: set either path or env")
	}
	if cfg.PublicKeyPath != "" {
		return LoadPublicKeyBase64(cfg.PublicKeyPath)
	}
	if cfg.PublicKeyEnv != "" {
		encoded, ok := readEnvValue(cfg.PublicKeyEnv)
		if !ok {
			return nil, fmt.Errorf("public key env not set: %s", cfg.PublicKeyEnv)
		}
		return ParsePublicKeyBase64(encoded)
	}
	return nil, fmt.Errorf("public key not configured")
}

func readEnvValue(name string) (string, bool) {
	if name == "" {
		return "", false
	}
	val, ok := os.LookupEnv(name)
	if !ok {
		return "", false
	}
	val = strings.TrimSpace(val)
	if val == "" {
		return "", false
	}
	return val, true
}
