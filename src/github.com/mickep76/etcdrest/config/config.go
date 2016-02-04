package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/kezhuw/toml"
	"gopkg.in/yaml.v2"

	"github.com/mickep76/etcdrest/log"
)

// Config struct.
type Config struct {
	TemplDir  string  `json:"templDir" yaml:"templDir" toml:"templDir"`
	SchemaURI string  `json:"schemaURI" yaml:"schemaURI" toml:"schemaURI"`
	Bind      string  `json:"bind,omitempty" yaml:"bind,omitempty" toml:"bind,omitempty"`
	Envelope  bool    `json:"envelope" yaml:"evelope" toml:"envelope"`
	Indent    bool    `json:"indent" yaml:"indent" toml:"indent"`
	Etcd      Etcd    `json:"etcd,omitempty" yaml:"etcd,omitempty" toml:"etcd,omitempty"`
	Routes    []Route `json:"routes,omitempty" yaml:"routes,omitempty" toml:"routes,omitempty"`
}

// Etcd struct.
type Etcd struct {
	Peers      string        `json:"peers,omitempty" yaml:"peers,omitempty" toml:"peers,omitempty"`
	Cert       string        `json:"cert,omitempty" yaml:"cert,omitempty" toml:"cert,omitempty"`
	Key        string        `json:"key,omitempty" yaml:"key,omitempty" toml:"key,omitempty"`
	CA         string        `json:"ca,omitempty" yaml:"ca,omitempty" toml:"peers,omitempty"`
	User       string        `json:"user,omitempty" yaml:"user,omitempty" toml:"user,omitempty"`
	Timeout    time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty" toml:"timeout,omitempty"`
	CmdTimeout time.Duration `json:"cmdTimeout,omitempty" yaml:"cmdTimeout,omitempty" toml:"cmdTimeout,omitempty"`
}

// Route struct.
type Route struct {
	Endpoint       string `json:"endpoint,omitempty"`
	Collection     string `json:"collection,omitempty"`
	CollectionPath string `json:"collectionPath,omitempty"`
	Resource       string `json:"resource,omitempty"`
	ResourcePath   string `json:"resourcePath,omitempty"`
	Type           string `json:"type,omitempty"`
	Template       string `json:"template,omitempty"`
	Path           string `json:"path,omitempty"`
	DirName        string `json:"dirName,omitempty"`
	Schema         string `json:"schema,omitempty"`
}

func New() *Config {
	cfg := Config{
		TemplDir:  "templates",
		SchemaURI: "file://schemas",
		Bind:      "0.0.0.0:8080",
		Envelope:  false,
		Indent:    true,
	}

	cfg.Etcd = Etcd{
		Peers:      "http://127.0.0.1:4001,http://127.0.0.1:2379",
		Timeout:    time.Second,
		CmdTimeout: 5 * time.Second,
	}

	cfg.Routes = []Route{}

	return &cfg
}

func (cfg *Config) Load(c *cli.Context) {
	// Enable debug.
	if c.GlobalBool("debug") {
		log.SetDebug()
	}

	// Default path for config file.
	u, _ := user.Current()
	cfgs := []string{
		u.HomeDir + "/.etcdrest.json",
		u.HomeDir + "/.etcdrest.yaml",
		u.HomeDir + "/.etcdrest.yml",
		u.HomeDir + "/.etcdrest.toml",
		u.HomeDir + "/.etcdrest.tml",
		"/etc/etcdrest.json",
		"/etc/etcdrest.yaml",
		"/etc/etcdrest.yml",
		"/etc/etcdrest.toml",
		"/etc/etcdrest.tml",
		"/app/etc/etcdrest.json",
		"/app/etc/etcdrest.yaml",
		"/app/etc/etcdrest.yml",
		"/app/etc/etcdrest.toml",
		"/app/etc/etcdrest.tml",
	}

	// Check if we have an arg. for config file and that it exist's.
	if c.GlobalString("config") != "" {
		if _, err := os.Stat(c.GlobalString("config")); os.IsNotExist(err) {
			log.Fatalf("Config file doesn't exist: %s", c.GlobalString("config"))
		}
		cfgs = append([]string{c.GlobalString("config")}, cfgs...)
	}

	for _, fn := range cfgs {
		if _, err := os.Stat(fn); os.IsNotExist(err) {
			continue
		}

		log.Infof("Using config file: %s", fn)

		// Load config file.
		b, err := ioutil.ReadFile(fn)
		if err != nil {
			log.Fatal(err.Error())
		}

		switch filepath.Ext(fn) {
		case ".json":
			if err := json.Unmarshal(b, cfg); err != nil {
				log.Fatal(err.Error())
			}
		case ".yaml", ".yml":
			if err := yaml.Unmarshal(b, cfg); err != nil {
				log.Fatal(err.Error())
			}
		case ".toml", ".tml":
			if err := toml.Unmarshal(b, cfg); err != nil {
				log.Fatal(err.Error())
			}
		default:
			log.Fatal("unsupported data format")
		}

		// Validate config using JSON schema.

		break
	}

	// Override configuration.
	if c.GlobalString("templ-dir") != "" {
		cfg.TemplDir = c.GlobalString("templ-dir")
	}

	if c.GlobalString("schema-uri") != "" {
		cfg.SchemaURI = c.GlobalString("schema-uri")
	}

	if c.GlobalString("bind") != "" {
		cfg.Bind = c.GlobalString("bind")
	}

	if c.GlobalBool("envelope") {
		cfg.Envelope = true
	}

	if c.GlobalBool("no-indent") {
		cfg.Indent = true
	}

	// Override etcd configuration.
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
		cfg.Etcd.CmdTimeout = c.GlobalDuration("command-timeout")
	}
}

func (cfg *Config) Print(f string) {
	switch strings.ToLower(f) {
	case "json":
		b, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			log.Fatal(err.Error())
		}
		fmt.Println(string(b))
	case "yaml":
		b, err := yaml.Marshal(cfg)
		if err != nil {
			log.Fatal(err.Error())
		}
		fmt.Println(string(b))
	case "toml":
		b := new(bytes.Buffer)
		if err := toml.NewEncoder(b).Encode(cfg); err != nil {
			log.Fatal(err.Error())
		}
		fmt.Println(b.String())
	default:
		log.Fatal("unsupported data format")
	}
}
