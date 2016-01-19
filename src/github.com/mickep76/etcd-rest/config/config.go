package config

import (
	"log"
	"os"
	"os/user"
	"time"

	"github.com/mickep76/iodatafmt"
)

// Config struct.
type Config struct {
	Bind      string   `json:"bind,omitempty" yaml:"bind,omitempty" toml:"bind,omitempty"`
	BaseURI   string   `json:"baseURI,omitempty" yaml:"baseURI,omitempty" toml:"baseURI,omitempty"`
	SchemaURI string   `json:"schemaURI,omitempty" yaml:"schemaURI,omitempty" toml:"schemaURI,omitempty"`
	Envelope  bool     `json:"envelope" yaml:"evelope" toml:"envelope"`
	Etcd      *Etcd    `json:"etcd,omitempty" yaml:"etcd,omitempty" toml:"etcd,omitempty"`
	Routes    *[]Route `json:"routes,omitempty" yaml:"routes,omitempty" toml:"routes,omitempty"`
}

// Etcd struct.
type Etcd struct {
	Peers          string        `json:"peers,omitempty" yaml:"peers,omitempty" toml:"peers,omitempty"`
	Cert           string        `json:"cert,omitempty" yaml:"cert,omitempty" toml:"cert,omitempty"`
	Key            string        `json:"key,omitempty" yaml:"key,omitempty" toml:"key,omitempty"`
	CA             string        `json:"ca,omitempty" yaml:"ca,omitempty" toml:"peers,omitempty"`
	User           string        `json:"user,omitempty" yaml:"user,omitempty" toml:"user,omitempty"`
	Timeout        time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty" toml:"timeout,omitempty"`
	CommandTimeout time.Duration `json:"commandTimeout,omitempty" yaml:"commandTimeout,omitempty" toml:"commandTimeout,omitempty"`
}

// Route struct.
type Route struct {
	Regexp   string `json:"regexp"`
	Path     string `json:"path"`
	Endpoint string `json:"endpoint"`
	Schema   string `json:"schema"`
}

func New() *Config {
	c := &Config{
		Bind:     "0.0.0.0:8080",
		Envelope: false,
	}

	c.Etcd = &Etcd{
		Peers: "http://127.0.0.1:4001,http://127.0.0.1:2379",
	}

	c.Routes = &[]Route{}

	return c
}

func (c *Config) Load(fn string) {
	// Default path for config file.
	u, _ := user.Current()
	cfgs := []string{
		"/etcd/etcdrest.json",
		"/etcd/etcdrest.yaml",
		"/etcd/etcdrest.toml",
		u.HomeDir + "/.etcdrest.json",
		u.HomeDir + "/.etcdrest.yaml",
		u.HomeDir + "/.etcdrest.toml",
	}

	// Check if we have an arg. for config file and that it exist's.
	if fn != "" {
		if _, err := os.Stat(fn); os.IsNotExist(err) {
			log.Fatalf("Config file doesn't exist: %s", fn)
		}
		cfgs = append([]string{fn}, cfgs...)
	}

	// Check if config file exists and load it.
	for _, fn := range cfgs {
		if _, err := os.Stat(fn); os.IsNotExist(err) {
			continue
		}
		log.Printf("Using config file: %s", fn)
		f, err := iodatafmt.FileFormat(fn)
		if err != nil {
			log.Fatal(err.Error())
		}
		if err := iodatafmt.LoadPtr(c, fn, f); err != nil {
			log.Fatal(err.Error())
		}
	}
}

func (c *Config) Print(fmt string) {
	f, err := iodatafmt.Format(fmt)
	if err != nil {
		log.Fatal(err.Error())
	}

	iodatafmt.Print(c, f)
}
