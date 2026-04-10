package common

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeResponsesInputStatusesNormalizesUnsupportedStatuses(t *testing.T) {
	input := json.RawMessage(`[
		{"type":"message","status":"failed","content":[{"type":"output_text","status":"cancelled","text":"oops"}]},
		{"type":"message","status":"completed"},
		{"type":"message","status":"in_progress"}
	]`)

	got, err := SanitizeResponsesInputStatuses(input)
	require.NoError(t, err)

	assert.JSONEq(t, `[
		{"type":"message","status":"incomplete","content":[{"type":"output_text","status":"incomplete","text":"oops"}]},
		{"type":"message","status":"completed"},
		{"type":"message","status":"in_progress"}
	]`, string(got))
}
