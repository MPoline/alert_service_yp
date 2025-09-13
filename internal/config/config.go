package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

type Duration time.Duration

type ServerConfig struct {
	Address         string   `json:"address"`
	StoreInterval   Duration `json:"store_interval"`
	FileStoragePath string   `json:"file_storage_path"`
	Restore         bool     `json:"restore"`
	DatabaseDSN     string   `json:"database_dsn"`
	Key             string   `json:"key"`
	CryptoKey       string   `json:"crypto_key"`
	ConfigFile      string   `json:"-"`
	TrustedSubnet   string   `json:"trusted_subnet"`
}

type AgentConfig struct {
	Address        string   `json:"address"`
	ReportInterval Duration `json:"report_interval"`
	PollInterval   Duration `json:"poll_interval"`
	CryptoKey      string   `json:"crypto_key"`
	Key            string   `json:"key"`
	ConfigFile     string   `json:"-"`
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case string:
		dur, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(dur)
	case float64:
		*d = Duration(time.Duration(value) * time.Second)
	default:
		return fmt.Errorf("invalid duration format: %v", v)
	}
	return nil
}

func (d Duration) ToDuration() time.Duration {
	return time.Duration(d)
}

func LoadServerConfig(configPath string) (*ServerConfig, error) {
	if configPath == "" {
		return nil, nil
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ServerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func LoadAgentConfig(configPath string) (*AgentConfig, error) {
	if configPath == "" {
		return nil, nil
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config AgentConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func MaskSensitive(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + "****" + value[len(value)-2:]
}

func IsIPInTrustedSubnet(ipStr, trustedSubnet string) (bool, error) {
	if trustedSubnet == "" {
		return true, nil
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	_, ipNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		return false, fmt.Errorf("invalid CIDR format: %s", trustedSubnet)
	}

	return ipNet.Contains(ip), nil
}
