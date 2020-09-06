package cmd

import (
	"bytes"
	"encoding/hex"
	"strconv"

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
		log.Printf("Client configured")
		imggen.Generate()
		for _, containerReg := range config.Registries {
			hub, err := registry.New(containerReg.URL, containerReg.Username, containerReg.Password)
			if err != nil {
				log.Fatalf("Error initializing a registry client: %v", err)
			}
			for i := 0; i < config.ImageGeneration.LayerNumber; i++ {
				hexval, _ := randomHex(10)
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
				}
			}
		}
	}}
