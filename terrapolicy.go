package terrapolicy

import (
	"errors"
	"fmt"
	"github.com/clearbank/terrapolicy/internals/file"
	"github.com/clearbank/terrapolicy/internals/policies"
	"github.com/clearbank/terrapolicy/internals/policies/providers"
	"github.com/clearbank/terrapolicy/internals/policies/resources"
	"github.com/clearbank/terrapolicy/internals/providers"
	"github.com/clearbank/terrapolicy/internals/terraform"
	"log"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type Args struct {
	Policy policies.Policy
	Flags  policies.PolicyExecutionFlags
	Dir    string
}

type PoliciesHandlerFunc func(args *Args) error

var POLICY_MAPPING_RESOURCES = map[string]policies.ResourcePolicyExecutor{
	"attributes_policy": &resource_policies.AttributesPolicy{},
}
var POLICY_MAPPING_PROVIDERS = map[string]policies.ProviderPolicyExecutor{
	"version_policy": &provider_policies.VersionPolicy{},
}

func TerraPolicy(args Args) error {
	log.Printf("[INFO] starting terrapolicy")

	if err := terraform.ValidateInitRun(args.Dir); err != nil {
		return fail(err, "terraform_init")
	}

	for _, handler := range []PoliciesHandlerFunc{runProvidersPolicies, runResourcePolicies} {
		if err := handler(&args); err != nil {
			return err
		}
	}

	return nil
}

func runProvidersPolicies(args *Args) error {
	log.Printf("[INFO] starting providers policies")

	out, err := terraform.GetTerraformVersionOutput(args.Dir)

	if err != nil {
		return fail(err, "terraform_output")
	}

	log.Printf("[DEBUG] tf output: %v", out)

	providers, err := providers.ParseTerraformOutput(&out)
	if err != nil {
		return fail(err, "terraform_providers")
	}

	log.Printf("[INFO] tf providers: %v", providers)

	for _, providerPolicy := range args.Policy.Providers {
		policyHandler := POLICY_MAPPING_PROVIDERS[providerPolicy.Type]

		if policyHandler == nil {
			return fail(fmt.Errorf("cannot locate mapping for: %v", providerPolicy.Type), "missing_policy_type")
		}

		log.Printf("[INFO] processing policy `%v`", providerPolicy.Type)
		if result, err := policyHandler.Execute(policies.ProviderPolicyPayload{
			Policy:           providerPolicy,
			WorkingDir:       args.Dir,
			Flags:            args.Flags,
			CurrentProviders: providers,
		}); err != nil {
			//any unhandled error should immediately stop execution
			return fail(err, "policy_setup_failure")
		} else if result.Outcome == policies.OUTCOME_FAIL {
			//maybe consider in the future grouping failed policies instead of terminating
			return warn(fmt.Errorf("policy failed with reason: %v", result.Reason), "policy_failure")
		} else if result.Outcome == policies.OUTCOME_REMEDIATE {
			//provider policy cannot remediate
			return fail(err, "policy_unabled_to_remediate")
		}
	}
	return nil
}

func runResourcePolicies(args *Args) error {
	log.Printf("[INFO] starting resource policies")
	paths, err := terraform.GetTerraformFilePaths(args.Dir)

	if err != nil {
		return fail(err, "read_files")
	} else {
		log.Printf("[DEBUG] paths: %v", paths)
	}

	remediations := make(map[string]*hclwrite.File)
	for _, path := range paths {
		log.Printf("[INFO] processing %v", path)
		hcl, err := file.ReadHCLFile(path)

		if err != nil {
			return fail(err, "read_hcl_files")
		}

		for _, resourcePolicy := range args.Policy.Resources {
			policyHandler := POLICY_MAPPING_RESOURCES[resourcePolicy.Type]

			if policyHandler == nil {
				return fail(fmt.Errorf("cannot locate mapping for %v", resourcePolicy.Type), "missing_policy_type")
			}

			log.Printf("[INFO] processing policy `%v`", resourcePolicy.Type)
			if result, err := policyHandler.Execute(policies.ResourcePolicyPayload{
				Hcl:        hcl,
				Policy:     resourcePolicy,
				FileName:   file.GetFilename(path),
				FilePath:   path,
				WorkingDir: args.Dir,
				Flags:      args.Flags,
			}); err != nil {
				//any unhandled error should immediately stop execution
				return fail(err, "policy_setup_failure")
			} else if result.Outcome == policies.OUTCOME_FAIL {
				//maybe consider in the future grouping failed policies instead of terminating
				return warn(fmt.Errorf("policy failed with reason: %v", result.Reason), "policy_failure")
			} else if result.Outcome == policies.OUTCOME_REMEDIATE {
				remediations[path] = hcl
			}
		}
	}

	for path, hcl := range remediations {
		text := string(hcl.Bytes())
		if err := file.ReplaceWithTerrapolicyFile(path, text, true); err != nil {
			return fail(err, "policy_remediation_failure")
		}
	}

	return nil
}

func fail(e error, code string) error {
	log.Printf("[ERROR] %v", e)
	return errors.New(code)
}

func warn(e error, code string) error {
	log.Printf("[WARN] %v", e)
	return errors.New(code)
}
