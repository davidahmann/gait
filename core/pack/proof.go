package pack

import (
	"bufio"
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	proofrecord "github.com/Clyra-AI/proof/core/record"
	proofschema "github.com/Clyra-AI/proof/core/schema"
	sign "github.com/Clyra-AI/proof/signing"

	"github.com/Clyra-AI/gait/core/zipx"
)

const proofRecordsFileName = "proof_records.jsonl"

func buildProofRecordsJSONL(
	packType string,
	sourceRef string,
	producerVersion string,
	createdAt time.Time,
	files []zipx.File,
	signingPrivateKey ed25519.PrivateKey,
) ([]byte, error) {
	candidates := make([]zipx.File, 0, len(files))
	for _, file := range files {
		if strings.TrimSpace(file.Path) == "" || file.Path == proofRecordsFileName {
			continue
		}
		candidates = append(candidates, file)
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].Path < candidates[j].Path })

	var lines [][]byte
	previousHash := ""
	for _, file := range candidates {
		recordItem, err := proofrecord.New(proofrecord.RecordOpts{
			RecordVersion: "1.0",
			Timestamp:     createdAt.UTC(),
			Source:        "gait.pack",
			SourceProduct: "gait",
			AgentID:       "gait",
			Type:          "tool_invocation",
			Event: map[string]any{
				"pack_type":         packType,
				"source_ref":        sourceRef,
				"producer_version":  producerVersion,
				"entry_path":        file.Path,
				"entry_type":        detectEntryType(file.Path),
				"entry_sha256":      sha256Hex(file.Data),
				"entry_size_bytes":  len(file.Data),
				"entry_permissions": int(file.Mode),
			},
			Controls: proofrecord.Controls{
				PermissionsEnforced: false,
			},
			Metadata: map[string]any{
				"artifact_kind": "gait.pack.entry",
			},
		})
		if err != nil {
			return nil, fmt.Errorf("create proof record for %s: %w", file.Path, err)
		}
		if previousHash != "" {
			recordItem.Integrity.PreviousRecordHash = previousHash
			recordHash, hashErr := proofrecord.ComputeHash(recordItem)
			if hashErr != nil {
				return nil, fmt.Errorf("compute chained proof hash for %s: %w", file.Path, hashErr)
			}
			recordItem.Integrity.RecordHash = recordHash
		}
		if len(signingPrivateKey) > 0 {
			recordSig := sign.SignBytes(signingPrivateKey, []byte(recordItem.Integrity.RecordHash))
			recordItem.Integrity.SigningKeyID = recordSig.KeyID
			recordItem.Integrity.Signature = "base64:" + recordSig.Sig
		}
		canonicalLine, err := canonicalJSON(recordItem)
		if err != nil {
			return nil, fmt.Errorf("encode proof record for %s: %w", file.Path, err)
		}
		if err := proofschema.ValidateRecord(canonicalLine, recordItem.RecordType); err != nil {
			return nil, fmt.Errorf("validate proof record for %s: %w", file.Path, err)
		}
		lines = append(lines, canonicalLine)
		previousHash = recordItem.Integrity.RecordHash
	}
	return bytes.Join(lines, []byte{'\n'}), nil
}

func verifyProofRecordsJSONL(raw []byte, options VerifyOptions) (int, []string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	signatureErrors := make([]string, 0)
	previousHash := ""
	verified := 0
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var recordItem proofrecord.Record
		if err := json.Unmarshal(line, &recordItem); err != nil {
			return verified, nil, fmt.Errorf("line %d parse record: %w", lineNo, err)
		}
		if err := proofrecord.Validate(&recordItem); err != nil {
			return verified, nil, fmt.Errorf("line %d validate record: %w", lineNo, err)
		}
		if err := proofschema.ValidateRecord(line, recordItem.RecordType); err != nil {
			return verified, nil, fmt.Errorf("line %d validate schema: %w", lineNo, err)
		}
		expectedHash, err := proofrecord.ComputeHash(&recordItem)
		if err != nil {
			return verified, nil, fmt.Errorf("line %d compute record hash: %w", lineNo, err)
		}
		if expectedHash != recordItem.Integrity.RecordHash {
			return verified, nil, fmt.Errorf(
				"line %d record hash mismatch: expected %s got %s",
				lineNo,
				expectedHash,
				recordItem.Integrity.RecordHash,
			)
		}

		if previousHash == "" {
			if recordItem.Integrity.PreviousRecordHash != "" {
				return verified, nil, fmt.Errorf("line %d expected empty previous_record_hash", lineNo)
			}
		} else if recordItem.Integrity.PreviousRecordHash != previousHash {
			return verified, nil, fmt.Errorf(
				"line %d previous_record_hash mismatch: expected %s got %s",
				lineNo,
				previousHash,
				recordItem.Integrity.PreviousRecordHash,
			)
		}
		previousHash = recordItem.Integrity.RecordHash

		if options.RequireSignature && len(options.PublicKey) > 0 {
			if strings.TrimSpace(recordItem.Integrity.Signature) == "" {
				signatureErrors = append(signatureErrors, fmt.Sprintf("proof_records line %d missing signature", lineNo))
				verified++
				continue
			}
			if strings.TrimSpace(recordItem.Integrity.SigningKeyID) == "" {
				signatureErrors = append(signatureErrors, fmt.Sprintf("proof_records line %d missing signing_key_id", lineNo))
				verified++
				continue
			}
			signatureValue := strings.TrimSpace(recordItem.Integrity.Signature)
			if !strings.HasPrefix(signatureValue, "base64:") {
				signatureErrors = append(signatureErrors, fmt.Sprintf("proof_records line %d unsupported signature prefix", lineNo))
				verified++
				continue
			}
			ok, verifyErr := sign.VerifyBytes(options.PublicKey, sign.Signature{
				Alg:   sign.AlgEd25519,
				KeyID: recordItem.Integrity.SigningKeyID,
				Sig:   strings.TrimPrefix(signatureValue, "base64:"),
			}, []byte(recordItem.Integrity.RecordHash))
			if verifyErr != nil {
				signatureErrors = append(signatureErrors, fmt.Sprintf("proof_records line %d signature verification: %v", lineNo, verifyErr))
				verified++
				continue
			}
			if !ok {
				signatureErrors = append(signatureErrors, fmt.Sprintf("proof_records line %d signature verification failed", lineNo))
				verified++
				continue
			}
		}
		verified++
	}
	if err := scanner.Err(); err != nil {
		return verified, nil, fmt.Errorf("read proof records: %w", err)
	}
	return verified, signatureErrors, nil
}

func hasHashMismatch(hashMismatches []HashMismatch, path string) bool {
	for _, item := range hashMismatches {
		if item.Path == path {
			return true
		}
	}
	return false
}
