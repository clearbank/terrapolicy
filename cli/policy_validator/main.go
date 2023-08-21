package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/clearbank/terrapolicy/internals/policies"
	"log"
	"os"
)

type Args struct {
	Config string
}

func main() {
	args, err := initArgs()
	assert_success(err)

	log.Printf("%v", args.Config)
	policy, err := policies.Parse(args.Config)
	assert_success(err)

	fmt.Printf("%+v\n", policy)
}

func assert_success(e error) {
	if e != nil {
		log.Fatalf("[ERROR]: %v\n", e)
	}
}

func initArgs() (Args, error) {
	args := Args{}
	programName := os.Args[0]
	programArgs := os.Args[1:]

	fs := flag.NewFlagSet(programName, flag.ExitOnError)
	fs.StringVar(&args.Config, "config", "", "Tags as a valid JSON document")

	if err := fs.Parse(programArgs); err != nil {
		return args, err
	}

	if args.Config == "" {
		return args, errors.New("must pass a config")
	}

	return args, nil
}
