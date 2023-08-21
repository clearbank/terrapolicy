package file

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/otiai10/copy"
	"go.uber.org/multierr"
)

func ReplaceWithTerrapolicyFile(path string, textContent string, rename bool) error {
	backupFilename := path + ".bak"

	if rename {
		taggedFilename := strings.TrimSuffix(path, filepath.Ext(path)) + ".terrapolicy.tf"
		if err := createFile(taggedFilename, textContent); err != nil {
			return err
		}
	}

	log.Print("[INFO] Backing up ", path, " to ", backupFilename)
	if err := os.Rename(path, backupFilename); err != nil {
		return err
	}

	if !rename {
		if err := createFile(path, textContent); err != nil {
			return err
		}
	}

	return nil
}

func ReadHCLFile(path string) (*hclwrite.File, error) {
	src, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	file, diagnostics := hclwrite.ParseConfig(src, path, hcl.InitialPos)
	if err := multierr.Combine(diagnostics.Errs()...); err != nil {
		return nil, err
	}

	return file, nil
}

func ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func GetFilePaths(glob string) ([]string, error) {
	tfFiles, err := doublestar.Glob(glob)
	if err != nil {
		return nil, err
	}

	for i, tfFile := range tfFiles {
		resolvedTfFile, err := filepath.EvalSymlinks(tfFile)
		if err != nil {
			return nil, err
		}

		tfFiles[i] = resolvedTfFile
	}

	return tfFiles, nil
}

func GetFullFilename(path string) string {
	_, filename := filepath.Split(path)
	return filename
}

func GetFilename(path string) string {
	_, filename := filepath.Split(path)
	filename = strings.TrimSuffix(filename, filepath.Ext(path))
	filename = strings.ReplaceAll(filename, ".", "-")
	return filename
}

func Copy(glob, output string) error {
	inputs, err := GetFilePaths(glob)
	if err != nil {
		return err
	}

	isOutDir := strings.HasSuffix(output, "/")

	for _, input := range inputs {
		if isOutDir {
			copy.Copy(input, output+GetFullFilename(input))
		} else {
			copy.Copy(input, output)
		}
	}

	return nil
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func createFile(path string, textContent string) error {
	log.Print("[INFO] Creating file ", path)
	return ioutil.WriteFile(path, []byte(textContent), 0644)
}
