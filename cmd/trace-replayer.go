package cmd

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"registry_benchmark/auth"
	registryconfig "registry_benchmark/config"
)

var authOnly bool

const python string = "python2"
const master string = "master.py"
const clnt string = "client.py"
const localhost string = "0.0.0.0"
const port string = "-p"
const p1 string = "8081"
const p2 string = "8082"
const command string = "-c"
const warmup string = "warmup"
const run string = "run"
const i string = "-i"
const conf string = "config.yaml"

func init() {
	traceReplayerCmd.Flags().BoolVarP(&authOnly, "auth-only", "a", false, "Obtain and store credentials in .env only")
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
				Password:   strings.ReplaceAll(password, "\n", ""),
				Repository: containerReg.Repository,
				URL:        strings.TrimSuffix(strings.TrimPrefix(containerReg.URL, "https://"), "/"),
			}

			err := registryconfig.SetTraceReplayerEnvVariables(config.TraceReplayerPath, traceReplayerConfig)
			if err != nil {
				log.Fatalf("Error while setting env variables: %v", err)
			}

			if !authOnly {
				err = runTraceReplayer(config.TraceReplayerPath)
				if err != nil {
					log.Fatalf("Error while running trace replayer: %v", err)
				}
			}
		}
	},
}

func runTraceReplayer(path string) error {
	// Run registry warmup
	log.Println("Starting warmup")
	warmupCommand := exec.Command(python, master, command, warmup, i, conf)
	warmupCommand.Dir = path
	log.Println(warmupCommand.Dir)
	err := warmupCommand.Run()
	if err != nil {
		return err
	}
	log.Println("Warmup done, starting clients...")

	// Run clients
	clientCommand1 := exec.Command(python, clnt, i, localhost, port, p1)
	clientCommand2 := exec.Command(python, clnt, i, localhost, port, p2)

	clientCommand1.Dir = path
	clientCommand2.Dir = path

	// Start clients async
	err = clientCommand1.Start()
	if err != nil {
		return err
	}
	log.Println("Client 1 started")

	err = clientCommand2.Start()
	if err != nil {
		return err
	}
	log.Println("Client 2 started")
	// Run master
	masterCommand := exec.Command(python, master, command, run, i, conf)
	masterCommand.Dir = path
	out, err := masterCommand.Output()
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	log.Println("Master finished, killing clients...")

	if err := clientCommand1.Process.Kill(); err != nil {
		log.Fatal("failed to kill client 1: ", err)
	}

	if err := clientCommand2.Process.Kill(); err != nil {
		log.Fatal("failed to kill client 2: ", err)
	}
	return nil
}
