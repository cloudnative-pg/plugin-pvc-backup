package wal

import (
	"context"
	"path"

	"github.com/cloudnative-pg/cnpg-i/pkg/wal"

	"github.com/cloudnative-pg/plugin-pvc-backup/internal/fileutils"
	"github.com/cloudnative-pg/plugin-pvc-backup/pkg/logging"
	"github.com/cloudnative-pg/plugin-pvc-backup/pkg/metadata"
	"github.com/cloudnative-pg/plugin-pvc-backup/pkg/pluginhelper"
)

const (
	basePath      = "/backup"
	walsDirectory = "wals"
)

// Archive copies one WAL file into the archive
func (Implementation) Archive(
	ctx context.Context,
	request *wal.WALArchiveRequest,
) (*wal.WALArchiveResult, error) {
	logging := logging.FromContext(ctx)

	helper, err := pluginhelper.NewFromCluster(metadata.Data.Name, request.ClusterDefinition)
	if err != nil {
		logging.Error(err, "Error while decoding cluster definition from CNPG")
		return nil, err
	}

	walName := path.Base(request.SourceFileName)
	destinationPath := getWALFilePath(helper.GetCluster().Name, walName)

	logging = logging.WithValues(
		"sourceFileName", request.SourceFileName,
		"destinationPath", destinationPath,
		"clusterName", helper.GetCluster().Name,
	)

	logging.Info("Archiving WAL File")
	err = fileutils.CopyFile(request.SourceFileName, destinationPath)
	if err != nil {
		logging.Error(err, "Error archiving WAL file")
	}

	return &wal.WALArchiveResult{}, err
}

// Restore copies WAL file from the archive to the data directory
func (Implementation) Restore(
	ctx context.Context,
	request *wal.WALRestoreRequest,
) (*wal.WALRestoreResult, error) {
	logging := logging.FromContext(ctx)

	helper, err := pluginhelper.NewFromCluster(metadata.Data.Name, request.ClusterDefinition)
	if err != nil {
		logging.Error(err, "Error while decoding cluster definition from CNPG")
		return nil, err
	}

	walFilePath := getWALFilePath(helper.GetCluster().Name, request.SourceWalName)
	logging = logging.WithValues(
		"clusterName", helper.GetCluster().Name,
		"walName", request.SourceWalName,
		"walFilePath", walFilePath,
		"destinationPath", request.DestinationFileName,
	)

	logging.Info("Restoring WAL File")
	err = fileutils.CopyFile(walFilePath, request.DestinationFileName)
	if err != nil {
		logging.Info("Restored WAL File", "err", err)
	}

	return &wal.WALRestoreResult{}, err
}
