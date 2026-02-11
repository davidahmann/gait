package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/davidahmann/gait/core/gate"
	"github.com/davidahmann/gait/core/policytest"
	"github.com/goccy/go-yaml"
)

type policyTestOutput struct {
	OK            bool     `json:"ok"`
	SchemaID      string   `json:"schema_id,omitempty"`
	SchemaVersion string   `json:"schema_version,omitempty"`
	CreatedAt     string   `json:"created_at,omitempty"`
	PolicyDigest  string   `json:"policy_digest,omitempty"`
	IntentDigest  string   `json:"intent_digest,omitempty"`
	Verdict       string   `json:"verdict,omitempty"`
	ReasonCodes   []string `json:"reason_codes,omitempty"`
	Violations    []string `json:"violations,omitempty"`
	MatchedRule   string   `json:"matched_rule,omitempty"`
	Summary       string   `json:"summary,omitempty"`
	Error         string   `json:"error,omitempty"`
}

type policyInitOutput struct {
	OK           bool     `json:"ok"`
	Template     string   `json:"template,omitempty"`
	PolicyPath   string   `json:"policy_path,omitempty"`
	NextCommands []string `json:"next_commands,omitempty"`
	Error        string   `json:"error,omitempty"`
}

type policyValidateOutput struct {
	OK             bool   `json:"ok"`
	SchemaID       string `json:"schema_id,omitempty"`
	SchemaVersion  string `json:"schema_version,omitempty"`
	PolicyDigest   string `json:"policy_digest,omitempty"`
	DefaultVerdict string `json:"default_verdict,omitempty"`
	RuleCount      int    `json:"rule_count,omitempty"`
	Summary        string `json:"summary,omitempty"`
	Error          string `json:"error,omitempty"`
}

type policyFmtOutput struct {
	OK             bool   `json:"ok"`
	Path           string `json:"path,omitempty"`
	Written        bool   `json:"written,omitempty"`
	Changed        bool   `json:"changed,omitempty"`
	PolicyDigest   string `json:"policy_digest,omitempty"`
	DefaultVerdict string `json:"default_verdict,omitempty"`
	RuleCount      int    `json:"rule_count,omitempty"`
	Formatted      string `json:"formatted,omitempty"`
	Error          string `json:"error,omitempty"`
}

//go:embed policy_templates/baseline-lowrisk.yaml
var policyTemplateBaselineLowRisk string

//go:embed policy_templates/baseline-mediumrisk.yaml
var policyTemplateBaselineMediumRisk string

//go:embed policy_templates/baseline-highrisk.yaml
var policyTemplateBaselineHighRisk string

var policyTemplates = map[string]string{
	"baseline-lowrisk":    policyTemplateBaselineLowRisk,
	"baseline-mediumrisk": policyTemplateBaselineMediumRisk,
	"baseline-highrisk":   policyTemplateBaselineHighRisk,
}

var policyTemplateAliases = map[string]string{
	"baseline-lowrisk":     "baseline-lowrisk",
	"baseline-mediumrisk":  "baseline-mediumrisk",
	"baseline-highrisk":    "baseline-highrisk",
	"baseline_low_risk":    "baseline-lowrisk",
	"baseline_medium_risk": "baseline-mediumrisk",
	"baseline_high_risk":   "baseline-highrisk",
	"low":                  "baseline-lowrisk",
	"medium":               "baseline-mediumrisk",
	"high":                 "baseline-highrisk",
}

func runPolicy(arguments []string) int {
	if hasExplainFlag(arguments) {
		return writeExplain("Initialize, validate, format, and test Gate policies deterministically before rollout.")
	}
	if len(arguments) == 0 {
		printPolicyUsage()
		return exitInvalidInput
	}
	switch arguments[0] {
	case "init":
		return runPolicyInit(arguments[1:])
	case "validate":
		return runPolicyValidate(arguments[1:])
	case "fmt":
		return runPolicyFmt(arguments[1:])
	case "test":
		return runPolicyTest(arguments[1:])
	default:
		printPolicyUsage()
		return exitInvalidInput
	}
}

