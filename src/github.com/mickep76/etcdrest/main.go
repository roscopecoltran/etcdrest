package main

import (
	"os"
	"time"

	"github.com/bgentry/speakeasy"
	"github.com/codegangsta/cli"

	"github.com/mickep76/etcdrest/etcd"
	"github.com/mickep76/etcdrest/log"
	"github.com/mickep76/etcdrest/server"
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
		cli.StringFlag{Name: "print-config, P", Usage: "Print config"},
	}
	app.Action = func(c *cli.Context) {
		runServer(c)
	}

	app.Run(os.Args)
}

func runServer(c *cli.Context) {
	if c.GlobalBool("debug") {
		log.SetDebug()
	}

	// Create etcd config.
	ec := etcd.New()
	ec.Peers(c.GlobalString("peers"))
	ec.Cert(c.GlobalString("cert"))
	ec.Key(c.GlobalString("key"))
	ec.CA(c.GlobalString("ca"))
	ec.Timeout(c.GlobalDuration("timeout"))
	ec.CmdTimeout(c.GlobalDuration("command-timeout"))

	// If user is set ask for password.
	if c.GlobalString("user") != "" {
		ec.User(c.GlobalString("user"))
		pass, err := speakeasy.Ask("Password: ")
		if err != nil {
			log.Fatal(err.Error())
		}
		ec.Pass(pass)
	}

	// Connect to etcd server.
	es, err := ec.Connect()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Create server config.
	sc := server.New(es)
	sc.Bind(c.GlobalString("bind"))
	sc.APIVersion(c.GlobalString("api-version"))
	sc.Envelope(c.GlobalBool("envelope"))
	//  sc.Indent(c.GlobalBool("indent"))

	sc.RouteEtcd("/api/hosts", "/hosts", "file://schemas/host.json")
	sc.RouteTemplate("/hosts/{name}", "host")

	// Start server.
	if err := sc.Run(); err != nil {
		log.Fatal(err.Error())
	}
}
