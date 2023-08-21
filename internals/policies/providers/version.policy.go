package provider_policies

import (
	"fmt"
	"github.com/clearbank/terrapolicy/internals/policies"
	"github.com/clearbank/terrapolicy/internals/providers"
	"log"
)

type VersionPolicyStrategy string
type VersionPolicy struct{}

const (
	minimum_version VersionPolicyStrategy = "minimum_version"
	exclude         VersionPolicyStrategy = "exclude"
	policy_name     string                = "version_policy"
)

func (s *VersionPolicy) Execute(payload policies.ProviderPolicyPayload) (policies.PolicyResult, error) {
	policy, result := payload.Policy, policies.PolicyResult{}

	targetProvider, targetValue, setStrategy :=
		policy.Params["provider"], policy.Params["value"], policy.Params["strategy"]
	targetVersions, err := parseVersion(targetValue)

	if err != nil {
		return result, err
	}

	log.Printf("[INFO] parsed version: %v", targetVersions)

	switch setStrategy.(string) {
	case string(exclude):
		if match(payload, targetProvider.(string), targetVersions, func(providerVersion, targetVersion providers.Version) bool {
			return providerVersion.Major == targetVersion.Major && providerVersion.Minor == targetVersion.Minor
		}) {
			result.Outcome = policies.OUTCOME_FAIL
			result.Reason = "Excluded version matched"
		}
	case string(minimum_version):
		if match(payload, targetProvider.(string), targetVersions, func(providerVersion, targetVersion providers.Version) bool {
			return providerVersion.Major <= targetVersion.Major && providerVersion.Minor <= targetVersion.Minor
		}) {
			result.Outcome = policies.OUTCOME_FAIL
			result.Reason = "Minimum provider version not met"
		}
	default:
		result.Outcome = policies.OUTCOME_FAIL
		result.Reason = "Unknown strategy"
	}

	return result, nil
}

func match(
	payload policies.ProviderPolicyPayload,
	targetProvider string,
	targetVersions []providers.Version,
	matchingFunc func(providerVersion providers.Version, targetVersion providers.Version) bool) bool {
	for p, v1 := range payload.CurrentProviders {
		if p == targetProvider {
			r := false
			for _, v2 := range targetVersions {
				r = r || matchingFunc(v1, v2)
			}
			return r
		}
	}
	return false
}

func parseVersion(version interface{}) ([]providers.Version, error) {
	var versions []providers.Version

	switch version := version.(type) {
	case string:
		versions = append(versions, providers.ParseStringVersion(version))
	case []interface{}:
		for _, s := range version {
			s, ok := s.(string)
			if !ok {
				return versions, fmt.Errorf("cannot parse version: %v %T", s, version)
			}

			versions = append(versions, providers.ParseStringVersion(s))
		}
	default:
		return versions, fmt.Errorf("cannot parse version: %v %T", version, version)
	}

	return versions, nil
}
