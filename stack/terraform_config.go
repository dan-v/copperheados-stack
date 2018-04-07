package stack

import (
	log "github.com/sirupsen/logrus"
	"gitlab.com/dan-v/unofficial-copperheados/templates"
)

const (
	LambdaSpotFunctionFilename = "lambda_spot_function.py"
	LambdaSpotZipFilename      = "lambda_spot.zip"
	ShellScriptFilename        = "chos.sh"
)

type TerraformConfig struct {
	Name                    string
	Region                  string
	Device                  string
	TempDir                 *TempDir
	ShellScriptFile         string
	ShellScriptBytes        []byte
	LambdaSpotZipFile       string
	LambdaSpotFunctionBytes []byte
	PreventShutdown         bool
}

func generateTerraformConfig(config StackConfig) (*TerraformConfig, error) {
	renderedLambdaSpotFunction, err := renderTemplate(templates.LambdaSpotFunctionTemplate, config)
	if err != nil {
		log.Fatalln("Failed to render Lambda spot function:", err)
	}

	renderedCopperheadShellScript, err := renderTemplate(templates.CopperheadShellScriptTemplate, config)
	if err != nil {
		log.Fatalln("Failed to render shell script:", err)
	}

	tempDir, err := NewTempDir("copperheados-stack")
	if err != nil {
		return nil, err
	}

	conf := TerraformConfig{
		Name:                    config.Name,
		Region:                  config.Region,
		Device:                  config.Device,
		TempDir:                 tempDir,
		ShellScriptFile:         tempDir.Path(ShellScriptFilename),
		ShellScriptBytes:        renderedCopperheadShellScript,
		LambdaSpotZipFile:       tempDir.Path(LambdaSpotZipFilename),
		LambdaSpotFunctionBytes: renderedLambdaSpotFunction,
		PreventShutdown:         config.PreventShutdown,
	}

	return &conf, nil
}
