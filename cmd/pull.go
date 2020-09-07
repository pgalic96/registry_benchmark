package cmd

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	influxclient "github.com/influxdata/influxdb1-client/v2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"registry_benchmark/imggen"
)

// Registry is the struct for single registry config
type Registry struct {
	Platform   string
	ImageURL   string `yaml:"image-url,omitempty"`
	URL        string `yaml:"registry-url,omitempty"`
	Username   string
	Password   string
	Repository string
}

// Config is the configuration for the benchmark
type Config struct {
	Registries      []Registry
	ImageGeneration imggen.ImgGen `yaml:"image-generation,omitempty"`
	ImageName       string        `yaml:"image-name,omitempty"`
	Iterations      int
	StorageURL      string `yaml:"storage-url,omitempty"`
}

// LoadConfig is the function for loading configuration from yaml file
func loadConfig() (*Config, error) {
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
	Use:   "pullbench",
	Short: "Benchmark docker pull",
	Long:  `pull executes a docker pull and measures time it takes for it.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Loading config file")
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
		var benchmarkData = make([][]string, len(config.Registries)*config.Iterations+1)
		dt := time.Now()
		csvFile, err := os.Create("pull-" + dt.String() + ".csv")
		if err != nil {
			log.Fatalf("failed creating file: %s", err)
		}
		var csvwriter = csv.NewWriter(csvFile)
		defer csvFile.Close()
		if writeToCSV == true {
			benchmarkData[0] = []string{"iteration", "platform", "image", "latency", "time"}
		}

		for x, registry := range config.Registries {
			bp, _ := influxclient.NewBatchPoints(influxclient.BatchPointsConfig{
				Database:  "docker_benchmark",
				Precision: "s",
			})

			tags := map[string]string{"platform": registry.Platform, "image": config.ImageName}

			for i := 0; i < config.Iterations; i++ {

				ctx := context.Background()
				cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
				if err != nil {
					panic(err)
				}
				_, _ = cli.ImageRemove(ctx, registry.ImageURL, types.ImageRemoveOptions{})
				if err != nil {
					panic(err)
				}
				start := time.Now()

				reader, err := cli.ImagePull(ctx, registry.ImageURL, types.ImagePullOptions{})
				if err != nil {
					panic(err)
				}
				io.Copy(os.Stdout, reader)

				elapsed := time.Since(start)

				fields := map[string]interface{}{
					"docker_pull_time": elapsed.Seconds(),
					"iteration_number": i,
				}
				if writeToCSV == true {
					benchmarkData[x*config.Iterations+1+i] = []string{strconv.Itoa(i), registry.Platform, config.ImageName, elapsed.String(), time.Now().Format("2006-01-02T15:04:05.999999-07:00")}
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
		if writeToCSV == true {
			for _, row := range benchmarkData {
				csvwriter.Write(row)
			}
		}
		csvwriter.Flush()
	}}
