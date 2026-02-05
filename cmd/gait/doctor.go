package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/davidahmann/gait/core/doctor"
	"github.com/davidahmann/gait/core/sign"
)

type doctorOutput struct {
	OK              bool           `json:"ok"`
	SchemaID        string         `json:"schema_id,omitempty"`
	SchemaVersion   string         `json:"schema_version,omitempty"`
	CreatedAt       string         `json:"created_at,omitempty"`
	ProducerVersion string         `json:"producer_version,omitempty"`
	Status          string         `json:"status,omitempty"`
	NonFixable      bool           `json:"non_fixable,omitempty"`
	Summary         string         `json:"summary,omitempty"`
	FixCommands     []string       `json:"fix_commands,omitempty"`
	Checks          []doctor.Check `json:"checks,omitempty"`
	Error           string         `json:"error,omitempty"`
}

func runDoctor(arguments []string) int {
	flagSet := flag.NewFlagSet("doctor", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	var workDir string
	var outputDir string
	var keyMode string
	var privateKeyPath string
	var privateKeyEnv string
	var publicKeyPath string
	var publicKeyEnv string
	var jsonOutput bool
	var helpFlag bool

	flagSet.StringVar(&workDir, "workdir", ".", "workspace path for checks")
	flagSet.StringVar(&outputDir, "output-dir", "./gait-out", "default output directory to check")
	flagSet.StringVar(&keyMode, "key-mode", string(sign.ModeDev), "key mode to validate: dev or prod")
	flagSet.StringVar(&privateKeyPath, "private-key", "", "path to base64 private key")
	flagSet.StringVar(&privateKeyEnv, "private-key-env", "", "env var containing base64 private key")
	flagSet.StringVar(&publicKeyPath, "public-key", "", "path to base64 public key")
	flagSet.StringVar(&publicKeyEnv, "public-key-env", "", "env var containing base64 public key")
	flagSet.BoolVar(&jsonOutput, "json", false, "emit JSON output")
	flagSet.BoolVar(&helpFlag, "help", false, "show help")

	if err := flagSet.Parse(arguments); err != nil {
		return writeDoctorOutput(jsonOutput, doctorOutput{OK: false, Error: err.Error()}, exitInvalidInput)
	}
	if helpFlag {
		printDoctorUsage()
		return exitOK
	}
	if len(flagSet.Args()) > 0 {
		return writeDoctorOutput(jsonOutput, doctorOutput{OK: false, Error: "unexpected positional arguments"}, exitInvalidInput)
	}

	result := doctor.Run(doctor.Options{
		WorkDir:         workDir,
		OutputDir:       outputDir,
		ProducerVersion: version,
		KeyMode:         sign.KeyMode(strings.ToLower(strings.TrimSpace(keyMode))),
		KeyConfig: sign.KeyConfig{
			PrivateKeyPath: privateKeyPath,
			PrivateKeyEnv:  privateKeyEnv,
			PublicKeyPath:  publicKeyPath,
			PublicKeyEnv:   publicKeyEnv,
		},
	})

	exitCode := exitOK
	if result.NonFixable {
		exitCode = exitMissingDependency
	}
	return writeDoctorOutput(jsonOutput, doctorOutput{
		OK:              !result.NonFixable,
		SchemaID:        result.SchemaID,
		SchemaVersion:   result.SchemaVersion,
		CreatedAt:       result.CreatedAt,
		ProducerVersion: result.ProducerVersion,
		Status:          result.Status,
		NonFixable:      result.NonFixable,
		Summary:         result.Summary,
		FixCommands:     result.FixCommands,
		Checks:          result.Checks,
	}, exitCode)
}

func writeDoctorOutput(jsonOutput bool, output doctorOutput, exitCode int) int {
	if jsonOutput {
		encoded, err := json.Marshal(output)
		if err != nil {
			fmt.Println(`{"ok":false,"error":"failed to encode output"}`)
			return exitInvalidInput
		}
		fmt.Println(string(encoded))
		return exitCode
	}

	if output.Error != "" {
		fmt.Printf("doctor error: %s\n", output.Error)
		return exitCode
	}
	fmt.Println(output.Summary)
	for _, check := range output.Checks {
		fmt.Printf("- %s: %s (%s)\n", check.Name, check.Status, check.Message)
		if check.FixCommand != "" {
			fmt.Printf("  fix: %s\n", check.FixCommand)
		}
	}
	return exitCode
}

func printDoctorUsage() {
	fmt.Println("Usage:")
	fmt.Println("  gait doctor [--workdir <path>] [--output-dir <path>] [--key-mode dev|prod] [--private-key <path>|--private-key-env <VAR>] [--public-key <path>|--public-key-env <VAR>] [--json]")
}
