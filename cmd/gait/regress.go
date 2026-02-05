package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/davidahmann/gait/core/regress"
)

type regressInitOutput struct {
	OK           bool     `json:"ok"`
	RunID        string   `json:"run_id,omitempty"`
	FixtureName  string   `json:"fixture_name,omitempty"`
	FixtureDir   string   `json:"fixture_dir,omitempty"`
	RunpackPath  string   `json:"runpack_path,omitempty"`
	ConfigPath   string   `json:"config_path,omitempty"`
	NextCommands []string `json:"next_commands,omitempty"`
	Error        string   `json:"error,omitempty"`
}

func runRegress(arguments []string) int {
	if len(arguments) == 0 {
		printRegressUsage()
		return exitInvalidInput
	}
	switch arguments[0] {
	case "init":
		return runRegressInit(arguments[1:])
	default:
		printRegressUsage()
		return exitInvalidInput
	}
}

func runRegressInit(arguments []string) int {
	flagSet := flag.NewFlagSet("regress-init", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	var from string
	var name string
	var jsonOutput bool
	var helpFlag bool

	flagSet.StringVar(&from, "from", "", "run_id or path")
	flagSet.StringVar(&name, "name", "", "fixture name override")
	flagSet.BoolVar(&jsonOutput, "json", false, "emit JSON output")
	flagSet.BoolVar(&helpFlag, "help", false, "show help")

	if err := flagSet.Parse(arguments); err != nil {
		return writeRegressInitOutput(jsonOutput, regressInitOutput{OK: false, Error: err.Error()}, exitInvalidInput)
	}
	if helpFlag {
		printRegressInitUsage()
		return exitOK
	}
	if len(flagSet.Args()) > 0 {
		return writeRegressInitOutput(jsonOutput, regressInitOutput{
			OK:    false,
			Error: "unexpected positional arguments",
		}, exitInvalidInput)
	}
	if strings.TrimSpace(from) == "" {
		return writeRegressInitOutput(jsonOutput, regressInitOutput{
			OK:    false,
			Error: "missing required --from <run_id|path>",
		}, exitInvalidInput)
	}

	runpackPath, err := resolveRunpackPath(from)
	if err != nil {
		return writeRegressInitOutput(jsonOutput, regressInitOutput{OK: false, Error: err.Error()}, exitInvalidInput)
	}

	result, err := regress.InitFixture(regress.InitOptions{
		SourceRunpackPath: runpackPath,
		FixtureName:       name,
		WorkDir:           ".",
	})
	if err != nil {
		return writeRegressInitOutput(jsonOutput, regressInitOutput{OK: false, Error: err.Error()}, exitInvalidInput)
	}

	return writeRegressInitOutput(jsonOutput, regressInitOutput{
		OK:           true,
		RunID:        result.RunID,
		FixtureName:  result.FixtureName,
		FixtureDir:   result.FixtureDir,
		RunpackPath:  result.RunpackPath,
		ConfigPath:   result.ConfigPath,
		NextCommands: result.NextCommands,
	}, exitOK)
}

func writeRegressInitOutput(jsonOutput bool, output regressInitOutput, exitCode int) int {
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
		fmt.Printf("regress init ok: fixture=%s run_id=%s\n", output.FixtureName, output.RunID)
		fmt.Printf("config: %s\n", output.ConfigPath)
		fmt.Printf("runpack: %s\n", output.RunpackPath)
		if len(output.NextCommands) > 0 {
			fmt.Printf("next: %s\n", strings.Join(output.NextCommands, " | "))
		}
		return exitCode
	}
	fmt.Printf("regress init error: %s\n", output.Error)
	return exitCode
}

func printRegressUsage() {
	fmt.Println("Usage:")
	fmt.Println("  gait regress init --from <run_id|path> [--name <fixture_name>] [--json]")
}

func printRegressInitUsage() {
	fmt.Println("Usage:")
	fmt.Println("  gait regress init --from <run_id|path> [--name <fixture_name>] [--json]")
}
