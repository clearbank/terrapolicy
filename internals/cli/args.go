package cli

import (
	"errors"
	"flag"

	"github.com/clearbank/terrapolicy/internals/file"
)

type Args struct {
	Config  string
	Strict  bool
	Verbose bool
	Dir     string
	Help    bool
	Version bool
}

var TERRAPOLICY_DEFAULT_POLICY_NAME = ".terrapolicy.yaml"

func ParseArgs(programName string, programArgs []string) (Args, error) {
	args := Args{}

	fs := flag.NewFlagSet(programName, flag.ContinueOnError)
	fs.StringVar(&args.Config, "config", "", "The locations of the yaml policy")
	fs.BoolVar(&args.Strict, "strict", false, "Fails if remediations cannot be applied")
	fs.BoolVar(&args.Verbose, "verbose", false, "Verbose logging")
	fs.BoolVar(&args.Help, "help", false, "Usage")
	fs.StringVar(&args.Dir, "dir", ".", "cwd")
	fs.BoolVar(&args.Version, "version", false, "Prints the version")

	err := fs.Parse(programArgs)

	if err != nil {
		return args, errors.New("args")
	}
	if args.Help {
		fs.Usage()
		return args, errors.New("args")
	}
	if args.Version {
		return args, nil
	}

	if args.Config != "" && !file.Exists(args.Config) {
		return args, errors.New("config_not_found")
	}

	if !file.Exists(args.Dir) {
		return args, errors.New("dir_not_found")
	}

	if args.Config == "" && !file.Exists(args.Dir+"/"+TERRAPOLICY_DEFAULT_POLICY_NAME) {
		return args, errors.New("default_config_not_found")
	} else {
		args.Config = args.Dir + "/" + TERRAPOLICY_DEFAULT_POLICY_NAME
	}

	return args, nil
}
