package pluginhelper

import (
	"net"
	"path"

	"github.com/cloudnative-pg/cnpg-i/pkg/identity"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/cloudnative-pg/plugin-pvc-backup/pkg/logging"
)

const unixNetwork = "unix"

// ServerEnricher is the type of functions that can add register
// service implementations in a GRPC server
type ServerEnricher func(*grpc.Server)

// CreateMainCmd creates a command to be used as the server side
// for the CNPG-I infrastructure
func CreateMainCmd(identityImpl identity.IdentityServer, enrichers ...ServerEnricher) *cobra.Command {
	cmd := &cobra.Command{
		Use: "pvc-backup",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ctx := logging.IntoContext(
				cmd.Context(),
				viper.GetBool("debug"))
			cmd.SetContext(ctx)
		},
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			logger := logging.FromContext(cmd.Context())

			identityResponse, err := identityImpl.GetPluginMetadata(
				cmd.Context(),
				&identity.GetPluginMetadataRequest{})
			if err != nil {
				logger.Error(err, "Error while querying the identity service")
				return err
			}

			pluginPath := viper.GetString("plugin-path")
			pluginName := identityResponse.Name
			pluginDisplayName := identityResponse.DisplayName
			pluginVersion := identityResponse.Version
			socketName := path.Join(pluginPath, identityResponse.Name)

			grpcServer := grpc.NewServer()
			identity.RegisterIdentityServer(
				grpcServer,
				identityImpl)
			for _, enrich := range enrichers {
				enrich(grpcServer)
			}

			listener, err := net.Listen(
				unixNetwork,
				socketName,
			)
			if err != nil {
				logger.Error(err, "While starting server")
				return err
			}

			logger.Info(
				"Starting plugin",
				"path", pluginPath,
				"name", pluginName,
				"displayName", pluginDisplayName,
				"version", pluginVersion,
				"socketName", socketName,
			)
			err = grpcServer.Serve(listener)
			if err != nil {
				logger.Error(err, "While terminatind server")
			}

			return err
		},
	}

	cmd.PersistentFlags().Bool(
		"debug",
		true,
		"Enable debugging mode",
	)
	_ = viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))

	cmd.Flags().String(
		"plugin-path",
		"/plugins",
		"The plugins socket path",
	)
	_ = viper.BindPFlag("plugin-path", cmd.Flags().Lookup("plugin-path"))

	return cmd
}
