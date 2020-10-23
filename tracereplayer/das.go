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

func RunTraceReplayerDas() {
	Config, _ = loadConfig("das-config.yaml")
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	homeDir = fmt.Sprintf("%s/", usr.HomeDir)

	// Bundle files
	traceFiles := getFilesInDir(Config.TracePath)
	traceReplayerFiles := getFilesInDir(Config.TraceReplayerPath)
	files := []string{Config.GcloudKey, Config.AwsCredentialsPath, Config.TraceReplayerConfigPath, "registry_benchmark"}
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
	session, err := newDASSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()
	log.Println("Established session at das")

	err = sendFilesToDas(session, output)
	if err != nil {
		log.Fatal(err)
	}
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
