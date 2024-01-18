// Package main is the entrypoint of the application
package main

import (
	"fmt"
	"os"

	"github.com/cloudnative-pg/cnpg-i/pkg/operator"
	"google.golang.org/grpc"

	"github.com/cloudnative-pg/plugin-pvc-backup/internal/identity"
	operatorImpl "github.com/cloudnative-pg/plugin-pvc-backup/internal/operator"
	"github.com/cloudnative-pg/plugin-pvc-backup/pkg/pluginhelper"
)

func main() {
	cmd := pluginhelper.CreateMainCmd(identity.Implementation{}, func(server *grpc.Server) {
		operator.RegisterOperatorServer(server, operatorImpl.Implementation{})
	})
	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
