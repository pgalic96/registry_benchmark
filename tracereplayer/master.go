package tracereplayer

import "os"

func runMasterPrefetch() error {
	jumpClient, client, err := newDASClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer jumpClient.Close()
	session, err := client.NewSession()
	if session != nil {
		defer session.Close()
	}
	if err != nil {
		return err
	}
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	masterPrefetchCmd := "srun -N1 /usr/bin/bash master-prefetch.sh"
	if err := session.Run(masterPrefetchCmd); err != nil {
		return err
	}
	return nil
}

func runMasterRun() error {
	jumpClient, client, err := newDASClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer jumpClient.Close()
	session, err := client.NewSession()
	if session != nil {
		defer session.Close()
	}
	if err != nil {
		return err
	}
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	masterRunCmd := "srun -N1 /usr/bin/bash master-run.sh"
	if err := session.Run(masterRunCmd); err != nil {
		return err
	}
	return nil
}
