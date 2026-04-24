package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"golang.org/x/sync/errgroup"
)

type ProxyProbeTarget struct {
	Name string
	URL  string
}

type ProxyProbeTargetResult struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	LatencyMs  int64  `json:"latency_ms"`
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
}

type ProxyProbeResult struct {
	ExitIP           string                   `json:"exit_ip"`
	City             string                   `json:"city"`
	Region           string                   `json:"region"`
	Country          string                   `json:"country"`
	CountryCode      string                   `json:"country_code"`
	AverageLatencyMs int64                    `json:"average_latency_ms"`
	Quality          string                   `json:"quality"`
	QualityScore     int                      `json:"quality_score"`
	SuccessCount     int                      `json:"success_count"`
	FailureCount     int                      `json:"failure_count"`
	ProbedAt         int64                    `json:"probed_at"`
	Targets          []ProxyProbeTargetResult `json:"targets"`
	Error            string                   `json:"error,omitempty"`
}

type proxyGeoLookupResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	IP          string `json:"ip"`
	City        string `json:"city"`
	Region      string `json:"region"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
}

var proxyProbeTargets = []ProxyProbeTarget{
	{Name: "OpenAI", URL: "https://api.openai.com/v1/models"},
	{Name: "Claude", URL: "https://api.anthropic.com/v1/models"},
	{Name: "Gemini", URL: "https://generativelanguage.googleapis.com/v1beta/models"},
	{Name: "Mistral", URL: "https://api.mistral.ai/v1/models"},
}

var proxyGeoLookupURL = "https://ipwho.is/"

const (
	proxyProbeRequestTimeout    = 8 * time.Second
	proxyProbeTargetConcurrency = 4
	proxyProbeProxyConcurrency  = 6
	proxyProbeUserAgent         = "new-api-proxy-pool-probe/1.0"
)

// SetProxyProbeTestConfig overrides probe targets for tests and returns a restore function.
func SetProxyProbeTestConfig(targets []ProxyProbeTarget, geoLookupURL string) func() {
	oldTargets := append([]ProxyProbeTarget(nil), proxyProbeTargets...)
	oldGeoLookupURL := proxyGeoLookupURL

	proxyProbeTargets = append([]ProxyProbeTarget(nil), targets...)
	proxyGeoLookupURL = geoLookupURL

	return func() {
		proxyProbeTargets = oldTargets
		proxyGeoLookupURL = oldGeoLookupURL
	}
}

func ProbeProxyPoolItems(ctx context.Context, proxies []system_setting.ProxyPoolItem) ([]ProxyProbeResult, error) {
	results := make([]ProxyProbeResult, len(proxies))
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(proxyProbeProxyConcurrency)

	for i, proxyItem := range proxies {
		i := i
		proxyItem := proxyItem

		group.Go(func() error {
			results[i] = probeSingleProxy(groupCtx, proxyItem)
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

func probeSingleProxy(ctx context.Context, proxyItem system_setting.ProxyPoolItem) ProxyProbeResult {
	result := ProxyProbeResult{
		Targets:  make([]ProxyProbeTargetResult, len(proxyProbeTargets)),
		ProbedAt: time.Now().Unix(),
	}

	client, err := GetHttpClientWithProxy(strings.TrimSpace(proxyItem.ProxyURL))
	if err != nil {
		result.Quality = "不可用"
		result.Error = err.Error()
		result.FailureCount = len(proxyProbeTargets)
		return result
	}

	var geoResult proxyGeoLookupResponse
	geoErr := probeJSON(ctx, client, proxyGeoLookupURL, &geoResult)
	if geoErr == nil && (geoResult.Success || strings.TrimSpace(geoResult.IP) != "") {
		result.ExitIP = strings.TrimSpace(geoResult.IP)
		result.City = strings.TrimSpace(geoResult.City)
		result.Region = strings.TrimSpace(geoResult.Region)
		result.Country = strings.TrimSpace(geoResult.Country)
		result.CountryCode = strings.TrimSpace(geoResult.CountryCode)
	}

	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(proxyProbeTargetConcurrency)

	for i, target := range proxyProbeTargets {
		i := i
		target := target
		group.Go(func() error {
			result.Targets[i] = probeTarget(groupCtx, client, target)
			return nil
		})
	}

	_ = group.Wait()
	result.SuccessCount, result.FailureCount, result.AverageLatencyMs, result.QualityScore, result.Quality = summarizeProxyProbeResults(result.Targets)
	if geoErr != nil && result.Error == "" {
		result.Error = geoErr.Error()
	}
	return result
}

func probeTarget(ctx context.Context, client *http.Client, target ProxyProbeTarget) ProxyProbeTargetResult {
	start := time.Now()
	reqCtx, cancel := context.WithTimeout(ctx, proxyProbeRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, target.URL, nil)
	if err != nil {
		return ProxyProbeTargetResult{
			Name:  target.Name,
			URL:   target.URL,
			Error: err.Error(),
		}
	}
	req.Header.Set("User-Agent", proxyProbeUserAgent)
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	latencyMs := time.Since(start).Milliseconds()
	if err != nil {
		return ProxyProbeTargetResult{
			Name:      target.Name,
			URL:       target.URL,
			LatencyMs: latencyMs,
			Success:   false,
			Error:     err.Error(),
		}
	}
	defer CloseResponseBodyGracefully(resp)
	_, _ = io.Copy(io.Discard, resp.Body)

	return ProxyProbeTargetResult{
		Name:       target.Name,
		URL:        target.URL,
		StatusCode: resp.StatusCode,
		LatencyMs:  latencyMs,
		Success:    true,
	}
}

func probeJSON(ctx context.Context, client *http.Client, targetURL string, dst any) error {
	reqCtx, cancel := context.WithTimeout(ctx, proxyProbeRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, targetURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", proxyProbeUserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer CloseResponseBodyGracefully(resp)
	return common.DecodeJson(resp.Body, dst)
}

func summarizeProxyProbeResults(targets []ProxyProbeTargetResult) (successCount int, failureCount int, averageLatencyMs int64, qualityScore int, quality string) {
	if len(targets) == 0 {
		return 0, 0, 0, 0, "未知"
	}

	var latencySum int64
	var latencyCount int64

	for _, target := range targets {
		if target.Success {
			successCount++
			if target.LatencyMs > 0 {
				latencySum += target.LatencyMs
				latencyCount++
			}
		}
	}

	failureCount = len(targets) - successCount
	if latencyCount > 0 {
		averageLatencyMs = latencySum / latencyCount
	}

	if successCount == 0 {
		return successCount, failureCount, averageLatencyMs, 0, "不可用"
	}

	qualityScore = 100
	if failureCount > 0 {
		qualityScore -= failureCount * 12
	}
	switch {
	case averageLatencyMs <= 500:
		qualityScore += 5
	case averageLatencyMs <= 1000:
		qualityScore -= 5
	case averageLatencyMs <= 2000:
		qualityScore -= 15
	default:
		qualityScore -= 30
	}
	if qualityScore < 0 {
		qualityScore = 0
	}
	if qualityScore > 100 {
		qualityScore = 100
	}

	switch {
	case failureCount == 0 && averageLatencyMs <= 500:
		quality = "优秀"
	case failureCount == 0 && averageLatencyMs <= 1000:
		quality = "良好"
	case failureCount <= 1 && averageLatencyMs <= 2000:
		quality = "一般"
	default:
		quality = "较差"
	}

	return successCount, failureCount, averageLatencyMs, qualityScore, quality
}
