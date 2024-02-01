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
