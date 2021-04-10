package tracereplayer

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	scp "github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func newDASClient() (*ssh.Client, *ssh.Client, error) {
	knownHostsDir := fmt.Sprintf("%s.ssh/known_hosts", homeDir)
	hostKeyCallback, err := knownhosts.New(knownHostsDir)
	if err != nil {
		return nil, nil, err
	}
	sshconfig := &ssh.ClientConfig{
		User: Config.DasCredentials.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(Config.VunetCredentials.Password),
		},
		HostKeyCallback: hostKeyCallback,
	}

	client, err := ssh.Dial("tcp", Config.VunetCredentials.SSHConnection, sshconfig)
	if err != nil {
		return nil, nil, err
	}
	log.Println("Succeded connection to jumphost")

	dasconfig := &ssh.ClientConfig{
		User: Config.DasCredentials.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(Config.DasCredentials.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := client.Dial("tcp", Config.DasCredentials.DasConnection)
	if err != nil {
		return nil, nil, err
	}
	log.Println("Succeeded dialing das")

	ncc, chans, reqs, err := ssh.NewClientConn(conn, Config.DasCredentials.DasConnection, dasconfig)
	if err != nil {
		return nil, nil, err
	}
	log.Println("Established new client connection to DAS")
	sClient := ssh.NewClient(ncc, chans, reqs)

	return client, sClient, nil
}

func sendFilesToDas(session *ssh.Session, filename string, fileLocation string) error {
	file, err := os.Open(filename)

	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		hostIn, _ := session.StdinPipe()
		defer hostIn.Close()
		fmt.Fprintf(hostIn, "C0664 %d %s\n", stat.Size(), filename)
		io.Copy(hostIn, file)
		fmt.Fprint(hostIn, "\x00")
		wg.Done()
	}()

	command := fmt.Sprintf("/usr/bin/scp -t /home/%s/", fileLocation)

	session.Run(command)
	wg.Wait()
	return nil
}

func deployEnvFileToDAS(filepath string) error {
	jumpClient, client, err := newDASClient()
	defer client.Close()
	defer jumpClient.Close()
	if err != nil {
		return err
	}
	scpClient, err := scp.NewClientBySSH(client)
	err = scpClient.Connect()
	if err != nil {
		return err
	}
	f, _ := os.Open(filepath)
	// Close client connection after the file has been copied
	defer scpClient.Close()

	// Close the file after it has been copied
	defer f.Close()

	err = scpClient.CopyFile(f, "/home/user/thesis/docker-performance/.env", "0655")

	if err != nil {
		fmt.Println("Error while copying file ", err)
	}
	return err
}
