package providers

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

const provider_pattern string = "(?m)provider (.+?) (.+?)$"
const provider_version_pattern string = `v(\d+).(\d+)\.?(\d+)?`
const string_version_pattern string = `(\d+).(\d+)\.?(\d+)?`
const terraform_version_pattern string = `Terraform v(\d+).(\d+)\.(\d+)`

var SUPPORTED_PROVIDERS = []string{"azurerm"}

func ParseTerraformOutput(terraformVersionOutput *string) (map[string]Version, error) {
	providerVersions := getProvidersVersions(terraformVersionOutput)
	providerVersions["terraform"] = getTerraformVersion(terraformVersionOutput)
	return providerVersions, nil
}

func IsSupportedResource(resourceType string) bool {
	switch getProviderByResource(resourceType) {
	case "azurerm":
		return true
	default:
		return false
	}
}

func ExtractProviderNameFromResourceType(resourceType string) (string, error) {
	s := strings.SplitN(resourceType, "_", 2)
	if len(s) < 2 {
		return "", fmt.Errorf("failed to detect a provider name: %s", resourceType)
	}
	return s[0], nil
}

func ParseStringVersion(s string) Version {
	return parseVersion(string_version_pattern, &s)
}

func getTerraformVersion(terraformVersionOutput *string) Version {
	return parseVersion(terraform_version_pattern, terraformVersionOutput)
}

func getProvidersVersions(terraformVersionOutput *string) map[string]Version {
	providersRegex := regexp.MustCompile(provider_pattern)
	matches := providersRegex.FindAllStringSubmatch(*terraformVersionOutput, -1)

	log.Printf("[DEBUG] provider matches: %v", matches)
	providers := make(map[string]Version)
	for _, s := range matches {
		if len(s) != 3 {
			continue
		}

		version := parseVersion(provider_version_pattern, &s[2])
		providers[s[1]] = version
	}

	return providers
}

func parseVersion(pattern string, s *string) Version {
	versionRegex := regexp.MustCompile(pattern)
	versionMatch := versionRegex.FindStringSubmatch(*s)[1:]

	versionFrags := make([]int, 3)
	for i := range versionFrags {
		versionFrags[i] = getVersionPart(versionMatch, i)
	}

	return Version{
		Major: versionFrags[0],
		Minor: versionFrags[1],
		Patch: versionFrags[2],
	}
}

func getVersionPart(parts []string, versionPart int) int {
	if len(parts)-1 < int(versionPart) {
		return -1
	}

	version, err := strconv.Atoi(parts[versionPart])
	if err != nil {
		return -1
	}

	return version
}

func getProviderByResource(resourceType string) string {
	if strings.HasPrefix(resourceType, "aws_") {
		return "aws"
	} else if strings.HasPrefix(resourceType, "google_") {
		return "gcp"
	} else if strings.HasPrefix(resourceType, "azurerm_") || strings.HasPrefix(resourceType, "azurestack_") {
		return "azurerm"
	}

	return ""
}
