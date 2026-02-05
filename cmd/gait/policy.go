package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/davidahmann/gait/core/gate"
	"github.com/davidahmann/gait/core/policytest"
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
	Summary       string   `json:"summary,omitempty"`
	Error         string   `json:"error,omitempty"`
}

func runPolicy(arguments []string) int {
	if hasExplainFlag(arguments) {
		return writeExplain("Run deterministic policy validation workflows against intent fixtures.")
	}
	if len(arguments) == 0 {
		printPolicyUsage()
		return exitInvalidInput
	}
	switch arguments[0] {
	case "test":
		return runPolicyTest(arguments[1:])
	default:
		printPolicyUsage()
		return exitInvalidInput
	}
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
		return writePolicyTestOutput(jsonOutput, policyTestOutput{OK: false, Error: err.Error()}, exitInvalidInput)
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
		return writePolicyTestOutput(jsonOutput, policyTestOutput{OK: false, Error: err.Error()}, exitInvalidInput)
	}
	intent, err := readIntentRequest(intentPath)
	if err != nil {
		return writePolicyTestOutput(jsonOutput, policyTestOutput{OK: false, Error: err.Error()}, exitInvalidInput)
	}

	runResult, err := policytest.Run(policytest.RunOptions{
		Policy:          policy,
		Intent:          intent,
		ProducerVersion: version,
	})
	if err != nil {
		return writePolicyTestOutput(jsonOutput, policyTestOutput{OK: false, Error: err.Error()}, exitInvalidInput)
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
		Summary:       runResult.Summary,
	}, exitCode)
}

func writePolicyTestOutput(jsonOutput bool, output policyTestOutput, exitCode int) int {
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
		fmt.Println(output.Summary)
		return exitCode
	}
	fmt.Printf("policy test error: %s\n", output.Error)
	return exitCode
}

func printPolicyUsage() {
	fmt.Println("Usage:")
	fmt.Println("  gait policy test <policy.yaml> <intent_fixture.json> [--json] [--explain]")
}

func printPolicyTestUsage() {
	fmt.Println("Usage:")
	fmt.Println("  gait policy test <policy.yaml> <intent_fixture.json> [--json] [--explain]")
}
