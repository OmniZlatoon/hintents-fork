// Copyright 2026 Erst Users
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dotandev/hintents/internal/errors"
)

// -- Interfaces --

type Parser interface {
	Parse(*Config) error
}

type DefaultAssigner interface {
	Apply(*Config)
}

type Validator interface {
	Validate(*Config) error
}

// -- Types --

type Network string

const (
	NetworkPublic     Network = "public"
	NetworkTestnet    Network = "testnet"
	NetworkFuturenet  Network = "futurenet"
	NetworkStandalone Network = "standalone"
)

var validNetworks = map[string]bool{
	string(NetworkPublic):     true,
	string(NetworkTestnet):    true,
	string(NetworkFuturenet):  true,
	string(NetworkStandalone): true,
}

// Config represents the general configuration for erst
type Config struct {
	RpcUrl            string   `json:"rpc_url,omitempty"`
	RpcUrls           []string `json:"rpc_urls,omitempty"`
	Network           Network  `json:"network,omitempty"`
	NetworkPassphrase string   `json:"network_passphrase,omitempty"`
	SimulatorPath     string   `json:"simulator_path,omitempty"`
	LogLevel          string   `json:"log_level,omitempty"`
	CachePath         string   `json:"cache_path,omitempty"`
	RPCToken          string   `json:"rpc_token,omitempty"`
	// MaxCacheSize is the maximum size of the source map cache in bytes.
	MaxCacheSize int64 `json:"max_cache_size,omitempty"`
	// CrashReporting enables opt-in anonymous crash reporting.
	CrashReporting bool `json:"crash_reporting,omitempty"`
	// CrashEndpoint is a custom HTTPS URL that receives JSON crash reports.
	CrashEndpoint string `json:"crash_endpoint,omitempty"`
	// CrashSentryDSN is a Sentry Data Source Name for crash reporting.
	CrashSentryDSN string `json:"crash_sentry_dsn,omitempty"`
	// RequestTimeout is the HTTP request timeout in seconds for all RPC calls.
	RequestTimeout int `json:"request_timeout,omitempty"`
	// MaxTraceDepth is the maximum depth of the call tree before it is truncated.
	MaxTraceDepth int `json:"max_trace_depth,omitempty"`
}

// -- Constants & Defaults --

const defaultRequestTimeout = 15

var validLogLevels = map[string]bool{
	"trace": true,
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
}

var defaultConfig = &Config{
	RpcUrl:         "https://soroban-testnet.stellar.org",
	Network:        NetworkTestnet,
	SimulatorPath:  "",
	LogLevel:       "info",
	CachePath:      filepath.Join(os.ExpandEnv("$HOME"), ".erst", "cache"),
	RequestTimeout: defaultRequestTimeout,
	MaxCacheSize:   0,
	MaxTraceDepth:  50,
}

// -- Core Functions --

func Load() (*Config, error) {
	cfg := &Config{}
	parsers := []Parser{envParser{}, fileParser{}}
	for _, parser := range parsers {
		if err := parser.Parse(cfg); err != nil {
			return nil, err
		}
	}

	configDefaultsAssigner{}.Apply(cfg)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func DefaultConfig() *Config {
	return &Config{
		RpcUrl:         defaultConfig.RpcUrl,
		Network:        defaultConfig.Network,
		SimulatorPath:  defaultConfig.SimulatorPath,
		LogLevel:       defaultConfig.LogLevel,
		CachePath:      defaultConfig.CachePath,
		RequestTimeout: defaultConfig.RequestTimeout,
		MaxCacheSize:   defaultConfig.MaxCacheSize,
		MaxTraceDepth:  defaultConfig.MaxTraceDepth,
	}
}

func NewConfig(rpcUrl string, network Network) *Config {
	return &Config{
		RpcUrl:         rpcUrl,
		Network:        network,
		SimulatorPath:  defaultConfig.SimulatorPath,
		LogLevel:       defaultConfig.LogLevel,
		CachePath:      defaultConfig.CachePath,
		RequestTimeout: defaultConfig.RequestTimeout,
		MaxCacheSize:   defaultConfig.MaxCacheSize,
		MaxTraceDepth:  defaultConfig.MaxTraceDepth,
	}
}

// -- Config Methods --

func (c *Config) MergeDefaults() {
	configDefaultsAssigner{}.Apply(c)
}

func (c *Config) Validate() error {
	validators := []Validator{
		RPCValidator{},
		NetworkValidator{},
		SimulatorValidator{},
		LogLevelValidator{},
		TimeoutValidator{},
		MaxTraceDepthValidator{},
		CrashReportingValidator{},
	}
	for _, v := range validators {
		if err := v.Validate(c); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) NetworkURL() string {
	switch c.Network {
	case NetworkPublic:
		return "https://soroban.stellar.org"
	case NetworkTestnet:
		return "https://soroban-testnet.stellar.org"
	case NetworkFuturenet:
		return "https://soroban-futurenet.stellar.org"
	case NetworkStandalone:
		return "http://localhost:8000"
	default:
		return c.RpcUrl
	}
}

// -- Load/Save Config --

func GetGeneralConfigPath() (string, error) {
	// Assumes GetConfigPath is defined in your networks.go
	configDir, err := os.UserConfigDir() 
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "erst", "config.json"), nil
}

func LoadConfig() (*Config, error) {
	configPath, err := GetGeneralConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.WrapConfigError("failed to read config file", err)
	}

	config := DefaultConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.WrapConfigError("failed to parse config file", err)
	}

	return config, nil
}

