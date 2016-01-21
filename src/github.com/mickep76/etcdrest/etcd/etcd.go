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

	"github.com/mickep76/etcdrest/config"
	"github.com/mickep76/etcdrest/log"
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

// Delete document.
func (etcd *Etcd) Delete(path string) (int, error) {
	if _, err := etcd.keyAPI.Delete(context.TODO(), path, &client.DeleteOptions{Recursive: true}); err != nil {
		// Pocument doesn't exist.
		if cerr, ok := err.(client.Error); ok && cerr.Code == 100 {
			return http.StatusNotFound, err.Error
		}
		
		// Error deleting document.
		return http.StatusInternalServerError, err
	}

	// Return success.
	return http.StatusOK, nil
}

// Get document.
func (etcd *Etcd) Get(path string) ([]byte, int, error) {
	res, err := srv.keyAPI.Get(context.TODO(), path, &client.GetOptions{Recursive: true})
	if err != nil {
		// Document doesn't exist.
		if cerr, ok := err.(client.Error); ok && cerr.Code == 100 {
			return http.StatusNotFound, err.Error
		}
		
		// Error retrieving document.
		return http.StatusInternalServerError, err
	}

	return etcdmap.Map(res.Node), nil
}

// Create document.
func (etcd *Etcd) Create(path string, doc []byte) (interface{}, int, error) {
	var d interface{}
	if err := json.Unmarshal(doc, &d); err != nil {
		return nil, http.StatusBadRequest, err
	}
		
	if err = etcdmap.Create(etcd.keyAPI, path, reflect.ValueOf(d)); err != nil {
		return nil, http.StatusInternalServerError, err
		return
	}

	return d, http.StatusOK, nil
}