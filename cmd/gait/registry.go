package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/davidahmann/gait/core/registry"
	"github.com/davidahmann/gait/core/sign"
)

type registryInstallOutput struct {
	OK           bool   `json:"ok"`
	Source       string `json:"source,omitempty"`
	PackName     string `json:"pack_name,omitempty"`
	PackVersion  string `json:"pack_version,omitempty"`
	Digest       string `json:"digest,omitempty"`
	MetadataPath string `json:"metadata_path,omitempty"`
	PinPath      string `json:"pin_path,omitempty"`
	Error        string `json:"error,omitempty"`
}

func runRegistry(arguments []string) int {
	if hasExplainFlag(arguments) {
		return writeExplain("Install signed registry policy packs from local or remote sources with allowlisted hosts, cache, and pinning.")
	}
	if len(arguments) == 0 {
		printRegistryUsage()
		return exitInvalidInput
	}
	switch arguments[0] {
	case "install":
		return runRegistryInstall(arguments[1:])
	default:
		printRegistryUsage()
		return exitInvalidInput
	}
}

func runRegistryInstall(arguments []string) int {
	if hasExplainFlag(arguments) {
		return writeExplain("Fetch a registry pack manifest, verify signatures and optional pin digest, then cache and pin locally.")
	}
	arguments = reorderInterspersedFlags(arguments, map[string]bool{
		"source":          true,
		"cache-dir":       true,
		"allow-host":      true,
		"pin":             true,
		"public-key":      true,
		"public-key-env":  true,
		"private-key":     true,
		"private-key-env": true,
	})
	flagSet := flag.NewFlagSet("registry-install", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	var source string
	var cacheDir string
	var allowHostsCSV string
	var pinDigest string
	var publicKeyPath string
	var publicKeyEnv string
	var privateKeyPath string
	var privateKeyEnv string
	var jsonOutput bool
	var helpFlag bool

	flagSet.StringVar(&source, "source", "", "registry pack source (path or URL)")
	flagSet.StringVar(&cacheDir, "cache-dir", "", "registry cache directory (default ~/.gait/registry)")
	flagSet.StringVar(&allowHostsCSV, "allow-host", "", "comma-separated remote host allowlist")
	flagSet.StringVar(&pinDigest, "pin", "", "expected digest (sha256:<hex>)")
	flagSet.StringVar(&publicKeyPath, "public-key", "", "path to base64 public key")
	flagSet.StringVar(&publicKeyEnv, "public-key-env", "", "env var containing base64 public key")
	flagSet.StringVar(&privateKeyPath, "private-key", "", "path to base64 private key (derive public)")
	flagSet.StringVar(&privateKeyEnv, "private-key-env", "", "env var containing base64 private key (derive public)")
	flagSet.BoolVar(&jsonOutput, "json", false, "emit JSON output")
	flagSet.BoolVar(&helpFlag, "help", false, "show help")

	if err := flagSet.Parse(arguments); err != nil {
		return writeRegistryInstallOutput(jsonOutput, registryInstallOutput{OK: false, Error: err.Error()}, exitInvalidInput)
	}
	if helpFlag {
		printRegistryInstallUsage()
		return exitOK
	}
	remaining := flagSet.Args()
	if source == "" && len(remaining) > 0 {
		source = remaining[0]
		remaining = remaining[1:]
	}
	if source == "" || len(remaining) > 0 {
		return writeRegistryInstallOutput(jsonOutput, registryInstallOutput{OK: false, Error: "expected --source <path|url>"}, exitInvalidInput)
	}

	publicKey, err := sign.LoadVerifyKey(sign.KeyConfig{
		PublicKeyPath:  publicKeyPath,
		PublicKeyEnv:   publicKeyEnv,
		PrivateKeyPath: privateKeyPath,
		PrivateKeyEnv:  privateKeyEnv,
	})
	if err != nil {
		return writeRegistryInstallOutput(jsonOutput, registryInstallOutput{OK: false, Error: err.Error()}, exitInvalidInput)
	}

	result, err := registry.Install(context.Background(), registry.InstallOptions{
		Source:     source,
		CacheDir:   cacheDir,
		PublicKey:  publicKey,
		AllowHosts: parseCSVList(allowHostsCSV),
		PinDigest:  pinDigest,
	})
	if err != nil {
		return writeRegistryInstallOutput(jsonOutput, registryInstallOutput{OK: false, Error: err.Error()}, exitVerifyFailed)
	}
	return writeRegistryInstallOutput(jsonOutput, registryInstallOutput{
		OK:           true,
		Source:       result.Source,
		PackName:     result.PackName,
		PackVersion:  result.PackVersion,
		Digest:       result.Digest,
		MetadataPath: result.MetadataPath,
		PinPath:      result.PinPath,
	}, exitOK)
}

func writeRegistryInstallOutput(jsonOutput bool, output registryInstallOutput, exitCode int) int {
	if jsonOutput {
		encoded, err := json.Marshal(output)
		if err != nil {
			fmt.Println(`{"ok":false,"error":"failed to encode output"}`)
			return exitInvalidInput
		}
		fmt.Println(string(encoded))
		return exitCode
	}
	if output.OK {
		fmt.Printf("registry install ok: %s@%s\n", output.PackName, output.PackVersion)
		fmt.Printf("digest: sha256:%s\n", output.Digest)
		return exitCode
	}
	fmt.Printf("registry install error: %s\n", output.Error)
	return exitCode
}

func printRegistryUsage() {
	fmt.Println("Usage:")
	fmt.Println("  gait registry install --source <path|url> --allow-host <csv> [--pin sha256:<hex>] [--cache-dir <path>] [--public-key <path>|--public-key-env <VAR>] [--json] [--explain]")
}

func printRegistryInstallUsage() {
	fmt.Println("Usage:")
	fmt.Println("  gait registry install --source <path|url> --allow-host <csv> [--pin sha256:<hex>] [--cache-dir <path>] [--public-key <path>|--public-key-env <VAR>] [--json] [--explain]")
}
