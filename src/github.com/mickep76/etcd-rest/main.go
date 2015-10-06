package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	etcd "github.com/coreos/etcd/client"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mickep76/etcdmap"
	"github.com/xeipuuv/gojsonschema"
)

// JSONError structure.
type JSONError struct {
	message string `json:"message"`
}

// Env variables.
type Env struct {
	Peers string
	Bind  string
}

// Route structure.
type Route struct {
	Path       string `json:"path"`
	Endpoint   string `json:"endpoint"`
	Schema     string `json:"schema"`
	SchemaJSON string `json:"schema_json"`
}

// getEnv variables.
func getEnv() Env {
	env := Env{}
	env.Peers = "http://127.0.0.1:4001,http://127.0.0.1:2379"
	env.Bind = "127.0.0.1:8080"

	for _, e := range os.Environ() {
		a := strings.Split(e, "=")
		switch a[0] {
		case "ETCD_PEERS":
			env.Peers = a[1]
		case "ETCD_REST_BIND":
			env.Bind = a[1]
		}
	}

	return env
}

// errToJSON convert error to JSON.
func errToJSON(err error) []byte {
	e := JSONError{
		message: err.Error(),
	}

	j, err := json.Marshal(&e)
	if err != nil {
		return []byte{}
	}

	return j
}

// getAllEntries return all entries for an endpoint.
func getAllEntries(route Route) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		kapi := etcd.NewKeysAPI(*client)
		res, err := kapi.Get(context.Background(), route.Endpoint, &etcd.GetOptions{Recursive: true})
		if err != nil {
			log.Fatal(err.Error())
		}
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusNoContent)
			w.Write(errToJSON(err))
			return
		}
		j, err := etcdmap.JSONIndent(res.Node, "	")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(j)
	}
}

// getEntry return a single entry for an endpoint.
func getEntry(route Route) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]

		kapi := etcd.NewKeysAPI(*client)
		res, err := kapi.Get(context.Background(), route.Endpoint+"/"+name, &etcd.GetOptions{Recursive: true})
		if err != nil {
			log.Fatal(err.Error())
		}
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		j, err := etcdmap.JSONIndent(res.Node, "	")
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errToJSON(err))
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(j)
	}
}

// createEntry for an endpoint.
func createEntry(route Route) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]

		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			panic(err)
		}
		if err := r.Body.Close(); err != nil {
			panic(err)
		}

		docLoader := gojsonschema.NewStringLoader(string(body))
		schemaLoader := gojsonschema.NewStringLoader(route.SchemaJSON)

		result, err := gojsonschema.Validate(schemaLoader, docLoader)
		if err != nil {
			log.Fatalf(err.Error())
		}

		if !result.Valid() {
			var errors []string
			for _, e := range result.Errors() {
				errors = append(errors, fmt.Sprintf("%s: %s\n", strings.Replace(e.Context().String("/"), "(root)", route.Endpoint+"/"+name, 1), e.Description()))
			}
			b, _ := json.MarshalIndent(&errors, "", "	")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write(b)
			return
		}

		if err = etcdmap.CreateJSON(client, route.Endpoint+"/"+name, body); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errToJSON(err))
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write(body)
	}
}

func deleteEntry(route Route) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]

		kapi := etcd.NewKeysAPI(*client)
		if _, err := kapi.Delete(context.Background(), route.Endpoint+"/"+name, &etcd.DeleteOptions{Recursive: true}); err != nil {
			log.Fatalf(err.Error())
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNoContent)
	}
}

var client *etcd.Client
var routes []Route

func main() {
	// Get env variables.
	env := getEnv()

	// Options.
	version := flag.Bool("version", false, "Version")
	peers := flag.String("peers", env.Peers, "Comma separated list of etcd nodes, can be set with env. variable ETCD_PEERS")
	bind := flag.String("bind", env.Bind, "Bind to address and port, can be set with env. variable ETCD_REST_BIND")
	flag.Parse()

	// Print version.
	if *version {
		fmt.Printf("etcd-rest %s\n", Version)
		os.Exit(0)
	}

	// Connect to etcd.
	log.Printf("Connecting to etcd: %s", *peers)
	cfg := etcd.Config{
		Endpoints:               strings.Split(*peers, ","),
		Transport:               etcd.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	client, err := etcd.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Get routes.
	kapi := etcd.NewKeysAPI(client)
	res, err := kapi.Get(context.Background(), "/routes", &etcd.GetOptions{Recursive: true})
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, v := range etcdmap.Map(res.Node) {
		switch reflect.ValueOf(v).Kind() {
		case reflect.Map:
			m := v.(map[string]interface{})

			path, ok := m["path"]
			endpoint, ok2 := m["endpoint"]
			schema, ok3 := m["schema"]
			if ok && ok2 && ok3 {

				kapi := etcd.NewKeysAPI(client)
				res, err := kapi.Get(context.Background(), schema.(string), &etcd.GetOptions{})
				if err != nil {
					log.Fatal(err.Error())
				}

				routes = append(routes, Route{
					Path:       path.(string),
					Endpoint:   endpoint.(string),
					Schema:     schema.(string),
					SchemaJSON: res.Node.Value,
				})
				log.Printf("Adding endpoint: %s", path)
			}
		}
	}

	// Check that there are any registered endpoints
	if len(routes) < 1 {
		log.Fatal("No registered endpoints")
	}

	// Create new router.
	r := mux.NewRouter()
	logr := handlers.LoggingHandler(os.Stdout, r)

	for _, e := range routes {
		r.HandleFunc(e.Path, getAllEntries(e)).
			Methods("GET")
		r.HandleFunc(e.Path+"/{name}", getEntry(e)).
			Methods("GET")
		r.HandleFunc(e.Path+"/{name}", createEntry(e)).
			Methods("PUT")
		r.HandleFunc(e.Path+"/{name}", deleteEntry(e)).
			Methods("DELETE")
	}

	log.Fatal(http.ListenAndServe(*bind, logr))
}
