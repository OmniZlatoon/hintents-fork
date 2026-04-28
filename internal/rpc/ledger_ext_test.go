// Copyright 2026 Erst Users
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"encoding/base64"
	"testing"

	"github.com/stellar/go-stellar-sdk/xdr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContractInstanceKey(t *testing.T) {
	contractID := "CCW67ZTM6S7C3S64Z6S3SLZ6S7C3S64Z6S3SLZ6S7C3S64Z6S3SLZ6S"
	key, err := NewContractInstanceKey(contractID)
	require.NoError(t, err)

	assert.Equal(t, xdr.LedgerEntryTypeContractData, key.Type)
	assert.NotNil(t, key.ContractData)
	assert.Equal(t, xdr.ScAddressTypeScAddressTypeContract, key.ContractData.Contract.Type)
	assert.Equal(t, xdr.ScValTypeScvLedgerKeyContractInstance, key.ContractData.Key.Type)
}

func TestNewContractCodeKey(t *testing.T) {
	wasmHash := [32]byte{1, 2, 3}
	key, err := NewContractCodeKey(wasmHash)
	require.NoError(t, err)

	assert.Equal(t, xdr.LedgerEntryTypeContractCode, key.Type)
	assert.NotNil(t, key.ContractCode)
	assert.Equal(t, xdr.Hash(wasmHash), key.ContractCode.Hash)
}

func TestParseWasmHashFromInstance(t *testing.T) {
	wasmHash := [32]byte{0xde, 0xad, 0xbe, 0xef}
	
	// Create a mock LedgerEntry for ContractInstance
	entry := xdr.LedgerEntry{
		Data: xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeContractData,
			ContractData: &xdr.ContractDataEntry{
				Contract: xdr.ScAddress{
					Type: xdr.ScAddressTypeScAddressTypeContract,
					ContractId: &xdr.Hash{},
				},
				Key: xdr.ScVal{
					Type: xdr.ScValTypeScvLedgerKeyContractInstance,
				},
				Val: xdr.ScVal{
					Type: xdr.ScValTypeScvContractInstance,
					Instance: &xdr.ScContractInstance{
						Executable: xdr.ContractExecutable{
							Type: xdr.ContractExecutableTypeContractExecutableWasm,
							WasmHash: (*xdr.Hash)(&wasmHash),
						},
					},
				},
			},
		},
	}

	xdrBytes, err := entry.MarshalBinary()
	require.NoError(t, err)
	entryXDR := base64.StdEncoding.EncodeToString(xdrBytes)

	parsedHash, ok := ParseWasmHashFromInstance(entryXDR)
	assert.True(t, ok)
	assert.Equal(t, wasmHash, parsedHash)

	// Test invalid data
	_, ok = ParseWasmHashFromInstance("invalid-base64")
	assert.False(t, ok)

	// Test wrong entry type
	entry.Data.Type = xdr.LedgerEntryTypeAccount
	xdrBytes, _ = entry.MarshalBinary()
	_, ok = ParseWasmHashFromInstance(base64.StdEncoding.EncodeToString(xdrBytes))
	assert.False(t, ok)
}
