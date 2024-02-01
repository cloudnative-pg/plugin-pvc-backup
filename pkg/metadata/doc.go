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

// Package metadata contains the metadata of this plugin
package metadata

import "github.com/cloudnative-pg/cnpg-i/pkg/identity"

// Data is the metadata of this plugin
// LEO: does this really belong here?
var Data = identity.GetPluginMetadataResponse{
	Name:          "pvc-backup.cloudnative-pg.io",
	Version:       "1.0.0",
	DisplayName:   "CNPG-I plugin to backup and recover using a PVC",
	ProjectUrl:    "https://github.com/cloudnative-pg/plugin-pvc-backup",
	RepositoryUrl: "https://github.com/cloudnative-pg/plugin-pvc-backup",
	License:       "Apache 2",
	LicenseUrl:    "https://github.com/cloudnative-pg/plugin-pvc-backup/blob/main/LICENSE",
	Maturity:      "alpha",
	Vendor:        "CloudNative-PG Community",
}
