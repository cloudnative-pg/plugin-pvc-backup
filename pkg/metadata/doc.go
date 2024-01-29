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
