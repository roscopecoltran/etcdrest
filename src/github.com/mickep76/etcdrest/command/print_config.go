package command

import (
	"github.com/codegangsta/cli"
)

func NewPrintConfigCmd() cli.Command {
	return cli.Command{
		Name:  "print-config",
		Usage: "Print configuration",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "format, f", EnvVar: "ETCDREST_FORMAT", Value: "JSON", Usage: "Data serialization format YAML, TOML or JSON"},
		},
		Action: func(c *cli.Context) {
			printConfigCmd(c)
		},
	}
}

func printConfigCmd(c *cli.Context) {
}
