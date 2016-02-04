package etcd

import (
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/mickep76/etcdmap"
	"golang.org/x/net/context"

	"github.com/mickep76/etcdrest/log"
)

// Config interface.
type Config interface {
	Peers(string) Config
	Cert(string) Config
	Key(string) Config
	CA(string) Config
	User(string) Config
	Pass(string) Config
	Timeout(time.Duration) Config
	CmdTimeout(time.Duration) Config
	Connect() (Session, error)
}

// Session interface.
type Session interface {
	Put(string, interface{}) (int, error)
	Delete(string) (int, error)
	Get(string, bool, string) (interface{}, int, error)
	GetKeys(...string) ([]string, int, error)
}

// config struct.
type config struct {
	peers      string
	cert       string
	key        string
	ca         string
	user       string
	pass       string
	timeout    time.Duration
	cmdTimeout time.Duration
}

// session struct.
type session struct {
	client  client.Client
	keysAPI client.KeysAPI
}

// New config constructor.
func New() Config {
	return &config{
		peers:      "http://127.0.0.1:4001,http://127.0.0.1:2379",
		timeout:    time.Second,
		cmdTimeout: time.Second * 5,
	}
}

func (c *config) Peers(peers string) Config {
	c.peers = peers
	return c
}

func (c *config) Cert(cert string) Config {
	c.cert = cert
	return c
}

func (c *config) Key(key string) Config {
	c.key = key
	return c
}

func (c *config) CA(ca string) Config {
	c.ca = ca
	return c
}

func (c *config) User(user string) Config {
	c.user = user
	return c
}

func (c *config) Pass(pass string) Config {
	c.pass = pass
	return c
}

func (c *config) Timeout(timeout time.Duration) Config {
	c.timeout = timeout
	return c
}

func (c *config) CmdTimeout(cmdTimeout time.Duration) Config {
	c.cmdTimeout = cmdTimeout
	return c
}

func (c *config) newTransport() (*http.Transport, error) {
	return transport.NewTransport(transport.TLSInfo{
		CAFile:   c.ca,
		CertFile: c.cert,
		KeyFile:  c.key,
	}, 30*time.Second)
}

func (c *config) newClient() (client.Client, error) {
	tr, err := c.newTransport()
	if err != nil {
		return nil, err
	}

	return client.New(client.Config{
		Transport:               tr,
		Endpoints:               strings.Split(c.peers, ","),
		HeaderTimeoutPerRequest: c.timeout,
		Username:                c.user,
		Password:                c.pass,
	})
}

func (c *config) Connect() (Session, error) {
	log.Infof("Connect to etcd peers: %s", c.peers)
	cl, err := c.newClient()
	if err != nil {
		return nil, err
	}

	return &session{
		client:  cl,
		keysAPI: client.NewKeysAPI(cl),
	}, nil
}

// Put document.
func (s *session) Put(p string, d interface{}) (int, error) {
	if err := etcdmap.Create(s.keysAPI, p, reflect.ValueOf(d)); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

// Get document.
func (s *session) Get(p string, table bool, dirName string) (interface{}, int, error) {
	res, err := s.keysAPI.Get(context.TODO(), p, &client.GetOptions{Recursive: true})
	if err != nil {
		// Document doesn't exist.
		if cerr, ok := err.(client.Error); ok && cerr.Code == 100 {
			return nil, http.StatusNotFound, err
		}

		// Error retrieving document.
		return nil, http.StatusInternalServerError, err
	}

	if table {
		arr := etcdmap.Array(res.Node, dirName)
		return arr, http.StatusOK, nil
	}

	return etcdmap.Map(res.Node), http.StatusOK, nil
}

// GetKeys substitute keys in order.
func (s *session) GetKeys(paths ...string) ([]string, int, error) {
	arr := []string{}
	for _, p := range paths {

		res, err := s.keysAPI.Get(context.TODO(), p, &client.GetOptions{Recursive: false})
		if err != nil {
			// Document doesn't exist.
			if cerr, ok := err.(client.Error); ok && cerr.Code == 100 {
				return []string{}, http.StatusNotFound, err
			}

			// Error retrieving document.
			return []string{}, http.StatusInternalServerError, err
		}

		if res.Node.Value != "" {
			arr = append(arr, res.Node.Value)
		}
	}
	return arr, http.StatusOK, nil
}

// Delete document.
func (s *session) Delete(path string) (int, error) {
	if _, err := s.keysAPI.Delete(context.TODO(), path, &client.DeleteOptions{Recursive: true}); err != nil {
		// Pocument doesn't exist.
		if cerr, ok := err.(client.Error); ok && cerr.Code == 100 {
			return http.StatusNotFound, err
		}

		// Error deleting document.
		return http.StatusInternalServerError, err
	}

	// Return success.
	return http.StatusOK, nil
}
