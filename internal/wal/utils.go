package wal

import "path"

func getWalPrefix(walName string) string {
	return walName[0:16]
}

func getClusterPath(clusterName string) string {
	return path.Join(basePath, clusterName)
}

func getWALPath(clusterName string) string {
	return path.Join(
		getClusterPath(clusterName),
		walsDirectory,
	)
}

func getWALFilePath(clusterName string, walName string) string {
	return path.Join(
		getWALPath(clusterName),
		getWalPrefix(walName),
		walName,
	)
}
