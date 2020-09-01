package cmd

import (
	"io/ioutil"
	"log"
	_ "crypto/sha512"
	digest "github.com/opencontainers/go-digest"
	"github.com/heroku/docker-registry-client/registry"
	"github.com/spf13/cobra"

	"github.com/pgalic96/registry_benchmark/imggen"
)

// Registry is the docker registry configured
var Hub registry.Registry
var username string
var password string
var url string
var repository string
func init() {
	pushCmd.Flags().StringVarP(&username, "username", "u", "", "docker username")
	pushCmd.Flags().StringVarP(&password, "password", "p", "", "docker password")
	pushCmd.Flags().StringVarP(&url, "url", "url", "https://registry-1.docker.io/", "docker registry address")
	pushCmd.Flags().StringVarP(&repository, "repository", "r", "", "registry repository")
	Hub, err = registry.New(url, username, password)
	rootCmd.AddCommand(pushCmd)
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Benchmark docker push with http",
	Long:  `push generates images and measures push latency`,
	Run: func(cmd *cobra.Command, args []string) {
		imggen.Generate()
		digest := digest.NewDigestFromHex(
			"sha256",
			"a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4",
		)
		exists, err := Hub.HasBlob(&repository)
		if err != nil {
			log.Fatalf("Error while checking if image exists: %v", err)
		}
		if !exists {
			file, err := ioutil.ReadFile('imggen')
			reader, err = hub.UploadBlob(&repository, digest, file)
		}
		if reader != nil {
			defer reader.Close()
		}
		if err != nil {
			return err
		}
	}}
