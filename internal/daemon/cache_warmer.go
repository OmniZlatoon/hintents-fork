// Copyright 2026 Erst Users
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"context"
	"time"

	"github.com/dotandev/hintents/internal/logger"
	"github.com/dotandev/hintents/internal/rpc"
)

// defaultHotContracts defines a list of high-traffic contracts that should be
// preemptively cached to reduce cold-start latency.
var defaultHotContracts = map[string][]string{
	"public": {
		"CAS3J7GYCCX3TP377666F6A6X6X6X6X6X6X6X6X6X6X6X6X6X6X6X6X", // Native XLM Wrapper
		"CCW67ZTM6S7C3S64Z6S3SLZ6S7C3S64Z6S3SLZ6S7C3S64Z6S3SLZ6S", // USDC (Placeholder)
	},
	"testnet": {
		"CDLZSTBBZUTG4F32CH7YJYSZQMHSFVP6XQ6YUKMCTM4S3YFEW6M6CSRE", // Community Test Contract
	},
}

// CacheWarmer is a background worker that preemptively fetches ledger entries
// for hot contracts and their WASM dependencies.
type CacheWarmer struct {
	client   *rpc.Client
	interval time.Duration
}

// NewCacheWarmer creates a new instance of the CacheWarmer.
func NewCacheWarmer(client *rpc.Client) *CacheWarmer {
	return &CacheWarmer{
		client:   client,
		interval: 10 * time.Minute, // Default interval
	}
}

// Start launches the warming loop in a blocking manner.
func (cw *CacheWarmer) Start(ctx context.Context) {
	ticker := time.NewTicker(cw.interval)
	defer ticker.Stop()

	logger.Logger.Info("Cache warmer background worker started", "interval", cw.interval)

	// Run initial warm
	cw.warm(ctx)

	for {
		select {
		case <-ctx.Done():
			logger.Logger.Info("Cache warmer shutting down")
			return
		case <-ticker.C:
			cw.warm(ctx)
		}
	}
}

// warm performs a single warming cycle.
func (cw *CacheWarmer) warm(ctx context.Context) {
	network := string(cw.client.Network)
	contracts := defaultHotContracts[network]
	if len(contracts) == 0 {
		return
	}

	logger.Logger.Debug("Starting cache warming cycle", "network", network, "hot_contracts", len(contracts))

	// Stage 1: Fetch Contract Instances
	var instanceKeys []string
	for _, id := range contracts {
		key, err := rpc.NewContractInstanceKey(id)
		if err != nil {
			logger.Logger.Warn("Invalid hot contract ID", "id", id, "error", err)
			continue
		}
		xdrKey, err := rpc.EncodeLedgerKey(key)
		if err != nil {
			continue
		}
		instanceKeys = append(instanceKeys, xdrKey)
	}

	if len(instanceKeys) == 0 {
		return
	}

	// GetLedgerEntries handles caching internally via its Client
	entries, err := cw.client.GetLedgerEntries(ctx, instanceKeys)
	if err != nil {
		logger.Logger.Error("Cache warmer failed to fetch instances", "error", err)
		return
	}

	// Stage 2: Resolve and Fetch WASM dependencies
	var codeKeys []string
	for _, entryXDR := range entries {
		wasmHash, ok := rpc.ParseWasmHashFromInstance(entryXDR)
		if !ok {
			continue
		}
		key, _ := rpc.NewContractCodeKey(wasmHash)
		xdrKey, _ := rpc.EncodeLedgerKey(key)
		codeKeys = append(codeKeys, xdrKey)
	}

	if len(codeKeys) > 0 {
		_, err = cw.client.GetLedgerEntries(ctx, codeKeys)
		if err != nil {
			logger.Logger.Error("Cache warmer failed to fetch WASM bytecode", "error", err)
			return
		}
	}

	logger.Logger.Info("Cache warmer cycle completed", 
		"instances_cached", len(entries), 
		"code_cached", len(codeKeys))
}
