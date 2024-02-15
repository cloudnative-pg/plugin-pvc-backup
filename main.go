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

// Package main is the entrypoint of the application
package main

import (
	"fmt"
	"os"

	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/pluginhelper"
	"github.com/cloudnative-pg/cnpg-i/pkg/backup"
	"github.com/cloudnative-pg/cnpg-i/pkg/operator"
	"github.com/cloudnative-pg/cnpg-i/pkg/wal"
	"google.golang.org/grpc"

	backupImpl "github.com/cloudnative-pg/plugin-pvc-backup/internal/backup"
	"github.com/cloudnative-pg/plugin-pvc-backup/internal/identity"
	operatorImpl "github.com/cloudnative-pg/plugin-pvc-backup/internal/operator"
	walImpl "github.com/cloudnative-pg/plugin-pvc-backup/internal/wal"
)

func main() {
	cmd := pluginhelper.CreateMainCmd(identity.Implementation{}, func(server *grpc.Server) {
		operator.RegisterOperatorServer(server, operatorImpl.Implementation{})
		wal.RegisterWALServer(server, walImpl.Implementation{})
		backup.RegisterBackupServer(server, backupImpl.Implementation{})
	})
	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
