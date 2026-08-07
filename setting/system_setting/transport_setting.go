package system_setting

import "github.com/QuantumNous/new-api/setting/config"

// TransportSetting controls the HTTP connection pool for relay (outbound) traffic.
// These values are applied to every http.Transport created by the relay proxy.
// When zero, the corresponding common package default (env var) is used.
type TransportSetting struct {
	MaxIdleConns        int `json:"max_idle_conns"`         // 0 = use env default (500)
	MaxIdleConnsPerHost int `json:"max_idle_conns_per_host"` // 0 = use env default (100)
	IdleConnTimeout     int `json:"idle_conn_timeout"`       // seconds; 0 = use env default (90)
}

var defaultTransportSetting = TransportSetting{
	MaxIdleConns:        0,
	MaxIdleConnsPerHost: 0,
	IdleConnTimeout:     0,
}

func init() {
	config.GlobalConfig.Register("transport_setting", &defaultTransportSetting)
}

// GetTransportSetting returns the current transport pool settings.
func GetTransportSetting() *TransportSetting {
	return &defaultTransportSetting
}