/*
Copyright The CloudNativePG Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package wal

import (
	"context"
	"path"

	"github.com/cloudnative-pg/cnpg-i/pkg/wal"

	"github.com/cloudnative-pg/plugin-pvc-backup/internal/backup/storage"
	"github.com/cloudnative-pg/plugin-pvc-backup/internal/fileutils"
	"github.com/cloudnative-pg/plugin-pvc-backup/pkg/logging"
	"github.com/cloudnative-pg/plugin-pvc-backup/pkg/metadata"
	"github.com/cloudnative-pg/plugin-pvc-backup/pkg/pluginhelper"
)

// Archive copies one WAL file into the archive
func (Implementation) Archive(
	ctx context.Context,
	request *wal.WALArchiveRequest,
) (*wal.WALArchiveResult, error) {
	contextLogger := logging.FromContext(ctx)

	helper, err := pluginhelper.NewDataBuilder(metadata.Data.Name, request.ClusterDefinition).Build()
	if err != nil {
		contextLogger.Error(err, "Error while decoding cluster definition from CNPG")
		return nil, err
	}

	walName := path.Base(request.SourceFileName)
	destinationPath := storage.GetWALFilePath(helper.GetCluster().Name, walName)

	contextLogger = contextLogger.WithValues(
		"sourceFileName", request.SourceFileName,
		"destinationPath", destinationPath,
		"clusterName", helper.GetCluster().Name,
	)

	contextLogger.Info("Archiving WAL File")
	err = fileutils.CopyFile(request.SourceFileName, destinationPath)
	if err != nil {
		contextLogger.Error(err, "Error archiving WAL file")
	}

	return &wal.WALArchiveResult{}, err
}

// Restore copies WAL file from the archive to the data directory
func (Implementation) Restore(
	ctx context.Context,
	request *wal.WALRestoreRequest,
) (*wal.WALRestoreResult, error) {
	contextLogger := logging.FromContext(ctx)

	helper, err := pluginhelper.NewDataBuilder(metadata.Data.Name, request.ClusterDefinition).Build()
	if err != nil {
		contextLogger.Error(err, "Error while decoding cluster definition from CNPG")
		return nil, err
	}

	walFilePath := storage.GetWALFilePath(helper.GetCluster().Name, request.SourceWalName)
	contextLogger = contextLogger.WithValues(
		"clusterName", helper.GetCluster().Name,
		"walName", request.SourceWalName,
		"walFilePath", walFilePath,
		"destinationPath", request.DestinationFileName,
	)

	contextLogger.Info("Restoring WAL File")
	err = fileutils.CopyFile(walFilePath, request.DestinationFileName)
	if err != nil {
		contextLogger.Info("Restored WAL File", "err", err)
	}

	return &wal.WALRestoreResult{}, err
}
