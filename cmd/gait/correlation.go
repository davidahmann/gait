package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync/atomic"
)

var correlationIDValue atomic.Value

func init() {
	correlationIDValue.Store("")
}

func newCorrelationID(arguments []string) string {
	if len(arguments) == 0 {
		return "000000000000000000000000"
	}
	normalized := make([]string, 0, len(arguments))
	for _, arg := range arguments {
		normalized = append(normalized, strings.TrimSpace(arg))
	}
	sum := sha256.Sum256([]byte(strings.Join(normalized, "\x1f")))
	return hex.EncodeToString(sum[:12])
}

func setCurrentCorrelationID(correlationID string) {
	correlationIDValue.Store(strings.TrimSpace(correlationID))
}

func currentCorrelationID() string {
	value, _ := correlationIDValue.Load().(string)
	return strings.TrimSpace(value)
}
