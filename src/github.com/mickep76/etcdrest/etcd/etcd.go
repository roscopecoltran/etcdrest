package etcd

import (
	"net/http"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/coreos/etcd/pkg/transport"
	"golang.org/x/net/context"
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
	Connect() Session
}

// Session interface.
type Session interface {
	Put(string, []byte) (interface{}, int, error)
	Delete(string) (int, error)
	Get(string) ([]byte, int, error))
}

// config struct.
type config struct {
	peers      string        `json:"peers,omitempty" yaml:"peers,omitempty" toml:"peers,omitempty"`
	cert       string        `json:"cert,omitempty" yaml:"cert,omitempty" toml:"cert,omitempty"`
	key        string        `json:"key,omitempty" yaml:"key,omitempty" toml:"key,omitempty"`
	ca         string        `json:"ca,omitempty" yaml:"ca,omitempty" toml:"peers,omitempty"`
	user       string        `json:"user,omitempty" yaml:"user,omitempty" toml:"user,omitempty"`
	timeout    time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty" toml:"timeout,omitempty"`
	cmdTimeout time.Duration `json:"cmdTimeout,omitempty" yaml:"cmdTimeout,omitempty" toml:"cmdTimeout,omitempty"`
}

// session struct.
type session struct {
	client  client.Client
	keysAPI client.KeysAPI
}

// NewConfig config constructor.
func NewConfig() Config {
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
		Endpoints:               strings.Split(c.Peers, ","),
		HeaderTimeoutPerRequest: c.Timeout,
		Username:                c.User,
		Password:                c.Pass,
	})
}

func (c *config) Connect() (Session, error) {
	cl, err := newClient()
	if err != nil {
		return nil, err
	}

	kapi, err := client.NewKeysAPI(cl)
	if err != nil {
		return nil, err
	}

	return &session{
		client:  cl,
		keysAPI: kapi,
	}
}

// Create document.
func (s *Session) Put(path string, doc []byte) (interface{}, int, error) {
	var d interface{}
	if err := json.Unmarshal(doc, &d); err != nil {
		return nil, http.StatusBadRequest, err
	}

	if err = etcdmap.Create(etcd.keyAPI, path, reflect.ValueOf(d)); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return d, http.StatusOK, nil
}

// Get document.
func (s *Session) Get(path string) ([]byte, int, error) {
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

// Delete document.
func (s *session) Delete(path string) (int, error) {
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
