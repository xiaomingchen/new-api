package common

import (
	"bytes"
	"encoding/json"

	appcommon "github.com/QuantumNous/new-api/common"
)

func SanitizeResponsesInputStatuses(input json.RawMessage) (json.RawMessage, error) {
	trimmed := bytes.TrimSpace(input)
	if len(trimmed) == 0 {
		return input, nil
	}
	if trimmed[0] != '{' && trimmed[0] != '[' {
		return input, nil
	}

	var decoded any
	if err := appcommon.Unmarshal(input, &decoded); err != nil {
		return nil, err
	}

	if !sanitizeResponsesInputStatusValue(decoded) {
		return input, nil
	}

	sanitized, err := appcommon.Marshal(decoded)
	if err != nil {
		return nil, err
	}
	return sanitized, nil
}

func sanitizeResponsesInputStatusValue(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		changed := false
		if status, ok := typed["status"].(string); ok {
			normalized := normalizeResponsesInputStatus(status)
			if normalized != status {
				typed["status"] = normalized
				changed = true
			}
		}
		for _, child := range typed {
			if sanitizeResponsesInputStatusValue(child) {
				changed = true
			}
		}
		return changed
	case []any:
		changed := false
		for _, child := range typed {
			if sanitizeResponsesInputStatusValue(child) {
				changed = true
			}
		}
		return changed
	default:
		return false
	}
}

func normalizeResponsesInputStatus(status string) string {
	switch status {
	case "", "in_progress", "completed", "incomplete":
		return status
	default:
		return "incomplete"
	}
}
