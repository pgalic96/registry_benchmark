package tracereplayer

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"registry_benchmark/auth"
	"registry_benchmark/config"
	"strings"

	"gopkg.in/yaml.v2"
)

// DeploymentConfig contains config for DAS deployment
type DeploymentConfig struct {
	GcloudKey               string           `yaml:"gcloud-key,omitempty"`
	AwsCredentialsPath      string           `yaml:"aws-credentials-path,omitempty"`
	TraceReplayerPath       string           `yaml:"trace-replayer-path,omitempty"`
	TracePath               string           `yaml:"trace-path,omitempty"`
	TraceReplayerConfigPath string           `yaml:"trace-replayer-config-path,omitempty"`
	VunetCredentials        VunetCredentials `yaml:"vunet-credentials"`
	DasCredentials          DasCredentials   `yaml:"das-credentials"`
	ClientScript            string           `yaml:"client-script"`
	PrefetchScript          string           `yaml:"prefetch-script"`
	RunScript               string           `yaml:"run-script"`
	ClientList              []string         `yaml:"das-client-nodes"`
}

type VunetCredentials struct {
	SSHConnection string `yaml:"ssh-connection"`
	Password      string
}

type DasCredentials struct {
	DasConnection string `yaml:"das-connection"`
	Username      string
	Password      string
}

// Config is a Deployment Config used across tracereplayer package
var Config *DeploymentConfig
var homeDir string
var registryConfig *config.Config

func DeployTraceReplayerDas() {
	Config, _ = loadConfig("das-config.yaml")
	registryConfig, _ = config.LoadConfig(Config.TraceReplayerConfigPath)
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	homeDir = fmt.Sprintf("%s/", usr.HomeDir)

	// Bundle files
	traceFiles := getFilesInDir(Config.TracePath)
	traceReplayerFiles := getFilesInDir(Config.TraceReplayerPath)
	files := []string{Config.GcloudKey, Config.AwsCredentialsPath, Config.TraceReplayerConfigPath, Config.PrefetchScript, Config.ClientScript, Config.RunScript}
	files = append(files, traceFiles...)
	files = append(files, traceReplayerFiles...)
	output := "done.zip"
	if err := zipFiles(output, files); err != nil {
		log.Fatalln(err)
	}
	log.Println("Zipped File:", output)

	// Copy files to DAS via scp
	log.Println("Copying bundled files to DAS...")
	//results := make(chan string, 10)
	jumpClient, client, err := newDASClient()
	if err != nil {
		log.Fatal(err)
	}
	session, err := client.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Established session at das")

	err = sendFilesToDas(session, output, Config.DasCredentials.Username)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Files copied successfully.")
	session.Close()
	client.Close()
	jumpClient.Close()

	log.Println("Unpacking files...")
	jumpClient, client, err = newDASClient()
	if err != nil {
		log.Fatal(err)
	}
	session, err = client.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	session.Run("/usr/bin/unzip -o /home/user/done.zip")
	session.Close()
	client.Close()
	jumpClient.Close()
	log.Println("Files unpacked")

	// resultFileList := make([]string, len(registryConfig.Registries))

	go func() {
		err := runClients()
		if err != nil {
			log.Fatalf("Error obtained while running clients: %v", err)
			return
		}
	}()

	for _, registry := range registryConfig.Registries {
		// Set env file
		username, password, _ := auth.ObtainRegistryCredentials(registry, "config-pull.yaml")
		traceReplayerConfig := config.TraceReplayerCredentials{
			Username:   username,
			Password:   strings.ReplaceAll(password, "\n", ""),
			Repository: registry.Repository,
			URL:        strings.TrimSuffix(strings.TrimPrefix(registry.URL, "https://"), "/"),
		}
		clientIPs := getClientIPs(Config.ClientList)
		err := config.SetTraceReplayerEnvVariables(traceReplayerConfig, registryConfig.ReplayerConfig, clientIPs)
		if err != nil {
			log.Fatalf("Error setting trace replayer env variables: %v", err)
		}
		log.Println("Sending env file to DAS")
		err = deployEnvFileToDAS(Config.TraceReplayerPath + "/.env")
		if err != nil {
			log.Fatalf("Error copying .env file to DAS: %v", err)
		}

		err = runMasterPrefetch()
		if err != nil {
			log.Fatalf("Error while running master prefetch: %v", err)
		}

		err = runMasterRun()
		if err != nil {
			log.Fatalf("Error while running master prefetch: %v", err)
		}
	}

	resultsFilePath, err := extractResults(registryConfig.ReplayerConfig.ResultsDir)
	if err != nil {
		log.Fatalf("Error while extracting results: %v", err)
	}

	log.Println("Successfully extracted results at following path: " + resultsFilePath)

}

func loadConfig(filename string) (*DeploymentConfig, error) {
	c := DeploymentConfig{}
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return &c, nil
}

func zipFiles(filename string, files []string) error {

	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if err = addFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {

	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = strings.TrimPrefix(filename, homeDir)

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

func getFilesInDir(pathname string) []string {
	var files []string

	err := filepath.Walk(pathname, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}

func getClientIPs(clientNames []string) []string {
	clientIPs := make([]string, len(clientNames))
	for i, clientName := range clientNames {
		clientIPs[i] = fmt.Sprintf("10.141.0.%s:8084", strings.TrimPrefix(clientName, "node0"))
		log.Println(clientIPs[i])
	}
	return clientIPs
}
