package tracereplayer

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

const python string = "python2"
const master string = "master.py"
const clnt string = "client.py"
const localhost string = "0.0.0.0"
const port string = "-p"
const command string = "-c"
const warmup string = "warmup"
const run string = "run"
const i string = "-i"
const conf string = "config.yaml"

// RunTraceReplayerLocal locally deploys the trace replayer
func RunTraceReplayerLocal(path string, clientPorts []string) error {
	// Run registry warmup
	log.Println("Starting warmup")
	warmupCommand := exec.Command(python, master, command, warmup, i, conf)
	warmupCommand.Stdout = os.Stdout
	warmupCommand.Stderr = os.Stderr
	warmupCommand.Dir = path
	log.Println(warmupCommand.Dir)
	err := warmupCommand.Run()
	if err != nil {
		return err
	}
	log.Println("Warmup done, starting clients...")

	var clientProcesses []*exec.Cmd
	var clientCommand *exec.Cmd
	for _, clientPort := range clientPorts {
		clientCommand = exec.Command(python, clnt, i, localhost, port, clientPort)
		clientCommand.Stdout = os.Stdout
		clientCommand.Stderr = os.Stderr
		clientCommand.Dir = path
		err = clientCommand.Start()
		if err != nil {
			return err
		}
		clientProcesses = append(clientProcesses, clientCommand)
		log.Println(fmt.Sprintf("Client on port %s started", clientPort))
	}
	for _, clientCommand := range clientProcesses {
		defer clientCommand.Process.Kill()
	}
	// Run master
	masterCommand := exec.Command(python, master, command, run, i, conf)
	masterCommand.Dir = path
	out, err := masterCommand.Output()
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	log.Println("Master finished, killing clients...")

	return nil
}
