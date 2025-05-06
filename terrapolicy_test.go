package terrapolicy

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/clearbank/terrapolicy/internals/cli"
	"github.com/clearbank/terrapolicy/internals/file"
	"github.com/clearbank/terrapolicy/internals/policies"

	"log"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/logutils"

	. "github.com/onsi/gomega"
)

var testLoc = "integration_tests/"
var tmpDir = "__tmp"

type Logger interface {
	Log(args ...any)
}

func TestTerraPolicy(t *testing.T) {
	if _, skip := os.LookupEnv("SKIP_INTEGRATION_TESTS"); skip {
		t.Skip("skipping integration test")
		return
	}

	prepare()

	var suites []TestSuite
	paths, _ := file.GetFilePaths(testLoc + "/*")

	for _, testBaseLocation := range paths {
		testPolicies, _ := file.GetFilePaths(testBaseLocation + "/policy_*.yaml")
		for _, policyToTest := range testPolicies {
			testLocation := tmpDir + "/" + file.GetFilename(testBaseLocation) + "/" + file.GetFilename(policyToTest) + "/"
			file.Copy(testBaseLocation+"/*.tf", testLocation)
			file.Copy(policyToTest, testLocation+cli.TERRAPOLICY_DEFAULT_POLICY_NAME)

			extraArgs, _ := file.ReadFile(testBaseLocation + "/args")
			suites = append(suites, TestSuite{
				location:  testLocation,
				pass:      shouldPass(policyToTest, t),
				extraArgs: string(extraArgs),
			})
		}
	}

	for _, suite := range suites {
		suite := suite

		t.Run(suite.location, func(t *testing.T) {
			if _, parallel := os.LookupEnv("PARALLEL"); parallel {
				t.Parallel() // marks each test case as capable of running in parallel with each other
			}
			if name, skip := os.LookupEnv("SUITE"); skip && !strings.Contains(t.Name(), name) {
				t.Skip("skipping")
			}

			g := NewWithT(t)

			itShouldRunTerraformInit(&suite, g, t)
			itShouldRunTerraPolicy(&suite, g, t)
			itShouldRunTerraformValidate(&suite, g, t)
		})
	}
}

func itShouldRunTerraformInit(suite *TestSuite, g *WithT, l Logger) {
	err := run("terraform", suite.location, "init", l)
	g.Expect(err).To(BeNil(), "terraform init failed")
}

func itShouldRunTerraPolicy(suite *TestSuite, g *WithT, l Logger) {
	stringArgs := fmt.Sprintf("-dir %v %v", suite.location, suite.extraArgs)
	cliArgs, err := cli.ParseArgs("terrapolicy", strings.Split(stringArgs, " "))
	g.Expect(err).To(BeNil(), "arguments failed to parse")

	l.Log(stringArgs, cliArgs)
	p, err := policies.Parse(cliArgs.Config)
	g.Expect(err).To(BeNil(), "policy failed to parse")

	err = TerraPolicy(Args{
		Policy: p,
		Flags: policies.PolicyExecutionFlags{
			Strict: cliArgs.Strict,
		},
		Dir: cliArgs.Dir,
	})

	g.Expect(err == nil).To(BeEquivalentTo(suite.pass), "wrong expected outcome")
}

func itShouldRunTerraformValidate(suite *TestSuite, g *WithT, l Logger) {
	err := run("terraform", suite.location, "validate", l)
	g.Expect(err).To(BeNil(), "terraform validate failed")
}

func run(prog string, entryDir string, cmd string, g Logger) error {
	println(prog, cmd)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command := exec.Command(prog, cmd)
	command.Dir = entryDir
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		g.Log(stderr.String())
		return err
	}

	println(stdout.String())

	return nil
}

func prepare() {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "TRACE", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("INFO"),
		Writer:   os.Stderr,
	}

	log.SetOutput(filter)
	hclog.DefaultOutput = io.Discard //suppress tfSchema logs
	os.RemoveAll(tmpDir)
}

func shouldPass(policyName string, t *testing.T) bool {
	frags := strings.Split(file.GetFilename(policyName), "_")
	exp := strings.ToLower(frags[len(frags)-1])

	switch string(exp) {
	case "ok":
		return true
	case "ko":
		return false
	default:
		t.Fatalf("policy name %v must end in ok|ko to determine the success/fail outcome of the policy", policyName)
		return false
	}
}

type TestSuite struct {
	location  string
	pass      bool
	extraArgs string
}
