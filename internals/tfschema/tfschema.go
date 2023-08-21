package tfschema

import (
	"errors"
	"fmt"
	"github.com/clearbank/terrapolicy/internals/providers"
	"github.com/clearbank/terrapolicy/internals/terraform"
	"github.com/clearbank/terrapolicy/internals/utils"
	"log"
	"strings"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/minamijoyo/tfschema/tfschema"
)

var providerToClientMapLock sync.Mutex
var providerToClientMap = map[string]tfschema.Client{}

//export forward
type Attribute = tfschema.Attribute

func GetSchemaForBlock(resource *hclwrite.Block, rootDir string) (map[string]*tfschema.Attribute, error) {
	resourceType := terraform.GetResourceType(resource)
	providerName, ok := isResourceSupported(resourceType)
	if !ok {
		log.Printf("[WARN] Resource %v not supported", resourceType)
		return nil, errors.New("not supported")
	}

	client, err := getTfSchemaClient(providerName, rootDir)

	if err != nil {
		return nil, err
	}

	typeSchema, err := client.GetResourceTypeSchema(resourceType)

	if err != nil {
		if strings.Contains(err.Error(), "Failed to find resource type") {
			log.Print("[WARN] Skipped ", resourceType, " as it is not YET supported")
			return nil, errors.New("not found")
		}

		return nil, err
	}

	return typeSchema.Attributes, nil
}

func getTfSchemaClient(providerName string, rootDir string) (tfschema.Client, error) {
	providerToClientMapLock.Lock()
	defer providerToClientMapLock.Unlock()

	hashKey := fmt.Sprintf("%v:%v", rootDir, providerName)
	client, exists := providerToClientMap[hashKey]

	if exists {
		log.Printf("[DEBUG] Using cached client for provider %v", hashKey)
		return client, nil
	}

	log.Printf("[DEBUG] Initiating client for provider %v", hashKey)

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
	})

	newClient, err := tfschema.NewClient(providerName, tfschema.Option{
		RootDir: rootDir,
		Logger:  logger,
	})
	if err != nil {
		return nil, err
	}

	providerToClientMap[hashKey] = newClient

	return newClient, nil
}

func isResourceSupported(resourceType string) (string, bool) {
	providerName, err := providers.ExtractProviderNameFromResourceType(resourceType)
	if err != nil {
		return "", false
	}

	return providerName, utils.Contains(providers.SUPPORTED_PROVIDERS, providerName)
}
