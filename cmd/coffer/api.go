package main

import (
	"git.sr.ht/~jakintosh/command-go/pkg/args"
)

var apiCmd = &args.Command{
	Name: "api",
	Help: "call HTTP API resources",
	Subcommands: []*args.Command{
		metricsCmd,
		ledgerCmd,
		patronsCmd,
		settingsCmd,
	},
}
