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
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/cloudnative-pg/cnpg-i/pkg/wal"

	"github.com/cloudnative-pg/plugin-pvc-backup/internal/backup/storage"
	"github.com/cloudnative-pg/plugin-pvc-backup/pkg/logging"
	"github.com/cloudnative-pg/plugin-pvc-backup/pkg/metadata"
	"github.com/cloudnative-pg/plugin-pvc-backup/pkg/pluginhelper"
)

type walStatMode string

const (
	walStatModeFirst = "first"
	walStatModeLast  = "last"
)

// Status gets the statistics of the WAL file archive
func (Implementation) Status(
	ctx context.Context,
	request *wal.WALStatusRequest,
) (*wal.WALStatusResult, error) {
	contextLogger := logging.FromContext(ctx)

	helper, err := pluginhelper.NewFromJSONCluster(metadata.Data.Name, request.ClusterDefinition)
	if err != nil {
		contextLogger.Error(err, "Error while decoding cluster definition from CNPG")
		return nil, err
	}

	walPath := storage.GetWALPath(helper.GetCluster().Name)
	contextLogger = contextLogger.WithValues(
		"walPath", walPath,
		"clusterName", helper.GetCluster().Name,
	)

	walDirEntries, err := os.ReadDir(walPath)
	if err != nil {
		contextLogger.Error(err, "Error while reading WALs directory")
		return nil, err
	}

	firstWal, err := getWALStat(helper.GetCluster().Name, walDirEntries, walStatModeFirst)
	if err != nil {
		contextLogger.Error(err, "Error while reading WALs directory (getting first WAL)")
		return nil, err
	}

	lastWal, err := getWALStat(helper.GetCluster().Name, walDirEntries, walStatModeLast)
	if err != nil {
		contextLogger.Error(err, "Error while reading WALs directory (getting first WAL)")
		return nil, err
	}

	return &wal.WALStatusResult{
		FirstWal: firstWal,
		LastWal:  lastWal,
	}, nil
}

func getWALStat(clusterName string, entries []fs.DirEntry, mode walStatMode) (string, error) {
	entry, ok := getEntry(entries, mode)
	if !ok {
		return "", nil
	}

	if !entry.IsDir() {
		return "", fmt.Errorf("%s is not a directory", entry)
	}

	entryAbsolutePath := path.Join(storage.GetWALPath(clusterName), entry.Name())
	subFolderEntries, err := os.ReadDir(entryAbsolutePath)
	if err != nil {
		return "", fmt.Errorf("while reading %s entries: %w", entry, err)
	}

	selectSubFolderEntry, ok := getEntry(subFolderEntries, mode)
	if !ok {
		return "", nil
	}

	return selectSubFolderEntry.Name(), nil
}

func getEntry(entries []fs.DirEntry, mode walStatMode) (fs.DirEntry, bool) {
	if len(entries) == 0 {
		return nil, false
	}

	switch mode {
	case walStatModeFirst:
		return entries[0], true

	case walStatModeLast:
		return entries[len(entries)-1], true

	default:
		return nil, false
	}
}
