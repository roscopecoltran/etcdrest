package main

// TODO
// Get Config default MIME
// Get MIME headers yaml/json/toml
// GET MIME in URL .json/.yaml/.toml

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mickep76/etcdmap"
	//	"github.com/xeipuuv/gojsonschema"
	"golang.org/x/net/context"
)

// Config struct.
type Config struct {
	Bind           string        `json:"bind,omitempty" yaml:"bind,omitempty" toml:"bind,omitempty"`
	BaseURI        string        `json:"baseURI,omitempty" yaml:"baseURI,omitempty" toml:"baseURI,omitempty"`
	SchemaURI      string        `json:"schemasURI,omitempty" yaml:"schemasURI,omitempty" toml:"schemasURI,omitempty"`
	Peers          string        `json:"peers,omitempty" yaml:"peers,omitempty" toml:"peers,omitempty"`
	Cert           string        `json:"cert,omitempty" yaml:"cert,omitempty" toml:"cert,omitempty"`
	Key            string        `json:"key,omitempty" yaml:"key,omitempty" toml:"key,omitempty"`
	CA             string        `json:"ca,omitempty" yaml:"ca,omitempty" toml:"peers,omitempty"`
	User           string        `json:"user,omitempty" yaml:"user,omitempty" toml:"user,omitempty"`
	Timeout        time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty" toml:"timeout,omitempty"`
	CommandTimeout time.Duration `json:"commandTimeout,omitempty" yaml:"commandTimeout,omitempty" toml:"commandTimeout,omitempty"`
	Envelope       bool          `json:"envelope,omitempty" yaml:"evelope,omitempty" toml:"envelope,omitempty"`
	Routes         []Route       `json:"routes" yaml:"routes" toml:"routes"`
}

// Route struct.
type Route struct {
	Regexp   string `json:"regexp"`
	Path     string `json:"path"`
	Endpoint string `json:"endpoint"`
	Schema   string `json:"schema"`
}

// Envelope struct.
type Envelope struct {
	Code   int         `json:"code"`
	Data   interface{} `json:"data,omitempty"`
	Errors []string    `json:"errors,omitempty"`
}

func write(c *Config, w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	envelope := c.Envelope
	switch strings.ToLower(r.URL.Query().Get("envelope")) {
	case "true":
		envelope = true
	case "false":
		envelope = false
	}

	if envelope == false {
		b, _ := json.MarshalIndent(data, "", "  ")
		w.Write(b)
		return
	}

	s := Envelope{
		Code: http.StatusOK,
		Data: data,
	}

	b, _ := json.MarshalIndent(&s, "", "  ")
	w.Write(b)
}

func writeErrors(c *Config, w http.ResponseWriter, r *http.Request, e []string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	envelope := c.Envelope
	switch strings.ToLower(r.URL.Query().Get("envelope")) {
	case "true":
		envelope = true
	case "false":
		envelope = false
	}

	if envelope == false {
		b, _ := json.MarshalIndent(e, "", "  ")
		w.Write(b)
		return
	}

	s := Envelope{
		Code:   code,
		Errors: e,
	}

	b, _ := json.MarshalIndent(&s, "", "  ")
	w.Write(b)
}

func writeError(c *Config, w http.ResponseWriter, r *http.Request, e string, code int) {
	//  writeErrors(c, w, r, []string{e}, code) {
}

func all(c *Config, route *Route, kapi client.KeysAPI) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// Retrieve data
		res, err := kapi.Get(context.TODO(), route.Path, &client.GetOptions{Recursive: true})
		if err != nil {
			// Directory doesn't exist
			if cerr, ok := err.(client.Error); ok && cerr.Code == 100 {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			// Error retrieving data
			writeError(c, w, r, err.Error(), http.StatusInternalServerError)
			return
		}

		m := etcdmap.Map(res.Node)
		write(c, w, r, m)
	}
}

func main() {
	config := &Config{
		Peers:    "http://127.0.0.1:4001,http://127.0.0.1:2379",
		Bind:     "0.0.0.0:8080",
		Envelope: false,
	}

	if os.Getenv("ETCDREST_BIND") != "" {
		config.Bind = os.Getenv("ETCDREST_BIND")
	}

	if os.Getenv("ETCDREST_BASE_URI") != "" {
		config.BaseURI = os.Getenv("ETCDREST_BASE_URI")
	}

	if os.Getenv("ETCDREST_PEERS") != "" {
		config.Peers = os.Getenv("ETCDREST_PEERS")
	}

	// Options.
	version := flag.Bool("version", false, "Version")
	//	configFile := flag.String("config", "", "Comma separated list of etcd nodes, env. variable D2B_ETCD_PEERS")
	peers := flag.String("peers", config.Peers, "Comma separated list of etcd nodes, env. variable D2B_ETCD_PEERS")
	bind := flag.String("bind", config.Bind, "Bind to address and port, env. variable D2B_BIND_ADDR")
	baseURI := flag.String("base-uri", config.BaseURI, "Server name to advertise, env. variable D2B_SERVER_NAME")
	//	schemaURI := flag.String("schemas-uri", config.SchemaURI, "Schemas directory, env. variable D2B_SCHEMAS_DIR")
	flag.Parse()

	// Print version.
	if *version {
		fmt.Printf("d2b-api %s\n", Version)
		os.Exit(0)
	}

	// Get Base URI
	if *baseURI == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Fatal(err.Error())
		}
		port := strings.Split(*bind, ":")[1]
		str := "http://" + hostname + ":" + port + "/v1"
		baseURI = &str
	}

	// Connect to etcd.
	log.Printf("Connecting to etcd: %s", *peers)
	cfg := client.Config{
		Endpoints:               strings.Split(*peers, ","),
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	kapi := client.NewKeysAPI(c)

	// Create new router.
	r := mux.NewRouter()

	route := Route{
		Regexp:   "^/hosts$",
		Path:     "/hosts",
		Endpoint: "/hosts",
	}

	r.HandleFunc("/hosts", all(config, &route, kapi)).
		Methods("GET")

	// Fire up the server
	log.Printf("Bind to: %s", *bind)
	logr := handlers.LoggingHandler(os.Stdout, r)
	log.Fatal(http.ListenAndServe(*bind, logr))
}
