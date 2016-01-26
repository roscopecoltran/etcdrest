package main

import (
	"os"
	"time"

	"github.com/codegangsta/cli"

	"github.com/mickep76/etcdrest/command"
)

func main() {
	app := cli.NewApp()
	app.Name = "etcdrest"
	app.Version = Version
	app.Usage = "REST API server with etcd as backend."
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "debug, d", Usage: "Debug"},
		cli.StringFlag{Name: "config, c", EnvVar: "ETCDREST_CONFIG", Usage: "Configuration file (/etc/etcdrest.json|yaml|toml or $HOME/.etcdrest.json|yaml|toml)"},
		cli.StringFlag{Name: "peers, p", Value: "http://127.0.0.1:4001,http://127.0.0.1:2379", EnvVar: "ETCDREST_PEERS", Usage: "Comma-delimited list of hosts in the cluster"},
		cli.StringFlag{Name: "cert", Value: "", EnvVar: "ETCDREST_CERT", Usage: "Identify HTTPS client using this SSL certificate file"},
		cli.StringFlag{Name: "key", Value: "", EnvVar: "ETCDREST_KEY", Usage: "Identify HTTPS client using this SSL key file"},
		cli.StringFlag{Name: "ca", Value: "", EnvVar: "ETCDREST_CA", Usage: "Verify certificates of HTTPS-enabled servers using this CA bundle"},
		cli.StringFlag{Name: "user, u", Value: "", EnvVar: "ETCDREST_USER", Usage: "Username"},
		cli.DurationFlag{Name: "timeout, t", Value: time.Second, Usage: "Connection timeout"},
		cli.DurationFlag{Name: "command-timeout, T", Value: 5 * time.Second, Usage: "Command timeout"},
		cli.StringFlag{Name: "bind, b", Value: "0.0.0.0:8080", EnvVar: "ETCDREST_BIND", Usage: "Bind address"},
		cli.StringFlag{Name: "api-version, V", Value: "v1", EnvVar: "ETCDREST_API_VERSION", Usage: "API Version"},
	}
	app.Commands = []cli.Command{
		command.NewServerCmd(),
		command.NewPrintConfigCmd(),
	}

	app.Run(os.Args)
}
