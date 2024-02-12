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
	contextLogger := logging.FromContext(ctx)

	helper, err := pluginhelper.NewFromCluster(metadata.Data.Name, request.ClusterDefinition)
	if err != nil {
		contextLogger.Error(err, "Error while decoding cluster definition from CNPG")
		return nil, err
	}

	backupObject, err := helper.DecodeBackup(request.BackupDefinition)
	if err != nil {
		contextLogger.Error(err, "Error while decoding backup definition from CNPG")
		return nil, err
	}

	cluster := helper.GetCluster()
	rep, err := newRepository(
		ctx,
		storage.GetBasePath(cluster.Name),
		storage.GetKopiaConfigFilePath(cluster.Name),
		storage.GetKopiaCacheDirectory(cluster.Name),
	)
	if err != nil {
		return nil, err
	}

	exec := newExecutor(
		cluster,
		backupObject,
		rep,
		podIP,
	)

	startedAt := time.Now()
	contextLogger.Info("Preparing physical backup")
	if err := exec.setBackupMode(ctx); err != nil {
		return nil, err
	}

	contextLogger.Info("Copying files")
	if err := exec.execBackup(ctx); err != nil {
		return nil, err
	}

	contextLogger.Info("Finishing backup")
	backupInfo, err := exec.unsetBackupMode(ctx)
	if err != nil {
		return nil, err
	}

	return &backup.BackupResult{
		BackupId:          backupInfo.BackupName,
		BackupName:        backupInfo.BackupName,
		StartedAt:         startedAt.Unix(),
		StoppedAt:         time.Now().Unix(),
		BeginWal:          exec.beginWal,
		EndWal:            exec.endWal,
		BeginLsn:          string(backupInfo.BeginLSN),
		EndLsn:            string(backupInfo.EndLSN),
		BackupLabelFile:   backupInfo.LabelFile,
		TablespaceMapFile: backupInfo.SpcmapFile,
		Online:            true,
	}, nil
}