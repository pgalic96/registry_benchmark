package tracereplayer

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func newDASSession() (*ssh.Session, error) {
	knownHostsDir := fmt.Sprintf("%s.ssh/known_hosts", homeDir)
	hostKeyCallback, err := knownhosts.New(knownHostsDir)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal("Failed to dial: ", err)
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
		log.Fatal(err)
	}
	log.Println("Succeeded dialing das")

	ncc, chans, reqs, err := ssh.NewClientConn(conn, Config.DasCredentials.DasConnection, dasconfig)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Established new client connection to DAS")
	sClient := ssh.NewClient(ncc, chans, reqs)

	return sClient.NewSession()
}

func sendFilesToDas(session *ssh.Session, filename string) error {
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

	command := fmt.Sprintf("/usr/bin/scp -t /home/%s/", Config.DasCredentials.Username)

	session.Run(command)
	wg.Wait()
	return nil
}
