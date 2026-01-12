package main

import (
	"git.sr.ht/~jakintosh/command-go/pkg/args"
	"git.sr.ht/~jakintosh/command-go/pkg/envs"
)

var apiCmd = &args.Command{
	Name:    "api",
	Help:    "call HTTP API resources",
	Options: envs.APIOptions,
	Subcommands: []*args.Command{
		metricsCmd,
		ledgerCmd,
		patronsCmd,
		settingsCmd,
	},
}
