package main

import (
	"fmt"
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
		cli.StringFlag{Name: "config, c", EnvVar: "ETCDREST_CONFIG", Usage: "Configuration file"},
		cli.StringFlag{Name: "peers, p", Value: "http://127.0.0.1:4001,http://127.0.0.1:2379", EnvVar: "ETCDREST_PEERS", Usage: "Comma-delimited list of hosts in the cluster"},
		cli.StringFlag{Name: "cert", Value: "", EnvVar: "ETCDREST_CERT", Usage: "Identify HTTPS client using this SSL certificate file"},
		cli.StringFlag{Name: "key", Value: "", EnvVar: "ETCDREST_KEY", Usage: "Identify HTTPS client using this SSL key file"},
		cli.StringFlag{Name: "ca", Value: "", EnvVar: "ETCDREST_CA", Usage: "Verify certificates of HTTPS-enabled servers using this CA bundle"},
		cli.StringFlag{Name: "user, u", Value: "", Usage: "User"},
		cli.DurationFlag{Name: "timeout, t", Value: time.Second, Usage: "Connection timeout"},
		cli.DurationFlag{Name: "command-timeout, T", Value: 5 * time.Second, Usage: "Command timeout"},
	}
	app.Commands = []cli.Command{
		{
			Name:   "server",
			Usage:  "Start REST API server",
			Flags:  []cli.Flag{},
			Action: server,
		},
		{
			Name:  "print-config",
			Usage: "Print configuration",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "format, f", EnvVar: "ETCDREST_FORMAT", Value: "JSON", Usage: "Data serialization format YAML, TOML or JSON"},
			},
			Action: printConfig,
		},
	}

	app.Run(os.Args)
}

func server(c *cli.Context) {
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

	// Get document from etcd.
	doc, _, err := es.Get("/")
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println(doc)

	// Create server config.
	sc := server.New()
	sc.Bind(c.GlobalString("bind"))
	sc.APIVersion(c.GlobalString("api-version"))
	sc.Envelope(c.GlobalString("envelope"))
	sc.Indent(c.GlobalString("indent"))

	//	sc.Run()
}

func printConfig(c *cli.Context) {
}
