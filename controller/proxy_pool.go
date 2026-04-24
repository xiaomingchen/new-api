package controller

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/system_setting"

	"github.com/gin-gonic/gin"
)

type ProxyPoolSaveRequest struct {
	Proxies []system_setting.ProxyPoolItem `json:"proxies"`
}

type ProxyPoolProbeRequest struct {
	Proxies []system_setting.ProxyPoolItem `json:"proxies"`
}

type ProxyPoolProbeResponse struct {
	Items []ProxyPoolResponseItem `json:"items"`
}

type ProxyPoolResponseItem struct {
	Id         string                    `json:"id"`
	Name       string                    `json:"name"`
	ProxyURL   string                    `json:"proxy_url"`
	UsageCount int                       `json:"usage_count"`
	Probe      *service.ProxyProbeResult `json:"probe,omitempty"`
}

type ProxyPoolListResponse struct {
	Items []ProxyPoolResponseItem `json:"items"`
}

func normalizeAndValidateProxyPoolItems(items []system_setting.ProxyPoolItem) ([]system_setting.ProxyPoolItem, error) {
	normalized := make([]system_setting.ProxyPoolItem, 0, len(items))
	seenIDs := make(map[string]struct{}, len(items))

	for _, item := range items {
		item.Id = strings.TrimSpace(item.Id)
		item.Name = strings.TrimSpace(item.Name)
		item.ProxyURL = strings.TrimSpace(item.ProxyURL)

		if item.Name == "" && item.ProxyURL == "" {
			continue
		}
		if item.Name == "" {
			return nil, fmt.Errorf("代理名称不能为空")
		}
		if item.ProxyURL == "" {
			return nil, fmt.Errorf("代理地址不能为空")
		}

		parsedURL, err := url.Parse(item.ProxyURL)
		if err != nil || parsedURL == nil || parsedURL.Host == "" {
			return nil, fmt.Errorf("代理地址必须是合法的 URL")
		}

		switch strings.ToLower(parsedURL.Scheme) {
		case "http", "https", "socks5", "socks5h":
		default:
			return nil, fmt.Errorf("代理地址只支持 http、https、socks5、socks5h")
		}

		if item.Id == "" {
			item.Id = "proxy-" + common.GetUUID()[:12]
		}
		for {
			if _, ok := seenIDs[item.Id]; !ok {
				break
			}
			item.Id = "proxy-" + common.GetUUID()[:12]
		}
		seenIDs[item.Id] = struct{}{}
		normalized = append(normalized, item)
	}

	return normalized, nil
}

func buildProxyPoolResponse(items []system_setting.ProxyPoolItem, probeResults []service.ProxyProbeResult) (ProxyPoolListResponse, error) {
	system_setting.NormalizeProxyPoolSetting()

	counts, err := model.CountProxyPoolUsage()
	if err != nil {
		return ProxyPoolListResponse{}, err
	}

	if len(items) == 0 {
		items = append([]system_setting.ProxyPoolItem(nil), system_setting.GetProxyPoolSetting().Proxies...)
	}

	responseItems := make([]ProxyPoolResponseItem, 0, len(items))
	for index, item := range items {
		responseItem := ProxyPoolResponseItem{
			Id:         item.Id,
			Name:       item.Name,
			ProxyURL:   item.ProxyURL,
			UsageCount: counts[item.Id],
		}
		if index < len(probeResults) {
			probe := probeResults[index]
			responseItem.Probe = &probe
		}
		responseItems = append(responseItems, responseItem)
	}

	return ProxyPoolListResponse{Items: responseItems}, nil
}

func GetProxyPools(c *gin.Context) {
	response, err := buildProxyPoolResponse(nil, nil)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    response,
	})
}

func UpdateProxyPools(c *gin.Context) {
	var req ProxyPoolSaveRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	normalized, err := normalizeAndValidateProxyPoolItems(req.Proxies)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	proxyPoolSetting := system_setting.ProxyPoolSetting{
		Proxies: normalized,
	}
	proxyPoolSettingBytes, err := common.Marshal(proxyPoolSetting.Proxies)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if err := model.UpdateOption("proxy_pool_setting.proxies", string(proxyPoolSettingBytes)); err != nil {
		common.ApiError(c, err)
		return
	}

	response, err := buildProxyPoolResponse(nil, nil)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "保存成功",
		"data":    response,
	})
}

func ProbeProxyPools(c *gin.Context) {
	var req ProxyPoolProbeRequest
	rawBody, err := c.GetRawData()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	items := append([]system_setting.ProxyPoolItem(nil), system_setting.GetProxyPoolSetting().Proxies...)
	trimmed := strings.TrimSpace(string(rawBody))
	if trimmed != "" {
		if err := common.Unmarshal(rawBody, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "无效的参数",
			})
			return
		}
		if len(req.Proxies) > 0 {
			items, err = normalizeAndValidateProxyPoolItems(req.Proxies)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": err.Error(),
				})
				return
			}
		}
	}

	probeResults, err := service.ProbeProxyPoolItems(c.Request.Context(), items)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	response, err := buildProxyPoolResponse(items, probeResults)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "探测完成",
		"data":    response,
	})
}
