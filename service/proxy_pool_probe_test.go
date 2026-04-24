package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/stretchr/testify/require"
)

func newForwardProxyServer(t *testing.T) *httptest.Server {
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

func newProbeTargetServer(t *testing.T, name string, delay time.Duration) (*httptest.Server, *int32) {
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

func newGeoLookupServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"success": true,
			"ip": "203.0.113.9",
			"city": "Hong Kong",
			"region": "Hong Kong",
			"country": "China",
			"country_code": "HK"
		}`)
	}))
}

func TestSummarizeProxyProbeResults(t *testing.T) {
	targets := []ProxyProbeTargetResult{
		{Name: "OpenAI", LatencyMs: 120, Success: true},
		{Name: "Claude", LatencyMs: 180, Success: true},
		{Name: "Gemini", LatencyMs: 250, Success: true},
		{Name: "Mistral", LatencyMs: 0, Success: false, Error: "timeout"},
	}

	successCount, failureCount, averageLatencyMs, qualityScore, quality := summarizeProxyProbeResults(targets)
	require.Equal(t, 3, successCount)
	require.Equal(t, 1, failureCount)
	require.Equal(t, int64(183), averageLatencyMs)
	require.Equal(t, 93, qualityScore)
	require.Equal(t, "一般", quality)
}

func TestProbeProxyPoolItems_UsesProxyAndGeoLookup(t *testing.T) {
	openaiServer, openaiHits := newProbeTargetServer(t, "OpenAI", 15*time.Millisecond)
	defer openaiServer.Close()
	claudeServer, claudeHits := newProbeTargetServer(t, "Claude", 20*time.Millisecond)
	defer claudeServer.Close()
	geminiServer, geminiHits := newProbeTargetServer(t, "Gemini", 10*time.Millisecond)
	defer geminiServer.Close()
	mistralServer, mistralHits := newProbeTargetServer(t, "Mistral", 12*time.Millisecond)
	defer mistralServer.Close()

	geoServer := newGeoLookupServer(t)
	defer geoServer.Close()

	proxyServer := newForwardProxyServer(t)
	defer proxyServer.Close()

	restore := SetProxyProbeTestConfig([]ProxyProbeTarget{
		{Name: "OpenAI", URL: openaiServer.URL},
		{Name: "Claude", URL: claudeServer.URL},
		{Name: "Gemini", URL: geminiServer.URL},
		{Name: "Mistral", URL: mistralServer.URL},
	}, geoServer.URL)
	t.Cleanup(restore)
	t.Cleanup(ResetProxyClientCache)

	results, err := ProbeProxyPoolItems(context.Background(), []system_setting.ProxyPoolItem{
		{
			Id:       "proxy-1",
			Name:     "Proxy 1",
			ProxyURL: proxyServer.URL,
		},
	})
	require.NoError(t, err)
	require.Len(t, results, 1)

	result := results[0]
	require.Equal(t, "203.0.113.9", result.ExitIP)
	require.Equal(t, "Hong Kong", result.City)
	require.Equal(t, "Hong Kong", result.Region)
	require.Equal(t, "China", result.Country)
	require.Equal(t, "HK", result.CountryCode)
	require.Equal(t, 4, result.SuccessCount)
	require.Equal(t, 0, result.FailureCount)
	require.Len(t, result.Targets, 4)
	require.NotEmpty(t, result.Quality)
	require.Greater(t, result.QualityScore, 0)

	targetHits := []struct {
		name string
		hits *int32
	}{
		{name: "OpenAI", hits: openaiHits},
		{name: "Claude", hits: claudeHits},
		{name: "Gemini", hits: geminiHits},
		{name: "Mistral", hits: mistralHits},
	}
	for index, target := range result.Targets {
		require.Equal(t, targetHits[index].name, target.Name)
		require.True(t, target.Success)
		require.Equal(t, targetHits[index].name, target.Name)
		require.Greater(t, target.LatencyMs, int64(0))
		require.Equal(t, 1, int(atomic.LoadInt32(targetHits[index].hits)))
	}
}
