package main

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/davidahmann/gait/core/doctor"
	"github.com/davidahmann/gait/core/runpack"
	schemagate "github.com/davidahmann/gait/core/schema/v1/gate"
	"github.com/davidahmann/gait/core/sign"
)

func TestRunDispatch(t *testing.T) {
	if code := run([]string{"gait"}); code != exitOK {
		t.Fatalf("run without args: expected %d got %d", exitOK, code)
	}
	if code := run([]string{"gait", "version"}); code != exitOK {
		t.Fatalf("run version: expected %d got %d", exitOK, code)
	}
	if code := run([]string{"gait", "unknown"}); code != exitInvalidInput {
		t.Fatalf("run unknown: expected %d got %d", exitInvalidInput, code)
	}
	if code := run([]string{"gait", "approve", "--help"}); code != exitOK {
		t.Fatalf("run approve help: expected %d got %d", exitOK, code)
	}
	if code := run([]string{"gait", "gate", "eval", "--help"}); code != exitOK {
		t.Fatalf("run gate help: expected %d got %d", exitOK, code)
	}
	if code := run([]string{"gait", "policy", "test", "--help"}); code != exitOK {
		t.Fatalf("run policy help: expected %d got %d", exitOK, code)
	}
	if code := run([]string{"gait", "trace", "verify", "--help"}); code != exitOK {
		t.Fatalf("run trace help: expected %d got %d", exitOK, code)
	}
	if code := run([]string{"gait", "regress", "init", "--help"}); code != exitOK {
		t.Fatalf("run regress init help: expected %d got %d", exitOK, code)
	}
	if code := run([]string{"gait", "regress", "run", "--help"}); code != exitOK {
		t.Fatalf("run regress run help: expected %d got %d", exitOK, code)
	}
	if code := run([]string{"gait", "run", "replay", "--help"}); code != exitOK {
		t.Fatalf("run replay help: expected %d got %d", exitOK, code)
	}
	if code := run([]string{"gait", "run", "diff", "--help"}); code != exitOK {
		t.Fatalf("run diff help: expected %d got %d", exitOK, code)
	}
	if code := run([]string{"gait", "verify", "--help"}); code != exitOK {
		t.Fatalf("run verify help: expected %d got %d", exitOK, code)
	}
	if code := run([]string{"gait", "doctor", "--help"}); code != exitOK {
		t.Fatalf("run doctor help: expected %d got %d", exitOK, code)
	}
}

