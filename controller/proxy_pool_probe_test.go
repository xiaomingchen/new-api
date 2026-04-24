package controller

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/stretchr/testify/require"
)

func newControllerProbeTargetServer(t *testing.T, name string, delay time.Duration) (*httptest.Server, *int32) {
	t.Helper()

	var hits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		time.Sleep(delay)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"name":%q,"ok":true}`, name)
	}))

	return server, &hits
}

func newControllerGeoLookupServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"success": true,
			"ip": "198.51.100.25",
			"city": "Singapore",
			"region": "Singapore",
			"country": "Singapore",
			"country_code": "SG"
		}`)
	}))
}

func newControllerForwardProxyServer(t *testing.T) *httptest.Server {
	t.Helper()

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodConnect {
			http.Error(w, "CONNECT is not supported in this test proxy", http.StatusMethodNotAllowed)
			return
		}

		outbound := r.Clone(r.Context())
		outbound.RequestURI = ""
		outbound.URL = r.URL
		outbound.Host = r.URL.Host
		outbound.Header.Del("Proxy-Connection")

		resp, err := transport.RoundTrip(outbound)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	}))
}

func TestProbeProxyPoolsReturnsProbeData(t *testing.T) {
	setupProxyPoolControllerTestDB(t)

	openaiServer, openaiHits := newControllerProbeTargetServer(t, "OpenAI", 12*time.Millisecond)
	defer openaiServer.Close()
	claudeServer, claudeHits := newControllerProbeTargetServer(t, "Claude", 15*time.Millisecond)
	defer claudeServer.Close()
	geminiServer, geminiHits := newControllerProbeTargetServer(t, "Gemini", 10*time.Millisecond)
	defer geminiServer.Close()
	mistralServer, mistralHits := newControllerProbeTargetServer(t, "Mistral", 8*time.Millisecond)
	defer mistralServer.Close()

	geoServer := newControllerGeoLookupServer(t)
	defer geoServer.Close()

	proxyServer := newControllerForwardProxyServer(t)
	defer proxyServer.Close()

	restore := service.SetProxyProbeTestConfig([]service.ProxyProbeTarget{
		{Name: "OpenAI", URL: openaiServer.URL},
		{Name: "Claude", URL: claudeServer.URL},
		{Name: "Gemini", URL: geminiServer.URL},
		{Name: "Mistral", URL: mistralServer.URL},
	}, geoServer.URL)
	t.Cleanup(restore)
	t.Cleanup(service.ResetProxyClientCache)

	ctx, recorder := newProxyPoolRequestContext(t, http.MethodPost, "/api/channel/proxy_pools/probe", ProxyPoolProbeRequest{
		Proxies: []system_setting.ProxyPoolItem{
			{
				Name:     "Proxy 1",
				ProxyURL: proxyServer.URL,
			},
		},
	})

	ProbeProxyPools(ctx)

	response := decodeProxyPoolAPIResponse(t, recorder)
	require.True(t, response.Success, response.Message)

	var data ProxyPoolProbeResponse
	require.NoError(t, common.Unmarshal(response.Data, &data))
	require.Len(t, data.Items, 1)

	item := data.Items[0]
	require.Equal(t, "Proxy 1", item.Name)
	require.NotNil(t, item.Probe)
	require.Equal(t, "Singapore", item.Probe.City)
	require.Equal(t, "SG", item.Probe.CountryCode)
	require.Equal(t, 4, item.Probe.SuccessCount)
	require.Equal(t, 0, item.Probe.FailureCount)
	require.Len(t, item.Probe.Targets, 4)
	require.Equal(t, 1, int(atomic.LoadInt32(openaiHits)))
	require.Equal(t, 1, int(atomic.LoadInt32(claudeHits)))
	require.Equal(t, 1, int(atomic.LoadInt32(geminiHits)))
	require.Equal(t, 1, int(atomic.LoadInt32(mistralHits)))
}
