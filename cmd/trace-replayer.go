package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"registry_benchmark/tracereplayer"

	"registry_benchmark/auth"
	registryconfig "registry_benchmark/config"
)

var authOnly bool
var deployment string
var RegistryConfig *registryconfig.Config

func init() {
	traceReplayerCmd.Flags().BoolVarP(&authOnly, "auth-only", "a", false, "Obtain and store credentials in .env only")
	traceReplayerCmd.Flags().StringVarP(&deployment, "deployment", "d", "local", "Specify deployment option (example: local, das, aws)")
	rootCmd.AddCommand(traceReplayerCmd)
}

func generateClientPorts(clientNumber int) (clientPorts []string) {
	for i := 0; i < clientNumber; i++ {
		clientPorts = append(clientPorts, fmt.Sprintf("808%d", i))
	}
	return clientPorts
}

var traceReplayerCmd = &cobra.Command{
	Use:   "trace-replayer",
	Short: "Benchmark registries using real world traces",
	Long:  "Use trace replayer to replay IBM traces",
	Run: func(cmd *cobra.Command, args []string) {
		RegistryConfig, _ := registryconfig.LoadConfig(YamlFilename)
		if deployment == "local" {
			for _, containerReg := range RegistryConfig.Registries {
				username, password, _ := auth.ObtainRegistryCredentials(containerReg, YamlFilename)

				traceReplayerConfig := registryconfig.TraceReplayerCredentials{
					Username:   username,
					Password:   strings.ReplaceAll(password, "\n", ""),
					Repository: containerReg.Repository,
					URL:        strings.TrimSuffix(strings.TrimPrefix(containerReg.URL, "https://"), "/"),
				}
				clientPorts := generateClientPorts(RegistryConfig.ReplayerConfig.ClientsNumber)
				var clients []string
				for _, port := range clientPorts {
					clients = append(clients, fmt.Sprintf("localhost:%s", port))
				}
				err := registryconfig.SetTraceReplayerEnvVariables(traceReplayerConfig, RegistryConfig.ReplayerConfig, clients)
				if err != nil {
					log.Fatalf("Error while setting env variables: %v", err)
				}

				if !authOnly {
					err = tracereplayer.RunTraceReplayerLocal(RegistryConfig.ReplayerConfig.TracePath, clientPorts)
					if err != nil {
						log.Fatalf("Error while running trace replayer: %v", err)
					}
				}
			}

		} else if deployment == "das" {
			log.Println("Deploying files to DAS...")
			tracereplayer.DeployTraceReplayerDas()
		}
	},
}
