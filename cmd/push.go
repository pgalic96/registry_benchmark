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
	"os"
	"time"

	influxclient "github.com/influxdata/influxdb1-client/v2"
	"github.com/joho/godotenv"
	"github.com/nokia/docker-registry-client/registry"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"

	"registry_benchmark/imggen"
)

var username string
var password string
var url string
var repository string

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func init() {
	pushCmd.Flags().StringVarP(&url, "url", "l", "https://registry-1.docker.io/", "docker registry address")
	pushCmd.Flags().StringVarP(&repository, "repository", "r", "", "registry repository")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

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

		hub, err := registry.New(url, os.Getenv("DOCKER_USERNAME"), os.Getenv("DOCKER_PASSWORD"))
		if err != nil {
			log.Fatalf("Error initializing a registry client: %v", err)
		}
		imggen.Generate()
		for i := 0; i < config.ImageGeneration.LayerNumber; i++ {
			val, _ := randomHex(64)
			digest := digest.NewDigestFromHex(
				"sha256",
				val,
			)

			// requestBody, err := json.Marshal(map[string]string{
			// 	"username": os.Getenv("DOCKER_USERNAME"),
			// 	"password": os.Getenv("DOCKER_PASSWORD"),
			// })

			// if err != nil {
			// 	log.Fatalln(err)
			// }
			// resp, err := http.Post("https://hub.docker.com/v2/users/login", "application/json", bytes.NewBuffer(requestBody))
			// log.Println(resp)

			log.Printf("Checking for blob in repository")
			exists, err := hub.HasBlob(repository, digest)
			if err != nil {
				log.Fatalf("Error while checking if image exists: %v", err)
			}
			if !exists {
				log.Printf("Blob not found")
				file, _ := ioutil.ReadFile("docker-layer-" + strconv.Itoa(i))
				start := time.Now()
				hub.UploadBlob(repository, digest, bytes.NewReader(file), nil)
				elapsed := time.Since(start)
				log.Printf("Blob uploaded successfully")
			}
		}
	}}
