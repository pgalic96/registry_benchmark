package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	influxclient "github.com/influxdata/influxdb1-client/v2"
	"github.com/spf13/cobra"
)

// Iterations is a number of iterations of docker pull
var Iterations int

// Image specifies which image to pull
var Image string

// Platform to be benchmarked
var Platform string

func init() {
	pullCmd.Flags().StringVarP(&Image, "img", "i", "", "Image to pull")
	pullCmd.Flags().StringVarP(&Platform, "platform", "p", "", "Registry platform to be benchmarked")
	pullCmd.Flags().IntVarP(&Iterations, "iter", "n", 10, "Iterations of the docker pull")
	rootCmd.AddCommand(pullCmd)
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Benchmark docker pull",
	Long:  `pull executes a docker pull and measures time it takes for it.`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Printf("Configuring influx client")
		c, err := influxclient.NewHTTPClient(influxclient.HTTPConfig{
			Addr: "http://ec2-3-121-232-205.eu-central-1.compute.amazonaws.com:8086",
		})
		if err != nil {
			fmt.Println("Error creating InfluxDB Client: ", err.Error())
		}
		defer c.Close()
		log.Printf("Client configured")

		bp, _ := influxclient.NewBatchPoints(influxclient.BatchPointsConfig{
			Database:  "docker_benchmark",
			Precision: "s",
		})

		tags := map[string]string{"platform": Platform, "image": Image}

		for i := 0; i < Iterations; i++ {

			ctx := context.Background()
			cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				panic(err)
			}
			_, _ = cli.ImageRemove(ctx, Image, types.ImageRemoveOptions{})
			if err != nil {
				panic(err)
			}
			start := time.Now()

			reader, err := cli.ImagePull(ctx, Image, types.ImagePullOptions{})
			if err != nil {
				panic(err)
			}
			io.Copy(os.Stdout, reader)

			elapsed := time.Since(start)

			fields := map[string]interface{}{
				"docker_pull_time": elapsed.Seconds(),
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
	}}
