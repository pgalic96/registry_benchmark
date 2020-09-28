package cmd

import (
	"log"
	"os/exec"

	"github.com/spf13/cobra"

	"registry_benchmark/auth"
	registryconfig "registry_benchmark/config"
)

func init() {
	rootCmd.AddCommand(traceReplayerCmd)
}

var traceReplayerCmd = &cobra.Command{
	Use:   "trace-replayer",
	Short: "Benchmark registries using real world traces",
	Long:  "Use trace replayer to replay IBM traces",
	Run: func(cmd *cobra.Command, args []string) {
		config, _ := registryconfig.LoadConfig(yamlFilename)

		for _, containerReg := range config.Registries {
			username, password, _ := auth.ObtainRegistryCredentials(containerReg, yamlFilename)

			traceReplayerConfig := registryconfig.TraceReplayerConfig{
				Username:   username,
				Password:   password,
				Repository: containerReg.Repository,
			}

			err := registryconfig.SetTraceReplayerEnvVariables(config.TraceReplayerPath, traceReplayerConfig)
			if err != nil {
				log.Fatalf("Error while setting env variables: %v", err)
			}

			// Run registry warmup
			warmupCommand := exec.Command("python2 master.py -c warmup -i config.yaml")
			warmupCommand.Dir = config.TraceReplayerPath
			warmupCommand.Run()

			// Run clients
			clientCommand1 := exec.Command("python2 client.py -i 0.0.0.0 -p 8081")
			clientCommand2 := exec.Command("python2 client.py -i 0.0.0.0 -p 8082")

			clientCommand1.Dir = config.TraceReplayerPath
			clientCommand2.Dir = config.TraceReplayerPath

			// Start clients async
			clientCommand1.Start()
			clientCommand2.Start()

			// Run master
			masterCommand := exec.Command("python2 master.py -c run -i config.yaml")
			masterCommand.Dir = config.TraceReplayerPath
			masterCommand.Run()
		}
	},
}
