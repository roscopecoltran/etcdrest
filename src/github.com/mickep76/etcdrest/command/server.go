package command

import (
	"github.com/bgentry/speakeasy"
	"github.com/codegangsta/cli"

	"github.com/mickep76/etcdrest/etcd"
	"github.com/mickep76/etcdrest/log"
	"github.com/mickep76/etcdrest/server"
)

func NewServerCmd() cli.Command {
	return cli.Command{
		Name:  "server",
		Usage: "Start server",
		Flags: []cli.Flag{},
		Action: func(c *cli.Context) {
			serverCmd(c)
		},
	}
}

func serverCmd(c *cli.Context) {
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
	//	sc.Indent(c.GlobalBool("indent"))

	sc.RouteEtcd("/api/hosts", "/hosts", "file://schemas/host.json")
	sc.RouteTemplate("/hosts/{name}", "host")

	// Start server.
	if err := sc.Run(); err != nil {
		log.Fatal(err.Error())
	}
}
