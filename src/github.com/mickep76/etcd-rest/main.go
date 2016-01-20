package main

import (
	"os"
	"time"

	"github.com/codegangsta/cli"

	"github.com/mickep76/etcd-rest/config"
	"github.com/mickep76/etcd-rest/server"
)

func main() {
	cfg := config.New()

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
			Name:  "server",
			Usage: "Start REST API server",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) {
				cfg.Load(c)
				server.Run(cfg)
			},
		},
		{
			Name:  "print-config",
			Usage: "Print configuration",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "format, f", EnvVar: "ETCDREST_FORMAT", Value: "JSON", Usage: "Data serialization format YAML, TOML or JSON"},
			},
			Action: func(c *cli.Context) {
				cfg.Load(c)
				cfg.Print(c)
			},
		},
	}

	app.Run(os.Args)

	/*
		// Get Base URI
		if cfg.BaseURI == "" && *baseURI == "" {
			hostname, err := os.Hostname()
			if err != nil {
				log.Fatal(err.Error())
			}
			port := strings.Split(*bind, ":")[1]
			str := "http://" + hostname + ":" + port + "/v1"
			cfg.BaseURI = str
		}
	*/
}
