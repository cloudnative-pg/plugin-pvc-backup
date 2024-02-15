package storage

import "path"

const (
	basePath      = "/backup"
	walsDirectory = "wals"
	baseDirectory = "base"
)

func getWalPrefix(walName string) string {
	return walName[0:16]
}

// getClusterPath gets the path where the files relative
// to a cluster are stored
func getClusterPath(clusterName string) string {
	return path.Join(basePath, clusterName)
}

// GetWALPath gets the path where the WALs relative
// to a cluster are stored
func GetWALPath(clusterName string) string {
	return path.Join(
		getClusterPath(clusterName),
		walsDirectory,
	)
}

// GetKopiaConfigFilePath gets the path where the
// kopia configuration file will be written
func GetKopiaConfigFilePath(clusterName string) string {
	return path.Join(
		getClusterPath(clusterName),
		".kopia.config",
	)
}

// GetKopiaCacheDirectory gets the path where the
// kopia cache will be written
func GetKopiaCacheDirectory(clusterName string) string {
	return path.Join(
		getClusterPath(clusterName),
		".kopia.cache",
	)
}

// GetBasePath gets the path where the WALs relative
// to a cluster are stored
func GetBasePath(clusterName string) string {
	return path.Join(
		getClusterPath(clusterName),
		baseDirectory,
	)
}

// GetWALFilePath gets the path where a certain WAL file
// should be stored
func GetWALFilePath(clusterName string, walName string) string {
	return path.Join(
		GetWALPath(clusterName),
		getWalPrefix(walName),
		walName,
	)
}
