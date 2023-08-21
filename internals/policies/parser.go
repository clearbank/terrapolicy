package policies

import (
	"errors"
	"github.com/clearbank/terrapolicy/internals/file"
	"gopkg.in/yaml.v2"
	"log"
)

func Parse(path string) (Policy, error) {
	policy := Policy{}
	data, err := file.ReadFile(path)

	if err != nil {
		log.Printf("[ERROR] %v", err)
		return policy, errors.New("config_read")
	}

	err = yaml.Unmarshal(data, &policy)

	if err != nil {
		log.Printf("[ERROR] %v", err)
		return policy, errors.New("unmarshal_error")
	}

	return policy, nil
}
