package resource_policies

import (
	"fmt"
	"github.com/clearbank/terrapolicy/internals/policies"
	"github.com/clearbank/terrapolicy/internals/terraform"
	"github.com/clearbank/terrapolicy/internals/tfschema"
	"log"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty/gocty"
)

type AttributesPolicyStrategy string
type AttributesPolicy struct{}

const (
	set_if_missing  AttributesPolicyStrategy = "set_if_missing"
	force_set       AttributesPolicyStrategy = "force_set"
	fail_if_missing AttributesPolicyStrategy = "fail_if_missing"
	fail_if_set     AttributesPolicyStrategy = "fail_if_set"
	policy_name     string                   = "attributes_policy"
)

func (s *AttributesPolicy) Execute(payload policies.ResourcePolicyPayload) (policies.PolicyResult, error) {
	policy, result := payload.Policy, policies.PolicyResult{}

	targetResource, targetAttribute, targetValue, setStrategy :=
		policy.Params["resource"], policy.Params["attribute"], policy.Params["value"], policy.Params["strategy"]

	for _, resource := range payload.Hcl.Body().Blocks() {
		switch t := resource.Type(); t {

		case "resource":
			currentResource := terraform.GetResourceType(resource)
			log.Printf("[DEBUG] processing resource %v", currentResource)

			if currentResource != targetResource {
				log.Printf("[DEBUG] resource %v not affected by policy", currentResource)
				continue
			}

			attributeIsSet := isAttributeSet(resource, targetAttribute.(string))
			if attributeIsSet && setStrategy.(string) == string(set_if_missing) {
				log.Printf("[DEBUG] attribute already found on resource. skipping due to strategy %v", setStrategy)
				continue
			}

			if (attributeIsSet && setStrategy.(string) == string(fail_if_set)) || (!attributeIsSet && setStrategy.(string) == string(fail_if_missing)) {
				log.Printf("[DEBUG] failed policy check. attribute set: %v policy: %v", attributeIsSet, setStrategy)
				result.Outcome = policies.OUTCOME_FAIL
				result.Reason = "Attribute non conformant"
				return result, nil
			}

			schema_attributes, err := tfschema.GetSchemaForBlock(resource, payload.WorkingDir)
			if err != nil {
				return result, fmt.Errorf("cannot retrieve schema: %v", err)
			}
			attribute_schema := schema_attributes[targetAttribute.(string)]
			if attribute_schema != nil {
				v, err := gocty.ToCtyValue(targetValue, attribute_schema.Type.Type)
				if err != nil {
					return result, fmt.Errorf("bad conversion: %v", err)
				}

				resource.Body().SetAttributeValue(targetAttribute.(string), v)
				log.Printf("[INFO] setting attribute %v set to %v", targetAttribute, targetValue)

				result.Outcome = policies.OUTCOME_REMEDIATE
			} else {
				if payload.Flags.Strict {
					result.Outcome = policies.OUTCOME_FAIL
					result.Reason = "Schema failure"
					return result, nil
				} else {
					log.Printf("[WARN] cannot retrive attribute from schema. continue due to strict mode off: %v", targetAttribute)
				}
			}
		default:
			log.Printf("[DEBUG] skipping resource type: %v", t)
		}
	}
	return result, nil
}

func isAttributeSet(block *hclwrite.Block, attribute string) bool {
	tagsAttribute := block.Body().GetAttribute(attribute)

	if tagsAttribute != nil {
		tokens := tagsAttribute.Expr().BuildTokens(hclwrite.Tokens{})
		return tokens != nil
	}

	return false
}
