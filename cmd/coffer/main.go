package main

import (
	cmd "git.sr.ht/~jakintosh/command-go"
)

const (
	BIN_NAME    = "coffer"
	AUTHOR      = "jakintosh"
	VERSION     = "0.1"
	DEFAULT_CFG = "~/.config/coffer" + BIN_NAME
)

func main() {
	root.Parse()
}

var root = &cmd.Command{
	Name:        BIN_NAME,
	Author:      AUTHOR,
	Version:     VERSION,
	Help:        "manage your coffer from the command line",
	Subcommands: []*cmd.Command{},
	Operands:    []cmd.Operand{},
	Options:     []cmd.Option{},
	Handler: func(i *cmd.Input) error {
		return nil
	},
}
