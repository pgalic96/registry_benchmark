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
)

func init() {
	rootCmd.AddCommand(layerPullCmd)
}

var layerPullCmd = &cobra.Command{
	Use:   "layerpull",
	Short: "Benchmark docker pull with http",
	Long:  `layerpull measures latency when pulling previously pushed generated image layers`,
	Run: func(cmd *cobra.Command, args []string) {
		config, _ := config.LoadConfig()

		var benchmarkData = make([][]string, len(config.Registries)*config.ImageGeneration.LayerNumber+1)
		benchmarkData[0] = []string{"platform", "layer", "latency"}

		for x, containerReg := range config.Registries {
			hub, err := auth.AuthenticateRegistry(containerReg)
			if err != nil {
				log.Fatalf("Error initializing a registry client: %v", err)
			}

			items, _ := ioutil.ReadDir(config.PullSourceFolder)
			for i, item := range items {
				digest := digest.NewDigestFromHex(
					"sha256",
					item.Name(),
				)
				log.Printf("Checking for blob in repository")
				if err != nil {
					log.Fatalf("Error while checking if image exists: %v", err)
				}

				log.Printf("Blob found")
				start := time.Now()
				reader, err := hub.DownloadBlob(containerReg.Repository, digest)
				elapsed := time.Since(start)
				if reader != nil {
					defer reader.Close()
				}
				if err != nil {
					log.Fatalf("Error while pulling layer: %s", err)
				}
				log.Printf("Blob downloaded successfully: %v", elapsed)
				benchmarkData[1+i+x*config.ImageGeneration.LayerNumber] = []string{containerReg.Platform, strconv.Itoa(i), elapsed.String()}

			}
		}
		if writeToCSV == true {
			dt := time.Now()
			csvFile, err := os.Create("pull-" + strconv.Itoa(config.ImageGeneration.ImgSizeMb) + "-" + strconv.Itoa(config.ImageGeneration.LayerNumber) + "-" + dt.String() + ".csv")
			if err != nil {
				log.Fatalf("failed creating file: %s", err)
			}
			var csvwriter = csv.NewWriter(csvFile)
			for _, row := range benchmarkData {
				csvwriter.Write(row)
			}
			csvwriter.Flush()
			csvFile.Close()
		}
	}}
