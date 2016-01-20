package etcd

import (
	"net/http"
	"strings"
	"time"

	"github.com/bgentry/speakeasy"
	"github.com/codegangsta/cli"
	"github.com/coreos/etcd/client"
	"github.com/coreos/etcd/pkg/transport"
	"golang.org/x/net/context"

	"github.com/mickep76/etcd-rest/config"
	"github.com/mickep76/etcd-rest/log"
)

func contextWithCommandTimeout(c *cli.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.GlobalDuration("command-timeout"))
}

func newTransport(cfg *config.Config) *http.Transport {
	tls := transport.TLSInfo{
		CAFile:   cfg.Etcd.CA,
		CertFile: cfg.Etcd.Cert,
		KeyFile:  cfg.Etcd.Key,
	}

	timeout := 30 * time.Second
	tr, err := transport.NewTransport(tls, timeout)
	if err != nil {
		log.Fatal(err.Error())
	}

	return tr
}

func newClient(cfg *config.Config) client.Client {
	config := client.Config{
		Transport:               newTransport(cfg),
		Endpoints:               strings.Split(cfg.Etcd.Peers, ","),
		HeaderTimeoutPerRequest: cfg.Etcd.Timeout,
	}

	if cfg.Etcd.User != "" {
		config.Username = cfg.Etcd.User
		var err error
		config.Password, err = speakeasy.Ask("Password: ")
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	cl, err := client.New(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	return cl
}

func NewKeyAPI(cfg *config.Config) client.KeysAPI {
	return client.NewKeysAPI(newClient(cfg))
}
