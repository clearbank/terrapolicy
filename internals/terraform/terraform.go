package terraform

import (
	"encoding/json"
	"fmt"
	"github.com/clearbank/terrapolicy/internals/file"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/clearbank/terrapolicy/internals/utils"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type ModulesJson struct {
	Modules []ModuleMetadata `json:"Modules"`
}

type ModuleMetadata struct {
	Key    string `json:"Key"`
	Source string `json:"Source"`
	Dir    string `json:"Dir"`
}

func GetTerraformVersionOutput(dir string) (string, error) {
	command := exec.Command("terraform", "version")
	command.Dir = dir
	output, err := command.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func ValidateInitRun(dir string) error {
	path := dir + `/.terraform`

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("terraform init must run before running terrapolicy")
		}

		return fmt.Errorf("couldn't determine if terraform init has run: %v", err)
	}

	return nil
}

func GetResourceType(resource *hclwrite.Block) string {
	return resource.Labels()[0]
}

func GetTerraformFilePaths(dir string) ([]string, error) {
	rootDir, tfFileMatcher := dir, "/*.tf"
	tfFiles, err := file.GetFilePaths(rootDir + tfFileMatcher)

	if err != nil {
		return nil, err
	}

	modulesDirs, err := getTerraformModulesDirPaths(rootDir)
	if err != nil {
		return nil, err
	}

	for _, moduleDir := range modulesDirs {
		matches, err := file.GetFilePaths(moduleDir + tfFileMatcher)
		if err != nil {
			return nil, err
		}

		tfFiles = append(tfFiles, matches...)
	}

	return utils.UniqString(tfFiles), nil
}

func getTerraformModulesDirPaths(dir string) ([]string, error) {
	var paths []string
	var modulesJson ModulesJson

	jsonFile, err := os.Open(dir + "/.terraform/modules/modules.json")
	//lint:ignore SA5001 not required to check file close status.
	defer jsonFile.Close()

	if os.IsNotExist(err) {
		return paths, nil
	}

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(byteValue, &modulesJson); err != nil {
		return nil, err
	}

	for _, module := range modulesJson.Modules {
		modulePath, err := filepath.EvalSymlinks(dir + "/" + module.Dir)

		if os.IsNotExist(err) {
			log.Print("[WARN] Module not found, skipping.", dir+"/"+module.Dir)
			continue
		}

		if err != nil {
			return nil, err
		}

		paths = append(paths, modulePath)
	}

	return paths, nil
}
