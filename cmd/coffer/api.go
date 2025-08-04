package main

import (
	cmd "git.sr.ht/~jakintosh/command-go"
)

var apiCmd = &cmd.Command{
	Name: "api",
	Help: "call HTTP API resources",
	Subcommands: []*cmd.Command{
		metricsCmd,
		ledgerCmd,
		patronsCmd,
		settingsCmd,
	},
}
