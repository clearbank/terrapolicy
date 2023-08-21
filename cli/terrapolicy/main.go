package main

import (
	"fmt"
	"github.com/clearbank/terrapolicy"
	"github.com/clearbank/terrapolicy/internals/cli"
	"github.com/clearbank/terrapolicy/internals/policies"
	"io"
	"log"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/logutils"
)

func initLogFiltering(verbose bool) {
	level := "INFO"
	if verbose {
		level = "DEBUG"
	}

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "TRACE", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(level),
		Writer:   os.Stderr,
	}

	log.SetOutput(filter)
	hclog.DefaultOutput = io.Discard //suppress tfSchema logs
}

func fail(e error) {
	log.Printf("\n[ERROR] execution failed due to error code: %v", e)
	os.Exit(1)
}

func main() {
	args, err := initArgs()

	if err != nil {
		switch err.Error() {
		case "args":
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "[ERROR] code: %v\n", err)
			os.Exit(1)
		}
	}

	initLogFiltering(args.Verbose)
	policy, err := policies.Parse(args.Config)

	if err != nil {
		fail(err)
	}

	err = terrapolicy.TerraPolicy(terrapolicy.Args{
		Policy: policy,
		Flags:  policies.PolicyExecutionFlags{Strict: args.Strict},
		Dir:    args.Dir,
	})

	if err != nil {
		fail(err)
	} else {
		log.Printf("[INFO] completed")
	}
}

func initArgs() (cli.Args, error) {
	programName := os.Args[0]
	programArgs := os.Args[1:]

	return cli.ParseArgs(programName, programArgs)
}
