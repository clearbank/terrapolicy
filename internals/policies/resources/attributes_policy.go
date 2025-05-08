package resource_policies

import (
	"fmt"
	"log"
	"strings"

	"github.com/clearbank/terrapolicy/internals/policies"
	"github.com/clearbank/terrapolicy/internals/terraform"
	"github.com/clearbank/terrapolicy/internals/tfschema"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
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
			log.Printf("[DEBUG] processing resource \"%v\"", currentResource)

			if currentResource != targetResource {
				log.Printf("[DEBUG] resource \"%v\" not affected by policy", currentResource)
				continue
			}

			attributePath := strings.Split(targetAttribute.(string), ".")
			attributeIsSet := isAttributeSet(resource, attributePath)
			if attributeIsSet && setStrategy.(string) == string(set_if_missing) {
				log.Printf("[DEBUG] attribute already found on resource. skipping due to strategy \"%v\"", setStrategy)
				continue
			}

			if (attributeIsSet && setStrategy.(string) == string(fail_if_set)) || (!attributeIsSet && setStrategy.(string) == string(fail_if_missing)) {
				log.Printf("[DEBUG] failed policy check. attribute set: %v policy: %v", attributeIsSet, setStrategy)
				result.Outcome = policies.OUTCOME_FAIL
				result.Reason = "Attribute non conformant"
				return result, nil
			}

			schema, err := tfschema.GetSchemaForBlock(resource, payload.WorkingDir)
			if err != nil {
				return result, fmt.Errorf("cannot retrieve schema: %v", err)
			}

			attributeType, err := getTypeForAttribute(schema, attributePath)
			if err != nil {
				return result, err
			}

			if attributeType != cty.NilType {
				v, err := gocty.ToCtyValue(targetValue, attributeType)
				if err != nil {
					return result, fmt.Errorf("bad conversion: %v", err)
				}

				setAttribute(resource.Body(), attributePath, v)
				log.Printf("[INFO] setting attribute \"%v\" set to %v", targetAttribute, targetValue)

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
			log.Printf("[DEBUG] skipping block of type: \"%v\"", t)
		}
	}
	return result, nil
}

func isAttributeSet(block *hclwrite.Block, path []string) bool {
	if len(path) == 0 {
		return false
	}

	if len(path) == 1 {
		tagsAttribute := block.Body().GetAttribute(path[0])

		if tagsAttribute != nil {
			tokens := tagsAttribute.Expr().BuildTokens(hclwrite.Tokens{})
			return tokens != nil
		}

		return false
	}

	var blocks []*hclwrite.Block
	for _, block := range block.Body().Blocks() {
		log.Printf("[DEBUG] checking block %v against path %v", block.Type(), path)
		if block.Type() == path[0] {
			blocks = append(blocks, block)
		}
	}

	var result bool = len(blocks) > 0
	for _, block := range blocks {
		result = result && isAttributeSet(block, path[1:])
	}

	return result
}

func getTypeForAttribute(schema *tfschema.Block, path []string) (cty.Type, error) {
	if len(path) == 1 {
		attributeSchema := schema.Attributes[path[0]]
		if attributeSchema != nil {
			return attributeSchema.Type.Type, nil
		}
		return cty.NilType, fmt.Errorf("cannot retrieve type for attribute: %v", path[0])
	}

	if nestedBlock, found := schema.BlockTypes[path[0]]; found {
		return getTypeForAttribute(&nestedBlock.Block, path[1:])
	}

	return cty.NilType, fmt.Errorf("cannot retrieve type for attribute: %v", path[0])
}

func setAttribute(body *hclwrite.Body, path []string, value cty.Value) {
	if len(path) == 0 {
		return
	}

	if len(path) == 1 {
		body.SetAttributeValue(path[0], value)
		return
	}

	var blocks []*hclwrite.Block
	for _, block := range body.Blocks() {
		if block.Type() == path[0] {
			blocks = append(blocks, block)
		}
	}

	if len(blocks) > 0 {
		for _, block := range blocks {
			setAttribute(block.Body(), path[1:], value)
		}
		return
	}

	block := body.AppendNewBlock(path[0], nil)
	setAttribute(block.Body(), path[1:], value)
}
