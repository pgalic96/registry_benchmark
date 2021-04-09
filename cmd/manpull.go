package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"registry_benchmark/auth"
	registryconfig "registry_benchmark/config"
)

func init() {
	rootCmd.AddCommand(manPullCmd)
}

var manPullCmd = &cobra.Command{
	Use:   "manpull",
	Short: "Check if correct manifest is fetched",
	Long:  `Check if the correct manifest is referenced by latest tag`,
	Run: func(cmd *cobra.Command, args []string) {
		config, _ := registryconfig.LoadConfig(YamlFilename)
		for _, containerReg := range config.Registries {
			hub, err := auth.AuthenticateRegistry(containerReg, YamlFilename)
			if err != nil {
				log.Printf("Error initializing a registry client: %v", err)
				continue
			}

			manifest, err := hub.ManifestV2(containerReg.Repository, "latest")
			if err != nil {
				log.Printf("Error when fetching manifest: %v", err)
				continue
			}

			layers := manifest.Layers
			for _, layer := range layers {
				log.Printf("LAYER DIGEST %s", layer.Digest.String())
			}
		}
	}}
