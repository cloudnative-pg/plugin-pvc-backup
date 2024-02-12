package backup

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"time"

	apiv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	"github.com/cloudnative-pg/cloudnative-pg/pkg/management/postgres/webserver"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"

	"github.com/cloudnative-pg/plugin-pvc-backup/pkg/logging"
)

var (
	errBackupNotStarted = fmt.Errorf("backup not started")
	errBackupNotStopped = fmt.Errorf("backup not stopped")
)

var backupModeBackoff = wait.Backoff{
	Steps:    10,
	Duration: 1 * time.Second,
	Factor:   5.0,
	Jitter:   0.1,
}

// Executor manages the execution of a backup
type Executor struct {
	backupClient webserver.BackupClient

	beginWal string
	endWal   string

	cluster              *apiv1.Cluster
	backup               *apiv1.Backup
	repository           *Repository
	backupClientEndpoint string
}

// Tablespace represent a tablespace location
type Tablespace struct {
	// Path is the path where the tablespaces data is stored
	Path string

	// Oid is the OID of the tablespace inside the database
	Oid string
}

// NewExecutor creates a new backup executor
func NewExecutor(cluster *apiv1.Cluster, backup *apiv1.Backup, repo *Repository, endpoint string) *Executor {
	return &Executor{
		backupClient:         webserver.NewBackupClient(),
		cluster:              cluster,
		backup:               backup,
		repository:           repo,
		backupClientEndpoint: endpoint,
	}
}

// Start starts a backup by setting PostgreSQL in backup mode
func (executor *Executor) Start(ctx context.Context) error {
	logger := logging.FromContext(ctx)

	var currentWALErr error
	executor.beginWal, currentWALErr = executor.getCurrentWALFile(ctx)
	if currentWALErr != nil {
		return currentWALErr
	}

	if err := executor.backupClient.Start(ctx, executor.backupClientEndpoint, webserver.StartBackupRequest{
		ImmediateCheckpoint: true,
		WaitForArchive:      true,
		BackupName:          executor.backup.GetName(),
		Force:               true,
	}); err != nil {
		logger.Error(err, "while requesting new backup on PostgreSQL")
		return err
	}

	logger.Info("Requesting PostgreSQL Backup mode")
	if err := retry.OnError(backupModeBackoff, retryOnBackupNotStarted, func() error {
		response, err := executor.backupClient.StatusWithErrors(ctx, executor.backupClientEndpoint)
		if err != nil {
			return err
		}

		if response.Data.Phase != webserver.Started {
			logger.V(4).Info("Backup still not started", "status", response.Data)
			return errBackupNotStarted
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("Backup Mode started")
	return nil
}

// Backup takes the snapshot of the data directory and the tablespace folder
func (executor *Executor) Backup(ctx context.Context) error {
	const snapshotTablespaceOidName = "oid"

	const (
		snapshotTypeName       = "type"
		snapshotTypeBase       = "base"
		snapshotTypeTablespace = "tablespace"
	)

	logger := logging.FromContext(ctx)

	tablespaces, err := executor.getTablespaces(ctx)
	if err != nil {
		return err
	}

	logger.Info("Taking snapshot of data directory")
	err = executor.repository.TakeSnapshot(ctx, pgDataLocation, map[string]string{
		snapshotTypeName: snapshotTypeBase,
	})
	if err != nil {
		return err
	}

	for i := range tablespaces {
		logger.Info("Taking snapshot of tablespace", "tablespace", tablespaces[i])
		err := executor.repository.TakeSnapshot(ctx, tablespaces[i].Path, map[string]string{
			snapshotTypeName:          snapshotTypeTablespace,
			snapshotTablespaceOidName: tablespaces[i].Oid,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// GetTablespaces read the list of tablespaces
func (*Executor) getTablespaces(ctx context.Context) ([]Tablespace, error) {
	logger := logging.FromContext(ctx)

	tblFolder := path.Join(pgDataLocation, tablespacesFolder)
	entries, err := os.ReadDir(tblFolder)
	if err != nil {
		return nil, err
	}
	result := make([]Tablespace, 0, len(entries))

	for i := range entries {
		fullPath, err := os.Readlink(path.Join(tblFolder, entries[i].Name()))
		if err != nil {
			logger.Error(err, "Error while reading tablespace link")
			return nil, err
		}

		if (entries[i].Type() & fs.ModeSymlink) != 0 {
			result = append(result, Tablespace{
				Oid:  entries[i].Name(),
				Path: fullPath,
			})
		}
	}

	return result, nil
}

// Stop stops a backup and resume PostgreSQL normal operation
func (executor *Executor) Stop(ctx context.Context) (*webserver.BackupResultData, error) {
	logger := logging.FromContext(ctx)

	err := executor.backupClient.Stop(ctx, executor.backupClientEndpoint, webserver.StopBackupRequest{
		BackupName: executor.backup.GetName(),
	})
	if err != nil {
		logger.Error(err, "while requesting new backup on PostgreSQL")
		return nil, err
	}

	logger.Info("Stopping PostgreSQL Backup mode")
	var backupStatus webserver.BackupResultData
	err = retry.OnError(backupModeBackoff, retryOnBackupNotStopped, func() error {
		response, err := executor.backupClient.StatusWithErrors(ctx, executor.backupClientEndpoint)
		if err != nil {
			return err
		}

		if response.Data.Phase != webserver.Completed {
			logger.V(4).Info("backup still not stopped", "status", response.Data)
			return errBackupNotStopped
		}

		backupStatus = *response.Data

		return nil
	})
	if err != nil {
		return nil, err
	}
	logger.Info("PostgreSQL Backup mode stopped")

	executor.endWal, err = executor.getCurrentWALFile(ctx)
	if err != nil {
		return nil, err
	}

	return &backupStatus, err
}

func retryOnBackupNotStarted(e error) bool {
	return e == errBackupNotStarted
}

func retryOnBackupNotStopped(e error) bool {
	return e == errBackupNotStopped
}

func (executor *Executor) getCurrentWALFile(ctx context.Context) (string, error) {
	const currentWALFileControlFile = "Latest checkpoint's REDO WAL file"

	controlDataOutput, err := getPgControlData(ctx)
	if err != nil {
		return "", err
	}

	return controlDataOutput[currentWALFileControlFile], nil
}
