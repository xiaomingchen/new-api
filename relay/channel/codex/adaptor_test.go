package codex

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertOpenAIResponsesRequestDefaultsInstructionsAndStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	adaptor := &Adaptor{}
	info := &relaycommon.RelayInfo{
		RelayMode: relayconstant.RelayModeResponses,
	}
	request := dto.OpenAIResponsesRequest{
		Model: "codex-mini-latest",
		Input: json.RawMessage(`[
			{"type":"message","status":"failed","content":[{"type":"output_text","status":"cancelled","text":"oops"}]},
			{"type":"message","status":"completed"},
			{"type":"message","status":"in_progress"}
		]`),
	}

	convertedAny, err := adaptor.ConvertOpenAIResponsesRequest(ctx, info, request)
	require.NoError(t, err)

	converted, ok := convertedAny.(dto.OpenAIResponsesRequest)
	require.True(t, ok)

	assert.JSONEq(t, string(request.Input), string(converted.Input))
	assert.JSONEq(t, `false`, string(converted.Store))
	assert.JSONEq(t, `""`, string(converted.Instructions))
}
