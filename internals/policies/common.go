package policies

import (
	"github.com/clearbank/terrapolicy/internals/providers"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type Policy struct {
	Providers []PolicyBlock `yaml:"providers"`
	Resources []PolicyBlock `yaml:"resources"`
}

type PolicyBlock struct {
	Type   string                 `yaml:"type"`
	Params map[string]interface{} `yaml:"params"`
}

type PolicyOutcome uint64

const (
	OUTCOME_SUCCESS PolicyOutcome = iota
	OUTCOME_FAIL
	OUTCOME_REMEDIATE
)

type PolicyResult struct {
	Outcome PolicyOutcome
	Reason  string
}

type PolicyExecutionFlags struct {
	Strict bool
}

type ResourcePolicyPayload struct {
	Hcl        *hclwrite.File
	Policy     PolicyBlock
	WorkingDir string
	FileName   string
	FilePath   string
	Flags      PolicyExecutionFlags
}

type ProviderPolicyPayload struct {
	Policy           PolicyBlock
	WorkingDir       string
	Flags            PolicyExecutionFlags
	CurrentProviders map[string]providers.Version
}

type ResourcePolicyExecutor interface {
	Execute(payload ResourcePolicyPayload) (PolicyResult, error)
}

type ProviderPolicyExecutor interface {
	Execute(payload ProviderPolicyPayload) (PolicyResult, error)
}
