package tracereplayer

import (
	"fmt"
	"io"
	"os"
	"time"
)

func extractResults(filepath string) (string, error) {
	jumpClient, client, err := newDASClient()
	defer jumpClient.Close()
	defer client.Close()
	if err != nil {
		return "", err
	}
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	r, err := session.StdoutPipe()
	if err != nil {
		return "", err
	}
	name := fmt.Sprintf("results/experiment-result-%v.tar.gz", time.Now().Unix())
	file, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	cmd := "tar -cz results"
	if err := session.Start(cmd); err != nil {
		return "", err
	}
	_, err = io.Copy(file, r)
	if err != nil {
		return "", err
	}
	if err := session.Wait(); err != nil {
		return "", err
	}
	return name, nil
}
