package backup

import (
	"context"
	"time"

	"github.com/cloudnative-pg/cnpg-i/pkg/backup"

	"github.com/cloudnative-pg/plugin-pvc-backup/internal/backup/storage"
	"github.com/cloudnative-pg/plugin-pvc-backup/pkg/logging"
	"github.com/cloudnative-pg/plugin-pvc-backup/pkg/metadata"
	"github.com/cloudnative-pg/plugin-pvc-backup/pkg/pluginhelper"
)

// Implementation is the implementation of the identity service
type Implementation struct {
	backup.BackupServer
}

// GetCapabilities gets the capabilities of the Backup service
func (Implementation) GetCapabilities(
	context.Context,
	*backup.BackupCapabilitiesRequest,
) (*backup.BackupCapabilitiesResult, error) {
	return &backup.BackupCapabilitiesResult{
		Capabilities: []*backup.BackupCapability{
			{
				Type: &backup.BackupCapability_Rpc{
					Rpc: &backup.BackupCapability_RPC{
						Type: backup.BackupCapability_RPC_TYPE_BACKUP,
					},
				},
			},
		},
	}, nil
}

// Backup take a physical backup using Kopia
func (Implementation) Backup(
	ctx context.Context,
	request *backup.BackupRequest,
) (*backup.BackupResult, error) {
	logging := logging.FromContext(ctx)

	helper, err := pluginhelper.NewFromCluster(metadata.Data.Name, request.ClusterDefinition)
	if err != nil {
		logging.Error(err, "Error while decoding cluster definition from CNPG")
		return nil, err
	}

	backupObject, err := helper.DecodeBackup(request.BackupDefinition)
	if err != nil {
		logging.Error(err, "Error while decoding backup definition from CNPG")
		return nil, err
	}

	repository, err := NewRepository(
		ctx,
		storage.GetBasePath(helper.GetCluster().Name),
		storage.GetKopiaConfigFilePath(helper.GetCluster().Name),
		storage.GetKopiaCacheDirectory(helper.GetCluster().Name),
	)
	if err != nil {
		return nil, err
	}

	executor := NewExecutor(
		helper.GetCluster(),
		backupObject,
		repository,
	)

	startedAt := time.Now()
	logging.Info("Preparing physical backup")
	if err := executor.Start(ctx); err != nil {
		return nil, err
	}

	logging.Info("Copying files")
	if err := executor.Backup(ctx); err != nil {
		return nil, err
	}

	logging.Info("Finishing backup")
	backupInfo, err := executor.Stop(ctx)
	if err != nil {
		return nil, err
	}
	stoppedAt := time.Now()

	return &backup.BackupResult{
		BackupId:          backupInfo.BackupName,
		BackupName:        backupInfo.BackupName,
		StartedAt:         startedAt.Unix(),
		StoppedAt:         stoppedAt.Unix(),
		BeginWal:          executor.beginWal,
		EndWal:            executor.endWal,
		BeginLsn:          string(backupInfo.BeginLSN),
		EndLsn:            string(backupInfo.EndLSN),
		BackupLabelFile:   backupInfo.LabelFile,
		TablespaceMapFile: backupInfo.SpcmapFile,
		Online:            true,
	}, nil
}
