package chos

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"gitlab.com/dan-v/unofficial-copperheados/templates"
)

var DarwinBinaryURL = "https://releases.hashicorp.com/terraform/0.11.1/terraform_0.11.1_darwin_amd64.zip"
var LinuxBinaryURL = "https://releases.hashicorp.com/terraform/0.11.1/terraform_0.11.1_linux_amd64.zip"
var WindowsBinaryURL = "https://releases.hashicorp.com/terraform/0.11.1/terraform_0.11.1_windows_amd64.zip"

type TerraformClient struct {
	configDir string
	tempDir   *TempDir
	stdout    io.Writer
	stderr    io.Writer
}

func NewTerraformClient(config *TerraformConfig, stdout, stderr io.Writer) (*TerraformClient, error) {
	if err := setupBinary(config.TempDir); err != nil {
		return nil, err
	}

	log.Info("Rendering Terraform templates in temp dir " + config.TempDir.path)
	terraformFile, err := renderTemplate(templates.TerraformTemplate, config)
	if err != nil {
		return nil, err
	}

	// write out terraform template
	configDir := config.TempDir.Path("config")
	if err := os.Mkdir(configDir, 0777); err != nil {
		return nil, err
	}
	configPath := config.TempDir.Path("config/main.tf")
	if err := ioutil.WriteFile(configPath, terraformFile, 0777); err != nil {
		return nil, err
	}

	// write out shell script
	err = ioutil.WriteFile(config.TempDir.Path(ShellScriptFilename), config.ShellScriptBytes, 0644)
	if err != nil {
		return nil, err
	}

	// write out spot lambda function and zip it up
	err = ioutil.WriteFile(config.TempDir.Path(LambdaSpotFunctionFilename), config.LambdaSpotFunctionBytes, 0644)
	if err != nil {
		return nil, err
	}
	files := []string{config.TempDir.Path(LambdaSpotFunctionFilename)}
	output := config.TempDir.Path(LambdaSpotZipFilename)
	err = zipFiles(output, files)
	if err != nil {
		return nil, err
	}

	// create client and run init
	client := &TerraformClient{
		tempDir:   config.TempDir,
		configDir: configDir,
		stdout:    stdout,
		stderr:    stderr,
	}
	devNull := bytes.NewBuffer(nil)
	if err := client.terraform([]string{"init"}, devNull); err != nil {
		io.Copy(stdout, devNull)
		return nil, err
	}
	return client, nil
}

func (client *TerraformClient) Apply() error {
	client.terraform([]string{
		"plan",
		"-input=false",
		"-out=tfplan",
	}, client.stdout)
	return client.terraform([]string{
		"apply",
		"tfplan",
	}, client.stdout)
}

func (client *TerraformClient) Destroy() error {
	return client.terraform([]string{
		"destroy",
		"-force",
	}, client.stdout)
}

func (client *TerraformClient) terraform(args []string, stdout io.Writer) error {
	cmd := exec.Command(client.tempDir.Path("terraform"), args...)
	cmd.Dir = client.configDir
	cmd.Stdout = stdout
	cmd.Stderr = client.stderr
	return cmd.Run()
}

func (client *TerraformClient) Cleanup() error {
	return os.RemoveAll(client.tempDir.path)
}

func getTerraformURL() (string, error) {
	os := runtime.GOOS
	if os == "darwin" {
		return DarwinBinaryURL, nil
	} else if os == "linux" {
		return LinuxBinaryURL, nil
	} else if os == "windows" {
		return WindowsBinaryURL, nil
	}
	return "", fmt.Errorf("unknown os: `%s`", os)
}

func setupBinary(tempDir *TempDir) error {
	fileHandler, err := os.Create(tempDir.Path("terraform.zip"))
	if err != nil {
		return err
	}
	defer fileHandler.Close()

	url, err := getTerraformURL()
	if err != nil {
		return err
	}

	log.Infoln("Downloading Terraform binary from URL:", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if _, err := io.Copy(fileHandler, resp.Body); err != nil {
		return err
	}
	if err := fileHandler.Sync(); err != nil {
		return err
	}

	err = unzip(tempDir.Path("terraform.zip"), tempDir.path)
	if err != nil {
		return err
	}

	if err := os.Chmod(tempDir.Path("terraform"), 0700); err != nil {
		return err
	}

	return nil
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, f.Mode())
		} else {
			var fdir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				fdir = fpath[:lastIndex]
			}

			err = os.MkdirAll(fdir, f.Mode())
			if err != nil {
				log.Fatal(err)
				return err
			}
			f, err := os.OpenFile(
				fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func zipFiles(filename string, files []string) error {

	newfile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newfile.Close()

	zipWriter := zip.NewWriter(newfile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {

		zipfile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer zipfile.Close()

		// Get the file information
		info, err := zipfile.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Change to deflate to gain better compression
		// see http://golang.org/pkg/archive/zip/#pkg-constants
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, zipfile)
		if err != nil {
			return err
		}
	}
	return nil
}

type TempDir struct {
	path string
}

func NewTempDir() (*TempDir, error) {
	path, err := ioutil.TempDir("", "chosdeploy")
	if err != nil {
		return nil, err
	}

	return &TempDir{
		path: path,
	}, nil
}

func (tempDir *TempDir) Save(filename string, contents []byte) (string, error) {
	path := filepath.Join(tempDir.path, filename)
	if err := ioutil.WriteFile(path, contents, 0700); err != nil {
		return "", err
	}

	return path, nil
}

func (tempDir *TempDir) Path(filename string) string {
	return filepath.Join(tempDir.path, filename)
}

func (tempDir *TempDir) Cleanup() error {
	return os.RemoveAll(tempDir.path)
}
