package cmd

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	influxclient "github.com/influxdata/influxdb1-client/v2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Registry is the struct for single registry config
type Registry struct {
	Platform  string
	Imagename string
	Imageurl  string
}

// Config is the configuration for the benchmark
type Config struct {
	Registries []Registry
	Iterations int
	Storageurl string
}

// LoadConfig is the function for loading configuration from yaml file
func LoadConfig() (*Config, error) {
	c := Config{}
	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return &c, nil
}

func init() {
	rootCmd.AddCommand(pullCmd)
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Benchmark docker pull",
	Long:  `pull executes a docker pull and measures time it takes for it.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Loading config file")
		config, err := LoadConfig()

		log.Printf("Configuring influx client")
		c, err := influxclient.NewHTTPClient(influxclient.HTTPConfig{
			Addr: config.Storageurl,
		})
		if err != nil {
			fmt.Println("Error creating InfluxDB Client: ", err.Error())
		}
		defer c.Close()
		log.Printf("Client configured")
		for _, registry := range config.Registries {
			bp, _ := influxclient.NewBatchPoints(influxclient.BatchPointsConfig{
				Database:  "docker_benchmark",
				Precision: "s",
			})

			tags := map[string]string{"platform": registry.Platform, "image": registry.Imagename}

			for i := 0; i < config.Iterations; i++ {

				ctx := context.Background()
				cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
				if err != nil {
					panic(err)
				}
				_, _ = cli.ImageRemove(ctx, registry.Imageurl, types.ImageRemoveOptions{})
				if err != nil {
					panic(err)
				}
				start := time.Now()

				reader, err := cli.ImagePull(ctx, registry.Imageurl, types.ImagePullOptions{})
				if err != nil {
					panic(err)
				}
				io.Copy(os.Stdout, reader)

				elapsed := time.Since(start)

				fields := map[string]interface{}{
					"docker_pull_time": elapsed.Seconds(),
					"iteration_number": i,
				}

				pt, err := influxclient.NewPoint("registry_pull", tags, fields, time.Now())
				if err != nil {
					fmt.Println("Error: ", err.Error())
				}
				bp.AddPoint(pt)

				log.Printf("Time for the pull %s", elapsed)
			}

			err = c.Write(bp)
			if err != nil {
				panic(err)
			}
		}
	}}
