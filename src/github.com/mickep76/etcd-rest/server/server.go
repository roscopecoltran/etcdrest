package server

import (
	"encoding/json"
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

	"github.com/mickep76/etcd-rest/config"
)

// Envelope struct.
type Envelope struct {
	Code   int         `json:"code"`
	Data   interface{} `json:"data,omitempty"`
	Errors []string    `json:"errors,omitempty"`
}

func write(c *config.Config, w http.ResponseWriter, r *http.Request, data interface{}) {
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

func writeErrors(c *config.Config, w http.ResponseWriter, r *http.Request, e []string, code int) {
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

func writeError(c *config.Config, w http.ResponseWriter, r *http.Request, e string, code int) {
	//  writeErrors(c, w, r, []string{e}, code) {
}

func all(c *config.Config, route *config.Route, kapi client.KeysAPI) func(w http.ResponseWriter, r *http.Request) {
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

func Run(cfg *config.Config) {
	// -- Split in separate func. --
	// Connect to etcd.
	log.Printf("Connecting to etcd: %s", cfg.Etcd.Peers)
	etcdCfg := client.Config{
		Endpoints:               strings.Split(cfg.Etcd.Peers, ","),
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	c, err := client.New(etcdCfg)
	if err != nil {
		log.Fatal(err)
	}

	kapi := client.NewKeysAPI(c)
	// -- Split in separate func. --

	// Create new router.
	r := mux.NewRouter()

	route := config.Route{
		Regexp:   "^/hosts$",
		Path:     "/hosts",
		Endpoint: "/hosts",
	}

	r.HandleFunc("/hosts", all(cfg, &route, kapi)).
		Methods("GET")

	// Fire up the server
	log.Printf("Bind to: %s", cfg.Bind)
	logr := handlers.LoggingHandler(os.Stdout, r)
	log.Fatal(http.ListenAndServe(cfg.Bind, logr))
}
