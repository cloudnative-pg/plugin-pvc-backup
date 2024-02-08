package wal

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/cloudnative-pg/cnpg-i/pkg/wal"

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
	logging := logging.FromContext(ctx)

	helper, err := pluginhelper.NewFromCluster(metadata.Data.Name, request.ClusterDefinition)
	if err != nil {
		logging.Error(err, "Error while decoding cluster definition from CNPG")
		return nil, err
	}

	walPath := getWALPath(helper.GetCluster().Name)
	logging = logging.WithValues(
		"walPath", walPath,
		"clusterName", helper.GetCluster().Name,
	)

	walDirEntries, err := os.ReadDir(walPath)
	if err != nil {
		logging.Error(err, "Error while reading WALs directory")
		return nil, err
	}

	firstWal, err := getWALStat(helper.GetCluster().Name, walDirEntries, walStatModeFirst)
	if err != nil {
		logging.Error(err, "Error while reading WALs directory (getting first WAL)")
		return nil, err
	}

	lastWal, err := getWALStat(helper.GetCluster().Name, walDirEntries, walStatModeLast)
	if err != nil {
		logging.Error(err, "Error while reading WALs directory (getting first WAL)")
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

	entryAbsolutePath := path.Join(getWALPath(clusterName), entry.Name())
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
