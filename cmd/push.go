package cmd

import (
	"bytes"
	"encoding/csv"
	"encoding/hex"
	"os"
	"strconv"
	"strings"

	// Blind
	"crypto/rand"
	_ "crypto/sha256"
	"io/ioutil"
	"log"
	"time"

	influxclient "github.com/influxdata/influxdb1-client/v2"
	"github.com/nokia/docker-registry-client/registry"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"

	"registry_benchmark/auth"
	"registry_benchmark/imggen"
)

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func init() {
	pullCmd.Flags().BoolVarP(&writeToCSV, "csv", "c", false, "write to local csv file")
	rootCmd.AddCommand(pushCmd)
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Benchmark docker push with http",
	Long:  `push generates images and measures push latency`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadConfig()

		log.Printf("Configuring influx client")
		c, err := influxclient.NewHTTPClient(influxclient.HTTPConfig{
			Addr: config.StorageURL,
		})
		if err != nil {
			log.Fatalf("Error creating InfluxDB Client: ", err.Error())
		}
		defer c.Close()

		var benchmarkData = make([][]string, len(config.Registries)*config.ImageGeneration.LayerNumber+1)
		dt := time.Now()
		csvFile, err := os.Create("push-" + strconv.Itoa(config.ImageGeneration.ImgSizeMb) + "-" + strconv.Itoa(config.ImageGeneration.LayerNumber) + "-" + dt.String() + ".csv")
		if err != nil {
			log.Fatalf("failed creating file: %s", err)
		}
		var csvwriter = csv.NewWriter(csvFile)
		defer csvFile.Close()
		if writeToCSV == true {
			benchmarkData[0] = []string{"platform", "layer", "latency"}
		}

		log.Printf("Client configured")
		imggen.Generate()
		for x, containerReg := range config.Registries {
			var password string
			if containerReg.Platform == "ecr" {
				token, err := auth.GetECRAuthorizationToken(containerReg.AccountID, containerReg.Region)
				if err != nil {
					log.Fatalf("Error obtaining aws authorization token: %v", err)
				}
				password = strings.TrimPrefix(token, "AWS:")
				log.Println(password)
			} else {
				password = containerReg.Password
			}
			hub, err := registry.New(containerReg.URL, containerReg.Username, password)
			if err != nil {
				log.Fatalf("Error initializing a registry client: %v", err)
			}
			for i := 0; i < config.ImageGeneration.LayerNumber; i++ {
				hexval, _ := randomHex(32)
				digest := digest.NewDigestFromHex(
					"sha256",
					hexval,
				)
				log.Printf("Checking for blob in repository")
				exists, err := hub.HasBlob(containerReg.Repository, digest)
				if err != nil {
					log.Fatalf("Error while checking if image exists: %v", err)
				}
				if !exists {
					log.Printf("Blob not found")
					file, _ := ioutil.ReadFile("docker-layer-" + strconv.Itoa(i))
					start := time.Now()
					hub.UploadBlob(containerReg.Repository, digest, bytes.NewReader(file), nil)
					elapsed := time.Since(start)
					log.Printf("Blob uploaded successfully: %v", elapsed)
					if writeToCSV == true {
						benchmarkData[1+i+x*config.ImageGeneration.LayerNumber] = []string{containerReg.Platform, strconv.Itoa(i), elapsed.String()}
					}
				}
			}
		}
		if writeToCSV == true {
			for _, row := range benchmarkData {
				csvwriter.Write(row)
			}
		}
		csvwriter.Flush()
	}}
