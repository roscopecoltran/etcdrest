package config

import (
	"os"
	"os/user"
	"time"

	"github.com/codegangsta/cli"
	"github.com/mickep76/iodatafmt"

	"github.com/mickep76/etcd-rest/log"
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
	cfg := &Config{
		Bind:     "0.0.0.0:8080",
		Envelope: false,
	}

	cfg.Etcd = &Etcd{
		Peers: "http://127.0.0.1:4001,http://127.0.0.1:2379",
	}

	cfg.Routes = &[]Route{}

	return cfg
}

func (cfg *Config) Load(c *cli.Context) {
	// Enable debug.
	if c.GlobalBool("debug") {
		log.SetDebug()
	}

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
	if c.GlobalString("config") != "" {
		if _, err := os.Stat(c.GlobalString("config")); os.IsNotExist(err) {
			log.Fatalf("Config file doesn't exist: %s", c.GlobalString("config"))
		}
		cfgs = append([]string{c.GlobalString("config")}, cfgs...)
	}

	// Check if config file exists and load it.
	for _, fn := range cfgs {
		if _, err := os.Stat(fn); os.IsNotExist(err) {
			continue
		}
		log.Infof("Using config file: %s", fn)
		f, err := iodatafmt.FileFormat(fn)
		if err != nil {
			log.Fatal(err.Error())
		}
		if err := iodatafmt.LoadPtr(c, fn, f); err != nil {
			log.Fatal(err.Error())
		}
	}

	// Override configuration with arguments
	if c.GlobalString("peers") != "" {
		cfg.Etcd.Peers = c.GlobalString("peers")
	}

	if c.GlobalString("cert") != "" {
		cfg.Etcd.Cert = c.GlobalString("cert")
	}

	if c.GlobalString("key") != "" {
		cfg.Etcd.Key = c.GlobalString("key")
	}

	if c.GlobalString("ca") != "" {
		cfg.Etcd.CA = c.GlobalString("ca")
	}

	if c.GlobalString("user") != "" {
		cfg.Etcd.User = c.GlobalString("user")
	}

	if c.GlobalDuration("timeout") != 0 {
		cfg.Etcd.Timeout = c.GlobalDuration("timeout")
	}

	if c.GlobalDuration("command-timeout") != 0 {
		cfg.Etcd.CommandTimeout = c.GlobalDuration("command-timeout")
	}
}

func (cfg *Config) Print(c *cli.Context) {
	f, err := iodatafmt.Format(c.String("format"))
	if err != nil {
		log.Fatal(err.Error())
	}

	iodatafmt.Print(cfg, f)
}
