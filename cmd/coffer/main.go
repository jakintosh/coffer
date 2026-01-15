package main

import (
	"log"
	"os"
	"path/filepath"

	"git.sr.ht/~jakintosh/command-go/pkg/args"
	"git.sr.ht/~jakintosh/command-go/pkg/envs"
	"git.sr.ht/~jakintosh/command-go/pkg/version"
)

const (
	BIN_NAME    = "coffer"
	AUTHOR      = "jakintosh"
	VERSION     = "0.1"
	DEFAULT_CFG = "~/.config/coffer"
)

func main() {
	root.Parse()
}

var root = &args.Command{
	Name: BIN_NAME,
	Config: &args.Config{
		Author:  AUTHOR,
		Version: VersionInfo.Version,
		HelpOption: &args.HelpOption{
			Short: 'h',
			Long:  "help",
		},
	},
	Help: "manage your coffer from the command line",
	Subcommands: []*args.Command{
		apiCmd,
		envs.Command(DEFAULT_CFG),
		serveCmd,
		statusCmd,
		version.Command(VersionInfo),
	},
	Options: envs.ConfigOptionsAnd(args.Option{
		Short: 'v',
		Long:  "verbose",
		Type:  args.OptionTypeFlag,
		Help:  "use verbose output",
	}),
}

func loadCredential(
	name string,
	credsDir string,
) string {
	credPath := filepath.Join(credsDir, name)
	cred, err := os.ReadFile(credPath)
	if err != nil {
		log.Fatalf("failed to load required credential '%s': %v\n", name, err)
	}
	return string(cred)
}