func runPolicyInit(arguments []string) int {
	if hasExplainFlag(arguments) {
		return writeExplain("Write a starter policy scaffold for low, medium, or high risk tool-control rollouts.")
	}
	arguments = reorderInterspersedFlags(arguments, map[string]bool{
		"out": true,
	})

	flagSet := flag.NewFlagSet("policy-init", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	var outPath string
	var force bool
	var jsonOutput bool
	var helpFlag bool

	flagSet.StringVar(&outPath, "out", "gait.policy.yaml", "output path for generated policy")
	flagSet.BoolVar(&force, "force", false, "overwrite existing output path")
	flagSet.BoolVar(&jsonOutput, "json", false, "emit JSON output")
	flagSet.BoolVar(&helpFlag, "help", false, "show help")

	if err := flagSet.Parse(arguments); err != nil {
		return writePolicyInitOutput(jsonOutput, policyInitOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}
	if helpFlag {
		printPolicyInitUsage()
		return exitOK
	}
	if len(flagSet.Args()) != 1 {
		return writePolicyInitOutput(jsonOutput, policyInitOutput{
			OK:    false,
			Error: "expected <template>, one of: baseline-lowrisk, baseline-mediumrisk, baseline-highrisk",
		}, exitInvalidInput)
	}

	templateKey := strings.ToLower(strings.TrimSpace(flagSet.Args()[0]))
	resolvedTemplate, ok := policyTemplateAliases[templateKey]
	if !ok {
		return writePolicyInitOutput(jsonOutput, policyInitOutput{
			OK:    false,
			Error: "unknown template: " + templateKey,
		}, exitInvalidInput)
	}
	templateBody, ok := policyTemplates[resolvedTemplate]
	if !ok {
		return writePolicyInitOutput(jsonOutput, policyInitOutput{
			OK:    false,
			Error: "template is not available: " + resolvedTemplate,
		}, exitInvalidInput)
	}

	trimmedOutPath := strings.TrimSpace(outPath)
	if trimmedOutPath == "" {
		return writePolicyInitOutput(jsonOutput, policyInitOutput{
			OK:    false,
			Error: "output path must not be empty",
		}, exitInvalidInput)
	}

	if !force {
		if _, err := os.Stat(trimmedOutPath); err == nil {
			return writePolicyInitOutput(jsonOutput, policyInitOutput{
				OK:    false,
				Error: "output path already exists (use --force to overwrite): " + trimmedOutPath,
			}, exitInvalidInput)
		}
	}

	parentDir := filepath.Dir(trimmedOutPath)
	if parentDir != "." && parentDir != "" {
		if err := os.MkdirAll(parentDir, 0o750); err != nil {
			return writePolicyInitOutput(jsonOutput, policyInitOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
		}
	}
	if err := os.WriteFile(trimmedOutPath, []byte(templateBody), 0o600); err != nil {
		return writePolicyInitOutput(jsonOutput, policyInitOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}

	nextCommands := []string{
		fmt.Sprintf("gait policy validate %s --json", trimmedOutPath),
		fmt.Sprintf("gait policy fmt %s --write --json", trimmedOutPath),
		fmt.Sprintf("gait policy test %s examples/policy/intents/intent_read.json --json", trimmedOutPath),
		fmt.Sprintf("gait policy test %s examples/policy/intents/intent_write.json --json", trimmedOutPath),
		fmt.Sprintf("gait policy test %s examples/policy/intents/intent_delete.json --json", trimmedOutPath),
	}

	return writePolicyInitOutput(jsonOutput, policyInitOutput{
		OK:           true,
		Template:     resolvedTemplate,
		PolicyPath:   trimmedOutPath,
		NextCommands: nextCommands,
	}, exitOK)
}

func runPolicyValidate(arguments []string) int {
	if hasExplainFlag(arguments) {
		return writeExplain("Parse and normalize one policy file with strict YAML checks and return deterministic metadata.")
	}
	arguments = reorderInterspersedFlags(arguments, nil)

	flagSet := flag.NewFlagSet("policy-validate", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	var jsonOutput bool
	var helpFlag bool

	flagSet.BoolVar(&jsonOutput, "json", false, "emit JSON output")
	flagSet.BoolVar(&helpFlag, "help", false, "show help")

	if err := flagSet.Parse(arguments); err != nil {
		return writePolicyValidateOutput(jsonOutput, policyValidateOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}
	if helpFlag {
		printPolicyValidateUsage()
		return exitOK
	}
	if len(flagSet.Args()) != 1 {
		return writePolicyValidateOutput(jsonOutput, policyValidateOutput{
			OK:    false,
			Error: "expected <policy.yaml>",
		}, exitInvalidInput)
	}

	policyPath := flagSet.Args()[0]
	policy, err := gate.LoadPolicyFile(policyPath)
	if err != nil {
		return writePolicyValidateOutput(jsonOutput, policyValidateOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}
	policyDigest, err := gate.PolicyDigest(policy)
	if err != nil {
		return writePolicyValidateOutput(jsonOutput, policyValidateOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}
	summary := fmt.Sprintf("policy validate ok: default=%s rules=%d digest=%s", policy.DefaultVerdict, len(policy.Rules), policyDigest)
	return writePolicyValidateOutput(jsonOutput, policyValidateOutput{
		OK:             true,
		SchemaID:       policy.SchemaID,
		SchemaVersion:  policy.SchemaVersion,
		PolicyDigest:   policyDigest,
		DefaultVerdict: policy.DefaultVerdict,
		RuleCount:      len(policy.Rules),
		Summary:        summary,
	}, exitOK)
}

func runPolicyFmt(arguments []string) int {
	if hasExplainFlag(arguments) {
		return writeExplain("Normalize one policy file and emit deterministic YAML formatting with optional write-back.")
	}
	arguments = reorderInterspersedFlags(arguments, nil)

	flagSet := flag.NewFlagSet("policy-fmt", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	var writeFlag bool
	var jsonOutput bool
	var helpFlag bool

	flagSet.BoolVar(&writeFlag, "write", false, "write formatted YAML back to the policy path")
	flagSet.BoolVar(&jsonOutput, "json", false, "emit JSON output")
	flagSet.BoolVar(&helpFlag, "help", false, "show help")

	if err := flagSet.Parse(arguments); err != nil {
		return writePolicyFmtOutput(jsonOutput, policyFmtOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}
	if helpFlag {
		printPolicyFmtUsage()
		return exitOK
	}
	if len(flagSet.Args()) != 1 {
		return writePolicyFmtOutput(jsonOutput, policyFmtOutput{
			OK:    false,
			Error: "expected <policy.yaml>",
		}, exitInvalidInput)
	}

	policyPath := strings.TrimSpace(flagSet.Args()[0])
	if policyPath == "" {
		return writePolicyFmtOutput(jsonOutput, policyFmtOutput{
			OK:    false,
			Error: "policy path must not be empty",
		}, exitInvalidInput)
	}
	content, err := os.ReadFile(policyPath) // #nosec G304 -- explicit local user input path.
	if err != nil {
		return writePolicyFmtOutput(jsonOutput, policyFmtOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}
	policy, err := gate.ParsePolicyYAML(content)
	if err != nil {
		return writePolicyFmtOutput(jsonOutput, policyFmtOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}
	formatted, err := formatPolicyYAML(policy)
	if err != nil {
		return writePolicyFmtOutput(jsonOutput, policyFmtOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}
	changed := string(content) != string(formatted)
	policyDigest, err := gate.PolicyDigest(policy)
	if err != nil {
		return writePolicyFmtOutput(jsonOutput, policyFmtOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}

	if writeFlag && changed {
		fileInfo, statErr := os.Stat(policyPath)
		mode := os.FileMode(0o600)
		if statErr == nil {
			mode = fileInfo.Mode().Perm()
		}
		if err := os.WriteFile(policyPath, formatted, mode); err != nil {
			return writePolicyFmtOutput(jsonOutput, policyFmtOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
		}
	}

	output := policyFmtOutput{
		OK:             true,
		Path:           policyPath,
		Written:        writeFlag && changed,
		Changed:        changed,
		PolicyDigest:   policyDigest,
		DefaultVerdict: policy.DefaultVerdict,
		RuleCount:      len(policy.Rules),
	}
	if !writeFlag && jsonOutput {
		output.Formatted = string(formatted)
	}

	if !writeFlag && !jsonOutput {
		fmt.Print(string(formatted))
		return exitOK
	}
	return writePolicyFmtOutput(jsonOutput, output, exitOK)
}

func runPolicyTest(arguments []string) int {
	if hasExplainFlag(arguments) {
		return writeExplain("Evaluate one intent fixture against one policy and return a deterministic verdict with reason codes.")
	}
	arguments = reorderInterspersedFlags(arguments, nil)

	flagSet := flag.NewFlagSet("policy-test", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	var jsonOutput bool
	var helpFlag bool

	flagSet.BoolVar(&jsonOutput, "json", false, "emit JSON output")
	flagSet.BoolVar(&helpFlag, "help", false, "show help")

	if err := flagSet.Parse(arguments); err != nil {
		return writePolicyTestOutput(jsonOutput, policyTestOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}
	if helpFlag {
		printPolicyTestUsage()
		return exitOK
	}
	if len(flagSet.Args()) != 2 {
		return writePolicyTestOutput(jsonOutput, policyTestOutput{
			OK:    false,
			Error: "expected <policy.yaml> <intent_fixture.json>",
		}, exitInvalidInput)
	}

	policyPath := flagSet.Args()[0]
	intentPath := flagSet.Args()[1]

	policy, err := gate.LoadPolicyFile(policyPath)
	if err != nil {
		return writePolicyTestOutput(jsonOutput, policyTestOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}
	intent, err := readIntentRequest(intentPath)
	if err != nil {
		return writePolicyTestOutput(jsonOutput, policyTestOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}

	runResult, err := policytest.Run(policytest.RunOptions{
		Policy:          policy,
		Intent:          intent,
		ProducerVersion: version,
	})
	if err != nil {
		return writePolicyTestOutput(jsonOutput, policyTestOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}

	exitCode := exitOK
	switch runResult.Result.Verdict {
	case "block":
		exitCode = exitPolicyBlocked
	case "require_approval":
		exitCode = exitApprovalRequired
	}

	return writePolicyTestOutput(jsonOutput, policyTestOutput{
		OK:            true,
		SchemaID:      runResult.Result.SchemaID,
		SchemaVersion: runResult.Result.SchemaVersion,
		CreatedAt:     runResult.Result.CreatedAt.UTC().Format(time.RFC3339Nano),
		PolicyDigest:  runResult.Result.PolicyDigest,
		IntentDigest:  runResult.Result.IntentDigest,
		Verdict:       runResult.Result.Verdict,
		ReasonCodes:   runResult.Result.ReasonCodes,
		Violations:    runResult.Result.Violations,
		MatchedRule:   runResult.Result.MatchedRule,
		Summary:       runResult.Summary,
	}, exitCode)
}

func writePolicyTestOutput(jsonOutput bool, output policyTestOutput, exitCode int) int {
	if jsonOutput {
		return writeJSONOutput(output, exitCode)
	}

	if output.OK {
		fmt.Println(output.Summary)
		return exitCode
	}
	fmt.Printf("policy test error: %s\n", output.Error)
	return exitCode
}

func writePolicyInitOutput(jsonOutput bool, output policyInitOutput, exitCode int) int {
	if jsonOutput {
		return writeJSONOutput(output, exitCode)
	}
	if output.OK {
		fmt.Printf("policy init ok: template=%s output=%s\n", output.Template, output.PolicyPath)
		if len(output.NextCommands) > 0 {
			fmt.Printf("next: %s\n", strings.Join(output.NextCommands, " | "))
		}
		return exitCode
	}
	fmt.Printf("policy init error: %s\n", output.Error)
	return exitCode
}

func writePolicyValidateOutput(jsonOutput bool, output policyValidateOutput, exitCode int) int {
	if jsonOutput {
		return writeJSONOutput(output, exitCode)
	}
	if output.OK {
		fmt.Println(output.Summary)
		return exitCode
	}
	fmt.Printf("policy validate error: %s\n", output.Error)
	return exitCode
}

func writePolicyFmtOutput(jsonOutput bool, output policyFmtOutput, exitCode int) int {
	if jsonOutput {
		return writeJSONOutput(output, exitCode)
	}
	if output.OK {
		fmt.Printf("policy fmt ok: path=%s changed=%t written=%t rules=%d\n", output.Path, output.Changed, output.Written, output.RuleCount)
		return exitCode
	}
	fmt.Printf("policy fmt error: %s\n", output.Error)
	return exitCode
}

func formatPolicyYAML(policy gate.Policy) ([]byte, error) {
	encoded, err := yaml.MarshalWithOptions(policy, yaml.Indent(2))
	if err != nil {
		return nil, fmt.Errorf("encode policy yaml: %w", err)
	}
	if len(encoded) == 0 || encoded[len(encoded)-1] != '\n' {
		encoded = append(encoded, '\n')
	}
	return encoded, nil
}

func printPolicyUsage() {
	fmt.Println("Usage:")
	fmt.Println("  gait policy init <baseline-lowrisk|baseline-mediumrisk|baseline-highrisk> [--out gait.policy.yaml] [--force] [--json] [--explain]")
	fmt.Println("  gait policy validate <policy.yaml> [--json] [--explain]")
	fmt.Println("  gait policy fmt <policy.yaml> [--write] [--json] [--explain]")
	fmt.Println("  gait policy test <policy.yaml> <intent_fixture.json> [--json] [--explain]")
}

func printPolicyInitUsage() {
	fmt.Println("Usage:")
	fmt.Println("  gait policy init <baseline-lowrisk|baseline-mediumrisk|baseline-highrisk> [--out gait.policy.yaml] [--force] [--json] [--explain]")
	fmt.Println("Aliases:")
	fmt.Println("  baseline_high_risk, baseline_medium_risk, baseline_low_risk, high, medium, low")
}

func printPolicyTestUsage() {
	fmt.Println("Usage:")
	fmt.Println("  gait policy test <policy.yaml> <intent_fixture.json> [--json] [--explain]")
}

func printPolicyValidateUsage() {
	fmt.Println("Usage:")
	fmt.Println("  gait policy validate <policy.yaml> [--json] [--explain]")
}

func printPolicyFmtUsage() {
	fmt.Println("Usage:")
	fmt.Println("  gait policy fmt <policy.yaml> [--write] [--json] [--explain]")
}
