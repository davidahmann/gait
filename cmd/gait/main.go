package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/davidahmann/gait/core/scout"
)

const version = "0.0.0-dev"

func main() {
	os.Exit(run(os.Args))
}

func run(arguments []string) int {
	startedAt := time.Now()
	correlationID := newCorrelationID(arguments)
	setCurrentCorrelationID(correlationID)
	command := normalizeAdoptionCommand(arguments)
	writeOperationalEventStart(command, correlationID, startedAt.UTC())
	exitCode := runDispatch(arguments)
	finishedAt := time.Now().UTC()
	elapsed := time.Since(startedAt)
	writeAdoptionEvent(command, exitCode, elapsed, finishedAt)
	writeOperationalEventEnd(command, correlationID, exitCode, elapsed, finishedAt)
	setCurrentCorrelationID("")
	return exitCode
}

func runDispatch(arguments []string) int {
	if len(arguments) < 2 {
		fmt.Println("gait", version)
		return exitOK
	}
	if arguments[1] == "--explain" {
		return writeExplain("Gait is an offline-first CLI for deterministic run artifacts, policy-gated tool execution, and reproducible debugging workflows.")
	}

	switch arguments[1] {
	case "approve":
		return runApprove(arguments[2:])
	case "delegate":
		return runDelegate(arguments[2:])
	case "demo":
		return runDemo(arguments[2:])
	case "doctor":
		return runDoctor(arguments[2:])
	case "gate":
		return runGate(arguments[2:])
	case "policy":
		return runPolicy(arguments[2:])
	case "keys":
		return runKeys(arguments[2:])
	case "trace":
		return runTrace(arguments[2:])
	case "regress":
		return runRegress(arguments[2:])
	case "run":
		return runCommand(arguments[2:])
	case "scout":
		return runScout(arguments[2:])
	case "guard":
		return runGuard(arguments[2:])
	case "incident":
		return runIncident(arguments[2:])
	case "registry":
		return runRegistry(arguments[2:])
	case "migrate":
		return runMigrate(arguments[2:])
	case "mcp":
		return runMCP(arguments[2:])
	case "verify":
		return runVerify(arguments[2:])
	case "version", "--version", "-v":
		if hasExplainFlag(arguments[2:]) {
			return writeExplain("Print the CLI version.")
		}
		fmt.Println("gait", version)
		return exitOK
	default:
		printUsage()
		return exitInvalidInput
	}
}

func normalizeAdoptionCommand(arguments []string) string {
	if len(arguments) < 2 {
		return "version"
	}
	command := strings.TrimSpace(arguments[1])
	if command == "" {
		return "unknown"
	}
	switch command {
	case "--version", "-v", "version":
		return "version"
	case "--explain":
		return "explain"
	case "gate", "policy", "keys", "trace", "regress", "run", "scout", "guard", "incident", "registry", "mcp", "doctor", "delegate":
		if len(arguments) > 2 {
			subcommand := strings.TrimSpace(arguments[2])
			if subcommand != "" && !strings.HasPrefix(subcommand, "-") {
				return command + " " + subcommand
			}
		}
	}
	return command
}

func writeAdoptionEvent(command string, exitCode int, elapsed time.Duration, now time.Time) {
	adoptionPath := strings.TrimSpace(os.Getenv("GAIT_ADOPTION_LOG"))
	if adoptionPath == "" {
		return
	}
	event := scout.NewAdoptionEvent(command, exitCode, elapsed, version, now)
	_ = scout.AppendAdoptionEvent(adoptionPath, event)
}

func writeOperationalEventStart(command string, correlationID string, now time.Time) {
	operationalPath := strings.TrimSpace(os.Getenv("GAIT_OPERATIONAL_LOG"))
	if operationalPath == "" {
		return
	}
	event := scout.NewOperationalStartEvent(command, correlationID, version, now)
	_ = scout.AppendOperationalEvent(operationalPath, event)
}

func writeOperationalEventEnd(command string, correlationID string, exitCode int, elapsed time.Duration, now time.Time) {
	operationalPath := strings.TrimSpace(os.Getenv("GAIT_OPERATIONAL_LOG"))
	if operationalPath == "" {
		return
	}
	category := "none"
	retryable := false
	if exitCode != exitOK {
		resolvedCategory := defaultErrorCategory(exitCode)
		category = string(resolvedCategory)
		retryable = defaultRetryable(resolvedCategory)
	}
	event := scout.NewOperationalEndEvent(command, correlationID, version, exitCode, category, retryable, elapsed, now)
	_ = scout.AppendOperationalEvent(operationalPath, event)
}