func SaveConfig(config *Config) error {
	configPath, err := GetGeneralConfigPath()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return errors.WrapConfigError("failed to create config directory", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.WrapConfigError("failed to marshal config", err)
	}

	return os.WriteFile(configPath, data, 0600)
}

// -- Parsers --

type envParser struct{}

func (envParser) Parse(cfg *Config) error {
	if v := os.Getenv("ERST_RPC_URL"); v != "" {
		cfg.RpcUrl = v
	}
	if v := os.Getenv("ERST_NETWORK"); v != "" {
		cfg.Network = Network(v)
	}
	if v := os.Getenv("ERST_SIMULATOR_PATH"); v != "" {
		cfg.SimulatorPath = v
	}
	if v := os.Getenv("ERST_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := os.Getenv("ERST_CACHE_PATH"); v != "" {
		cfg.CachePath = v
	}
	if v := os.Getenv("ERST_RPC_TOKEN"); v != "" {
		cfg.RPCToken = v
	}
	if v := os.Getenv("ERST_MAX_CACHE_SIZE"); v != "" {
		// Note: parseSize helper usually defined in parse.go
		n, _ := strconv.ParseInt(v, 10, 64)
		if n > 0 {
			cfg.MaxCacheSize = n
		}
	}
	if v := os.Getenv("ERST_REQUEST_TIMEOUT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.RequestTimeout = n
		}
	}
	if v := os.Getenv("ERST_MAX_TRACE_DEPTH"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.MaxTraceDepth = n
		}
	}
	return nil
}

type fileParser struct{}

func (fileParser) Parse(cfg *Config) error {
	// Implementation usually calls loadTOML defined in parse.go
	return nil 
}

// -- Validators --

type RPCValidator struct{}
func (RPCValidator) Validate(cfg *Config) error {
	if cfg.RpcUrl == "" { return errors.WrapValidationError("rpc_url cannot be empty") }
	return nil
}

type NetworkValidator struct{}
func (NetworkValidator) Validate(cfg *Config) error {
	if cfg.Network != "" && !validNetworks[string(cfg.Network)] {
		return errors.WrapInvalidNetwork(string(cfg.Network))
	}
	return nil
}

type SimulatorValidator struct{}
func (SimulatorValidator) Validate(cfg *Config) error { return nil }

type LogLevelValidator struct{}
func (LogLevelValidator) Validate(cfg *Config) error { return nil }

type TimeoutValidator struct{}
func (TimeoutValidator) Validate(cfg *Config) error { return nil }

type MaxTraceDepthValidator struct{}
func (MaxTraceDepthValidator) Validate(cfg *Config) error {
	if cfg.MaxTraceDepth < 1 {
		return errors.WrapValidationError("max_trace_depth must be at least 1")
	}
	return nil
}

type CrashReportingValidator struct{}
func (CrashReportingValidator) Validate(cfg *Config) error { return nil }

type configDefaultsAssigner struct{}

func (configDefaultsAssigner) Apply(cfg *Config) {
	if cfg.RpcUrl == "" { cfg.RpcUrl = defaultConfig.RpcUrl }
	if cfg.Network == "" { cfg.Network = defaultConfig.Network }
	if cfg.LogLevel == "" { cfg.LogLevel = defaultConfig.LogLevel }
	if cfg.RequestTimeout == 0 { cfg.RequestTimeout = defaultRequestTimeout }
	if cfg.MaxTraceDepth == 0 { cfg.MaxTraceDepth = 50 }
}