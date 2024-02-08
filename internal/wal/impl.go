package wal

import (
	"context"

	"github.com/cloudnative-pg/cnpg-i/pkg/wal"
)

// Implementation is the implementation of the identity service
type Implementation struct {
	wal.WALServer
}

// GetCapabilities gets the capabilities of the WAL service
func (Implementation) GetCapabilities(
	context.Context,
	*wal.WALCapabilitiesRequest,
) (*wal.WALCapabilitiesResult, error) {
	return &wal.WALCapabilitiesResult{
		Capabilities: []*wal.WALCapability{
			{
				Type: &wal.WALCapability_Rpc{
					Rpc: &wal.WALCapability_RPC{
						Type: wal.WALCapability_RPC_TYPE_ARCHIVE_WAL,
					},
				},
			},
			{
				Type: &wal.WALCapability_Rpc{
					Rpc: &wal.WALCapability_RPC{
						Type: wal.WALCapability_RPC_TYPE_RESTORE_WAL,
					},
				},
			},
			{
				Type: &wal.WALCapability_Rpc{
					Rpc: &wal.WALCapability_RPC{
						Type: wal.WALCapability_RPC_TYPE_STATUS,
					},
				},
			},
			{
				Type: &wal.WALCapability_Rpc{
					Rpc: &wal.WALCapability_RPC{
						Type: wal.WALCapability_RPC_TYPE_SET_FIRST_REQUIRED,
					},
				},
			},
		},
	}, nil
}