func TestMainEntrypoint(t *testing.T) {
	if os.Getenv("GAIT_TEST_MAIN") == "1" {
		os.Args = []string{"gait", "version"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMainEntrypoint")
	cmd.Env = append(os.Environ(), "GAIT_TEST_MAIN=1")
	if err := cmd.Run(); err != nil {
		t.Fatalf("run child process: %v", err)
	}
}

func TestRunpackPathHelpers(t *testing.T) {
	workDir := t.TempDir()
	runpackDir := filepath.Join(workDir, "gait-out")
	if err := os.MkdirAll(runpackDir, 0o750); err != nil {
		t.Fatalf("mkdir runpack dir: %v", err)
	}
	runpackPath := filepath.Join(runpackDir, "runpack_run_demo.zip")
	if err := os.WriteFile(runpackPath, []byte("zip"), 0o600); err != nil {
		t.Fatalf("write runpack: %v", err)
	}

	withWorkingDir(t, workDir)
	resolved, err := resolveRunpackPath("run_demo")
	if err != nil {
		t.Fatalf("resolve run id: %v", err)
	}
	if resolved != filepath.Join(".", "gait-out", "runpack_run_demo.zip") {
		t.Fatalf("unexpected resolved path: %s", resolved)
	}

	resolved, err = resolveRunpackPath(runpackPath)
	if err != nil {
		t.Fatalf("resolve path: %v", err)
	}
	if resolved != runpackPath {
		t.Fatalf("unexpected resolved runpack path: %s", resolved)
	}

	if !looksLikePath("a/b") {
		t.Fatalf("expected path with separator to match")
	}
	if !looksLikePath("thing.zip") {
		t.Fatalf("expected .zip to match")
	}
	if looksLikePath("run_demo") {
		t.Fatalf("unexpected path detection")
	}
	if !fileExists(runpackPath) {
		t.Fatalf("expected file to exist")
	}
	if fileExists(runpackDir) {
		t.Fatalf("directory should not be treated as file")
	}
}

func TestDemoVerifyReplayDiffAndRegressFlow(t *testing.T) {
	workDir := t.TempDir()
	withWorkingDir(t, workDir)

	if code := runDemo(nil); code != exitOK {
		t.Fatalf("demo: expected %d got %d", exitOK, code)
	}

	if code := runVerify([]string{"--json", "run_demo"}); code != exitOK {
		t.Fatalf("verify: expected %d got %d", exitOK, code)
	}
	if code := runReplay([]string{"--json", "run_demo"}); code != exitOK {
		t.Fatalf("replay: expected %d got %d", exitOK, code)
	}
	if code := runReplay([]string{"--real-tools", "run_demo"}); code != exitUnsafeReplay {
		t.Fatalf("unsafe replay guard: expected %d got %d", exitUnsafeReplay, code)
	}

	diffPath := filepath.Join(workDir, "diff.json")
	if code := runDiff([]string{"--json", "--output", diffPath, "run_demo", "run_demo"}); code != exitOK {
		t.Fatalf("diff: expected %d got %d", exitOK, code)
	}
	if _, err := os.Stat(diffPath); err != nil {
		t.Fatalf("diff output: %v", err)
	}

	if code := runRegressInit([]string{"--from", "run_demo", "--json"}); code != exitOK {
		t.Fatalf("regress init: expected %d got %d", exitOK, code)
	}
	if code := runRegressRun([]string{"--json", "--junit", "junit.xml"}); code != exitOK {
		t.Fatalf("regress run pass: expected %d got %d", exitOK, code)
	}

	fixturePath := filepath.Join(workDir, "fixtures", "run_demo", "fixture.json")
	rawFixture, err := os.ReadFile(fixturePath) // #nosec G304
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixture map[string]any
	if err := json.Unmarshal(rawFixture, &fixture); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	fixture["expected_replay_exit_code"] = float64(exitVerifyFailed)
	encodedFixture, err := json.MarshalIndent(fixture, "", "  ")
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	if err := os.WriteFile(fixturePath, append(encodedFixture, '\n'), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	if code := runRegressRun([]string{"--json"}); code != exitRegressFailed {
		t.Fatalf("regress run fail: expected %d got %d", exitRegressFailed, code)
	}
}

func TestGatePolicyTraceApproveAndDoctor(t *testing.T) {
	workDir := t.TempDir()
	withWorkingDir(t, workDir)

	privateKeyPath := filepath.Join(workDir, "private.key")
	writePrivateKey(t, privateKeyPath)

	intentPath := filepath.Join(workDir, "intent.json")
	writeIntentFixture(t, intentPath, "tool.write")

	policyAllowPath := filepath.Join(workDir, "policy_allow.yaml")
	mustWriteFile(t, policyAllowPath, strings.Join([]string{
		"schema_id: gait.gate.policy",
		"schema_version: 1.0.0",
		"default_verdict: allow",
		"rules:",
		"  - name: allow-write",
		"    priority: 1",
		"    effect: allow",
		"    match:",
		"      tool_names: [tool.write]",
		"      target_kinds: [path]",
		"    reason_codes: [allow_rule]",
	}, "\n")+"\n")

	tracePath := filepath.Join(workDir, "trace_allow.json")
	if code := runGateEval([]string{
		"--policy", policyAllowPath,
		"--intent", intentPath,
		"--trace-out", tracePath,
		"--key-mode", "prod",
		"--private-key", privateKeyPath,
		"--json",
	}); code != exitOK {
		t.Fatalf("gate eval allow: expected %d got %d", exitOK, code)
	}
	if code := runTraceVerify([]string{"--path", tracePath, "--private-key", privateKeyPath, "--json"}); code != exitOK {
		t.Fatalf("trace verify: expected %d got %d", exitOK, code)
	}

	if code := runPolicyTest([]string{"--json", policyAllowPath, intentPath}); code != exitOK {
		t.Fatalf("policy test allow: expected %d got %d", exitOK, code)
	}

	policyBlockPath := filepath.Join(workDir, "policy_block.yaml")
	mustWriteFile(t, policyBlockPath, strings.Join([]string{
		"default_verdict: allow",
		"rules:",
		"  - name: block-write",
		"    priority: 1",
		"    effect: block",
		"    match:",
		"      tool_names: [tool.write]",
		"    reason_codes: [blocked_write]",
	}, "\n")+"\n")
	if code := runPolicyTest([]string{policyBlockPath, intentPath, "--json"}); code != exitPolicyBlocked {
		t.Fatalf("policy test block: expected %d got %d", exitPolicyBlocked, code)
	}

	policyApprovalPath := filepath.Join(workDir, "policy_approval.yaml")
	mustWriteFile(t, policyApprovalPath, "default_verdict: require_approval\n")
	if code := runGateEval([]string{
		"--policy", policyApprovalPath,
		"--intent", intentPath,
		"--trace-out", filepath.Join(workDir, "trace_approval.json"),
		"--key-mode", "prod",
		"--private-key", privateKeyPath,
		"--json",
	}); code != exitApprovalRequired {
		t.Fatalf("gate eval approval: expected %d got %d", exitApprovalRequired, code)
	}

	tokenPath := filepath.Join(workDir, "approval_token.json")
	if code := runApprove([]string{
		"--intent-digest", strings.Repeat("a", 64),
		"--policy-digest", strings.Repeat("b", 64),
		"--ttl", "1h",
		"--scope", "tool:tool.write",
		"--approver", "alice",
		"--reason-code", "change-ticket",
		"--out", tokenPath,
		"--json",
	}); code != exitOK {
		t.Fatalf("approve: expected %d got %d", exitOK, code)
	}
	if _, err := os.Stat(tokenPath); err != nil {
		t.Fatalf("approval token path: %v", err)
	}
	if code := runApprove([]string{"--ttl", "bad", "--json"}); code != exitInvalidInput {
		t.Fatalf("approve invalid: expected %d got %d", exitInvalidInput, code)
	}

	repoRoot := repoRootFromPackageDir(t)
	if code := runDoctor([]string{
		"--workdir", repoRoot,
		"--output-dir", filepath.Join(workDir, "gait-out"),
		"--json",
	}); code != exitOK {
		t.Fatalf("doctor valid: expected %d got %d", exitOK, code)
	}
	if code := runDoctor([]string{
		"--workdir", workDir,
		"--output-dir", filepath.Join(workDir, "gait-out"),
		"--json",
	}); code != exitMissingDependency {
		t.Fatalf("doctor missing schemas: expected %d got %d", exitMissingDependency, code)
	}
}

func TestCommandRoutersAndHelpers(t *testing.T) {
	if code := runGate(nil); code != exitInvalidInput {
		t.Fatalf("runGate no args: expected %d got %d", exitInvalidInput, code)
	}
	if code := runGate([]string{"unknown"}); code != exitInvalidInput {
		t.Fatalf("runGate unknown: expected %d got %d", exitInvalidInput, code)
	}
	if code := runPolicy(nil); code != exitInvalidInput {
		t.Fatalf("runPolicy no args: expected %d got %d", exitInvalidInput, code)
	}
	if code := runPolicy([]string{"unknown"}); code != exitInvalidInput {
		t.Fatalf("runPolicy unknown: expected %d got %d", exitInvalidInput, code)
	}
	if code := runTrace(nil); code != exitInvalidInput {
		t.Fatalf("runTrace no args: expected %d got %d", exitInvalidInput, code)
	}
	if code := runTrace([]string{"unknown"}); code != exitInvalidInput {
		t.Fatalf("runTrace unknown: expected %d got %d", exitInvalidInput, code)
	}
	if code := runRegress(nil); code != exitInvalidInput {
		t.Fatalf("runRegress no args: expected %d got %d", exitInvalidInput, code)
	}
	if code := runRegress([]string{"unknown"}); code != exitInvalidInput {
		t.Fatalf("runRegress unknown: expected %d got %d", exitInvalidInput, code)
	}
	if code := runCommand(nil); code != exitInvalidInput {
		t.Fatalf("runCommand no args: expected %d got %d", exitInvalidInput, code)
	}
	if code := runCommand([]string{"unknown"}); code != exitInvalidInput {
		t.Fatalf("runCommand unknown: expected %d got %d", exitInvalidInput, code)
	}

	if joined := joinCSV([]string{"a", "b", "c"}); joined != "a,b,c" {
		t.Fatalf("joinCSV mismatch: %s", joined)
	}
	if values := parseCSV("a, b, ,c"); len(values) != 3 || values[1] != "b" {
		t.Fatalf("parseCSV mismatch: %#v", values)
	}
	merged := mergeUniqueSorted([]string{"b", "a", "a"}, []string{"c", "b", " "})
	if strings.Join(merged, ",") != "a,b,c" {
		t.Fatalf("mergeUniqueSorted mismatch: %#v", merged)
	}

	if !hasAnyKeySource(sign.KeyConfig{PrivateKeyPath: "x"}) {
		t.Fatalf("expected key source detection")
	}
	if hasAnyKeySource(sign.KeyConfig{}) {
		t.Fatalf("unexpected key source detection")
	}
}

func TestValidationBranches(t *testing.T) {
	workDir := t.TempDir()
	withWorkingDir(t, workDir)

	if code := runDemo([]string{"--help"}); code != exitOK {
		t.Fatalf("runDemo help: expected %d got %d", exitOK, code)
	}
	if code := runGateEval([]string{}); code != exitInvalidInput {
		t.Fatalf("runGateEval missing args: expected %d got %d", exitInvalidInput, code)
	}
	if code := runPolicyTest([]string{"--json"}); code != exitInvalidInput {
		t.Fatalf("runPolicyTest missing positional args: expected %d got %d", exitInvalidInput, code)
	}
	if code := runTraceVerify([]string{"--json"}); code != exitInvalidInput {
		t.Fatalf("runTraceVerify missing path: expected %d got %d", exitInvalidInput, code)
	}
	if code := runVerify([]string{"--json"}); code != exitInvalidInput {
		t.Fatalf("runVerify missing path: expected %d got %d", exitInvalidInput, code)
	}
	if code := runRegressInit([]string{"--json"}); code != exitInvalidInput {
		t.Fatalf("runRegressInit missing --from: expected %d got %d", exitInvalidInput, code)
	}
	if code := runRegressRun([]string{"--json", "extra"}); code != exitInvalidInput {
		t.Fatalf("runRegressRun positional args: expected %d got %d", exitInvalidInput, code)
	}
	if code := runDiff([]string{"--json", "left-only"}); code != exitInvalidInput {
		t.Fatalf("runDiff positional args: expected %d got %d", exitInvalidInput, code)
	}
	if code := runDemo(nil); code != exitOK {
		t.Fatalf("runDemo setup: expected %d got %d", exitOK, code)
	}
	if code := runVerify([]string{"--json", "--require-signature", "run_demo"}); code != exitVerifyFailed {
		t.Fatalf("runVerify require signature: expected %d got %d", exitVerifyFailed, code)
	}

	policyPath := filepath.Join(workDir, "policy.yaml")
	intentPath := filepath.Join(workDir, "intent.json")
	mustWriteFile(t, policyPath, "default_verdict: require_approval\n")
	writeIntentFixture(t, intentPath, "tool.write")

	if code := runGateEval([]string{
		"--policy", policyPath,
		"--intent", intentPath,
		"--approval-token", filepath.Join(workDir, "missing-token.json"),
		"--json",
	}); code != exitInvalidInput {
		t.Fatalf("runGateEval missing approval token path: expected %d got %d", exitInvalidInput, code)
	}

	privateKeyPath := filepath.Join(workDir, "private.key")
	writePrivateKey(t, privateKeyPath)
	if code := runGateEval([]string{
		"--policy", policyPath,
		"--intent", intentPath,
		"--key-mode", "prod",
		"--private-key", privateKeyPath,
		"--approval-public-key-env", "MISSING_PUBLIC_KEY",
		"--approval-token", filepath.Join(workDir, "missing-token.json"),
		"--json",
	}); code != exitInvalidInput {
		t.Fatalf("runGateEval missing verify key env: expected %d got %d", exitInvalidInput, code)
	}

	if code := runVerify([]string{"--json", "--public-key-env", "MISSING_PUBLIC_KEY", "run_demo"}); code != exitInvalidInput {
		t.Fatalf("runVerify missing env key: expected %d got %d", exitInvalidInput, code)
	}

	gaitOutAsFile := filepath.Join(workDir, "gait-out")
	if err := os.RemoveAll(gaitOutAsFile); err != nil {
		t.Fatalf("remove gait-out dir: %v", err)
	}
	mustWriteFile(t, gaitOutAsFile, "not-a-dir\n")
	if code := runDemo(nil); code != exitInvalidInput {
		t.Fatalf("runDemo with invalid output dir: expected %d got %d", exitInvalidInput, code)
	}
}

func TestOutputWritersAndUsagePrinters(t *testing.T) {
	if code := writeApproveOutput(true, approveOutput{OK: true, TokenPath: "token.json"}, exitOK); code != exitOK {
		t.Fatalf("writeApproveOutput json: expected %d got %d", exitOK, code)
	}
	if code := writeApproveOutput(false, approveOutput{OK: false, Error: "bad"}, exitInvalidInput); code != exitInvalidInput {
		t.Fatalf("writeApproveOutput text: expected %d got %d", exitInvalidInput, code)
	}

	if code := writeGateEvalOutput(true, gateEvalOutput{OK: true, Verdict: "allow"}, exitOK); code != exitOK {
		t.Fatalf("writeGateEvalOutput json: expected %d got %d", exitOK, code)
	}
	if code := writeGateEvalOutput(false, gateEvalOutput{OK: false, Error: "bad"}, exitInvalidInput); code != exitInvalidInput {
		t.Fatalf("writeGateEvalOutput text: expected %d got %d", exitInvalidInput, code)
	}
	if code := writeGateEvalOutput(false, gateEvalOutput{
		OK:          true,
		Verdict:     "allow",
		TracePath:   "trace.json",
		ReasonCodes: []string{"r1"},
		Violations:  []string{"v1"},
		Warnings:    []string{"w1"},
	}, exitOK); code != exitOK {
		t.Fatalf("writeGateEvalOutput text ok: expected %d got %d", exitOK, code)
	}

	if code := writePolicyTestOutput(true, policyTestOutput{OK: true, Summary: "ok"}, exitOK); code != exitOK {
		t.Fatalf("writePolicyTestOutput json: expected %d got %d", exitOK, code)
	}
	if code := writePolicyTestOutput(false, policyTestOutput{OK: false, Error: "bad"}, exitInvalidInput); code != exitInvalidInput {
		t.Fatalf("writePolicyTestOutput text: expected %d got %d", exitInvalidInput, code)
	}

	if code := writeRegressInitOutput(true, regressInitOutput{OK: true, FixtureName: "f"}, exitOK); code != exitOK {
		t.Fatalf("writeRegressInitOutput json: expected %d got %d", exitOK, code)
	}
	if code := writeRegressInitOutput(false, regressInitOutput{OK: false, Error: "bad"}, exitInvalidInput); code != exitInvalidInput {
		t.Fatalf("writeRegressInitOutput text: expected %d got %d", exitInvalidInput, code)
	}
	if code := writeRegressInitOutput(false, regressInitOutput{
		OK:           true,
		FixtureName:  "fixture",
		RunID:        "run_demo",
		ConfigPath:   "gait.yaml",
		RunpackPath:  "fixtures/run_demo/runpack.zip",
		NextCommands: []string{"gait regress run --json"},
	}, exitOK); code != exitOK {
		t.Fatalf("writeRegressInitOutput text ok: expected %d got %d", exitOK, code)
	}

	if code := writeRegressRunOutput(true, regressRunOutput{OK: true, FixtureSet: "default"}, exitOK); code != exitOK {
		t.Fatalf("writeRegressRunOutput json: expected %d got %d", exitOK, code)
	}
	if code := writeRegressRunOutput(false, regressRunOutput{OK: false, Error: "bad"}, exitInvalidInput); code != exitInvalidInput {
		t.Fatalf("writeRegressRunOutput text err: expected %d got %d", exitInvalidInput, code)
	}
	if code := writeRegressRunOutput(false, regressRunOutput{
		OK:         true,
		FixtureSet: "default",
		Graders:    3,
		Output:     "regress_result.json",
		JUnit:      "junit.xml",
	}, exitOK); code != exitOK {
		t.Fatalf("writeRegressRunOutput text ok: expected %d got %d", exitOK, code)
	}
	if code := writeRegressRunOutput(false, regressRunOutput{OK: false, FixtureSet: "default", Failed: 1}, exitRegressFailed); code != exitRegressFailed {
		t.Fatalf("writeRegressRunOutput text fail: expected %d got %d", exitRegressFailed, code)
	}

	if code := writeDiffOutput(true, diffOutput{OK: true}, exitOK); code != exitOK {
		t.Fatalf("writeDiffOutput json: expected %d got %d", exitOK, code)
	}
	if code := writeDiffOutput(false, diffOutput{OK: false, Error: "bad"}, exitInvalidInput); code != exitInvalidInput {
		t.Fatalf("writeDiffOutput text err: expected %d got %d", exitInvalidInput, code)
	}
	if code := writeDiffOutput(false, diffOutput{
		OK: false,
		Summary: runpack.DiffSummary{
			RunIDLeft:    "left",
			RunIDRight:   "right",
			FilesChanged: []string{"manifest.json"},
		},
	}, exitVerifyFailed); code != exitVerifyFailed {
		t.Fatalf("writeDiffOutput text fail: expected %d got %d", exitVerifyFailed, code)
	}

	if code := writeReplayOutput(true, replayOutput{OK: true}, exitOK); code != exitOK {
		t.Fatalf("writeReplayOutput json: expected %d got %d", exitOK, code)
	}
	if code := writeReplayOutput(false, replayOutput{OK: false, Error: "bad"}, exitInvalidInput); code != exitInvalidInput {
		t.Fatalf("writeReplayOutput text err: expected %d got %d", exitInvalidInput, code)
	}
	if code := writeReplayOutput(false, replayOutput{
		OK:             false,
		RunID:          "run_demo",
		MissingResults: []string{"intent_2"},
		Warnings:       []string{"warn"},
	}, exitVerifyFailed); code != exitVerifyFailed {
		t.Fatalf("writeReplayOutput text fail: expected %d got %d", exitVerifyFailed, code)
	}

	if code := writeTraceVerifyOutput(true, traceVerifyOutput{OK: true, Path: "trace.json"}, exitOK); code != exitOK {
		t.Fatalf("writeTraceVerifyOutput json: expected %d got %d", exitOK, code)
	}
	if code := writeTraceVerifyOutput(false, traceVerifyOutput{OK: false, Error: "bad"}, exitInvalidInput); code != exitInvalidInput {
		t.Fatalf("writeTraceVerifyOutput text err: expected %d got %d", exitInvalidInput, code)
	}
	if code := writeTraceVerifyOutput(false, traceVerifyOutput{OK: false, Path: "trace.json", SignatureStatus: "failed"}, exitVerifyFailed); code != exitVerifyFailed {
		t.Fatalf("writeTraceVerifyOutput text fail: expected %d got %d", exitVerifyFailed, code)
	}

	if code := writeDoctorOutput(true, doctorOutput{OK: true, Summary: "ok"}, exitOK); code != exitOK {
		t.Fatalf("writeDoctorOutput json: expected %d got %d", exitOK, code)
	}
	if code := writeDoctorOutput(false, doctorOutput{
		OK:      true,
		Summary: "doctor summary",
		Checks: []doctor.Check{
			{Name: "check1", Status: "pass", Message: "ok"},
		},
	}, exitOK); code != exitOK {
		t.Fatalf("writeDoctorOutput text ok: expected %d got %d", exitOK, code)
	}
	if code := writeDoctorOutput(false, doctorOutput{Error: "bad"}, exitInvalidInput); code != exitInvalidInput {
		t.Fatalf("writeDoctorOutput text error: expected %d got %d", exitInvalidInput, code)
	}

	if code := writeVerifyOutput(true, verifyOutput{OK: true, Path: "runpack.zip"}, exitOK); code != exitOK {
		t.Fatalf("writeVerifyOutput json: expected %d got %d", exitOK, code)
	}
	if code := writeVerifyOutput(false, verifyOutput{OK: false, Error: "bad"}, exitInvalidInput); code != exitInvalidInput {
		t.Fatalf("writeVerifyOutput text err: expected %d got %d", exitInvalidInput, code)
	}
	if code := writeVerifyOutput(false, verifyOutput{
		OK:              false,
		Path:            "runpack.zip",
		MissingFiles:    []string{"manifest.json"},
		HashMismatches:  []runpack.HashMismatch{{Path: "run.json"}},
		SignatureStatus: "failed",
		SignatureErrors: []string{"bad sig"},
	}, exitVerifyFailed); code != exitVerifyFailed {
		t.Fatalf("writeVerifyOutput text fail: expected %d got %d", exitVerifyFailed, code)
	}

	printUsage()
	printApproveUsage()
	printDemoUsage()
	printDoctorUsage()
	printGateUsage()
	printGateEvalUsage()
	printPolicyUsage()
	printPolicyTestUsage()
	printTraceUsage()
	printTraceVerifyUsage()
	printRegressUsage()
	printRegressInitUsage()
	printRegressRunUsage()
	printRunUsage()
	printReplayUsage()
	printDiffUsage()
	printVerifyUsage()
}

func withWorkingDir(t *testing.T, path string) {
	t.Helper()
	current, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(path); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(current)
	})
}

func writePrivateKey(t *testing.T, path string) {
	t.Helper()
	kp, err := sign.GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString(kp.Private)
	if err := os.WriteFile(path, []byte(encoded+"\n"), 0o600); err != nil {
		t.Fatalf("write private key: %v", err)
	}
}

func writeIntentFixture(t *testing.T, path, toolName string) {
	t.Helper()
	intent := schemagate.IntentRequest{
		SchemaID:        "gait.gate.intent_request",
		SchemaVersion:   "1.0.0",
		CreatedAt:       time.Date(2026, time.February, 5, 0, 0, 0, 0, time.UTC),
		ProducerVersion: "test",
		ToolName:        toolName,
		Args: map[string]any{
			"path": "/tmp/out.txt",
		},
		Targets: []schemagate.IntentTarget{
			{Kind: "path", Value: "/tmp/out.txt"},
		},
		ArgProvenance: []schemagate.IntentArgProvenance{
			{ArgPath: "$.path", Source: "user"},
		},
		Context: schemagate.IntentContext{
			Identity:  "alice",
			Workspace: "/tmp",
			RiskClass: "high",
		},
	}
	raw, err := json.MarshalIndent(intent, "", "  ")
	if err != nil {
		t.Fatalf("marshal intent: %v", err)
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o600); err != nil {
		t.Fatalf("write intent: %v", err)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func repoRootFromPackageDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
