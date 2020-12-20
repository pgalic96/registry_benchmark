package tracereplayer

import (
	"fmt"
	"os"
	"strings"
)

func runClients() error {
	jumpClient, client, err := newDASClient()
	defer client.Close()
	defer jumpClient.Close()
	if err != nil {
		return err
	}
	session, err := client.NewSession()
	if session != nil {
		defer session.Close()
	}
	if err != nil {
		return err
	}
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	clientCommand := fmt.Sprintf("srun -N%d --nodelist=%s -t 180  /usr/bin/bash client.sh", registryConfig.ReplayerConfig.ClientsNumber, strings.Join(Config.ClientList, ","))
	if err := session.Run(clientCommand); err != nil {
		return err
	}
	return nil
}
