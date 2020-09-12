package cmd

import (
	"bytes"
	"encoding/csv"
	"io/ioutil"
	"os"
	"strconv"

	"log"
	"time"

	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"

	"registry_benchmark/auth"
	registryconfig "registry_benchmark/config"
	"registry_benchmark/imggen"
)

func init() {
	rootCmd.AddCommand(pushCmd)
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Benchmark docker push with http",
	Long:  `push generates images and measures push latency`,
	Run: func(cmd *cobra.Command, args []string) {
		config, _ := registryconfig.LoadConfig(yamlFilename)

		var benchmarkData = make([][]string, len(config.Registries)*config.ImageGeneration.LayerNumber+1)
		benchmarkData[0] = []string{"platform", "layer", "latency"}

		filepath := imggen.Generate(yamlFilename)
		for x, containerReg := range config.Registries {

			hub, err := auth.AuthenticateRegistry(containerReg, yamlFilename)
			if err != nil {
				log.Fatalf("Error initializing a registry client: %v", err)
			}
			items, _ := ioutil.ReadDir(filepath)
			for i, item := range items {
				digest := digest.NewDigestFromHex(
					"sha256",
					item.Name(),
				)
				log.Printf("Checking for blob in repository")
				exists, err := hub.HasBlob(containerReg.Repository, digest)
				if err != nil {
					log.Fatalf("Error while checking if image exists: %v", err)
				}
				if !exists {
					log.Printf("Blob not found")
					file, _ := ioutil.ReadFile(filepath + item.Name())
					start := time.Now()
					hub.UploadBlob(containerReg.Repository, digest, bytes.NewReader(file), nil)
					elapsed := time.Since(start)
					log.Printf("Blob uploaded successfully: %v", elapsed)
					benchmarkData[1+i+x*config.ImageGeneration.LayerNumber] = []string{containerReg.Platform, strconv.Itoa(i), elapsed.String()}
				}
			}
		}
		if writeToCSV == true {
			dt := time.Now()
			var csvFile *os.File
			if cronJob == true {
				csvFile, _ = imggen.Create("long-running/push-" + strconv.Itoa(config.ImageGeneration.ImgSizeMb) + "-" + strconv.Itoa(config.ImageGeneration.LayerNumber) + "-" + dt.String() + ".csv")
			} else {
				csvFile, _ = os.Create("push-" + strconv.Itoa(config.ImageGeneration.ImgSizeMb) + "-" + strconv.Itoa(config.ImageGeneration.LayerNumber) + "-" + dt.String() + ".csv")
			}
			var csvwriter = csv.NewWriter(csvFile)
			for _, row := range benchmarkData {
				csvwriter.Write(row)
			}
			csvwriter.Flush()
			csvFile.Close()
		}
	}}
