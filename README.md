# Terrapolicy

Terrapolicy is a CLI tool / go module that allows the enforcement of policies and remediations at terraform files level. The tool is based and heavily inspired on [terratag](https://github.com/env0/terratag)

## Rationale

When enforcing policies and remediations in the cloud, you can find yourself having resources drifting from your terraform state and files. Terrapolicy aims as solving this problem adjusting file format and properties in terraform, before the state is computed and applied. This can be particularily useful if your terraform files and resources are all following the same release workflow, and terrapolicy can be applied centrally to all files before being processed for a `terraform plan` command. It follows the same concept as Policy-as-code introduced by [checkov.io](https://www.checkov.io/)

# Requirements

* Go >= 1.20

## Getting started

```bash
go install ./cli/terrapolicy/
```

# Policies

See [docs](./docs/samples/policy.yaml) for examples

**version_policy**

| parameter                | type            | descr                                                                  |
| ------------------------ | --------------- | ---------------------------------------------------------------------- |
| provider                 | string          | the name of the provider to match. Based on `terraform version` output |
| value                    | string,string[] | the value to check against                                             |
| strategy                 | string          | minimum_version,exclude                                                |
| strategy.minimum_version |                 | provider must be >= of provider value                                  |
| strategy.exclude         |                 | fails policy is provider matches exacly the provided value             |

**attributes_policy**

| parameter                | type   | descr                                                |
| ------------------------ | ------ | ---------------------------------------------------- |
| resource                 | string | the name of the resource                             |
| value                    | any    | the value to set for remediation types               |
| attribute                | string | the attribute to check against on the resource       |
| strategy                 | string | fail_if_missing,fail_if_set,set_if_missing,force_set |
| strategy.fail_if_missing |        | fails policy if attribute is missing on resource     |
| strategy.fail_if_set     |        | fails policy if attribute is set on resource         |
| strategy.set_if_missing  |        | sets attribute on resource if missing                |
| strategy.force_set       |        | always sets attribute on resource                    |

# Test

```bash
go test -v
```