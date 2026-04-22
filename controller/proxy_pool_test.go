package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type proxyPoolAPIResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func setupProxyPoolControllerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	gin.SetMode(gin.TestMode)
	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	common.RedisEnabled = false

	oldOptionMap := common.OptionMap
	common.OptionMap = make(map[string]string)
	t.Cleanup(func() {
		common.OptionMap = oldOptionMap
	})

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	model.DB = db
	model.LOG_DB = db

	if err := db.AutoMigrate(&model.Option{}, &model.Channel{}); err != nil {
		t.Fatalf("failed to migrate proxy pool tables: %v", err)
	}

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

func newProxyPoolRequestContext(t *testing.T, method string, target string, body any) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	payload, err := common.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, target, bytes.NewReader(payload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	return ctx, recorder
}

func decodeProxyPoolAPIResponse(t *testing.T, recorder *httptest.ResponseRecorder) proxyPoolAPIResponse {
	t.Helper()

	var response proxyPoolAPIResponse
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	return response
}

func TestUpdateProxyPoolsStoresRawArrayJSON(t *testing.T) {
	db := setupProxyPoolControllerTestDB(t)

	oldSetting := append([]system_setting.ProxyPoolItem(nil), system_setting.GetProxyPoolSetting().Proxies...)
	t.Cleanup(func() {
		system_setting.GetProxyPoolSetting().Proxies = oldSetting
	})

	ctx, recorder := newProxyPoolRequestContext(t, http.MethodPut, "/api/option/proxy_pools", ProxyPoolSaveRequest{
		Proxies: []system_setting.ProxyPoolItem{
			{
				Name:     "Pool A",
				ProxyURL: "http://proxy-a.local:8080",
			},
		},
	})

	UpdateProxyPools(ctx)

	response := decodeProxyPoolAPIResponse(t, recorder)
	require.True(t, response.Success, response.Message)

	var option model.Option
	require.NoError(t, db.First(&option, "key = ?", "proxy_pool_setting.proxies").Error)
	require.Equal(t, "array", common.GetJsonType(json.RawMessage(option.Value)))

	var items []system_setting.ProxyPoolItem
	require.NoError(t, common.UnmarshalJsonStr(option.Value, &items))
	require.Len(t, items, 1)
	require.Equal(t, "Pool A", items[0].Name)
	require.Equal(t, "http://proxy-a.local:8080", items[0].ProxyURL)
}
