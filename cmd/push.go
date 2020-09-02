package cmd

import (
	// Blind

	"bytes"
	_ "crypto/sha256"
	"io/ioutil"
	"log"
	"os"

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
		log.Printf("username: %v", os.Getenv("DOCKER_USERNAME"))
		log.Printf("password: %v", os.Getenv("DOCKER_PASSWORD"))
		hub, err := registry.New(url, os.Getenv("DOCKER_USERNAME"), os.Getenv("DOCKER_PASSWORD"))
		if err != nil {
			log.Fatalf("Error initializing a registry client: %v", err)
		}
		imggen.Generate()
		digest := digest.NewDigestFromHex(
			"sha256",
			"a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4",
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
		log.Printf("Blob not found")
		if !exists {
			file, _ := ioutil.ReadFile("docker-layer-test")
			hub.UploadBlob(repository, digest, bytes.NewReader(file), nil)
			log.Printf("Blob uploaded successfully")
		}
	}}
