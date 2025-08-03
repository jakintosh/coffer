package main

import (
	cmd "git.sr.ht/~jakintosh/command-go"
)

var apiCmd = &cmd.Command{
	Name: "api",
	Help: "Call HTTP API resources",
	Subcommands: []*cmd.Command{
		healthCmd,
		metricsCmd,
		ledgerCmd,
		patronsCmd,
		settingsCmd,
	},
}
