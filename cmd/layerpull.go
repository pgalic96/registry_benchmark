package cmd

import (
	"encoding/csv"
	"io/ioutil"
	"os"
	"strconv"

	// Blind
	_ "crypto/sha256"
	"log"
	"time"

	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"

	"registry_benchmark/auth"
	"registry_benchmark/config"
	"registry_benchmark/imggen"
)

func init() {
	rootCmd.AddCommand(layerPullCmd)
}

var layerPullCmd = &cobra.Command{
	Use:   "layerpull",
	Short: "Benchmark docker pull with http",
	Long:  `layerpull measures latency when pulling previously pushed generated image layers`,
	Run: func(cmd *cobra.Command, args []string) {
		config, _ := config.LoadConfig(YamlFilename)

		var benchmarkData = make([][]string, len(config.Registries)*config.ImageGeneration.LayerNumber+1)
		benchmarkData[0] = []string{"platform", "layer", "latency"}

		for x, containerReg := range config.Registries {
			hub, err := auth.AuthenticateRegistry(containerReg, YamlFilename)
			if err != nil {
				log.Printf("Error initializing a registry client: %v", err)
				continue
			}

			items, _ := ioutil.ReadDir(config.PullSourceFolder)
			for i, item := range items {
				digest := digest.NewDigestFromHex(
					"sha256",
					item.Name(),
				)
				log.Printf("Checking for blob in repository")
				exists, err := hub.HasBlob(containerReg.Repository, digest)
				if err != nil {
					log.Printf("Error while checking if image exists: %v", err)
					continue
				}
				if exists {
					log.Printf("Blob found")

					start := time.Now()
					reader, err := hub.DownloadBlob(containerReg.Repository, digest)
					if reader != nil {
						defer reader.Close()
					}
					elapsed := time.Since(start)
					if err != nil {
						log.Printf("Error while pulling layer: %s", err)
						continue
					}
					log.Printf("Blob downloaded successfully: %v", elapsed)
					benchmarkData[1+i+x*config.ImageGeneration.LayerNumber] = []string{containerReg.Platform, strconv.Itoa(i), elapsed.String()}
				}
			}
		}
		if writeToCSV == true {
			dt := time.Now()
			var csvFile *os.File
			if cronJob == true {
				csvFile, _ = imggen.Create("long-running/pull-" + strconv.Itoa(config.ImageGeneration.ImgSizeMb) + "-" + strconv.Itoa(config.ImageGeneration.LayerNumber) + "-" + dt.String() + ".csv")
			} else {
				csvFile, _ = os.Create("pull-" + strconv.Itoa(config.ImageGeneration.ImgSizeMb) + "-" + strconv.Itoa(config.ImageGeneration.LayerNumber) + "-" + dt.String() + ".csv")
			}
			var csvwriter = csv.NewWriter(csvFile)
			for _, row := range benchmarkData {
				csvwriter.Write(row)
			}
			csvwriter.Flush()
			csvFile.Close()
		}
	}}
