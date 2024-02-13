package executor

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

	executed bool
}

// GetBeginWal returns the beginWal value, panics if the executor was not executed
func (executor *Executor) GetBeginWal() string {
	if !executor.executed {
		panic("beginWal: please run take backup before trying to access this value")
	}
	return executor.beginWal
}

// GetEndWal returns the endWal value, panics if the executor was not executed
func (executor *Executor) GetEndWal() string {
	if !executor.executed {
		panic("endWal: please run take backup before trying to access this value")
	}
	return executor.endWal
}

// tablespace represent a tablespace location
type tablespace struct {
	// path is the path where the tablespaces data is stored
	path string

	// oid is the OID of the tablespace inside the database
	oid string
}

// newExecutor creates a new backup Executor
func newExecutor(cluster *apiv1.Cluster, backup *apiv1.Backup, repo *Repository, endpoint string) *Executor {
	return &Executor{
		backupClient:         webserver.NewBackupClient(),
		cluster:              cluster,
		backup:               backup,
		repository:           repo,
		backupClientEndpoint: endpoint,
	}
}

// NewLocalExecutor creates a new backup Executor
func NewLocalExecutor(cluster *apiv1.Cluster, backup *apiv1.Backup, repo *Repository) *Executor {
	return newExecutor(cluster, backup, repo, podIP)
}

// TakeBackup executes a backup. Returns the result and any error encountered
func (executor *Executor) TakeBackup(ctx context.Context) (*webserver.BackupResultData, error) {
	defer func() {
		executor.executed = true
	}()

	contextLogger := logging.FromContext(ctx)
	contextLogger.Info("Preparing physical backup")
	if err := executor.setBackupMode(ctx); err != nil {
		return nil, err
	}

	contextLogger.Info("Copying files")
	if err := executor.execSnapshot(ctx); err != nil {
		return nil, err
	}

	contextLogger.Info("Finishing backup")
	return executor.unsetBackupMode(ctx)
}

// setBackupMode starts a backup by setting PostgreSQL in backup mode
func (executor *Executor) setBackupMode(ctx context.Context) error {
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

// execSnapshot takes the snapshot of the data directory and the tablespace folder
func (executor *Executor) execSnapshot(ctx context.Context) error {
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
	err = executor.repository.takeSnapshot(ctx, pgDataLocation, map[string]string{
		snapshotTypeName: snapshotTypeBase,
	})
	if err != nil {
		return err
	}

	for i := range tablespaces {
		logger.Info("Taking snapshot of tablespace", "tablespace", tablespaces[i])
		err := executor.repository.takeSnapshot(ctx, tablespaces[i].path, map[string]string{
			snapshotTypeName:          snapshotTypeTablespace,
			snapshotTablespaceOidName: tablespaces[i].oid,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// GetTablespaces read the list of tablespaces
func (*Executor) getTablespaces(ctx context.Context) ([]tablespace, error) {
	logger := logging.FromContext(ctx)

	tblFolder := path.Join(pgDataLocation, tablespacesFolder)
	entries, err := os.ReadDir(tblFolder)
	if err != nil {
		return nil, err
	}
	result := make([]tablespace, 0, len(entries))

	for i := range entries {
		fullPath, err := os.Readlink(path.Join(tblFolder, entries[i].Name()))
		if err != nil {
			logger.Error(err, "Error while reading tablespace link")
			return nil, err
		}

		if (entries[i].Type() & fs.ModeSymlink) != 0 {
			result = append(result, tablespace{
				oid:  entries[i].Name(),
				path: fullPath,
			})
		}
	}

	return result, nil
}

// unsetBackupMode stops a backup and resume PostgreSQL normal operation
func (executor *Executor) unsetBackupMode(ctx context.Context) (*webserver.BackupResultData, error) {
	logger := logging.FromContext(ctx)

	if err := executor.backupClient.Stop(ctx, executor.backupClientEndpoint, webserver.StopBackupRequest{
		BackupName: executor.backup.GetName(),
	}); err != nil {
		logger.Error(err, "while requesting new backup on PostgreSQL")
		return nil, err
	}

	logger.Info("Stopping PostgreSQL Backup mode")
	var backupStatus webserver.BackupResultData
	if err := retry.OnError(backupModeBackoff, retryOnBackupNotStopped, func() error {
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
	}); err != nil {
		return nil, err
	}
	logger.Info("PostgreSQL Backup mode stopped")

	var err error
	executor.endWal, err = executor.getCurrentWALFile(ctx)
	if err != nil {
		return nil, err
	}

	return &backupStatus, nil
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
