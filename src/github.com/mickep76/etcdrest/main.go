package main

// Args by reference will change struct directly go-flags
// Use class struct directly req. to export it
// Default -> File -> Env -> Arg
// Validate config file with JSON schema

import (
	"os"

	"github.com/bgentry/speakeasy"
	"github.com/codegangsta/cli"

	"github.com/mickep76/etcdrest/config"
	"github.com/mickep76/etcdrest/etcd"
	"github.com/mickep76/etcdrest/log"
	"github.com/mickep76/etcdrest/server"
)

func main() {
	cfg := config.New()

	app := cli.NewApp()
	app.Name = "etcdrest"
	app.Version = Version
	app.Usage = "REST API server with etcd as backend."
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "debug, d", Usage: "Debug"},
		cli.StringFlag{Name: "config, c", EnvVar: "ETCDREST_CONFIG", Usage: "Configuration file (/etc/etcdrest.json|yaml|toml or $HOME/.etcdrest.json|yaml|toml)"},
		cli.StringFlag{Name: "peers, p", Value: cfg.Etcd.Peers, EnvVar: "ETCDREST_PEERS", Usage: "Comma-delimited list of hosts in the cluster"},
		cli.StringFlag{Name: "cert", Value: "", EnvVar: "ETCDREST_CERT", Usage: "Identify HTTPS client using this SSL certificate file"},
		cli.StringFlag{Name: "key", Value: "", EnvVar: "ETCDREST_KEY", Usage: "Identify HTTPS client using this SSL key file"},
		cli.StringFlag{Name: "ca", Value: "", EnvVar: "ETCDREST_CA", Usage: "Verify certificates of HTTPS-enabled servers using this CA bundle"},
		cli.StringFlag{Name: "user, u", Value: "", EnvVar: "ETCDREST_USER", Usage: "Username"},
		cli.DurationFlag{Name: "timeout, t", Value: cfg.Etcd.Timeout, Usage: "Connection timeout"},
		cli.DurationFlag{Name: "command-timeout, T", Value: cfg.Etcd.CmdTimeout, Usage: "Command timeout"},
		cli.StringFlag{Name: "bind, b", Value: cfg.Bind, EnvVar: "ETCDREST_BIND", Usage: "Bind address"},
		cli.StringFlag{Name: "api-version, V", Value: cfg.APIVersion, EnvVar: "ETCDREST_API_VERSION", Usage: "API Version"},
		cli.BoolFlag{Name: "envelope", Usage: "Enable default data envelope in a response"},
		cli.BoolFlag{Name: "no-indent", Usage: "Disable default indent in a response"},
		cli.StringFlag{Name: "print-config", Value: "json", Usage: "Print config using either format json, yaml or toml"},
	}
	app.Action = func(c *cli.Context) {
		runServer(c, cfg)
	}

	app.Run(os.Args)
}

func runServer(c *cli.Context, cfg *config.Config) {
	// Set debug.
	if c.GlobalBool("debug") {
		log.SetDebug()
	}

	cfg.Load(c)

	// Print configuration.
	if c.GlobalString("print-config") != "" {
		cfg.Print(c.GlobalString("print-config"))
		os.Exit(0)
	}

	// Create etcd config.
	ec := etcd.New()
	ec.Peers(cfg.Etcd.Peers)
	ec.Cert(cfg.Etcd.Cert)
	ec.Key(cfg.Etcd.Key)
	ec.CA(cfg.Etcd.CA)
	ec.Timeout(cfg.Etcd.Timeout)
	ec.CmdTimeout(cfg.Etcd.CmdTimeout)

	// If user is set ask for password.
	if cfg.Etcd.User != "" {
		ec.User(cfg.Etcd.User)
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
	sc.Bind(cfg.Bind)
	sc.APIVersion(cfg.APIVersion)
	sc.Envelope(cfg.Envelope)
	sc.Indent(cfg.Indent)

	for _, route := range cfg.Routes {
		switch route.Type {
		case "api":
			sc.RouteEtcd(route.Endpoint, route.Path, route.Schema)
		case "template":
			sc.RouteTemplate(route.Endpoint, route.Template)
		default:
			log.Fatalf("Unknown type: %s for endpoint: %s", route.Type, route.Endpoint)
		}
	}

	// Check routes.
	if len(cfg.Routes) < 1 {
		log.Fatal("No routes specified.")
	}

	// Start server.
	if err := sc.Run(); err != nil {
		log.Fatal(err.Error())
	}
}
